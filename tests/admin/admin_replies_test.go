package admin_test

import (
	"context"
	"net/http"
	"testing"
	"time"

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

func TestAdminReplyList(t *testing.T) {
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

			adminACtx, adminAAcc := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminBCtx, adminBAcc := e2e.WithAccount(root, aw, seed.Account_002_Frigg)
			adminASession := sh.WithSession(adminACtx)
			adminBSession := sh.WithSession(adminBCtx)

			handle := xid.New().String()
			memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: handle,
				Token:      "password",
			})
			tests.Ok(t, err, memberResp)
			memberID := utils.Must(xid.FromString(memberResp.JSON200.Id))
			memberSession := sh.WithSession(e2e.WithAccountID(root, account.AccountID(memberID)))

			channelSuffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "admin-replies-" + channelSuffix,
				Slug:        "admin-replies-" + channelSuffix,
				Description: "channel for admin reply filter tests",
			}, adminASession)
			tests.Ok(t, err, channelResp)

			threadResp, err := cl.ChannelThreadCreateWithResponse(root, channelResp.JSON200.Id, openapi.ThreadInitialProps{
				Title: "Thread for admin reply filtering",
				Body:  opt.New[openapi.PostContent]("thread body").Ptr(),
			}, adminASession)
			tests.Ok(t, err, threadResp)
			threadSlug := threadResp.JSON200.Slug
			threadID := threadResp.JSON200.Id

			// Admin A replies, admin B replies, and a non-admin replies. Only the
			// two admin replies should ever surface from this endpoint.
			replyA, err := cl.ReplyCreateWithResponse(root, threadSlug, openapi.ReplyInitialProps{
				Body: "reply from admin A",
			}, adminASession)
			tests.Ok(t, err, replyA)

			replyB, err := cl.ReplyCreateWithResponse(root, threadSlug, openapi.ReplyInitialProps{
				Body: "reply from admin B",
			}, adminBSession)
			tests.Ok(t, err, replyB)

			memberReply, err := cl.ReplyCreateWithResponse(root, threadSlug, openapi.ReplyInitialProps{
				Body: "reply from a normal member",
			}, memberSession)
			tests.Ok(t, err, memberReply)

			idsOf := func(resp *openapi.AdminReplyListResponse) []string {
				ids := make([]string, 0, len(resp.JSON200.Replies))
				for _, rep := range resp.JSON200.Replies {
					ids = append(ids, rep.Id)
				}
				return ids
			}

			t.Run("non_admin_is_forbidden", func(t *testing.T) {
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{}, memberSession)
				require.NoError(t, err)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode())
			})

			t.Run("admin_list_endpoint_returns_admins", func(t *testing.T) {
				resp, err := cl.AdminAccountListWithResponse(root, adminASession)
				tests.Ok(t, err, resp)

				handles := make([]string, 0, len(resp.JSON200.Admins))
				for _, adm := range resp.JSON200.Admins {
					handles = append(handles, adm.Handle)
				}

				a.Contains(handles, adminAAcc.Handle)
				a.Contains(handles, adminBAcc.Handle)
				a.NotContains(handles, handle, "non-admin must not appear in the admin list")
			})

			t.Run("unfiltered_returns_only_admin_replies", func(t *testing.T) {
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{}, adminASession)
				tests.Ok(t, err, resp)

				ids := idsOf(resp)
				a.Contains(ids, replyA.JSON200.Id)
				a.Contains(ids, replyB.JSON200.Id)
				a.NotContains(ids, memberReply.JSON200.Id, "non-admin reply must be excluded")
				a.NotContains(ids, threadID, "the thread itself must never appear as a reply")
			})

			t.Run("filtering_by_admin_excludes_other_admins", func(t *testing.T) {
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					RepliedBy: opt.New(openapi.Identifier(adminAAcc.ID.String())).Ptr(),
				}, adminASession)
				tests.Ok(t, err, resp)

				ids := idsOf(resp)
				a.Contains(ids, replyA.JSON200.Id, "admin A's reply must be present")
				a.NotContains(ids, replyB.JSON200.Id, "admin B's reply must NOT appear when filtering by admin A")
				a.NotContains(ids, memberReply.JSON200.Id)

				for _, rep := range resp.JSON200.Replies {
					a.Equal(adminAAcc.ID.String(), rep.Author.Id, "every row must be authored by the filtered admin")
				}
			})

			t.Run("reply_rows_carry_thread_context", func(t *testing.T) {
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					RepliedBy: opt.New(openapi.Identifier(adminAAcc.ID.String())).Ptr(),
				}, adminASession)
				tests.Ok(t, err, resp)
				r.NotEmpty(resp.JSON200.Replies)

				var found *openapi.Reply
				for i, rep := range resp.JSON200.Replies {
					if rep.Id == replyA.JSON200.Id {
						found = &resp.JSON200.Replies[i]
						break
					}
				}
				r.NotNil(found, "admin A's reply should be in the result")

				a.Equal(threadID, found.RootId, "reply must point at its thread")
				a.Equal("Thread for admin reply filtering", found.Title, "reply must carry the thread title for display")
				a.NotEmpty(found.RootSlug, "reply must carry the thread slug for permalinks")
			})

			t.Run("date_range_filters_on_reply_date", func(t *testing.T) {
				// A window ending before the replies were made must exclude them,
				// even though the thread and replies all exist.
				past := time.Now().UTC().Add(-24 * time.Hour)
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					CreatedBefore: &past,
				}, adminASession)
				tests.Ok(t, err, resp)
				a.Empty(resp.JSON200.Replies, "replies created now must fall outside a window that ended yesterday")

				// A window covering now must include them.
				from := time.Now().UTC().Add(-1 * time.Hour)
				to := time.Now().UTC().Add(1 * time.Hour)
				inRange, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					CreatedAfter:  &from,
					CreatedBefore: &to,
				}, adminASession)
				tests.Ok(t, err, inRange)
				a.Contains(idsOf(inRange), replyA.JSON200.Id)
			})

			t.Run("admin_and_date_filters_compose", func(t *testing.T) {
				from := time.Now().UTC().Add(-1 * time.Hour)
				to := time.Now().UTC().Add(1 * time.Hour)
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					RepliedBy:     opt.New(openapi.Identifier(adminBAcc.ID.String())).Ptr(),
					CreatedAfter:  &from,
					CreatedBefore: &to,
				}, adminASession)
				tests.Ok(t, err, resp)

				ids := idsOf(resp)
				a.Contains(ids, replyB.JSON200.Id)
				a.NotContains(ids, replyA.JSON200.Id, "admin filter must still apply alongside the date range")
			})

			t.Run("filtering_by_non_admin_is_rejected", func(t *testing.T) {
				resp, err := cl.AdminReplyListWithResponse(root, &openapi.AdminReplyListParams{
					RepliedBy: opt.New(openapi.Identifier(memberID.String())).Ptr(),
				}, adminASession)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode(),
					"a non-admin account is not a valid replied_by value")
			})
		}))
	}))
}
