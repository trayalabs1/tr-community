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
	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

// TestFeedCacheDisable verifies that FEED_CACHE_DISABLE bypasses the channel
// feed's cached ordered-sequence layer entirely: a thread created after the
// feed has already been read (and, under normal caching, would populate the
// 30-minute cache) must still appear on the very next read.
func TestFeedCacheDisable(t *testing.T) {
	// The channel thread-list endpoint always uses the v4 sentiment-ranking
	// SQL (list_threads.go), which is Postgres-only syntax (EXP, EXTRACT) and
	// fails against this repo's default SQLite e2e harness — the same
	// pre-existing, accepted limitation as TestChannelThreads. Run with a
	// Postgres DATABASE_URL and remove this skip to verify.
	t.Skip("requires Postgres — channel thread listing uses Postgres-only ranking SQL; set DATABASE_URL to a Postgres instance and remove this skip to run")

	t.Parallel()

	integration.Test(t, &config.Config{FeedCacheDisable: true}, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			channelResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "feed-cache-disable-" + suffix,
					Slug:        "feed-cache-disable-" + suffix,
					Description: "channel for feed cache disable test",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			mkThread := func(title string) string {
				resp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title,
						Body:       opt.New[openapi.PostContent]("body for " + title).Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
					}, adminSession),
				)(t, http.StatusOK)
				return resp.JSON200.Id
			}

			firstID := mkThread("first " + suffix)

			// First read: under normal caching this would populate the
			// 30-minute feed-sequence cache for this channel.
			list1 := tests.AssertRequest(
				cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, adminSession),
			)(t, http.StatusOK)
			ids1 := make([]string, len(list1.JSON200.Threads))
			for i, th := range list1.JSON200.Threads {
				ids1[i] = th.Id
			}
			r.Contains(ids1, firstID)

			secondID := mkThread("second " + suffix)

			// Second read, immediately after: with the cache in its normal
			// mode this would still only see the sequence cached above and
			// miss the newly created thread for up to 30 minutes. With
			// FEED_CACHE_DISABLE set, it must be visible immediately.
			list2 := tests.AssertRequest(
				cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, adminSession),
			)(t, http.StatusOK)
			ids2 := make([]string, len(list2.JSON200.Threads))
			for i, th := range list2.JSON200.Threads {
				ids2[i] = th.Id
			}
			r.Contains(ids2, secondID, "feed cache disable must serve fresh results, not the previously cached sequence")
			r.Contains(ids2, firstID)
		}))
	}))
}
