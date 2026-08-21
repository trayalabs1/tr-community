package ranker_job_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/resources/settings"
	"github.com/Southclaws/storyden/app/services/sentiment/ranker_job"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

// TestApplyDailyIncrementsUsesConfiguredWeights verifies that admin-updated
// like/reply weights (via the settings repository) are actually picked up by
// ApplyDailyIncrements, rather than the hardcoded defaults.
func TestApplyDailyIncrementsUsesConfiguredWeights(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		db *ent.Client,
		aw *account_writer.Writer,
		settingsRepo *settings.SettingsRepository,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			_, liker := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			likerID := xid.ID(liker.ID)

			suffix := xid.New().String()
			channelResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "ranker-job-weights-" + suffix,
					Slug:        "ranker-job-weights-" + suffix,
					Description: "channel for ranker job weights tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			threadResp := tests.AssertRequest(
				cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Title:      "weights test " + suffix,
					Body:       opt.New[openapi.PostContent]("body").Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
				}, adminSession),
			)(t, http.StatusOK)
			postID, err := xid.FromString(threadResp.JSON200.Id)
			r.NoError(err)

			err = db.PostSentiment.Create().
				SetPostID(postID).
				SetSentimentTag("neutral").
				SetPositivityScore(50).
				SetCategory("NA").
				SetFeedValueScore(50).
				SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
				SetRankScore(0).
				OnConflictColumns(ent_post_sentiment.FieldPostID).
				UpdateNewValues().
				Exec(root)
			r.NoError(err)

			_, err = db.LikePost.Create().
				SetPostID(postID).
				SetAccountID(likerID).
				Save(root)
			r.NoError(err)

			// Set a custom like weight far from the default (2.0) so the
			// test fails loudly if the configured value isn't being read.
			_, err = settingsRepo.Set(root, settings.Settings{
				Services: opt.New(settings.ServiceSettings{
					FeedRanking: opt.New(settings.FeedRankingServiceSettings{
						LikeWeight: opt.New(10.0),
					}),
				}),
			})
			r.NoError(err)

			since := time.Now().Add(-24 * time.Hour)

			updated, err := ranker_job.ApplyDailyIncrements(root, db, settingsRepo, since)
			r.NoError(err)
			r.Equal(1, updated)

			ps, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(postID)).Only(root)
			r.NoError(err)
			r.InDelta(10.0, ps.RankScore, 0.001,
				"rank_score should reflect the configured like_weight=10.0 (1 like * 10.0), not the default 2.0")
		}))
	}))
}
