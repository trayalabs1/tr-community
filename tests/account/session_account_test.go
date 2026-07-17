package account_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
)

func TestGetByIDForSessionReturnsAccountWithRoles(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		accountQuery *account_querier.Querier,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			handle := "session-" + xid.New().String()
			signup, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{Identifier: handle, Token: "password"})
			r.NoError(err)
			r.Equal(http.StatusOK, signup.StatusCode())
			id := account.AccountID(utils.Must(xid.FromString(signup.JSON200.Id)))

			acc, err := accountQuery.GetByIDForSession(root, id)
			r.NoError(err)
			r.Equal(id, acc.ID)
			r.Equal(handle, acc.Handle)
			r.NotEmpty(acc.Roles)
		}))
	}))
}
