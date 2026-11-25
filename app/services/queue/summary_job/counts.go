package summary_job

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/resources/visibility"
	"github.com/Southclaws/storyden/internal/ent"
	ent_node "github.com/Southclaws/storyden/internal/ent/node"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/ent/predicate"
	ent_raq "github.com/Southclaws/storyden/internal/ent/replyadminqueue"
)

// excludeBAHPosts and excludeFeedbackPosts mirror thread_querier's
// ExcludeBAHPosts/ExcludeFeedbackPosts, matching the default filters the
// queue screen applies to its "Pending Review" and "Pending Reply" tabs.
func excludeBAHPosts() predicate.Post {
	return predicate.Post(func(s *sql.Selector) {
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("COALESCE(" + s.C(ent_post.FieldMetadata) + "->>'post_category', '') != 'BAH'")
		}))
	})
}

func excludeFeedbackPosts() predicate.Post {
	return predicate.Post(func(s *sql.Selector) {
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("COALESCE(" + s.C(ent_post.FieldMetadata) + "->>'post_category', '') != 'feedback'")
		}))
	})
}

// pendingReplyWindow matches the default lookback window used by the queue
// screen's "Pending Reply" and "Pending Reply to Reply" tabs.
const pendingReplyWindow = 2 * 24 * time.Hour

type Counts struct {
	PendingReview       int
	PendingReply        int
	PendingReplyToReply int
}

// ComputeCounts computes the pending counts for each admin queue as of `now`.
func ComputeCounts(ctx context.Context, db *ent.Client, now time.Time) (Counts, error) {
	reviewThreads, err := db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.VisibilityEQ(ent_post.Visibility(visibility.VisibilityReview.String())),
			excludeBAHPosts(),
			excludeFeedbackPosts(),
		).
		Count(ctx)
	if err != nil {
		return Counts{}, fault.Wrap(err, fctx.With(ctx))
	}

	reviewNodes, err := db.Node.Query().
		Where(
			ent_node.DeletedAtIsNil(),
			ent_node.VisibilityEQ(ent_node.Visibility(visibility.VisibilityReview.String())),
		).
		Count(ctx)
	if err != nil {
		return Counts{}, fault.Wrap(err, fctx.With(ctx))
	}

	windowStart := now.Add(-pendingReplyWindow)

	pendingReply, err := db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.VisibilityEQ(ent_post.Visibility(visibility.VisibilityPublished.String())),
			ent_post.CreatedAtGTE(windowStart),
			ent_post.CreatedAtLTE(now),
			ent_post.Not(ent_post.HasPosts()),
			excludeBAHPosts(),
			excludeFeedbackPosts(),
		).
		Count(ctx)
	if err != nil {
		return Counts{}, fault.Wrap(err, fctx.With(ctx))
	}

	pendingReplyToReply, err := db.ReplyAdminQueue.Query().
		Where(
			ent_raq.CreatedAtGTE(windowStart),
			ent_raq.CreatedAtLTE(now),
		).
		Count(ctx)
	if err != nil {
		return Counts{}, fault.Wrap(err, fctx.With(ctx))
	}

	return Counts{
		PendingReview:       reviewThreads + reviewNodes,
		PendingReply:        pendingReply,
		PendingReplyToReply: pendingReplyToReply,
	}, nil
}
