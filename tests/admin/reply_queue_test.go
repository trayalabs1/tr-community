package admin_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
	"github.com/Southclaws/storyden/tests"
)

func TestReplyAdminQueue(t *testing.T) {
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

			handle := xid.New().String()
			memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: handle,
				Token:      "password",
			})
			tests.Ok(t, err, memberResp)
			memberID := utils.Must(xid.FromString(memberResp.JSON200.Id))
			memberCtx := sh.WithSession(e2e.WithAccountID(root, account.AccountID(memberID)))

			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "test-channel-" + xid.New().String(),
			}, sh.WithSession(adminCtx))
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			threadResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title: "Test thread",
				Body:  opt.New[openapi.PostContent]("Thread body").Ptr(),
			}, sh.WithSession(adminCtx))
			tests.Ok(t, err, threadResp)
			threadSlug := threadResp.JSON200.Slug

			replyResp, err := cl.ReplyCreateWithResponse(root, threadSlug, openapi.ReplyInitialProps{
				Body: "This is my question about the product",
			}, memberCtx)
			tests.Ok(t, err, replyResp)
			replyID := replyResp.JSON200.Id

			t.Run("non_admin_cannot_list_queue", func(t *testing.T) {
				resp, err := cl.AdminReplyQueueListWithResponse(root, &openapi.AdminReplyQueueListParams{},
					memberCtx)
				require.NoError(t, err)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode())
			})

			t.Run("admin_can_list_queue_and_sees_reply", func(t *testing.T) {
				resp, err := cl.AdminReplyQueueListWithResponse(root, &openapi.AdminReplyQueueListParams{},
					sh.WithSession(adminCtx))
				tests.Ok(t, err, resp)
				require.NotNil(t, resp.JSON200)

				ids := make([]openapi.Identifier, len(resp.JSON200.ReplyQueueEntries))
				for i, e := range resp.JSON200.ReplyQueueEntries {
					ids[i] = e.ReplyId
				}
				assert.Contains(t, ids, replyID)
			})

			t.Run("admin_reply_does_not_enter_queue", func(t *testing.T) {
				_, err := cl.ReplyCreateWithResponse(root, threadSlug, openapi.ReplyInitialProps{
					Body: "Admin response",
				}, sh.WithSession(adminCtx))
				require.NoError(t, err)

				resp, err := cl.AdminReplyQueueListWithResponse(root, &openapi.AdminReplyQueueListParams{},
					sh.WithSession(adminCtx))
				tests.Ok(t, err, resp)

				for _, e := range resp.JSON200.ReplyQueueEntries {
					assert.NotEqual(t, "Admin response", e.ContentSnippet)
				}
			})

			t.Run("admin_can_dismiss_entry", func(t *testing.T) {
				listResp, err := cl.AdminReplyQueueListWithResponse(root, &openapi.AdminReplyQueueListParams{},
					sh.WithSession(adminCtx))
				tests.Ok(t, err, listResp)

				var entryID openapi.Identifier
				for _, e := range listResp.JSON200.ReplyQueueEntries {
					if e.ReplyId == replyID {
						entryID = e.Id
						break
					}
				}
				require.NotEmpty(t, entryID, "expected queue entry for member reply")

				dismissResp, err := cl.AdminReplyQueueDismissWithResponse(root, entryID,
					sh.WithSession(adminCtx))
				require.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, dismissResp.StatusCode())

				listResp2, err := cl.AdminReplyQueueListWithResponse(root, &openapi.AdminReplyQueueListParams{},
					sh.WithSession(adminCtx))
				tests.Ok(t, err, listResp2)
				for _, e := range listResp2.JSON200.ReplyQueueEntries {
					assert.NotEqual(t, replyID, e.ReplyId)
				}
			})
		}))
	}))
}
