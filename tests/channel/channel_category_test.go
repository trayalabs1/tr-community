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

func TestChannelCategories(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			// Create test users: owner, admin, member, non-member
			owner := "owner-" + xid.New().String()
			ownerResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: owner,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, ownerResp.StatusCode())
			ownerID := account.AccountID(utils.Must(xid.FromString(ownerResp.JSON200.Id)))
			ownerSession := sh.WithSession(e2e.WithAccountID(root, ownerID))

			admin := "admin-" + xid.New().String()
			adminResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: admin,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, adminResp.StatusCode())
			adminID := account.AccountID(utils.Must(xid.FromString(adminResp.JSON200.Id)))
			adminSession := sh.WithSession(e2e.WithAccountID(root, adminID))

			member := "member-" + xid.New().String()
			memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: member,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, memberResp.StatusCode())
			memberID := account.AccountID(utils.Must(xid.FromString(memberResp.JSON200.Id)))
			memberSession := sh.WithSession(e2e.WithAccountID(root, memberID))

			nonMember := "nonmember-" + xid.New().String()
			nonMemberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: nonMember,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, nonMemberResp.StatusCode())
			nonMemberID := account.AccountID(utils.Must(xid.FromString(nonMemberResp.JSON200.Id)))
			nonMemberSession := sh.WithSession(e2e.WithAccountID(root, nonMemberID))

			// Test: Create a channel
			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Test Channel",
				Slug:        "test-channel",
				Description: "A test channel for category operations",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			channelID := string(createResp.JSON200.Id)

			// Add admin and member to channel
			_, err = cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(adminID.String()),
				Role:      openapi.ChannelMemberAddRoleAdmin,
			}, ownerSession)
			r.NoError(err)

			_, err = cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			r.NoError(err)

			// Test: List categories in empty channel - should be empty
			listResp, err := cl.ChannelCategoryListWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp.StatusCode())
			r.NotNil(listResp.JSON200)
			r.Len(listResp.JSON200.Categories, 0)

			// Test: Owner creates a category in the channel
			catName := "General Discussion"
			catSlug := "general-discussion"
			catDesc := "A category for general topics"
			catColour := "#FF6B6B"
			createCatResp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        catName,
				Slug:        &catSlug,
				Description: catDesc,
				Colour:      catColour,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createCatResp.StatusCode())
			r.NotNil(createCatResp.JSON200)
			r.Equal(catName, createCatResp.JSON200.Name)
			r.Equal(catSlug, createCatResp.JSON200.Slug)
			r.Equal(catDesc, createCatResp.JSON200.Description)
			r.Equal(catColour, createCatResp.JSON200.Colour)

			// Test: Admin can create categories
			cat2Name := "Announcements"
			cat2Slug := "announcements"
			cat2Desc := "Official announcements"
			cat2Colour := "#4ECDC4"
			createCat2Resp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        cat2Name,
				Slug:        &cat2Slug,
				Description: cat2Desc,
				Colour:      cat2Colour,
			}, adminSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createCat2Resp.StatusCode())
			category2Slug := createCat2Resp.JSON200.Slug

			// Test: Regular member cannot create categories
			createCat3Resp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        "Unauthorized",
				Description: "Should fail",
				Colour:      "#000000",
			}, memberSession)
			r.NoError(err)
			r.True(createCat3Resp.StatusCode() == http.StatusForbidden || createCat3Resp.StatusCode() == http.StatusInternalServerError)

			// Test: Non-member cannot create categories
			createCat4Resp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        "Unauthorized",
				Description: "Should fail",
				Colour:      "#000000",
			}, nonMemberSession)
			r.NoError(err)
			r.True(createCat4Resp.StatusCode() == http.StatusForbidden || createCat4Resp.StatusCode() == http.StatusInternalServerError)

			// Test: List categories - should now have 2 categories
			listResp2, err := cl.ChannelCategoryListWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp2.StatusCode())
			r.Len(listResp2.JSON200.Categories, 2)

			// Test: Member can list categories
			listResp3, err := cl.ChannelCategoryListWithResponse(root, channelID, memberSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp3.StatusCode())
			r.Len(listResp3.JSON200.Categories, 2)

			// Test: Non-member cannot list categories in channel
			listResp4, err := cl.ChannelCategoryListWithResponse(root, channelID, nonMemberSession)
			r.NoError(err)
			r.True(listResp4.StatusCode() == http.StatusForbidden || listResp4.StatusCode() == http.StatusInternalServerError)

			// Test: Get a specific category
			getCatResp, err := cl.ChannelCategoryGetWithResponse(root, channelID, catSlug, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, getCatResp.StatusCode())
			r.NotNil(getCatResp.JSON200)

			// Test: Member can get category
			getCatResp2, err := cl.ChannelCategoryGetWithResponse(root, channelID, catSlug, memberSession)
			r.NoError(err)
			r.Equal(http.StatusOK, getCatResp2.StatusCode())

			// Test: Non-member cannot get category
			getCatResp3, err := cl.ChannelCategoryGetWithResponse(root, channelID, catSlug, nonMemberSession)
			r.NoError(err)
			r.True(getCatResp3.StatusCode() == http.StatusForbidden || getCatResp3.StatusCode() == http.StatusInternalServerError)

			// Test: Owner updates category
			newName := "General Chat"
			newDesc := "Updated description"
			newColour := "#95E1D3"
			updateCatResp, err := cl.ChannelCategoryUpdateWithResponse(root, channelID, catSlug, openapi.CategoryMutableProps{
				Name:        &newName,
				Description: &newDesc,
				Colour:      &newColour,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, updateCatResp.StatusCode())
			r.Equal(newName, updateCatResp.JSON200.Name)
			r.Equal(newDesc, updateCatResp.JSON200.Description)
			r.Equal(newColour, updateCatResp.JSON200.Colour)

			// Test: Admin can update category
			adminNewName := "Official News"
			adminUpdateResp, err := cl.ChannelCategoryUpdateWithResponse(root, channelID, category2Slug, openapi.CategoryMutableProps{
				Name: &adminNewName,
			}, adminSession)
			r.NoError(err)
			r.Equal(http.StatusOK, adminUpdateResp.StatusCode())

			// Test: Member cannot update category
			updateCat2Resp, err := cl.ChannelCategoryUpdateWithResponse(root, channelID, catSlug, openapi.CategoryMutableProps{
				Name: &newName,
			}, memberSession)
			r.NoError(err)
			r.True(updateCat2Resp.StatusCode() == http.StatusForbidden || updateCat2Resp.StatusCode() == http.StatusInternalServerError)

			// Test: Delete category - need another category to move posts to
			deleteCatResp, err := cl.ChannelCategoryDeleteWithResponse(root, channelID, catSlug, openapi.CategoryDeleteProps{
				MoveTo: openapi.Identifier(createCat2Resp.JSON200.Id),
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, deleteCatResp.StatusCode())

			// Test: Category no longer accessible after delete
			getCatResp4, err := cl.ChannelCategoryGetWithResponse(root, channelID, catSlug, ownerSession)
			r.NoError(err)
			r.NotEqual(http.StatusOK, getCatResp4.StatusCode())

			// Test: Final category list should have 1 category
			listResp5, err := cl.ChannelCategoryListWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp5.StatusCode())
			r.Len(listResp5.JSON200.Categories, 1)
		}))
	}))
}
