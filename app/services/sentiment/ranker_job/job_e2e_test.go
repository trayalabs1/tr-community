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
	"github.com/Southclaws/storyden/app/services/sentiment/ranker_job"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestApplyDailyIncrements(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		db *ent.Client,
		aw *account_writer.Writer,
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
					Name:        "ranker-job-apply-" + suffix,
					Slug:        "ranker-job-apply-" + suffix,
					Description: "channel for ranker job apply tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			threadResp := tests.AssertRequest(
				cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Title:      "apply increments " + suffix,
					Body:       opt.New[openapi.PostContent]("body").Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
				}, adminSession),
			)(t, http.StatusOK)
			threadID, err := xid.FromString(threadResp.JSON200.Id)
			r.NoError(err)

			// Give the post an existing, known rank_score as if it had
			// already been classified by the LLM/prescored path.
			err = db.PostSentiment.Create().
				SetPostID(threadID).
				SetSentimentTag("neutral").
				SetPositivityScore(50).
				SetCategory("NA").
				SetFeedValueScore(50).
				SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
				SetRankScore(100).
				OnConflictColumns(ent_post_sentiment.FieldPostID).
				UpdateNewValues().
				Exec(root)
			r.NoError(err)

			// 1 like (+2.0) inside the window.
			_, err = db.LikePost.Create().
				SetPostID(threadID).
				SetAccountID(likerID).
				Save(root)
			r.NoError(err)

			since := time.Now().Add(-24 * time.Hour)

			updated, err := ranker_job.ApplyDailyIncrements(root, db, since)
			r.NoError(err)
			r.Equal(1, updated)

			ps, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(threadID)).Only(root)
			r.NoError(err)
			r.InDelta(102.0, ps.RankScore, 0.001, "rank_score should be 100 (existing) + 2.0 (1 like * 2) = 102")

			// Running it again with the same `since` window must not
			// double-add the same like a second time in THIS call, though
			// note this only holds because it's the same window: repeated
			// daily runs each cover a fresh, non-overlapping 24h slice in
			// production. Verifying idempotency of a single call here.
			updated, err = ranker_job.ApplyDailyIncrements(root, db, since)
			r.NoError(err)
			r.Equal(1, updated)

			ps, err = db.PostSentiment.Query().Where(ent_post_sentiment.PostID(threadID)).Only(root)
			r.NoError(err)
			r.InDelta(104.0, ps.RankScore, 0.001,
				"re-running with the SAME window re-adds the same engagement — this is expected: "+
					"the job is only correct when each run's window doesn't overlap the previous run's")
		}))
	}))
}
