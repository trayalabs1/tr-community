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

func TestChannelMembership(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			// Create test users: owner, admin, member1, member2, non-member
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

			member1 := "member1-" + xid.New().String()
			member1Resp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: member1,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, member1Resp.StatusCode())
			member1ID := account.AccountID(utils.Must(xid.FromString(member1Resp.JSON200.Id)))
			member1Session := sh.WithSession(e2e.WithAccountID(root, member1ID))

			member2 := "member2-" + xid.New().String()
			member2Resp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: member2,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, member2Resp.StatusCode())
			member2ID := account.AccountID(utils.Must(xid.FromString(member2Resp.JSON200.Id)))
			member2Session := sh.WithSession(e2e.WithAccountID(root, member2ID))

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
				Description: "A test channel for membership operations",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			channelID := string(createResp.JSON200.Id)

			// Test: List members - should have just the owner
			listResp, err := cl.ChannelMemberListWithResponse(root, channelID, nil, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp.StatusCode())
			r.NotNil(listResp.JSON200)
			r.Len(listResp.JSON200.Members, 1)
			r.Equal("owner", string(listResp.JSON200.Members[0].Role))
			r.Equal(ownerID.String(), string(listResp.JSON200.Members[0].Account.Id))

			// Test: Owner adds member with admin role
			addAdminResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(adminID.String()),
				Role:      openapi.ChannelMemberAddRoleAdmin,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, addAdminResp.StatusCode())
			r.Equal("admin", string(addAdminResp.JSON200.Role))

			// Test: Owner adds member with member role
			addMember1Resp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member1ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, addMember1Resp.StatusCode())
			r.Equal("member", string(addMember1Resp.JSON200.Role))

			// Test: List members - should now have 3 members
			listResp2, err := cl.ChannelMemberListWithResponse(root, channelID, nil, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp2.StatusCode())
			r.Len(listResp2.JSON200.Members, 3)

			// Test: Admin can add members
			addMember2Resp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member2ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			r.NoError(err)
			r.Equal(http.StatusOK, addMember2Resp.StatusCode())

			// Test: Regular member cannot add members
			addFailResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(nonMemberID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, member1Session)
			r.NoError(err)
			r.True(addFailResp.StatusCode() == http.StatusForbidden || addFailResp.StatusCode() == http.StatusInternalServerError)

			// Test: Update member role - promote member1 to moderator
			updateRoleResp, err := cl.ChannelMemberUpdateRoleWithResponse(root, channelID, member1ID.String(), openapi.ChannelMemberRoleUpdate{
				Role: "moderator",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, updateRoleResp.StatusCode())
			r.Equal("moderator", string(updateRoleResp.JSON200.Role))

			// Test: Non-member joins public channel
			joinResp, err := cl.ChannelJoinWithResponse(root, channelID, nonMemberSession)
			r.NoError(err)
			r.Equal(http.StatusOK, joinResp.StatusCode())
			r.Equal("member", string(joinResp.JSON200.Role))

			// Test: Member leaves channel
			leaveResp, err := cl.ChannelLeaveWithResponse(root, channelID, member2Session)
			r.NoError(err)
			r.Equal(http.StatusOK, leaveResp.StatusCode())

			// Test: List members after joins and leaves
			listResp3, err := cl.ChannelMemberListWithResponse(root, channelID, nil, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp3.StatusCode())
			r.Len(listResp3.JSON200.Members, 4) // owner, admin, member1 (moderator), nonMember

			// Test: Owner removes a member
			removeResp, err := cl.ChannelMemberRemoveWithResponse(root, channelID, nonMemberID.String(), ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, removeResp.StatusCode())

			// Test: Final member list
			listResp4, err := cl.ChannelMemberListWithResponse(root, channelID, nil, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp4.StatusCode())
			r.Len(listResp4.JSON200.Members, 3) // owner, admin, member1 (moderator)
		}))
	}))
}

