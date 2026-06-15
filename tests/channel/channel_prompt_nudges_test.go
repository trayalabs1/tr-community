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

func TestChannelPromptNudges(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
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

			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Prompt Channel",
				Slug:        "prompt-channel-" + xid.New().String(),
				Description: "A channel with posting prompts",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			r.NotNil(createResp.JSON200)
			channelID := string(createResp.JSON200.Id)

			meta := openapi.Metadata{
				"prompt_nudges": []any{
					map[string]any{"icon": "kit", "text": "Just got my Kit 1 — starting today!"},
					map[string]any{"icon": "clock", "text": "When will I see results?"},
				},
			}

			updateResp, err := cl.ChannelUpdateWithResponse(root, channelID, openapi.ChannelMutableProps{
				Meta: &meta,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, updateResp.StatusCode())
			r.NotNil(updateResp.JSON200)
			r.NotNil(updateResp.JSON200.Meta)

			// Round-trip through GET to confirm persistence.
			getResp, err := cl.ChannelGetWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, getResp.StatusCode())
			r.NotNil(getResp.JSON200.Meta)

			nudges, ok := (*getResp.JSON200.Meta)["prompt_nudges"].([]any)
			r.True(ok, "prompt_nudges should round-trip as a list")
			r.Len(nudges, 2)

			first, ok := nudges[0].(map[string]any)
			r.True(ok)
			r.Equal("kit", first["icon"])
			r.Equal("Just got my Kit 1 — starting today!", first["text"])
		}))
	}))
}
