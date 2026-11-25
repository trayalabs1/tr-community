package thread_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestThreadListCategoryAndSentimentFilters(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "cat-sentiment-" + suffix,
				Slug:        "cat-sentiment-" + suffix,
				Description: "channel for category/sentiment filter tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			published := openapi.Published
			bahMeta := openapi.Metadata{"post_category": "BAH"}
			feedbackMeta := openapi.Metadata{"post_category": "feedback"}

			// BAH and feedback are prescored at creation with the "neutral" tag.
			bahResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Streak post",
				Body:       opt.New[openapi.PostContent]("streak body").Ptr(),
				Meta:       &bahMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, bahResp)
			bahID := bahResp.JSON200.Id

			feedbackResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Feedback post",
				Body:       opt.New[openapi.PostContent]("feedback body").Ptr(),
				Meta:       &feedbackMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, feedbackResp)
			feedbackID := feedbackResp.JSON200.Id

			plainResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Plain post",
				Body:       opt.New[openapi.PostContent]("plain body").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, plainResp)
			plainID := plainResp.JSON200.Id

			noReplies := true

			t.Run("post_categories_single_value", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					PostCategories: &[]string{"BAH"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, bahID)
				assert.NotContains(t, ids, feedbackID)
				assert.NotContains(t, ids, plainID)
			})

			t.Run("post_categories_multiple_values", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					PostCategories: &[]string{"BAH", "feedback"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, bahID)
				assert.Contains(t, ids, feedbackID)
				assert.NotContains(t, ids, plainID)
			})

			t.Run("sentiments_neutral_matches_prescored", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:  &noReplies,
					Sentiments: &[]string{"neutral"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				require.Contains(t, ids, bahID)
				assert.Contains(t, ids, feedbackID)
			})

			t.Run("sentiments_positive_excludes_neutral_prescored", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:  &noReplies,
					Sentiments: &[]string{"positive"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.NotContains(t, ids, bahID)
				assert.NotContains(t, ids, feedbackID)
			})

			t.Run("category_and_sentiment_combined", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					NoReplies:      &noReplies,
					PostCategories: &[]string{"BAH"},
					Sentiments:     &[]string{"neutral"},
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, bahID)
				assert.NotContains(t, ids, feedbackID)
			})
		}))
	}))
}
