package channel_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

// Interchannel sharing (pointer model): a share is a thread created with a
// reference_post_id pointing at the original. It is a real post in the
// destination channel; interaction is routed to the original, so replies/likes
// on the share row itself are rejected.
func TestInterchannelThreadSharingPointer(t *testing.T) {
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

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)   // admin
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur) // non-admin

			adminSession := sh.WithSession(adminCtx)
			memberSession := sh.WithSession(memberCtx)

			chSrc, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Month 8 Warriors", Slug: "m8-warriors", Description: "source",
			}, adminSession)
			tests.Ok(t, err, chSrc)
			sourceChannel := chSrc.JSON200.Id

			chDst, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Month 2 Community", Slug: "m2-community", Description: "destination",
			}, adminSession)
			tests.Ok(t, err, chDst)
			destChannel := chDst.JSON200.Id

			addResp, err := cl.ChannelMemberAddWithResponse(root, destChannel, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addResp)

			// Original thread in the source channel (admin authors it; the admin
			// created the channel so has access).
			orig, err := cl.ChannelThreadCreateWithResponse(root, sourceChannel, openapi.ThreadInitialProps{
				Title:      "Hair regrowth finally visible",
				Body:       opt.New("<p>8 months of consistency.</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, orig)
			originalID := orig.JSON200.Id
			originalSlug := orig.JSON200.Slug

			t.Run("non-admin cannot create a share", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, destChannel, openapi.ThreadInitialProps{
					Title:           "shared",
					Body:            opt.New("<p>subtitle</p>").Ptr(),
					Visibility:      opt.New(openapi.Published).Ptr(),
					ReferencePostId: (*openapi.Identifier)(&originalID),
				}, memberSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "non-admin share must be rejected")
			})

			var shareID string
			var shareSlug string
			t.Run("admin creates a share into the destination channel", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, destChannel, openapi.ThreadInitialProps{
					Title:           "Hair regrowth finally visible",
					Body:            opt.New("<p>Sharing this for cohort 2</p>").Ptr(),
					Visibility:      opt.New(openapi.Published).Ptr(),
					ReferencePostId: (*openapi.Identifier)(&originalID),
				}, adminSession)
				tests.Ok(t, err, resp)
				shareID = resp.JSON200.Id
				shareSlug = resp.JSON200.Slug

				r.NotNil(resp.JSON200.ReferencePostId, "share must carry reference_post_id")
				r.Equal(originalID, string(*resp.JSON200.ReferencePostId))
			})

			t.Run("share is fetchable and points at the original", func(t *testing.T) {
				resp, err := cl.ThreadGetWithResponse(root, shareID, nil, memberSession)
				tests.Ok(t, err, resp)
				r.NotNil(resp.JSON200.ReferencePostId)
				r.Equal(originalID, string(*resp.JSON200.ReferencePostId))
			})

			t.Run("original has no reference (it is a normal thread)", func(t *testing.T) {
				resp, err := cl.ThreadGetWithResponse(root, originalID, nil, memberSession)
				tests.Ok(t, err, resp)
				r.Nil(resp.JSON200.ReferencePostId)
			})

			t.Run("interaction on the share row is rejected — reply", func(t *testing.T) {
				resp, err := cl.ChannelReplyCreateWithResponse(root, destChannel, openapi.ThreadMark(shareSlug), openapi.ReplyInitialProps{
					Body: "<p>should be rejected</p>",
				}, memberSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "reply on a share row must be rejected")
			})

			t.Run("interaction on the share row is rejected — like", func(t *testing.T) {
				resp, err := cl.LikePostAddWithResponse(root, shareID, memberSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "like on a share row must be rejected")
			})

			t.Run("interaction on the ORIGINAL still works — reply", func(t *testing.T) {
				// Admin is a member of the source channel (created it).
				resp, err := cl.ChannelReplyCreateWithResponse(root, sourceChannel, openapi.ThreadMark(originalSlug), openapi.ReplyInitialProps{
					Body: "<p>reply on original</p>",
				}, adminSession)
				tests.Ok(t, err, resp)
			})
		}))
	}))
}
