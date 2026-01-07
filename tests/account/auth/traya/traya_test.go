package traya_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/channel_membership"
	"github.com/Southclaws/storyden/app/services/authentication/provider/traya"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestTrayaAuth(t *testing.T) {
	t.Parallel()

	mockTrayaAPI := setupMockTrayaAPI()
	defer mockTrayaAPI.Close()

	integration.Test(t, nil, e2e.Setup(), fx.Decorate(func(cfg config.Config) config.Config {
		u, _ := url.Parse(mockTrayaAPI.URL)
		cfg.TrayaAPIServerURL = *u
		return cfg
	}), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		accountQuery *account_querier.Querier,
		channelRepo *channel.Repository,
		membershipRepo *channel_membership.Repository,
		cfg config.Config,
	) {
		lc.Append(fx.StartHook(func() {
			createTestChannels(t, root, channelRepo)

			t.Run("authenticate_new_user", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				token := "valid-token-1"

				auth, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: token,
				})
				tests.Ok(t, err, auth)

				a.NotEmpty(auth.HTTPResponse.Header.Get("Set-Cookie"))
				accountID := account.AccountID(openapi.GetAccountID(auth.JSON200.Id))

				acc, err := accountQuery.GetByID(root, accountID)
				r.NoError(err)
				a.Equal("priyankarjha70@gmail.com", acc.EmailAddresses[0].Email.Address)
				a.Equal("Priyankar Jha", acc.Name)

				memberships, err := membershipRepo.ListByAccount(root, accountID)
				r.NoError(err)
				r.Len(memberships, 3)

				channelSlugs := make([]string, len(memberships))
				for i, m := range memberships {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs[i] = ch.Slug
				}

				a.Contains(channelSlugs, "general")
				a.Contains(channelSlugs, "month-0-to-3")
				a.Contains(channelSlugs, "month-3-to-6")
			})

			t.Run("authenticate_existing_user", func(t *testing.T) {
				a := assert.New(t)

				token := "valid-token-1"

				auth1, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: token,
				})
				tests.Ok(t, err, auth1)

				accountID1 := account.AccountID(openapi.GetAccountID(auth1.JSON200.Id))

				auth2, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: token,
				})
				tests.Ok(t, err, auth2)

				accountID2 := account.AccountID(openapi.GetAccountID(auth2.JSON200.Id))

				a.Equal(accountID1, accountID2)
			})

			t.Run("authenticate_invalid_token", func(t *testing.T) {
				a := assert.New(t)

				auth, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: "invalid-token",
				})
				a.NoError(err)
				a.True(auth.StatusCode() == http.StatusUnauthorized || auth.StatusCode() == http.StatusForbidden)
			})

			t.Run("authenticate_high_kit_count", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				token := "valid-token-high-kit"

				auth, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: token,
				})
				tests.Ok(t, err, auth)

				accountID := account.AccountID(openapi.GetAccountID(auth.JSON200.Id))

				memberships, err := membershipRepo.ListByAccount(root, accountID)
				r.NoError(err)
				r.Len(memberships, 4)

				channelSlugs := make([]string, len(memberships))
				for i, m := range memberships {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs[i] = ch.Slug
				}

				a.Contains(channelSlugs, "general")
				a.Contains(channelSlugs, "month-0-to-3")
				a.Contains(channelSlugs, "month-3-to-6")
				a.Contains(channelSlugs, "month-6-to-9")
			})
		}))
	}))
}

func setupMockTrayaAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/community-user" {
			http.NotFound(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		// Extract token from "Bearer <token>" format
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == "invalid-token" || authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var response traya.TrayaUserResponse

		switch token {
		case "valid-token-1":
			response = traya.TrayaUserResponse{
				User: struct {
					ID          string  `json:"id"`
					Email       string  `json:"email"`
					PhoneNumber string  `json:"phone_number"`
					FirstName   string  `json:"first_name"`
					LastName    *string `json:"last_name"`
					Gender      string  `json:"gender"`
					CreatedAt   string  `json:"createdAt"`
				}{
					ID:          "98a13e6d-7af6-4bd0-9314-064aceb1b589",
					Email:       "priyankarjha70@gmail.com",
					PhoneNumber: "+917352020264",
					FirstName:   "Priyankar Jha",
					LastName:    nil,
					Gender:      "M",
					CreatedAt:   "2025-06-01T05:15:05.564Z",
				},
				Case: struct {
					ID                      string  `json:"id"`
					LatestOrderCount        int     `json:"latest_order_count"`
					LatestOrderDate         string  `json:"latest_order_date"`
					LatestOrderDisplayID    string  `json:"latest_order_display_id"`
					LatestOrderID           string  `json:"latest_order_id"`
					LatestOrderStatus       string  `json:"latest_order_status"`
					LatestOrderDeliveryDate *string `json:"latest_order_delivery_date"`
					FormStatus              string  `json:"form_status"`
				}{
					ID:                   "ab2f7b49-5628-41dd-bf56-4c963023d058",
					LatestOrderCount:     5,
					LatestOrderDate:      "2025-12-28T01:12:59.909Z",
					LatestOrderDisplayID: "#T9362642",
					LatestOrderID:        "d70eb299-d679-4857-9c94-8a7509a7cea3",
					LatestOrderStatus:    "shipped",
					FormStatus:           "draft",
				},
				ChatURL:                "https://wa.me/918828006272?text=Hi! I want to speak to a hair expert now!",
				TotalKitCount:          5,
				RunningMonthForHairKit: 5,
				CustomerSlug: struct {
					SlugName any `json:"slugName"`
				}{
					SlugName: "o5_o_to_d",
				},
			}
		case "valid-token-high-kit":
			response = traya.TrayaUserResponse{
				User: struct {
					ID          string  `json:"id"`
					Email       string  `json:"email"`
					PhoneNumber string  `json:"phone_number"`
					FirstName   string  `json:"first_name"`
					LastName    *string `json:"last_name"`
					Gender      string  `json:"gender"`
					CreatedAt   string  `json:"createdAt"`
				}{
					ID:          "high-kit-user-id",
					Email:       "highkit@storyden.org",
					PhoneNumber: "+911234567890",
					FirstName:   "High Kit",
					LastName:    stringPtr("User"),
					Gender:      "M",
					CreatedAt:   "2025-01-01T00:00:00.000Z",
				},
				Case: struct {
					ID                      string  `json:"id"`
					LatestOrderCount        int     `json:"latest_order_count"`
					LatestOrderDate         string  `json:"latest_order_date"`
					LatestOrderDisplayID    string  `json:"latest_order_display_id"`
					LatestOrderID           string  `json:"latest_order_id"`
					LatestOrderStatus       string  `json:"latest_order_status"`
					LatestOrderDeliveryDate *string `json:"latest_order_delivery_date"`
					FormStatus              string  `json:"form_status"`
				}{
					ID:                   "high-kit-case-id",
					LatestOrderCount:     8,
					LatestOrderDate:      "2025-12-28T01:12:59.909Z",
					LatestOrderDisplayID: "#T9999999",
					LatestOrderID:        "high-kit-order-id",
					LatestOrderStatus:    "shipped",
					FormStatus:           "completed",
				},
				ChatURL:                "https://wa.me/918828006272",
				TotalKitCount:          8,
				RunningMonthForHairKit: 8,
				CustomerSlug: struct {
					SlugName any `json:"slugName"`
				}{
					SlugName: "o8_o_to_d",
				},
			}
		default:
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func createTestChannels(t *testing.T, ctx context.Context, repo *channel.Repository) {
	channels := []struct {
		name string
		slug string
	}{
		{"General", "general"},
		{"Month 0-3", "month-0-to-3"},
		{"Month 3-6", "month-3-to-6"},
		{"Month 6-9", "month-6-to-9"},
		{"Month 9-12", "month-9-to-12"},
	}

	for _, ch := range channels {
		_, err := repo.Create(ctx, ch.name, ch.slug, "Test channel for "+ch.slug, channel.VisibilityPrivate)
		if err != nil {
			t.Logf("Warning: failed to create channel %s: %v", ch.slug, err)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
