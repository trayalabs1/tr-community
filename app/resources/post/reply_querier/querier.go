package reply_querier

import (
	"context"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/internal/ent"
	ent_account "github.com/Southclaws/storyden/internal/ent/account"
	"github.com/Southclaws/storyden/internal/ent/asset"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

type Querier struct {
	db *ent.Client
}

func New(db *ent.Client) *Querier {
	return &Querier{db: db}
}

func (d *Querier) Get(ctx context.Context, id post.ID) (*reply.Reply, error) {
	p, err := d.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(id))).
		WithAuthor().
		WithRoot(func(pq *ent.PostQuery) {
			pq.WithAuthor()
		}).
		WithAssets(func(aq *ent.AssetQuery) {
			aq.Order(asset.ByUpdatedAt(), asset.ByCreatedAt())
		}).
		WithReplyTo(func(pq *ent.PostQuery) {
			pq.WithAuthor()
			pq.Where(ent_post.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	return reply.Map(p)
}

func (d *Querier) GetMany(ctx context.Context, ids ...post.ID) ([]*reply.Reply, error) {
	if len(ids) == 0 {
		return []*reply.Reply{}, nil
	}

	xids := dt.Map(ids, func(id post.ID) xid.ID { return xid.ID(id) })

	posts, err := d.db.Post.
		Query().
		Where(ent_post.IDIn(xids...)).
		WithAuthor().
		WithRoot(func(pq *ent.PostQuery) {
			pq.WithAuthor()
		}).
		WithAssets(func(aq *ent.AssetQuery) {
			aq.Order(asset.ByUpdatedAt(), asset.ByCreatedAt())
		}).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	replies := make([]*reply.Reply, 0, len(posts))
	for _, p := range posts {
		r, err := reply.Map(p)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
		replies = append(replies, r)
	}

	return replies, nil
}

type ListFilters struct {
	// AuthorIDs restricts results to replies written by these accounts. Empty
	// means no author restriction.
	AuthorIDs []xid.ID

	// CreatedAfter and CreatedBefore bound the reply's own creation time, not
	// the creation time of the thread it belongs to.
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// List returns replies matching the filters, newest first. Only posts with a
// root are considered, so threads themselves are never returned.
func (d *Querier) List(
	ctx context.Context,
	page pagination.Parameters,
	filters ListFilters,
) (pagination.Result[*reply.Reply], error) {
	query := d.db.Post.
		Query().
		Where(
			ent_post.RootPostIDNotNil(),
			ent_post.DeletedAtIsNil(),
		)

	if len(filters.AuthorIDs) > 0 {
		query = query.Where(ent_post.HasAuthorWith(ent_account.IDIn(filters.AuthorIDs...)))
	}
	// created_at is persisted with the writer's local offset, and SQLite compares
	// the resulting strings by wall-clock digits rather than by instant. Bounds
	// arrive from the API as UTC, so shift them into the local zone to keep the
	// comparison against the same wall clock the rows were written in.
	if filters.CreatedAfter != nil {
		query = query.Where(ent_post.CreatedAtGTE(filters.CreatedAfter.Local()))
	}
	if filters.CreatedBefore != nil {
		query = query.Where(ent_post.CreatedAtLTE(filters.CreatedBefore.Local()))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return pagination.Result[*reply.Reply]{}, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	posts, err := query.
		Order(ent.Desc(ent_post.FieldCreatedAt)).
		Limit(page.Limit()).
		Offset(page.Offset()).
		WithAuthor().
		WithRoot(func(pq *ent.PostQuery) {
			pq.WithAuthor()
		}).
		WithAssets(func(aq *ent.AssetQuery) {
			aq.Order(asset.ByUpdatedAt(), asset.ByCreatedAt())
		}).
		All(ctx)
	if err != nil {
		return pagination.Result[*reply.Reply]{}, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	replies, err := dt.MapErr(posts, reply.Map)
	if err != nil {
		return pagination.Result[*reply.Reply]{}, fault.Wrap(err, fctx.With(ctx))
	}

	return pagination.NewPageResult(page, total, replies), nil
}

func (d *Querier) Probe(ctx context.Context, id post.ID) (*reply.ReplyRef, error) {
	p, err := d.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(id))).
		Only(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	return reply.MapRef(p), nil
}
