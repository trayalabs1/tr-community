package admin_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestAdminFeedRankingSettings(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			getBefore := tests.AssertRequest(
				cl.AdminSettingsGetWithResponse(root, adminSession),
			)(t, http.StatusOK)
			r.NotNil(getBefore.JSON200.Services)
			r.NotNil(getBefore.JSON200.Services.FeedRanking, "defaults should already populate feed_ranking on a fresh instance")
			r.InDelta(1.0, *getBefore.JSON200.Services.FeedRanking.WPositivity, 0.001)
			r.InDelta(2.5, *getBefore.JSON200.Services.FeedRanking.FormatMultiplier, 0.001)

			wPositivity := 3.5
			likeWeight := 7.0
			updateResp := tests.AssertRequest(
				cl.AdminSettingsUpdateWithResponse(root, openapi.AdminSettingsUpdateJSONRequestBody{
					Services: &openapi.AdminSettingsServiceProps{
						FeedRanking: &openapi.FeedRankingServiceSettings{
							WPositivity: &wPositivity,
							LikeWeight:  &likeWeight,
						},
					},
				}, adminSession),
			)(t, http.StatusOK)

			r.NotNil(updateResp.JSON200.Services.FeedRanking)
			r.InDelta(wPositivity, *updateResp.JSON200.Services.FeedRanking.WPositivity, 0.001)
			r.InDelta(likeWeight, *updateResp.JSON200.Services.FeedRanking.LikeWeight, 0.001)
			// Fields not included in this update must retain their previous
			// values (partial-update semantics via mergo), not reset to zero.
			r.InDelta(2.5, *updateResp.JSON200.Services.FeedRanking.FormatMultiplier, 0.001,
				"fields omitted from the update must be preserved, not zeroed")

			getAfter := tests.AssertRequest(
				cl.AdminSettingsGetWithResponse(root, adminSession),
			)(t, http.StatusOK)
			r.InDelta(wPositivity, *getAfter.JSON200.Services.FeedRanking.WPositivity, 0.001,
				"a subsequent Get must reflect the just-written value (cache invalidated on write)")
			r.InDelta(likeWeight, *getAfter.JSON200.Services.FeedRanking.LikeWeight, 0.001)
		}))
	}))
}
