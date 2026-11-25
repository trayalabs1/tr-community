package thread_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestChannelThreadStatsDailyUsers(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			a := assert.New(t)
			r := require.New(t)

			ownerCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			ownerSession := sh.WithSession(ownerCtx)

			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Daily Users " + uuid.NewString(),
				Slug: "daily-users-" + uuid.NewString(),
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, channelResp.StatusCode())
			channelID := channelResp.JSON200.Id

			t.Run("requires_session", func(t *testing.T) {
				resp, err := cl.ChannelThreadStatsDailyUsersWithResponse(root, string(channelID))
				r.NoError(err)
				a.Equal(http.StatusUnauthorized, resp.StatusCode())
			})

			t.Run("returns_count_for_member", func(t *testing.T) {
				resp, err := cl.ChannelThreadStatsDailyUsersWithResponse(root, string(channelID), ownerSession)
				r.NoError(err)
				r.Equal(http.StatusOK, resp.StatusCode())
				r.NotNil(resp.JSON200)
				a.GreaterOrEqual(resp.JSON200.Count, 0)
			})
		}))
	}))
}
