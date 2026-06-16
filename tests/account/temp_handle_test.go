package account_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
)

func TestAdminRegenerateTempHandles(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		accountWrite *account_writer.Writer,
		accountQuery *account_querier.Querier,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminHandle := "admin-" + xid.New().String()
			admin, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{Identifier: adminHandle, Token: "password"})
			r.NoError(err)
			r.Equal(http.StatusOK, admin.StatusCode())
			adminID := account.AccountID(utils.Must(xid.FromString(admin.JSON200.Id)))
			adminSession := sh.WithSession(e2e.WithAccountID(root, adminID))
			_, err = accountWrite.Update(root, adminID, account_writer.SetAdmin(true))
			r.NoError(err)

			// Create an account and force it into a temporary-handle state.

			tempSignup, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{Identifier: "temped-" + xid.New().String(), Token: "password"})
			r.NoError(err)
			r.Equal(http.StatusOK, tempSignup.StatusCode())
			tempID := account.AccountID(utils.Must(xid.FromString(tempSignup.JSON200.Id)))

			tempHandle := "temp_" + xid.New().String()
			_, err = accountWrite.Update(root, tempID,
				account_writer.SetHandle(tempHandle),
				account_writer.SetName("Priyanka"))
			r.NoError(err)

			// Non-admin cannot trigger the regeneration.

			randomSignup, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{Identifier: "rando-" + xid.New().String(), Token: "password"})
			r.NoError(err)
			randomID := account.AccountID(utils.Must(xid.FromString(randomSignup.JSON200.Id)))
			randomSession := sh.WithSession(e2e.WithAccountID(root, randomID))

			forbidden, err := cl.AdminRegenerateTempHandlesWithResponse(root, &openapi.AdminRegenerateTempHandlesParams{}, randomSession)
			r.NoError(err)
			r.Equal(http.StatusForbidden, forbidden.StatusCode())

			// Admin drains all batches using the returned keyset cursor.

			batch := 100
			var cursor *string
			var totalUpdated int
			for {
				res, err := cl.AdminRegenerateTempHandlesWithResponse(root,
					&openapi.AdminRegenerateTempHandlesParams{Limit: &batch, After: cursor}, adminSession)
				r.NoError(err)
				r.Equal(http.StatusOK, res.StatusCode())
				r.NotNil(res.JSON200)

				totalUpdated += res.JSON200.Updated

				if res.JSON200.NextCursor == nil {
					break
				}
				cursor = res.JSON200.NextCursor
			}
			r.GreaterOrEqual(totalUpdated, 1)

			// The temp account now has a regenerated handle derived from its name.

			acc, err := accountQuery.GetByID(root, tempID)
			r.NoError(err)
			r.False(strings.HasPrefix(acc.Handle, "temp_"), "handle should no longer be temporary, got %q", acc.Handle)
			r.True(strings.HasPrefix(acc.Handle, "priy"), "handle should derive from name, got %q", acc.Handle)

			// An invalid cursor is rejected.

			bad := "not-an-xid"
			badRes, err := cl.AdminRegenerateTempHandlesWithResponse(root,
				&openapi.AdminRegenerateTempHandlesParams{After: &bad}, adminSession)
			r.NoError(err)
			r.Equal(http.StatusBadRequest, badRes.StatusCode())
		}))
	}))
}
