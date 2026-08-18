package summary_job_test

import (
	"context"
	"testing"
	"time"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/services/queue/summary_job"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/internal/utils"
	"github.com/Southclaws/storyden/tests"
)

func TestComputeCounts(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		db *ent.Client,
		aw *account_writer.Writer,
		raqw *reply_admin_queue_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			handle := xid.New().String()
			memberResp, err := cl.AuthPasswordSignupWithResponse(root, nil, openapi.AuthPair{
				Identifier: handle,
				Token:      "password",
			})
			tests.Ok(t, err, memberResp)
			memberID := utils.Must(xid.FromString(memberResp.JSON200.Id))
			memberCtx := sh.WithSession(e2e.WithAccountID(root, account.AccountID(memberID)))

			channelSuffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "queue-summary-" + channelSuffix,
				Slug:        "queue-summary-" + channelSuffix,
				Description: "channel for queue summary job tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			addMemberResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession)
			tests.Ok(t, err, addMemberResp)

			before, err := summary_job.ComputeCounts(root, db, time.Now())
			r.NoError(err)

			// A thread pending review.
			reviewResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Pending review thread",
				Body:       opt.New("<p>body</p>").Ptr(),
				Visibility: opt.New(openapi.Review).Ptr(),
			}, memberCtx)
			tests.Ok(t, err, reviewResp)

			// A published thread with no replies yet, pending reply.
			pendingReplyResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Pending reply thread",
				Body:       opt.New("<p>body</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, memberCtx)
			tests.Ok(t, err, pendingReplyResp)

			// An admin thread that a member replies to.
			adminThreadResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Admin thread",
				Body:       opt.New("<p>body</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, adminThreadResp)
			adminThreadID := utils.Must(xid.FromString(adminThreadResp.JSON200.Id))

			replyResp, err := cl.ReplyCreateWithResponse(root, adminThreadResp.JSON200.Slug, openapi.ReplyInitialProps{
				Body: "A question for the admin",
			}, memberCtx)
			tests.Ok(t, err, replyResp)
			replyID := utils.Must(xid.FromString(replyResp.JSON200.Id))

			// The reply-admin-queue entry itself is normally populated
			// asynchronously by an event consumer, whose timing isn't
			// relevant to what we're testing here (the counting query, not
			// the enqueue pipeline), so it's written directly.
			_, err = raqw.Enqueue(root, post.ID(replyID), post.ID(adminThreadID), utils.Must(xid.FromString(channelID)), "a question for the admin")
			r.NoError(err)

			after, err := summary_job.ComputeCounts(root, db, time.Now())
			r.NoError(err)

			r.Equal(before.PendingReview+1, after.PendingReview)
			r.Equal(before.PendingReply+1, after.PendingReply)
			r.Equal(before.PendingReplyToReply+1, after.PendingReplyToReply)

			// A BAH (streak) review thread and a feedback review thread,
			// matching the queue screen's default excludeBAH/excludeFeedback
			// filters, must not inflate the pending review count.
			bahResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "BAH thread",
				Body:       opt.New("<p>body</p>").Ptr(),
				Visibility: opt.New(openapi.Review).Ptr(),
			}, memberCtx)
			tests.Ok(t, err, bahResp)
			bahThreadID := utils.Must(xid.FromString(bahResp.JSON200.Id))
			_, err = db.Post.UpdateOneID(bahThreadID).SetMetadata(map[string]any{"post_category": "BAH"}).Save(root)
			r.NoError(err)

			feedbackResp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Feedback thread",
				Body:       opt.New("<p>body</p>").Ptr(),
				Visibility: opt.New(openapi.Review).Ptr(),
			}, memberCtx)
			tests.Ok(t, err, feedbackResp)
			feedbackThreadID := utils.Must(xid.FromString(feedbackResp.JSON200.Id))
			_, err = db.Post.UpdateOneID(feedbackThreadID).SetMetadata(map[string]any{"post_category": "feedback"}).Save(root)
			r.NoError(err)

			afterExcluded, err := summary_job.ComputeCounts(root, db, time.Now())
			r.NoError(err)
			r.Equal(after.PendingReview, afterExcluded.PendingReview, "BAH and feedback posts must be excluded from the pending review count")
		}))
	}))
}
