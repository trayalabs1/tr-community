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

// TestApplyDailyIncrementsBulkUpdatesDistinctDeltas verifies that
// ApplyDailyIncrements' single bulk UPDATE correctly applies a different
// rank_score delta per post in one call, rather than mixing them up or only
// updating one row (the failure mode a naive shared-Update()/AddRankScore
// call would have).
func TestApplyDailyIncrementsBulkUpdatesDistinctDeltas(t *testing.T) {
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

			_, liker1 := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			_, liker2 := e2e.WithAccount(root, aw, seed.Account_004_Loki)
			_, liker3 := e2e.WithAccount(root, aw, seed.Account_005_Þórr)

			suffix := xid.New().String()
			channelResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "ranker-job-bulk-" + suffix,
					Slug:        "ranker-job-bulk-" + suffix,
					Description: "channel for ranker job bulk update tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			createPost := func(title string, baseRankScore float64) xid.ID {
				threadResp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title + " " + suffix,
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
					SetPrimaryTopic("NA").
					SetFeedValueScore(50).
					SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
					SetRankScore(baseRankScore).
					OnConflictColumns(ent_post_sentiment.FieldPostID).
					UpdateNewValues().
					Exec(root)
				r.NoError(err)

				return postID
			}

			// Three posts, three different starting scores, three different
			// like counts, so each row needs a distinct delta applied.
			postNoLikes := createPost("bulk no likes", 10)
			postOneLike := createPost("bulk one like", 20)
			postThreeLikes := createPost("bulk three likes", 30)

			_, err := db.LikePost.Create().
				SetPostID(postOneLike).
				SetAccountID(xid.ID(liker1.ID)).
				Save(root)
			r.NoError(err)

			for _, liker := range []xid.ID{xid.ID(liker1.ID), xid.ID(liker2.ID), xid.ID(liker3.ID)} {
				_, err = db.LikePost.Create().
					SetPostID(postThreeLikes).
					SetAccountID(liker).
					Save(root)
				r.NoError(err)
			}

			since := time.Now().Add(-24 * time.Hour)

			updated, err := ranker_job.ApplyDailyIncrements(root, db, settingsRepo, since)
			r.NoError(err)
			r.Equal(2, updated, "only the 2 posts with new likes should be counted as updated")

			psNoLikes, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(postNoLikes)).Only(root)
			r.NoError(err)
			r.InDelta(10.0, psNoLikes.RankScore, 0.001, "post with no likes must be untouched")

			psOneLike, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(postOneLike)).Only(root)
			r.NoError(err)
			r.InDelta(22.0, psOneLike.RankScore, 0.001, "20 (existing) + 2.0 (1 like * default weight 2.0) = 22")

			psThreeLikes, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(postThreeLikes)).Only(root)
			r.NoError(err)
			r.InDelta(36.0, psThreeLikes.RankScore, 0.001, "30 (existing) + 6.0 (3 likes * default weight 2.0) = 36")
		}))
	}))
}
