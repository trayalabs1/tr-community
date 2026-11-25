package thread_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

// TestFeedRankingV4Ordering verifies the v4 final_score feed ordering: freshness
// decay (EXP(-age/86400)) and the 0.1 negative-sentiment multiplier both take
// priority over raw content_score (rank_score).
//
// This is skipped by default: the feed-ordering SQL added for v4 uses
// Postgres-only syntax (EXP, EXTRACT(EPOCH FROM ...), NOW()) via query.Modify in
// app/resources/post/thread_querier/list_threads.go, and this repo's e2e harness
// (internal/integration/fx.go) boots SQLite by default, which does not support
// this syntax. The old v1 ranking SQL had the same Postgres-only limitation, so
// this is a pre-existing, accepted characteristic, not a new gap introduced here.
func TestFeedRankingV4Ordering(t *testing.T) {
	t.Skip("requires Postgres — feed-ordering SQL uses Postgres-only syntax (EXP, EXTRACT); set DATABASE_URL to a Postgres instance and remove this skip to run")

	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
		db *ent.Client,
	) {
		lc.Append(fx.StartHook(func() {
			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			chanResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "feed-ranking-v4-" + suffix,
					Slug:        "feed-ranking-v4-" + suffix,
					Description: "channel for feed ranking v4 ordering tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := chanResp.JSON200.Id

			published := openapi.Published

			mkThread := func(title string) xid.ID {
				resp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title,
						Body:       opt.New[openapi.PostContent]("placeholder body content for ranking test post " + title).Ptr(),
						Visibility: &published,
					}, adminSession),
				)(t, http.StatusOK)

				id, err := xid.FromString(resp.JSON200.Id)
				require.NoError(t, err)
				return id
			}

			setSentiment := func(postID xid.ID, tag string, category scorer.Category, rankScore float64) {
				require.NoError(t, db.PostSentiment.Create().
					SetPostID(postID).
					SetSentimentTag(tag).
					SetPositivityScore(50).
					SetPrimaryTopic(string(category)).
					SetFeedValueScore(int(rankScore)).
					SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
					SetRankScore(rankScore).
					OnConflictColumns(ent_post_sentiment.FieldPostID).
					UpdateNewValues().
					Exec(root))
			}

			backdate := func(postID xid.ID, age time.Duration) {
				require.NoError(t, db.Post.UpdateOneID(postID).
					Modify(func(u *sql.UpdateBuilder) {
						u.Set(ent_post.FieldCreatedAt, time.Now().Add(-age))
					}).
					Exec(root))
			}

			// Post 1: high content_score, but stale (96h old) — heavy freshness decay
			// should push it down despite its high rank_score.
			highScoreOld := mkThread("high score old " + suffix)
			setSentiment(highScoreOld, "positive", scorer.CategoryResultsProgress, 400)
			backdate(highScoreOld, 96*time.Hour)

			// Post 2: low content_score, fresh — minimal decay should let it win
			// over the stale high-score post.
			lowScoreFresh := mkThread("low score fresh " + suffix)
			setSentiment(lowScoreFresh, "neutral", scorer.CategoryNA, 80)

			// Post 3: negative sentiment, fresh, otherwise high content_score — the
			// 0.1 sentiment multiplier should tank its rank below the neutral fresh
			// post despite an even higher rank_score.
			negativeFresh := mkThread("negative fresh high content " + suffix)
			setSentiment(negativeFresh, "negative", scorer.CategorySideEffects, 300)

			listResp := tests.AssertRequest(
				cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, adminSession),
			)(t, http.StatusOK)

			ids := make([]string, len(listResp.JSON200.Threads))
			for i, th := range listResp.JSON200.Threads {
				ids[i] = th.Id
			}

			lowIdx := indexOf(ids, lowScoreFresh.String())
			highOldIdx := indexOf(ids, highScoreOld.String())
			negativeIdx := indexOf(ids, negativeFresh.String())

			require.GreaterOrEqual(t, lowIdx, 0, "fresh low-score post must appear in the channel feed")
			require.GreaterOrEqual(t, highOldIdx, 0, "stale high-score post must appear in the channel feed")
			require.GreaterOrEqual(t, negativeIdx, 0, "fresh negative post must appear in the channel feed")

			assert.Less(t, lowIdx, highOldIdx,
				"a fresh low-content_score post should outrank a stale high-content_score post due to freshness decay")
			assert.Less(t, lowIdx, negativeIdx,
				"a neutral post should outrank an equally-fresh negative-sentiment post due to the 0.1 sentiment multiplier")
		}))
	}))
}

func indexOf(haystack []string, needle string) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return -1
}
