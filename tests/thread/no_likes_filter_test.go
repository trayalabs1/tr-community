package thread_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestThreadListNoLikesFilter(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			userCtx, _ := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			adminSession := sh.WithSession(adminCtx)
			userSession := sh.WithSession(userCtx)

			suffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "no-likes-" + suffix,
				Slug:        "no-likes-" + suffix,
				Description: "channel for no_likes filter tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			published := openapi.Published
			bahMeta := openapi.Metadata{"post_category": "BAH"}

			likedResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Liked streak",
				Body:       opt.New[openapi.PostContent]("liked").Ptr(),
				Meta:       &bahMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, likedResp)
			likedID := likedResp.JSON200.Id

			unlikedResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Unliked streak",
				Body:       opt.New[openapi.PostContent]("unliked").Ptr(),
				Meta:       &bahMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, unlikedResp)
			unlikedID := unlikedResp.JSON200.Id

			// Give one thread a like.
			likeResp, err := cl.LikePostAddWithResponse(root, likedID, userSession)
			tests.Ok(t, err, likeResp)

			noLikes := true
			noReplies := true

			t.Run("no_likes_returns_only_unliked", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					NoLikes:        &noLikes,
					PostCategories: &[]string{"BAH"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, unlikedID)
				assert.NotContains(t, ids, likedID)
			})

			t.Run("no_likes_with_sentiment", func(t *testing.T) {
				// BAH posts are prescored neutral, so both match the neutral
				// sentiment, but only the unliked one survives no_likes.
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					NoLikes:        &noLikes,
					PostCategories: &[]string{"BAH"},
					Sentiments:     &[]string{"neutral"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, unlikedID)
				assert.NotContains(t, ids, likedID)
			})

			t.Run("without_no_likes_both_present", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					PostCategories: &[]string{"BAH"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, unlikedID)
				assert.Contains(t, ids, likedID)
			})
		}))
	}))
}
