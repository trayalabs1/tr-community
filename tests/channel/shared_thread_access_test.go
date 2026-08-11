package channel_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
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

// A thread shared into a channel should be reachable by that channel's members.
// The deep-link gate can only allow this if the thread response says which
// channels the thread has been shared into, so ThreadGet exposes the reverse
// reference_post_id lookup as shared_to_channel_ids.
func TestSharedThreadChannelExposure(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			adminSession := sh.WithSession(adminCtx)
			memberSession := sh.WithSession(memberCtx)

			chSrc, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "shared-src", Slug: "shared-src", Description: "source",
			}, adminSession)
			tests.Ok(t, err, chSrc)
			sourceChannel := chSrc.JSON200.Id

			chDst, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "shared-dst", Slug: "shared-dst", Description: "destination",
			}, adminSession)
			tests.Ok(t, err, chDst)
			destChannel := chDst.JSON200.Id

			// The member joins ONLY the destination channel.
			addResp, err := cl.ChannelMemberAddWithResponse(root, destChannel, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addResp)

			orig, err := cl.ChannelThreadCreateWithResponse(root, sourceChannel, openapi.ThreadInitialProps{
				Title:      "Original in the source channel",
				Body:       opt.New("<p>original body</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, orig)
			originalID := orig.JSON200.Id

			t.Run("before sharing, the original lists no shared channels", func(t *testing.T) {
				got, err := cl.ThreadGetWithResponse(root, orig.JSON200.Slug, nil, memberSession)
				tests.Ok(t, err, got)
				a.Empty(got.JSON200.SharedToChannelIds,
					"an unshared thread must not report any shared-into channels")
			})

			share, err := cl.ChannelThreadCreateWithResponse(root, destChannel, openapi.ThreadInitialProps{
				Title:           "Shared into the destination channel",
				Body:            opt.New("<p>share subtitle</p>").Ptr(),
				Visibility:      opt.New(openapi.Published).Ptr(),
				ReferencePostId: (*openapi.Identifier)(&originalID),
			}, adminSession)
			tests.Ok(t, err, share)

			t.Run("the original reports the channel it was shared into", func(t *testing.T) {
				got, err := cl.ThreadGetWithResponse(root, orig.JSON200.Slug, nil, memberSession)
				tests.Ok(t, err, got)

				r.NotNil(got.JSON200.SharedToChannelIds)
				ids := *got.JSON200.SharedToChannelIds

				a.Contains(ids, openapi.Identifier(destChannel),
					"the destination channel must be listed so the gate can allow its members")
				a.NotContains(ids, openapi.Identifier(sourceChannel),
					"the thread's own home channel is not a shared-into channel")
			})

			t.Run("home channel is still reported separately", func(t *testing.T) {
				got, err := cl.ThreadGetWithResponse(root, orig.JSON200.Slug, nil, adminSession)
				tests.Ok(t, err, got)
				a.Equal(sourceChannel, got.JSON200.ChannelId,
					"sharing must not move the original")
			})

			t.Run("the share row itself reports no shared channels", func(t *testing.T) {
				got, err := cl.ThreadGetWithResponse(root, share.JSON200.Slug, nil, memberSession)
				tests.Ok(t, err, got)
				a.Empty(got.JSON200.SharedToChannelIds,
					"a share is not itself shared anywhere; only originals accumulate shares")
			})

			// Mirrors what the /t/locate gate computes: the member can reach the
			// original because it is shared into a channel they belong to, even
			// though its home channel is invisible to them.
			t.Run("member can resolve access via the shared channel", func(t *testing.T) {
				got, err := cl.ThreadGetWithResponse(root, orig.JSON200.Slug, nil, memberSession)
				tests.Ok(t, err, got)

				list, err := cl.ChannelListWithResponse(root, memberSession)
				tests.Ok(t, err, list)

				accessible := map[string]bool{}
				for _, c := range list.JSON200.Channels {
					accessible[c.Id] = true
				}

				a.False(accessible[got.JSON200.ChannelId],
					"the home channel must be invisible to this member")

				r.NotNil(got.JSON200.SharedToChannelIds)
				reachable := false
				for _, id := range *got.JSON200.SharedToChannelIds {
					if accessible[string(id)] {
						reachable = true
					}
				}
				a.True(reachable,
					"the member must reach the thread through the channel it was shared into")
			})

			t.Run("an unrelated member gets no access via shares", func(t *testing.T) {
				outsiderCtx, _ := e2e.WithAccount(root, aw, seed.Account_004_Loki)
				outsiderSession := sh.WithSession(outsiderCtx)

				got, err := cl.ThreadGetWithResponse(root, orig.JSON200.Slug, nil, outsiderSession)
				tests.Ok(t, err, got)

				list, err := cl.ChannelListWithResponse(root, outsiderSession)
				tests.Ok(t, err, list)

				accessible := map[string]bool{}
				for _, c := range list.JSON200.Channels {
					accessible[c.Id] = true
				}

				a.False(accessible[got.JSON200.ChannelId], "home channel not accessible")

				for _, id := range *got.JSON200.SharedToChannelIds {
					a.False(accessible[string(id)],
						"an outsider must not reach the thread through any shared channel")
				}
			})
		}))
	}))
}
