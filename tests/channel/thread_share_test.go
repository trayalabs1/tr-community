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

func TestInterchannelThreadSharing(t *testing.T) {
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

			// Odin is a seeded administrator - the only role permitted to share.
			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			memberACtx, memberA := e2e.WithAccount(root, aw, seed.Account_002_Frigg)
			memberBCtx, memberB := e2e.WithAccount(root, aw, seed.Account_003_Baldur)

			adminSession := sh.WithSession(adminCtx)
			memberASession := sh.WithSession(memberACtx)
			memberBSession := sh.WithSession(memberBCtx)

			// Two cohorts: A (source) and B (destination).
			chA, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Cohort A", Slug: "cohort-a", Description: "source cohort",
			}, adminSession)
			tests.Ok(t, err, chA)
			channelA := chA.JSON200.Id

			chB, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Cohort B", Slug: "cohort-b", Description: "destination cohort",
			}, adminSession)
			tests.Ok(t, err, chB)
			channelB := chB.JSON200.Id

			// Members belong to their respective cohorts.
			addA, err := cl.ChannelMemberAddWithResponse(root, channelA, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberA.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addA)

			addB, err := cl.ChannelMemberAddWithResponse(root, channelB, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberB.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addB)

			// A published root thread lives in cohort A.
			thread, err := cl.ChannelThreadCreateWithResponse(root, channelA, openapi.ThreadInitialProps{
				Title:      "Shared thread",
				Body:       opt.New("<p>hello from cohort A</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, memberASession)
			tests.Ok(t, err, thread)
			threadID := thread.JSON200.Id
			threadSlug := thread.JSON200.Slug

			subtitle := "Featured for cohort B"

			t.Run("non-admin cannot share", func(t *testing.T) {
				// Baldur (memberB) is the only non-admin of the seeded accounts.
				resp, err := cl.ThreadShareCreateWithResponse(root, openapi.ThreadMark(threadSlug), openapi.ThreadShareCreateProps{
					Channels: []openapi.Identifier{openapi.Identifier(channelB)},
					Subtitle: &subtitle,
				}, memberBSession)
				r.NoError(err)
				r.NotNil(resp)
				r.GreaterOrEqual(resp.StatusCode(), 400, "non-admin share must be rejected")
			})

			t.Run("admin shares the thread into cohort B", func(t *testing.T) {
				resp, err := cl.ThreadShareCreateWithResponse(root, openapi.ThreadMark(threadSlug), openapi.ThreadShareCreateProps{
					Channels: []openapi.Identifier{openapi.Identifier(channelB)},
					Subtitle: &subtitle,
				}, adminSession)
				tests.Ok(t, err, resp)
			})

			// NOTE: The channel feed listing (ChannelThreadList) uses Postgres-only
			// sentiment-ranking SQL (INTERVAL literals) that SQLite cannot execute,
			// so destination-feed membership is asserted end-to-end under
			// `task test:e2e` (Postgres). Here we verify the share exists via the
			// share-list endpoint, which exercises the same HasChannel/share join.

			t.Run("share list reflects the destination", func(t *testing.T) {
				resp, err := cl.ThreadShareListWithResponse(root, openapi.ThreadMark(threadSlug), adminSession)
				tests.Ok(t, err, resp)
				r.Len(resp.JSON200.Shares, 1)
				r.Equal(channelB, resp.JSON200.Shares[0].Channel.Id)
				r.NotNil(resp.JSON200.Shares[0].Subtitle)
				r.Equal(subtitle, *resp.JSON200.Shares[0].Subtitle)
			})

			var replyBID string
			t.Run("cohort B member can reply to the shared thread", func(t *testing.T) {
				resp, err := cl.ChannelReplyCreateWithResponse(root, channelB, openapi.ThreadMark(threadSlug), openapi.ReplyInitialProps{
					Body: "<p>reply from cohort B</p>",
				}, memberBSession)
				tests.Ok(t, err, resp)
				replyBID = resp.JSON200.Id
			})

			t.Run("cohort A member can reply to the same thread", func(t *testing.T) {
				resp, err := cl.ChannelReplyCreateWithResponse(root, channelA, openapi.ThreadMark(threadSlug), openapi.ReplyInitialProps{
					Body: "<p>reply from cohort A</p>",
				}, memberASession)
				tests.Ok(t, err, resp)
			})

			t.Run("both cohorts' replies appear in one comment tree, B tagged with origin", func(t *testing.T) {
				resp, err := cl.ThreadGetWithResponse(root, threadID, nil, memberBSession)
				tests.Ok(t, err, resp)

				var replyB *openapi.Reply
				var replyA *openapi.Reply
				bodies := map[string]struct{}{}
				for i := range resp.JSON200.Replies.Replies {
					rep := resp.JSON200.Replies.Replies[i]
					bodies[rep.Body] = struct{}{}
					if rep.Id == replyBID {
						replyB = &resp.JSON200.Replies.Replies[i]
					}
					if rep.Body == "<body><p>reply from cohort A</p></body>" {
						replyA = &resp.JSON200.Replies.Replies[i]
					}
				}
				r.Contains(bodies, "<body><p>reply from cohort B</p></body>")
				r.Contains(bodies, "<body><p>reply from cohort A</p></body>")

				// Every reply carries its true origin cohort; the client decides
				// (viewer-relative) whether to render the tag.
				r.NotNil(replyB, "cohort B reply must be present")
				r.NotNil(replyB.OriginChannel, "cohort B reply must carry its origin cohort")
				r.Equal(channelB, replyB.OriginChannel.Id)

				r.NotNil(replyA, "cohort A reply must be present")
				r.NotNil(replyA.OriginChannel, "cohort A reply must carry its origin cohort")
				r.Equal(channelA, replyA.OriginChannel.Id)
			})

			t.Run("admin can pin then unpin the thread in cohort B", func(t *testing.T) {
				pinResp, err := cl.ThreadSharePinWithResponse(root, channelB, openapi.ThreadMark(threadSlug), openapi.ThreadSharePinProps{
					Pinned: true,
				}, adminSession)
				tests.Ok(t, err, pinResp)
				r.True(pinResp.JSON200.Pinned)

				unpinResp, err := cl.ThreadSharePinWithResponse(root, channelB, openapi.ThreadMark(threadSlug), openapi.ThreadSharePinProps{
					Pinned: false,
				}, adminSession)
				tests.Ok(t, err, unpinResp)
				r.False(unpinResp.JSON200.Pinned)
			})

			t.Run("unshare removes the thread from cohort B and re-share succeeds", func(t *testing.T) {
				del, err := cl.ThreadShareDeleteWithResponse(root, openapi.ThreadMark(threadSlug), channelB, adminSession)
				tests.Ok(t, err, del)

				listResp, err := cl.ThreadShareListWithResponse(root, openapi.ThreadMark(threadSlug), adminSession)
				tests.Ok(t, err, listResp)
				r.Empty(listResp.JSON200.Shares, "unshare must clear the active share")

				// Re-sharing into the same cohort must not trip the unique index
				// (partial index excludes soft-deleted rows).
				reshare, err := cl.ThreadShareCreateWithResponse(root, openapi.ThreadMark(threadSlug), openapi.ThreadShareCreateProps{
					Channels: []openapi.Identifier{openapi.Identifier(channelB)},
					Subtitle: &subtitle,
				}, adminSession)
				tests.Ok(t, err, reshare)

				listAgain, err := cl.ThreadShareListWithResponse(root, openapi.ThreadMark(threadSlug), adminSession)
				tests.Ok(t, err, listAgain)
				r.Len(listAgain.JSON200.Shares, 1, "re-share must recreate one active share")
			})
		}))
	}))
}
