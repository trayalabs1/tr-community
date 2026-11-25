package notification_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/account/notification"
	"github.com/Southclaws/storyden/app/resources/account/notification/notify_writer"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestNotificationListPagination(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
		nw *notify_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			userCtx, userAcc := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			userSession := sh.WithSession(userCtx)

			_, otherAcc := e2e.WithAccount(root, aw, seed.Account_003_Baldur)

			const total = 55
			for i := 0; i < total; i++ {
				_, err := nw.Notification(root, userAcc.ID, notification.EventPostLike, opt.NewEmpty[datagraph.ItemRef](), opt.New(otherAcc.ID))
				r.NoError(err)
			}

			_, err := nw.Notification(root, otherAcc.ID, notification.EventPostLike, opt.NewEmpty[datagraph.ItemRef](), opt.New(userAcc.ID))
			r.NoError(err)

			page1, err := cl.NotificationListWithResponse(root, &openapi.NotificationListParams{}, userSession)
			tests.Ok(t, err, page1)
			a.Len(page1.JSON200.Notifications, 50)
			a.Equal(1, page1.JSON200.CurrentPage)
			a.Equal(50, page1.JSON200.PageSize)
			a.Equal(50, page1.JSON200.Results)
			a.Equal(2, page1.JSON200.TotalPages)
			r.NotNil(page1.JSON200.NextPage)
			a.Equal(2, *page1.JSON200.NextPage)

			pageTwo := "2"
			page2, err := cl.NotificationListWithResponse(root, &openapi.NotificationListParams{Page: &pageTwo}, userSession)
			tests.Ok(t, err, page2)
			a.Len(page2.JSON200.Notifications, 5)
			a.Equal(2, page2.JSON200.CurrentPage)
			a.Equal(5, page2.JSON200.Results)
			a.Equal(2, page2.JSON200.TotalPages)
			a.Nil(page2.JSON200.NextPage)

			seen := map[string]struct{}{}
			for _, n := range page1.JSON200.Notifications {
				seen[n.Id] = struct{}{}
			}
			for _, n := range page2.JSON200.Notifications {
				seen[n.Id] = struct{}{}
			}
			a.Len(seen, total, "pages must not overlap and must cover every notification of the owner only")
		}))
	}))
}

func TestNotificationListStatusFilter(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
		nw *notify_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			userCtx, userAcc := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			userSession := sh.WithSession(userCtx)

			var ids []string
			for i := 0; i < 3; i++ {
				n, err := nw.Notification(root, userAcc.ID, notification.EventFollow, opt.NewEmpty[datagraph.ItemRef](), opt.New(userAcc.ID))
				r.NoError(err)
				ids = append(ids, n.ID.String())
			}

			statusRead := openapi.Read
			upd, err := cl.NotificationUpdateWithResponse(root, ids[0], openapi.NotificationMutableProps{Status: &statusRead}, userSession)
			tests.Ok(t, err, upd)

			list := func(statuses ...openapi.NotificationStatus) []openapi.Notification {
				params := &openapi.NotificationListParams{}
				if len(statuses) > 0 {
					s := openapi.NotificationStatusQuery(statuses)
					params.Status = &s
				}
				resp, err := cl.NotificationListWithResponse(root, params, userSession)
				tests.Ok(t, err, resp)
				return resp.JSON200.Notifications
			}

			unread := list(openapi.Unread)
			a.Len(unread, 2)
			for _, n := range unread {
				a.Equal(openapi.Unread, n.Status)
				a.NotEqual(ids[0], n.Id)
			}

			read := list(openapi.Read)
			r.Len(read, 1)
			a.Equal(ids[0], read[0].Id)
			a.Equal(openapi.Read, read[0].Status)

			a.Len(list(openapi.Read, openapi.Unread), 3)
			a.Len(list(), 3)
		}))
	}))
}
