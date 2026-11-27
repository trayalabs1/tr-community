package post_querier

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

func (q *Querier) IsCrosspost(ctx context.Context, id post.ID) (bool, error) {
	p, err := q.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(id))).
		Select(ent_post.FieldMetadata).
		Only(ctx)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx))
	}

	_, ok := p.Metadata[post.MetaCrosspostSourceID]
	return ok, nil
}
