package ranker_job

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/ent/likepost"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

const (
	likeWeight  = 2.0
	replyWeight = 0.5
)

// GetDailyIncrements returns, per post, the rank_score delta owed for likes
// and replies created at or after `since` — the same weights the live v4
// ranking formula would have applied (likes ×2, replies ×0.5), computed once
// per day instead of live per feed read. Only posts with at least one new
// like or reply in the window appear in the result.
func GetDailyIncrements(ctx context.Context, db *ent.Client, since time.Time) (map[xid.ID]float64, error) {
	deltas := make(map[xid.ID]float64)

	var likeCounts []struct {
		PostID xid.ID `json:"post_id"`
		Count  int    `json:"count"`
	}
	err := db.LikePost.
		Query().
		Where(likepost.CreatedAtGTE(since)).
		GroupBy(likepost.FieldPostID).
		Aggregate(ent.Count()).
		Scan(ctx, &likeCounts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	for _, lc := range likeCounts {
		deltas[lc.PostID] += float64(lc.Count) * likeWeight
	}

	var replyCounts []struct {
		RootPostID xid.ID `json:"root_post_id"`
		Count      int    `json:"count"`
	}
	err = db.Post.
		Query().
		Where(
			ent_post.RootPostIDNotNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.CreatedAtGTE(since),
		).
		GroupBy(ent_post.FieldRootPostID).
		Aggregate(ent.Count()).
		Scan(ctx, &replyCounts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	for _, rc := range replyCounts {
		deltas[rc.RootPostID] += float64(rc.Count) * replyWeight
	}

	return deltas, nil
}
