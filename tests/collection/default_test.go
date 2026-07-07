package collection_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/rs/xid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/opt"
	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
	"github.com/Southclaws/storyden/tests"
)

func TestCollectionDefault(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			// createThread spins up a fresh account that owns its own channel and
			// category, then posts a thread there, returning the account handle,
			// its authenticated session, and the thread IDs.
			createThreads := func(t *testing.T, n int) (string, openapi.RequestEditorFn, []string) {
				handle := xid.New().String()
				acc, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{Identifier: handle, Token: "password"})
				tests.Ok(t, err, acc)
				session := sh.WithSession(e2e.WithAccountID(root, account.AccountID(utils.Must(xid.FromString(acc.JSON200.Id)))))

				channel, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "Default Collection Channel",
					Slug:        "default-collection-" + xid.New().String(),
					Description: "channel for default collection tests",
				}, session)
				tests.Ok(t, err, channel)

				cat, err := cl.ChannelCategoryCreateWithResponse(root, channel.JSON200.Id, openapi.CategoryInitialProps{
					Colour:      "",
					Description: "cat",
					Name:        xid.New().String(),
				}, session)
				tests.Ok(t, err, cat)

				ids := make([]string, 0, n)
				for i := 0; i < n; i++ {
					thread, err := cl.ChannelThreadCreateWithResponse(root, channel.JSON200.Id, openapi.ThreadInitialProps{
						Title:      "thread",
						Body:       opt.New("<p>this is a thread</p>").Ptr(),
						Category:   opt.New(cat.JSON200.Id).Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
					}, session)
					tests.Ok(t, err, thread)
					ids = append(ids, thread.JSON200.Id)
				}

				return handle, session, ids
			}

			t.Run("unauthenticated", func(t *testing.T) {
				t.Parallel()

				_, _, ids := createThreads(t, 1)

				add, err := cl.CollectionAddPostToDefaultWithResponse(root, ids[0])
				tests.Status(t, err, add, http.StatusUnauthorized)
			})

			t.Run("creates_default_then_reuses_it", func(t *testing.T) {
				t.Parallel()
				a := assert.New(t)
				r := require.New(t)

				handle, session, ids := createThreads(t, 2)

				// No collections exist yet: first save creates the default.
				add1, err := cl.CollectionAddPostToDefaultWithResponse(root, ids[0], session)
				tests.Ok(t, err, add1)
				a.Equal("Saved", add1.JSON200.Name)
				r.Len(add1.JSON200.Items, 1)

				list1, err := cl.CollectionListWithResponse(root, &openapi.CollectionListParams{AccountHandle: &handle}, session)
				tests.Ok(t, err, list1)
				defaults1 := lo.Filter(list1.JSON200.Collections, func(c openapi.Collection, _ int) bool { return c.IsDefault })
				r.Len(defaults1, 1)
				a.Equal("Saved", defaults1[0].Name)
				a.Contains(defaults1[0].Slug, "saved")

				// Second save reuses the same default collection.
				add2, err := cl.CollectionAddPostToDefaultWithResponse(root, ids[1], session)
				tests.Ok(t, err, add2)
				a.Equal(add1.JSON200.Id, add2.JSON200.Id)
				r.Len(add2.JSON200.Items, 2)

				list2, err := cl.CollectionListWithResponse(root, &openapi.CollectionListParams{AccountHandle: &handle}, session)
				tests.Ok(t, err, list2)
				defaults2 := lo.Filter(list2.JSON200.Collections, func(c openapi.Collection, _ int) bool { return c.IsDefault })
				r.Len(defaults2, 1)
				a.Equal(add1.JSON200.Id, defaults2[0].Id)
			})

			t.Run("concurrent_first_saves_yield_one_default", func(t *testing.T) {
				t.Parallel()
				r := require.New(t)

				const n = 8
				handle, session, ids := createThreads(t, n)

				var wg sync.WaitGroup
				for i := 0; i < n; i++ {
					wg.Add(1)
					go func(postID string) {
						defer wg.Done()
						_, _ = cl.CollectionAddPostToDefaultWithResponse(root, postID, session)
					}(ids[i])
				}
				wg.Wait()

				list, err := cl.CollectionListWithResponse(root, &openapi.CollectionListParams{AccountHandle: &handle}, session)
				tests.Ok(t, err, list)
				defaults := lo.Filter(list.JSON200.Collections, func(c openapi.Collection, _ int) bool { return c.IsDefault })
				r.Len(defaults, 1, "the partial unique index must guarantee exactly one default collection")
			})
		}))
	}))
}
