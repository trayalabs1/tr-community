package thread_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestThreadListBAHFilter(t *testing.T) {
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
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "bah-filter-" + suffix,
				Slug:        "bah-filter-" + suffix,
				Description: "channel for BAH filter tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			bahMeta := openapi.Metadata{"post_category": "BAH"}
			published := openapi.Published
			review := openapi.Review

			// Published BAH thread, no replies.
			bahPublishedResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "BAH published streak",
				Body:       opt.New[openapi.PostContent]("streak body").Ptr(),
				Meta:       &bahMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, bahPublishedResp)
			bahPublishedSlug := bahPublishedResp.JSON200.Slug
			bahPublishedID := bahPublishedResp.JSON200.Id

			// Review BAH thread, no replies — must be visible to admin requesting review.
			bahReviewResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "BAH review streak",
				Body:       opt.New[openapi.PostContent]("review streak body").Ptr(),
				Meta:       &bahMeta,
				Visibility: &review,
			}, adminSession)
			tests.Ok(t, err, bahReviewResp)
			bahReviewID := bahReviewResp.JSON200.Id

			// Published non-BAH thread, no replies.
			plainResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Plain published",
				Body:       opt.New[openapi.PostContent]("plain body").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, plainResp)
			plainID := plainResp.JSON200.Id

			visibilities := openapi.VisibilityParam{openapi.Published, openapi.Review}
			noReplies := true

			t.Run("bah_only_returns_only_bah_including_review", func(t *testing.T) {
				bahOnly := true
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					NoReplies:  &noReplies,
					BahOnly:    &bahOnly,
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, bahPublishedID)
				assert.Contains(t, ids, bahReviewID)
				assert.NotContains(t, ids, plainID)
			})

			t.Run("exclude_bah_excludes_all_bah", func(t *testing.T) {
				excludeBah := true
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					NoReplies:  &noReplies,
					ExcludeBah: &excludeBah,
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, plainID)
				assert.NotContains(t, ids, bahPublishedID)
				assert.NotContains(t, ids, bahReviewID)
			})

			t.Run("admin_reply_removes_thread_from_no_replies", func(t *testing.T) {
				bahOnly := true

				before, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					NoReplies:  &noReplies,
					BahOnly:    &bahOnly,
				}, adminSession)
				tests.Ok(t, err, before)
				require.Contains(t, threadIDs(before.JSON200.Threads), bahPublishedID)

				replyResp, err := cl.ReplyCreateWithResponse(root, bahPublishedSlug, openapi.ReplyInitialProps{
					Body: "Great consistency, keep it up!",
				}, adminSession)
				tests.Ok(t, err, replyResp)

				after, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					NoReplies:  &noReplies,
					BahOnly:    &bahOnly,
				}, adminSession)
				tests.Ok(t, err, after)
				assert.NotContains(t, threadIDs(after.JSON200.Threads), bahPublishedID)
			})
		}))
	}))
}

func threadIDs(threads openapi.ThreadList) []openapi.Identifier {
	ids := make([]openapi.Identifier, len(threads))
	for i, t := range threads {
		ids[i] = t.Id
	}
	return ids
}
