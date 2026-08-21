package ranker_job

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/settings"
	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/ent/likepost"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

const (
	DefaultLikeWeight  = 2.0
	DefaultReplyWeight = 0.5
)

// EngagementWeights holds the admin-configurable per-like/per-reply weights
// used by GetDailyIncrements.
type EngagementWeights struct {
	Like  float64
	Reply float64
}

func DefaultEngagementWeights() EngagementWeights {
	return EngagementWeights{
		Like:  DefaultLikeWeight,
		Reply: DefaultReplyWeight,
	}
}

// LoadEngagementWeights reads the current admin-configured like/reply
// weights, falling back to DefaultEngagementWeights() for any value that
// hasn't been set.
func LoadEngagementWeights(ctx context.Context, repo *settings.SettingsRepository) (EngagementWeights, error) {
	s, err := repo.Get(ctx)
	if err != nil {
		return EngagementWeights{}, fault.Wrap(err, fctx.With(ctx))
	}

	fr := s.Services.OrZero().FeedRanking.OrZero()

	return EngagementWeights{
		Like:  fr.LikeWeight.Or(DefaultLikeWeight),
		Reply: fr.ReplyWeight.Or(DefaultReplyWeight),
	}, nil
}

// GetDailyIncrements returns, per post, the rank_score delta owed for likes
// and replies created at or after `since` — the same weights the live v4
// ranking formula would have applied, computed once per day instead of live
// per feed read. Only posts with at least one new like or reply in the
// window appear in the result.
func GetDailyIncrements(ctx context.Context, db *ent.Client, since time.Time, w EngagementWeights) (map[xid.ID]float64, error) {
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
		deltas[lc.PostID] += float64(lc.Count) * w.Like
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
		deltas[rc.RootPostID] += float64(rc.Count) * w.Reply
	}

	return deltas, nil
}
