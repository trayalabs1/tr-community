package thread_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestThreadListFeedbackFilter(t *testing.T) {
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
				Name:        "feedback-filter-" + suffix,
				Slug:        "feedback-filter-" + suffix,
				Description: "channel for feedback filter tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			feedbackMeta := openapi.Metadata{"post_category": "feedback"}
			bahMeta := openapi.Metadata{"post_category": "BAH"}
			published := openapi.Published
			review := openapi.Review

			// Published feedback thread, no replies.
			feedbackPublishedResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Feedback published",
				Body:       opt.New[openapi.PostContent]("feedback body").Ptr(),
				Meta:       &feedbackMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, feedbackPublishedResp)
			feedbackPublishedID := feedbackPublishedResp.JSON200.Id

			// Review feedback thread, no replies — must be visible to admin requesting review.
			feedbackReviewResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Feedback review",
				Body:       opt.New[openapi.PostContent]("review feedback body").Ptr(),
				Meta:       &feedbackMeta,
				Visibility: &review,
			}, adminSession)
			tests.Ok(t, err, feedbackReviewResp)
			feedbackReviewID := feedbackReviewResp.JSON200.Id

			// Published BAH thread — should survive an exclude_feedback filter.
			bahResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "BAH published",
				Body:       opt.New[openapi.PostContent]("streak body").Ptr(),
				Meta:       &bahMeta,
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, bahResp)
			bahID := bahResp.JSON200.Id

			// Published non-categorised thread.
			plainResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Plain published",
				Body:       opt.New[openapi.PostContent]("plain body").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, plainResp)
			plainID := plainResp.JSON200.Id

			visibilities := openapi.VisibilityParam{openapi.Published, openapi.Review}
			noReplies := true

			t.Run("exclude_feedback_excludes_all_feedback", func(t *testing.T) {
				excludeFeedback := true
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility:      &visibilities,
					NoReplies:       &noReplies,
					ExcludeFeedback: &excludeFeedback,
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.NotContains(t, ids, feedbackPublishedID)
				assert.NotContains(t, ids, feedbackReviewID)
				assert.Contains(t, ids, bahID)
				assert.Contains(t, ids, plainID)
			})

			t.Run("without_exclude_feedback_includes_feedback", func(t *testing.T) {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					NoReplies:  &noReplies,
				}, adminSession)
				tests.Ok(t, err, resp)

				ids := threadIDs(resp.JSON200.Threads)
				assert.Contains(t, ids, feedbackPublishedID)
				assert.Contains(t, ids, feedbackReviewID)
			})
		}))
	}))
}
