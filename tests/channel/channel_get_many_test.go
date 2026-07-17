package channel_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
)

func TestChannelGetMany(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		channelRepo *channel.Repository,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			owner := "owner-" + xid.New().String()
			ownerResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: owner,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, ownerResp.StatusCode())
			ownerID := account.AccountID(utils.Must(xid.FromString(ownerResp.JSON200.Id)))
			ownerSession := sh.WithSession(e2e.WithAccountID(root, ownerID))

			suffix := xid.New().String()
			ids := make([]channel.ChannelID, 0, 3)
			for _, name := range []string{"alpha", "beta", "gamma"} {
				createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name: name + "-" + suffix,
					Slug: name + "-" + suffix,
				}, ownerSession)
				r.NoError(err)
				r.Equal(http.StatusOK, createResp.StatusCode())
				ids = append(ids, channel.ChannelID(utils.Must(xid.FromString(string(createResp.JSON200.Id)))))
			}

			got, err := channelRepo.GetMany(root, ids...)
			r.NoError(err)
			r.Len(got, 3)
			for _, id := range ids {
				ch, ok := got[xid.ID(id)]
				r.True(ok)
				r.Equal(id, ch.ID)
			}

			empty, err := channelRepo.GetMany(root)
			r.NoError(err)
			r.Empty(empty)
		}))
	}))
}
