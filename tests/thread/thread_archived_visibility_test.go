package thread_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Southclaws/dt"
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

func TestThreadArchivedVisibility(t *testing.T) {
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
			memberCtx, memberAcc := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			otherCtx, _ := e2e.WithAccount(root, aw, seed.Account_002_Frigg)

			sessionAdmin := sh.WithSession(adminCtx)
			sessionMember := sh.WithSession(memberCtx)
			sessionOther := sh.WithSession(otherCtx)

			chanCreate := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "Archived testing channel",
					Slug:        "archived-testing-channel",
					Description: "channel for archived-visibility tests",
				}, sessionAdmin),
			)(t, http.StatusOK)
			channelID := chanCreate.JSON200.Id

			tests.AssertRequest(
				cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(memberAcc.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, sessionAdmin),
			)(t, http.StatusOK)

			reviewThread := tests.AssertRequest(
				cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Body:       opt.New("<p>A thread to archive</p>").Ptr(),
					Title:      "Archive me",
					Visibility: opt.New(openapi.Review).Ptr(),
				}, sessionMember),
			)(t, http.StatusOK)
			threadID := reviewThread.JSON200.Id
			threadSlug := reviewThread.JSON200.Slug

			archived := openapi.Archived
			archiveResp := tests.AssertRequest(
				cl.ThreadUpdateWithResponse(root, threadSlug, openapi.ThreadMutableProps{
					Visibility: &archived,
				}, sessionAdmin),
			)(t, http.StatusOK)
			r.Equal(openapi.Archived, archiveResp.JSON200.Visibility, "thread should be archived")

			t.Run("admin_review_queue_excludes_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
						Visibility: &[]openapi.Visibility{openapi.Review},
					}, sessionAdmin),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.NotContains(ids, threadID, "archived thread should NOT appear in admin review queue")
			})

			t.Run("admin_default_feed_excludes_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{}, sessionAdmin),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.NotContains(ids, threadID, "archived thread should NOT appear in admin default feed")
			})

			t.Run("owner_default_feed_includes_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{}, sessionMember),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.Contains(ids, threadID, "owner should see their own archived thread in default feed")
			})

			t.Run("owner_profile_includes_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
						Author: &memberAcc.Handle,
					}, sessionMember),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.Contains(ids, threadID, "owner should see their own archived thread on their profile")
			})

			t.Run("other_member_does_not_see_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{}, sessionOther),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.NotContains(ids, threadID, "other member should NOT see another user's archived thread")
			})

			t.Run("unauthenticated_user_does_not_see_archived", func(t *testing.T) {
				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{}),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.NotContains(ids, threadID, "unauthenticated user should NOT see archived thread")
			})

			t.Run("admin_can_restore_archived_to_review", func(t *testing.T) {
				review := openapi.Review
				restoreResp := tests.AssertRequest(
					cl.ThreadUpdateWithResponse(root, threadSlug, openapi.ThreadMutableProps{
						Visibility: &review,
					}, sessionAdmin),
				)(t, http.StatusOK)
				a.Equal(openapi.Review, restoreResp.JSON200.Visibility, "thread should be back in review")

				threadList := tests.AssertRequest(
					cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
						Visibility: &[]openapi.Visibility{openapi.Review},
					}, sessionAdmin),
				)(t, http.StatusOK)

				ids := dt.Map(threadList.JSON200.Threads, func(th openapi.ThreadReference) string { return th.Id })
				a.Contains(ids, threadID, "restored thread should reappear in admin review queue")
			})
		}))
	}))
}
