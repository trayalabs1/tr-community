package thread_querier

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

// quoteSQLLiteral wraps a value as a single-quoted SQL string literal with
// embedded quotes doubled. Used for metadata predicate values that must be
// inlined into raw SQL (ent's b.Arg placeholder substitution does not play
// well with the way we hand-build JSON path predicates here).
func quoteSQLLiteral(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// HasRecentChannelPrescored reports whether a published, non-deleted root post
// with matching metadata (post_category, type) exists in the channel since the
// given time, excluding excludeThreadID. Used to gate the per-category cooldown
// (BAH, feedback, etc.) at thread creation.
func (d *Querier) HasRecentChannelPrescored(
	ctx context.Context,
	channelID xid.ID,
	since time.Time,
	excludeThreadID xid.ID,
	category string,
	postType string,
) (bool, error) {
	count, err := d.db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.ChannelID(channelID),
			ent_post.CreatedAtGTE(since),
			ent_post.IDNEQ(excludeThreadID),
			ent_post.VisibilityEQ(ent_post.VisibilityPublished),
			predicate.Post(func(s *sql.Selector) {
				metaCol := s.C(ent_post.FieldMetadata)
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString(metaCol)
					b.WriteString("->>'post_category' = ")
					b.WriteString(quoteSQLLiteral(category))
					b.WriteString(" AND ")
					b.WriteString(metaCol)
					b.WriteString("->>'type' = ")
					b.WriteString(quoteSQLLiteral(postType))
				}))
			}),
		).
		Count(ctx)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	return count > 0, nil
}

// HasRecentChannelBAH is a back-compat wrapper that gates the BAH cooldown
// regardless of meta.type. Prefer HasRecentChannelPrescored for new code.
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
			ent_post.VisibilityEQ(ent_post.VisibilityPublished),
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
