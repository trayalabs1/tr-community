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

			t.Run("ThreadGet exposes the home channel name (for the shared card pill)", func(t *testing.T) {
				// The share card names the source channel from the referenced
				// thread's own payload — no separate (membership-gated) lookup.
				resp, err := cl.ThreadGetWithResponse(root, originalID, nil, memberSession)
				tests.Ok(t, err, resp)
				r.NotNil(resp.JSON200.Channel, "thread must carry its home channel ref")
				r.Equal("Month 8 Warriors", resp.JSON200.Channel.Name)
			})

			t.Run("cannot share into the thread's own channel", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, sourceChannel, openapi.ThreadInitialProps{
					Title:           "self share",
					Body:            opt.New("<p>x</p>").Ptr(),
					Visibility:      opt.New(openapi.Published).Ptr(),
					ReferencePostId: (*openapi.Identifier)(&originalID),
				}, adminSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "sharing into the home channel must be rejected")
			})

			t.Run("cannot share into a channel it is already shared into", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, destChannel, openapi.ThreadInitialProps{
					Title:           "dup share",
					Body:            opt.New("<p>x</p>").Ptr(),
					Visibility:      opt.New(openapi.Published).Ptr(),
					ReferencePostId: (*openapi.Identifier)(&originalID),
				}, adminSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "duplicate share into the same channel must be rejected")
			})

			t.Run("cannot share a share (re-share)", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, sourceChannel, openapi.ThreadInitialProps{
					Title:           "share of a share",
					Body:            opt.New("<p>x</p>").Ptr(),
					Visibility:      opt.New(openapi.Published).Ptr(),
					ReferencePostId: (*openapi.Identifier)(&shareID),
				}, adminSession)
				r.NoError(err)
				r.GreaterOrEqual(resp.StatusCode(), 400, "sharing a share must be rejected")
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

			// Member joins the source channel so they can reply on the original.
			addSrc, err := cl.ChannelMemberAddWithResponse(root, sourceChannel, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addSrc)

			var memberReplyID string
			var adminReplyID string
			t.Run("interaction on the ORIGINAL still works — reply", func(t *testing.T) {
				memberReply, err := cl.ChannelReplyCreateWithResponse(root, sourceChannel, openapi.ThreadMark(originalSlug), openapi.ReplyInitialProps{
					Body: "<p>member reply</p>",
				}, memberSession)
				tests.Ok(t, err, memberReply)
				memberReplyID = memberReply.JSON200.Id

				adminReply, err := cl.ChannelReplyCreateWithResponse(root, sourceChannel, openapi.ThreadMark(originalSlug), openapi.ReplyInitialProps{
					Body: "<p>admin reply</p>",
				}, adminSession)
				tests.Ok(t, err, adminReply)
				adminReplyID = adminReply.JSON200.Id
			})

			t.Run("member reply is cohort-tagged; admin reply is not", func(t *testing.T) {
				resp, err := cl.ThreadGetWithResponse(root, originalID, nil, adminSession)
				tests.Ok(t, err, resp)

				byID := map[string]*openapi.Reply{}
				for i := range resp.JSON200.Replies.Replies {
					rep := &resp.JSON200.Replies.Replies[i]
					byID[rep.Id] = rep
				}

				memberReply := byID[memberReplyID]
				r.NotNil(memberReply, "member reply must be present")
				r.NotNil(memberReply.CohortChannel, "member reply must carry a cohort tag")

				adminReply := byID[adminReplyID]
				r.NotNil(adminReply, "admin reply must be present")
				r.Nil(adminReply.CohortChannel, "admin reply must NOT carry a cohort tag")
			})
		}))
	}))
}
