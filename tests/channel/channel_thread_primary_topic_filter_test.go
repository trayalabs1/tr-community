package channel_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

// TestChannelThreadListFiltersByPrimaryTopic verifies the primary_topics
// query param on ChannelThreadList restricts results to posts whose
// post_sentiments.primary_topic matches one of the requested values.
//
// Skipped by default: ChannelThreadList always sets UseSentimentRanking,
// whose ORDER BY (app/resources/post/thread_querier/list_threads.go) uses
// Postgres-only syntax (EXP, EXTRACT(EPOCH FROM ...), NOW()), which this
// repo's e2e harness (internal/integration/fx.go) can't run against its
// default SQLite database. Same pre-existing, accepted limitation as
// TestFeedRankingV4Ordering in tests/thread/feed_ranking_v4_test.go.
func TestChannelThreadListFiltersByPrimaryTopic(t *testing.T) {
	t.Skip("requires Postgres — ChannelThreadList's ranking ORDER BY uses Postgres-only syntax (EXP, EXTRACT); set DATABASE_URL to a Postgres instance and remove this skip to run. Verified passing against a scratch Postgres database during development.")

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

			suffix := xid.New().String()
			channelResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "primary-topic-filter-" + suffix,
					Slug:        "primary-topic-filter-" + suffix,
					Description: "channel for primary_topics filter tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			createThread := func(title string) xid.ID {
				resp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title + " " + suffix,
						Body:       opt.New[openapi.PostContent]("body").Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
					}, adminSession),
				)(t, http.StatusOK)
				id, err := xid.FromString(resp.JSON200.Id)
				r.NoError(err)
				return id
			}

			setPrimaryTopic := func(postID xid.ID, topic string) {
				err := db.PostSentiment.Create().
					SetPostID(postID).
					SetSentimentTag("neutral").
					SetPrimaryTopic(topic).
					SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
					OnConflictColumns(ent_post_sentiment.FieldPostID).
					UpdateNewValues().
					Exec(root)
				r.NoError(err)
			}

			hairfall := createThread("hairfall thread")
			setPrimaryTopic(hairfall, "HAIRFALL CONCERNS")

			sideEffects := createThread("side effects thread")
			setPrimaryTopic(sideEffects, "SIDE EFFECTS")

			unscored := createThread("unscored thread")
			_ = unscored

			listResp := tests.AssertRequest(
				cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{
					PrimaryTopics: &[]string{"HAIRFALL CONCERNS"},
				}, adminSession),
			)(t, http.StatusOK)

			ids := make([]string, len(listResp.JSON200.Threads))
			for i, th := range listResp.JSON200.Threads {
				ids[i] = th.Id
			}

			r.Contains(ids, hairfall.String(), "thread classified as the requested topic must be included")
			r.NotContains(ids, sideEffects.String(), "thread classified as a different topic must be excluded")
			r.NotContains(ids, unscored.String(), "unscored thread must be excluded when filtering by topic")

			listAllResp := tests.AssertRequest(
				cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, adminSession),
			)(t, http.StatusOK)

			idsAll := make([]string, len(listAllResp.JSON200.Threads))
			for i, th := range listAllResp.JSON200.Threads {
				idsAll[i] = th.Id
			}
			r.Contains(idsAll, hairfall.String())
			r.Contains(idsAll, sideEffects.String())
			r.Contains(idsAll, unscored.String(), "without a topic filter, unscored threads must still be returned")
		}))
	}))
}
