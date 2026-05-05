package thread_querier

import (
	"context"
	"time"

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
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

type SelfRecentThread struct {
	Thread       *thread.Thread
	SentimentTag string
	PrimaryTopic string
}

func (d *Querier) ListSelfRecent(
	ctx context.Context,
	channelID xid.ID,
	accountID account.AccountID,
	since time.Time,
) ([]SelfRecentThread, error) {
	results, err := d.db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.ChannelID(channelID),
			ent_post.AccountPosts(xid.ID(accountID)),
			ent_post.CreatedAtGTE(since),
		).
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
		Order(ent_post.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	hydrated, err := d.hydrateThreads(ctx, results, opt.New(accountID))
	if err != nil {
		return nil, err
	}

	out := make([]SelfRecentThread, 0, len(hydrated))
	for i, p := range results {
		entry := SelfRecentThread{Thread: hydrated[i]}
		if p.Edges.Sentiment != nil {
			if p.Edges.Sentiment.SentimentTag != nil {
				entry.SentimentTag = *p.Edges.Sentiment.SentimentTag
			}
			if p.Edges.Sentiment.PrimaryTopic != nil {
				entry.PrimaryTopic = *p.Edges.Sentiment.PrimaryTopic
			}
		}
		out = append(out, entry)
	}
	return out, nil
}

func (d *Querier) ListSimilarFor(
	ctx context.Context,
	channelID xid.ID,
	excludeThreadID post.ID,
	excludeAccountID account.AccountID,
	sentimentTag string,
	primaryTopic string,
	limit int,
	requesterAccountID opt.Optional[account.AccountID],
) ([]*thread.Thread, error) {
	if sentimentTag == "" || primaryTopic == "" || sentimentTag == "negative" {
		return []*thread.Thread{}, nil
	}
	if limit <= 0 {
		limit = 2
	}

	results, err := d.db.Post.Query().
		Where(
			ent_post.RootPostIDIsNil(),
			ent_post.DeletedAtIsNil(),
			ent_post.ChannelID(channelID),
			ent_post.IDNEQ(xid.ID(excludeThreadID)),
			ent_post.AccountPostsNEQ(xid.ID(excludeAccountID)),
			ent_post.VisibilityEQ(ent_post.VisibilityPublished),
			ent_post.HasSentimentWith(
				ent_post_sentiment.SentimentTagEQ(sentimentTag),
				ent_post_sentiment.PrimaryTopicEQ(primaryTopic),
			),
		).
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
		Order(ent_post.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	return d.hydrateThreads(ctx, results, requesterAccountID)
}

func (d *Querier) hydrateThreads(
	ctx context.Context,
	posts []*ent.Post,
	accountID opt.Optional[account.AccountID],
) ([]*thread.Thread, error) {
	if len(posts) == 0 {
		return []*thread.Thread{}, nil
	}

	ids := dt.Map(posts, func(p *ent.Post) xid.ID { return p.ID })

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
	threads, err := dt.MapErr(posts, mapper)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return threads, nil
}
