package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestChannelScoreUnscored(t *testing.T) {
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
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)

			adminSession := sh.WithSession(adminCtx)
			memberSession := sh.WithSession(memberCtx)

			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Score Test Channel",
				Slug:        "score-test-channel",
				Description: "Channel for score-unscored testing",
			}, adminSession)
			tests.Ok(t, err, createResp)
			channelID := createResp.JSON200.Id

			catResp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        "General",
				Description: "General topics",
				Colour:      "#ff0000",
			}, adminSession)
			tests.Ok(t, err, catResp)
			categoryID := catResp.JSON200.Id

			addMemberResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addMemberResp)

			_, err = cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Thread 1",
				Body:       opt.New("<p>First thread body</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			r.NoError(err)

			_, err = cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Thread 2",
				Body:       opt.New("<p>Second thread body</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			r.NoError(err)

			t.Run("admin_can_score_unscored", func(t *testing.T) {
				resp, err := cl.ChannelScoreUnscoredWithResponse(root, channelID,
					&openapi.ChannelScoreUnscoredParams{},
					adminSession,
				)
				tests.Ok(t, err, resp)
				r.True(resp.JSON200.Success)
				r.GreaterOrEqual(resp.JSON200.PostsEnqueued, 0)
				r.GreaterOrEqual(resp.JSON200.DurationMs, int64(0))
			})

			t.Run("admin_with_include_failed", func(t *testing.T) {
				includeFailed := true
				resp, err := cl.ChannelScoreUnscoredWithResponse(root, channelID,
					&openapi.ChannelScoreUnscoredParams{
						IncludeFailed: &includeFailed,
					},
					adminSession,
				)
				tests.Ok(t, err, resp)
				r.True(resp.JSON200.Success)
			})

			t.Run("admin_with_date_range", func(t *testing.T) {
				after := time.Now().Add(-24 * time.Hour)
				before := time.Now().Add(24 * time.Hour)
				resp, err := cl.ChannelScoreUnscoredWithResponse(root, channelID,
					&openapi.ChannelScoreUnscoredParams{
						CreatedAfter:  &after,
						CreatedBefore: &before,
					},
					adminSession,
				)
				tests.Ok(t, err, resp)
				r.True(resp.JSON200.Success)
			})

			t.Run("non_admin_forbidden", func(t *testing.T) {
				resp, err := cl.ChannelScoreUnscoredWithResponse(root, channelID,
					&openapi.ChannelScoreUnscoredParams{},
					memberSession,
				)
				r.NoError(err)
				r.Equal(403, resp.StatusCode())
			})
		}))
	}))
}
