package thread_querier

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

func (d *Querier) HasRecentChannelBAH(
	ctx context.Context,
	channelID xid.ID,
	since time.Time,
	excludeThreadID xid.ID,
) (bool, error) {
	count, err := d.db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.ChannelID(channelID),
			ent_post.CreatedAtGTE(since),
			ent_post.IDNEQ(excludeThreadID),
			ent_post.VisibilityIn(
				ent_post.VisibilityPublished,
				ent_post.VisibilityReview,
			),
			predicate.Post(func(s *sql.Selector) {
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString(s.C(ent_post.FieldMetadata) + "->>'post_category' = 'BAH'")
				}))
			}),
		).
		Count(ctx)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	return count > 0, nil
}
