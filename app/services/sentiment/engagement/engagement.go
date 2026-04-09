package engagement

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/ent/likepost"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

type Counts struct {
	Likes   int
	Replies int
}

func GetBulk(ctx context.Context, db *ent.Client, postIDs []xid.ID) (map[xid.ID]Counts, error) {
	if len(postIDs) == 0 {
		return map[xid.ID]Counts{}, nil
	}

	result := make(map[xid.ID]Counts, len(postIDs))

	var likeCounts []struct {
		PostID xid.ID `json:"post_id"`
		Count  int    `json:"count"`
	}
	err := db.LikePost.
		Query().
		Where(likepost.PostIDIn(postIDs...)).
		GroupBy(likepost.FieldPostID).
		Aggregate(ent.Count()).
		Scan(ctx, &likeCounts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	for _, lc := range likeCounts {
		c := result[lc.PostID]
		c.Likes = lc.Count
		result[lc.PostID] = c
	}

	var replyCounts []struct {
		RootPostID xid.ID `json:"root_post_id"`
		Count      int    `json:"count"`
	}
	err = db.Post.
		Query().
		Where(
			ent_post.RootPostIDIn(postIDs...),
			ent_post.DeletedAtIsNil(),
		).
		GroupBy(ent_post.FieldRootPostID).
		Aggregate(ent.Count()).
		Scan(ctx, &replyCounts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	for _, rc := range replyCounts {
		c := result[rc.RootPostID]
		c.Replies = rc.Count
		result[rc.RootPostID] = c
	}

	return result, nil
}
