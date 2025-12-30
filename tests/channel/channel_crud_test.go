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

func TestChannelCRUD(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			// Create test users
			owner := "owner-" + xid.New().String()
			ownerResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: owner,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, ownerResp.StatusCode())
			ownerID := account.AccountID(utils.Must(xid.FromString(ownerResp.JSON200.Id)))
			ownerSession := sh.WithSession(e2e.WithAccountID(root, ownerID))

			member := "member-" + xid.New().String()
			memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: member,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, memberResp.StatusCode())
			memberID := account.AccountID(utils.Must(xid.FromString(memberResp.JSON200.Id)))
			memberSession := sh.WithSession(e2e.WithAccountID(root, memberID))

			// Test: Create a channel
			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Test Channel",
				Slug:        "test-channel",
				Description: "A test channel for CRUD operations",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			r.NotNil(createResp.JSON200)

			channelID := string(createResp.JSON200.Id)
			r.NotEmpty(channelID)
			r.Equal("Test Channel", createResp.JSON200.Name)
			r.Equal("test-channel", createResp.JSON200.Slug)
			r.Equal("A test channel for CRUD operations", createResp.JSON200.Description)
			r.Equal("public", string(createResp.JSON200.Visibility))

			// Test: Owner can get their channel
			getResp, err := cl.ChannelGetWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, getResp.StatusCode())
			r.NotNil(getResp.JSON200)
			r.Equal(channelID, string(getResp.JSON200.Id))
			r.Equal("Test Channel", getResp.JSON200.Name)

			// Test: Non-member cannot get private channel
			// First update to private
			visibility := openapi.Private
			updateResp, err := cl.ChannelUpdateWithResponse(root, channelID, openapi.ChannelMutableProps{
				Visibility: &visibility,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, updateResp.StatusCode())
			r.NotNil(updateResp.JSON200)
			r.Equal("private", string(updateResp.JSON200.Visibility))

			// Non-member tries to access
			getNonMemberResp, err := cl.ChannelGetWithResponse(root, channelID, memberSession)
			r.NoError(err)
			// Should get 403 or 500 (wrapped fault error)
			r.True(getNonMemberResp.StatusCode() == http.StatusForbidden || getNonMemberResp.StatusCode() == http.StatusInternalServerError)

			// Test: Update channel properties
			newName := "Updated Channel"
			newDesc := "Updated description"
			visibilityPublic := openapi.Public
			updatePropsResp, err := cl.ChannelUpdateWithResponse(root, channelID, openapi.ChannelMutableProps{
				Name:        &newName,
				Description: &newDesc,
				Visibility:  &visibilityPublic,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, updatePropsResp.StatusCode())
			r.NotNil(updatePropsResp.JSON200)
			r.Equal("Updated Channel", updatePropsResp.JSON200.Name)
			r.Equal("Updated description", updatePropsResp.JSON200.Description)
			r.Equal("public", string(updatePropsResp.JSON200.Visibility))

			// Test: Delete channel
			deleteResp, err := cl.ChannelDeleteWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, deleteResp.StatusCode())

			// Test: Channel no longer exists
			getDeletedResp, err := cl.ChannelGetWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.NotEqual(http.StatusOK, getDeletedResp.StatusCode())
		}))
	}))
}
