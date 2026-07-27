package thread_querier

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/alitto/pond/v2"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/collection/collection_item_status"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/internal/ent"
	ent_asset "github.com/Southclaws/storyden/internal/ent/asset"
	"github.com/Southclaws/storyden/internal/ent/collection"
	"github.com/Southclaws/storyden/internal/ent/link"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

func (d *Querier) ListShares(
	ctx context.Context,
	accountID opt.Optional[account.AccountID],
	opts ...Query,
) ([]*thread.Thread, error) {
	query := d.db.Post.Query().Where(ent_post.RootPostIDIsNil())
	queryOptions := threadListOptions{
		q: query,
	}

	for _, fn := range opts {
		fn(&queryOptions)
	}

	query.
		WithCategory().
		WithAuthor().
		WithAssets(func(aq *ent.AssetQuery) {
			aq.Order(ent_asset.ByUpdatedAt(), ent_asset.ByCreatedAt())
		}).
		WithCollections(func(cq *ent.CollectionQuery) {
			cq.WithOwner().Order(collection.ByUpdatedAt(), collection.ByCreatedAt())
		}).
		WithLink(func(lq *ent.LinkQuery) {
			lq.WithFaviconImage().WithPrimaryImage()
			lq.WithAssets().Order(link.ByCreatedAt(sql.OrderDesc()))
		}).
		WithSentiment().
		Order(ent_post.ByCreatedAt(sql.OrderDesc()))

	result, err := query.All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	if len(result) == 0 {
		return []*thread.Thread{}, nil
	}

	ids := dt.Map(result, func(p *ent.Post) xid.ID { return p.ID })

	pool := pond.NewGroup()

	var readStates post.ReadStateMap
	pool.SubmitErr(func() error {
		r, err := d.getReadStatus(ctx, ids, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		readStates = r
		return nil
	})

	var repliesMap post.PostRepliesMap
	pool.SubmitErr(func() error {
		r, err := d.getRepliesStatus(ctx, ids, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		repliesMap = r
		return nil
	})

	var likesMap post.PostLikesMap
	pool.SubmitErr(func() error {
		r, err := d.getLikesStatus(ctx, ids, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		likesMap = r
		return nil
	})

	var collectionsMap collection_item_status.CollectionStatusMap
	pool.SubmitErr(func() error {
		r, err := d.getCollectionsStatus(ctx, ids, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		collectionsMap = r
		return nil
	})

	if err := pool.Wait(); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	mapper := thread.Mapper(nil, readStates, likesMap, collectionsMap, repliesMap, nil)
	threads, err := dt.MapErr(result, mapper)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return threads, nil
}
