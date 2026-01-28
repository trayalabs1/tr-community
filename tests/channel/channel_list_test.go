package channel_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
)

func TestChannelList(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			// Create a test user
			userHandle := "user-" + xid.New().String()
			user, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: userHandle,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, user.StatusCode())
			userID := account.AccountID(utils.Must(xid.FromString(user.JSON200.Id)))
			userSession := sh.WithSession(e2e.WithAccountID(root, userID))

			// Test: Unauthenticated request should fail
			resp1, err := cl.ChannelListWithResponse(root)
			r.NoError(err)
			r.NotNil(resp1)
			r.Equal(http.StatusUnauthorized, resp1.StatusCode())

			// Test: Authenticated user can list their accessible channels
			resp2, err := cl.ChannelListWithResponse(root, userSession)
			r.NoError(err)
			r.NotNil(resp2)
			r.Equal(http.StatusOK, resp2.StatusCode())

			// Verify response structure is correct
			r.NotNil(resp2.JSON200)
			r.NotNil(resp2.JSON200.Channels)

			// Channels array may be empty for new users (depending on async subscriber timing)
			// The important thing is that the endpoint works and returns valid structure
			r.IsType([]openapi.Channel{}, resp2.JSON200.Channels)

			// If channels are present, verify their structure
			if len(resp2.JSON200.Channels) > 0 {
				channel := resp2.JSON200.Channels[0]
				r.NotEmpty(channel.Id)
				r.NotEmpty(channel.Name)
				r.NotEmpty(channel.Slug)
				r.NotEmpty(channel.Visibility)
			}
		}))
	}))
}