func TestChannelMembershipGet(t *testing.T) {
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

			nonMember := "nonmember-" + xid.New().String()
			nonMemberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: nonMember,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, nonMemberResp.StatusCode())
			nonMemberID := account.AccountID(utils.Must(xid.FromString(nonMemberResp.JSON200.Id)))
			nonMemberSession := sh.WithSession(e2e.WithAccountID(root, nonMemberID))

			// Create a channel
			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Test Channel",
				Slug:        "test-channel-membership",
				Description: "A test channel for membership get operation",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			channelID := string(createResp.JSON200.Id)

			// Test: Owner gets their own membership
			ownerMembershipResp, err := cl.ChannelMembershipGetWithResponse(root, channelID, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, ownerMembershipResp.StatusCode())
			r.NotNil(ownerMembershipResp.JSON200)
			r.Equal("owner", string(ownerMembershipResp.JSON200.Role))
			r.Equal(ownerID.String(), string(ownerMembershipResp.JSON200.Account.Id))

			// Add a member
			addMemberResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, addMemberResp.StatusCode())

			// Test: Member gets their own membership
			memberMembershipResp, err := cl.ChannelMembershipGetWithResponse(root, channelID, memberSession)
			r.NoError(err)
			r.Equal(http.StatusOK, memberMembershipResp.StatusCode())
			r.NotNil(memberMembershipResp.JSON200)
			r.Equal("member", string(memberMembershipResp.JSON200.Role))
			r.Equal(memberID.String(), string(memberMembershipResp.JSON200.Account.Id))

			// Test: Non-member cannot get membership (should return 404)
			nonMemberMembershipResp, err := cl.ChannelMembershipGetWithResponse(root, channelID, nonMemberSession)
			r.NoError(err)
			r.Equal(http.StatusNotFound, nonMemberMembershipResp.StatusCode())
		}))
	}))
}

func TestChannelMemberListPagination(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			// Create owner
			owner := "owner-" + xid.New().String()
			ownerResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: owner,
				Token:      "password",
			})
			r.NoError(err)
			r.Equal(http.StatusOK, ownerResp.StatusCode())
			ownerID := account.AccountID(utils.Must(xid.FromString(ownerResp.JSON200.Id)))
			ownerSession := sh.WithSession(e2e.WithAccountID(root, ownerID))

			// Create channel
			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Test Channel Pagination",
				Slug:        "test-channel-pagination",
				Description: "A test channel for pagination",
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, createResp.StatusCode())
			channelID := string(createResp.JSON200.Id)

			// Add 15 members to test pagination
			memberIDs := make([]string, 15)
			for i := 0; i < 15; i++ {
				memberHandle := "member-" + xid.New().String()
				memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
					Identifier: memberHandle,
					Token:      "password",
				})
				r.NoError(err)
				r.Equal(http.StatusOK, memberResp.StatusCode())
				memberID := string(memberResp.JSON200.Id)
				memberIDs[i] = memberID

				addResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(memberID),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, ownerSession)
				r.NoError(err)
				r.Equal(http.StatusOK, addResp.StatusCode())
			}

			// Test: Get first page with limit 10
			page1 := 1
			limit1 := 10
			listResp1, err := cl.ChannelMemberListWithResponse(root, channelID, &openapi.ChannelMemberListParams{
				Page:  &page1,
				Limit: &limit1,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp1.StatusCode())
			r.NotNil(listResp1.JSON200)
			r.Len(listResp1.JSON200.Members, 10)
			r.Equal(16, listResp1.JSON200.Total) // 15 members + 1 owner
			r.Equal(1, listResp1.JSON200.Page)
			r.Equal(10, listResp1.JSON200.Limit)

			// Test: Get second page with limit 10
			page2 := 2
			listResp2, err := cl.ChannelMemberListWithResponse(root, channelID, &openapi.ChannelMemberListParams{
				Page:  &page2,
				Limit: &limit1,
			}, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp2.StatusCode())
			r.NotNil(listResp2.JSON200)
			r.Len(listResp2.JSON200.Members, 6) // Remaining 6 members
			r.Equal(16, listResp2.JSON200.Total)
			r.Equal(2, listResp2.JSON200.Page)

			// Test: Get all members with default pagination
			listResp3, err := cl.ChannelMemberListWithResponse(root, channelID, nil, ownerSession)
			r.NoError(err)
			r.Equal(http.StatusOK, listResp3.StatusCode())
			r.NotNil(listResp3.JSON200)
			r.Len(listResp3.JSON200.Members, 16)
			r.Equal(16, listResp3.JSON200.Total)
			r.Equal(1, listResp3.JSON200.Page)
			r.Equal(50, listResp3.JSON200.Limit) // Default limit
		}))
	}))
}
