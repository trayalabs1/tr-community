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
				r.Len(memberships, 5) // 1 cohort + 4 topic channels

				channelSlugs := make([]string, len(memberships))
				for i, m := range memberships {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs[i] = ch.Slug
				}

				// Male, orderCount=5, within 60 days -> month-5-masters + 4 male topic channels
				a.Contains(channelSlugs, "month-5-masters")
				a.Contains(channelSlugs, "general")
				a.Contains(channelSlugs, "stress-sleep-nutrition")
				a.Contains(channelSlugs, "digestion-metabolism-gut-health")
				a.Contains(channelSlugs, "dandruff-hair-health")
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
				r.Len(memberships, 5) // 1 cohort + 4 topic channels

				channelSlugs := make([]string, len(memberships))
				for i, m := range memberships {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs[i] = ch.Slug
				}

				// Male, orderCount=8, within 60 days -> month-6-8-veterans + 4 male topic channels
				a.Contains(channelSlugs, "month-6-8-veterans")
				a.Contains(channelSlugs, "general")
				a.Contains(channelSlugs, "stress-sleep-nutrition")
				a.Contains(channelSlugs, "digestion-metabolism-gut-health")
				a.Contains(channelSlugs, "dandruff-hair-health")
			})

			t.Run("channel_transition_on_kit_count_change", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				// First authenticate with kit count 5
				token := "valid-token-1"
				auth1, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: token,
				})
				tests.Ok(t, err, auth1)

				accountID := account.AccountID(openapi.GetAccountID(auth1.JSON200.Id))

				// Verify initial memberships (kit count 5)
				memberships1, err := membershipRepo.ListByAccount(root, accountID)
				r.NoError(err)

				channelSlugs1 := make([]string, len(memberships1))
				for i, m := range memberships1 {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs1[i] = ch.Slug
				}

				a.Contains(channelSlugs1, "month-5-masters")
				a.NotContains(channelSlugs1, "month-6-8-veterans")

				// Authenticate again with updated kit count (same user, different token)
				auth2, err := cl.AuthTrayaTokenWithResponse(root, &openapi.AuthTrayaTokenParams{
					Token: "valid-token-1-updated",
				})
				tests.Ok(t, err, auth2)

				accountID2 := account.AccountID(openapi.GetAccountID(auth2.JSON200.Id))
				a.Equal(accountID, accountID2) // Should be same user

				// Verify updated memberships (kit count 8)
				memberships2, err := membershipRepo.ListByAccount(root, accountID2)
				r.NoError(err)
				r.Len(memberships2, 5) // Still 1 cohort + 4 topic channels

				channelSlugs2 := make([]string, len(memberships2))
				for i, m := range memberships2 {
					ch, err := channelRepo.Get(root, channel.ChannelID(m.ChannelID))
					r.NoError(err)
					channelSlugs2[i] = ch.Slug
				}

				// Verify transition: removed from month-5-masters, added to month-6-8-veterans
				a.Contains(channelSlugs2, "month-6-8-veterans")
				a.NotContains(channelSlugs2, "month-5-masters")

				// Topic channels should remain the same
				a.Contains(channelSlugs2, "general")
				a.Contains(channelSlugs2, "stress-sleep-nutrition")
				a.Contains(channelSlugs2, "digestion-metabolism-gut-health")
				a.Contains(channelSlugs2, "dandruff-hair-health")
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
		case "valid-token-1-updated":
			// Same user as valid-token-1, but with updated kit count (simulating user progression)
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
					ID:          "98a13e6d-7af6-4bd0-9314-064aceb1b589", // Same ID as valid-token-1
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
					LatestOrderCount:     8, // Updated from 5 to 8
					LatestOrderDate:      "2025-12-28T01:12:59.909Z",
					LatestOrderDisplayID: "#T9362642",
					LatestOrderID:        "d70eb299-d679-4857-9c94-8a7509a7cea3",
					LatestOrderStatus:    "shipped",
					FormStatus:           "draft",
				},
				ChatURL:                "https://wa.me/918828006272?text=Hi! I want to speak to a hair expert now!",
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
		// Male cohort channels
		{"Month 1 Warriors", "month-1-warriors"},
		{"Month 2 Titans", "month-2-titans"},
		{"Month 3 Pioneers", "month-3-pioneers"},
		{"Month 4 Mavericks", "month-4-mavericks"},
		{"Month 5 Masters", "month-5-masters"},
		{"Month 6-8 Veterans", "month-6-8-veterans"},
		{"Month 8+ Legends", "month-8-plus-legends"},

		// Female cohort channels
		{"Month 1 Heroines", "month-1-heroines"},
		{"Month 2 Divas", "month-2-divas"},
		{"Month 3 Icons", "month-3-icons"},
		{"Month 4 Anchors", "month-4-anchors"},
		{"Month 5 Elites", "month-5-elites"},
		{"Month 6-8 Champions", "month-6-8-champions"},
		{"Month 8+ Queens", "month-8-plus-queens"},

		// Male topic channels
		{"General", "general"},
		{"Stress Sleep Nutrition", "stress-sleep-nutrition"},
		{"Digestion Metabolism Gut Health", "digestion-metabolism-gut-health"},
		{"Dandruff Hair Health", "dandruff-hair-health"},

		// Female topic channels
		{"Hormones PCOS", "hormones-pcos"},
		{"Stress Sleep Nutrition Female", "stress-sleep-nutrition-female"},
		{"Digestion Metabolism Gut Female", "digestion-metabolism-gut-female"},
		{"Dandruff Hair Health Female", "dandruff-hair-health-female"},
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
