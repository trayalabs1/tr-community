package thread_querier

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/alitto/pond/v2"
	"github.com/rs/xid"
	"github.com/samber/lo"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/collection/collection_item_status"
	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reaction"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/resources/profile"
	"github.com/Southclaws/storyden/app/resources/tag/tag_ref"
	"github.com/Southclaws/storyden/internal/ent"
	ent_account "github.com/Southclaws/storyden/internal/ent/account"
	ent_asset "github.com/Southclaws/storyden/internal/ent/asset"
	ent_channel "github.com/Southclaws/storyden/internal/ent/channel"
	"github.com/Southclaws/storyden/internal/ent/collection"
	"github.com/Southclaws/storyden/internal/ent/link"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_react "github.com/Southclaws/storyden/internal/ent/react"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	ent_tag "github.com/Southclaws/storyden/internal/ent/tag"
	ent_threadshare "github.com/Southclaws/storyden/internal/ent/threadshare"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/kv"
)

func (d *Querier) List(
	ctx context.Context,
	page int,
	size int,
	accountID opt.Optional[account.AccountID],
	opts ...Query,
) (*Result, error) {
	if size < 1 {
		size = 1
	}

	if size > 100 {
		size = 100
	}

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
		WithSentiment()

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if queryOptions.useSentimentRanking {
		rankExpr := `CASE
			WHEN %[1]s = 'positive' AND %[2]s > NOW() - INTERVAL '24 hours' THEN 1
			WHEN %[1]s = 'neutral' AND %[2]s > NOW() - INTERVAL '24 hours' THEN 2
			WHEN %[1]s = 'positive' AND %[2]s > NOW() - INTERVAL '48 hours' THEN 3
			WHEN %[1]s = 'neutral' AND %[2]s > NOW() - INTERVAL '48 hours' THEN 4
			WHEN %[1]s = 'positive' AND %[2]s > NOW() - INTERVAL '72 hours' THEN 5
			WHEN %[1]s = 'neutral' AND %[2]s > NOW() - INTERVAL '72 hours' THEN 6
			WHEN %[1]s = 'positive' AND %[2]s > NOW() - INTERVAL '96 hours' THEN 7
			WHEN %[1]s = 'neutral' AND %[2]s > NOW() - INTERVAL '96 hours' THEN 8
			WHEN %[1]s = 'positive' THEN 9
			WHEN %[1]s = 'neutral' THEN 10
			WHEN %[1]s = 'negative' THEN 11
			ELSE 12
		END ASC`

		if queryOptions.ignorePinned {
			query.Modify(func(s *sql.Selector) {
				t := sql.Table(ent_post_sentiment.Table)
				s.LeftJoin(t).On(s.C(ent_post.FieldID), t.C(ent_post_sentiment.FieldPostID))
				sentimentCol := t.C(ent_post_sentiment.FieldSentimentTag)
				createdAtCol := s.C(ent_post.FieldCreatedAt)
				s.OrderExpr(
					sql.Expr(fmt.Sprintf(rankExpr, sentimentCol, createdAtCol)),
					sql.Expr("COALESCE("+t.C(ent_post_sentiment.FieldRankScore)+", -1) DESC"),
					sql.Expr(s.C(ent_post.FieldCreatedAt)+" DESC"),
				)
			})
		} else {
			query.Modify(func(s *sql.Selector) {
				t := sql.Table(ent_post_sentiment.Table)
				s.LeftJoin(t).On(s.C(ent_post.FieldID), t.C(ent_post_sentiment.FieldPostID))
				sentimentCol := t.C(ent_post_sentiment.FieldSentimentTag)
				createdAtCol := s.C(ent_post.FieldCreatedAt)
				s.OrderExpr(
					sql.Expr(effectivePinRankExpr(s, queryOptions.channelID)+" DESC"),
					sql.Expr(fmt.Sprintf(rankExpr, sentimentCol, createdAtCol)),
					sql.Expr("COALESCE("+t.C(ent_post_sentiment.FieldRankScore)+", -1) DESC"),
					sql.Expr(s.C(ent_post.FieldCreatedAt)+" DESC"),
				)
			})
		}
	} else {
		query.Modify(func(s *sql.Selector) {
			s.OrderExpr(sql.Expr(s.C(ent_post.FieldCreatedAt) + " DESC"))
		})
	}

	query.
		Limit(size + 1).
		Offset(page * size)

	result, err := query.All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
	}

	isNextPage := len(result) > size
	nextPage := opt.NewSafe(page+1, isNextPage)
	totalPages := int(math.Ceil(float64(total) / float64(size)))

	if len(result) == 0 {
		return &Result{
			PageSize:    size,
			Results:     0,
			TotalPages:  totalPages,
			CurrentPage: page,
			NextPage:    nextPage,
			Threads:     []*thread.Thread{},
		}, nil
	}

	if isNextPage {
		result = result[:len(result)-1]
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

	err = pool.Wait()
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	mapper := thread.Mapper(nil, readStates, likesMap, collectionsMap, repliesMap, nil)
	threads, err := dt.MapErr(result, mapper)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if channelID, ok := queryOptions.channelID.Get(); ok {
		if err := d.decorateShares(ctx, channelID, threads); err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	return &Result{
		PageSize:    size,
		Results:     len(threads),
		TotalPages:  totalPages,
		CurrentPage: page,
		NextPage:    nextPage,
		Threads:     threads,
	}, nil
}

// decorateShares attaches interchannel-share context (subtitle, source channel,
// and the admin who shared) to any thread that appears in this channel's feed by
// virtue of a share rather than living in the channel itself. Threads owned by
// the channel are left untouched.
func (d *Querier) decorateShares(ctx context.Context, channelID xid.ID, threads []*thread.Thread) error {
	shared := make([]*thread.Thread, 0)
	postIDs := make([]xid.ID, 0)
	for _, t := range threads {
		if t.ChannelID != channelID {
			shared = append(shared, t)
			postIDs = append(postIDs, xid.ID(t.ID))
		}
	}

	if len(postIDs) == 0 {
		return nil
	}

	shares, err := d.db.ThreadShare.Query().
		Where(
			ent_threadshare.PostIDIn(postIDs...),
			ent_threadshare.ChannelID(channelID),
			ent_threadshare.DeletedAtIsNil(),
		).
		WithChannel().
		WithAccount().
		All(ctx)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	byPost := make(map[xid.ID]*ent.ThreadShare, len(shares))
	sourceChannelIDs := make([]xid.ID, 0, len(shared))
	for _, s := range shares {
		byPost[s.PostID] = s
	}
	for _, t := range shared {
		if _, ok := byPost[xid.ID(t.ID)]; ok && !t.ChannelID.IsZero() {
			sourceChannelIDs = append(sourceChannelIDs, t.ChannelID)
		}
	}

	sourceChannels := make(map[xid.ID]*channel.Channel)
	if len(sourceChannelIDs) > 0 {
		chans, err := d.db.Channel.Query().
			Where(ent_channel.IDIn(sourceChannelIDs...)).
			All(ctx)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		for _, c := range chans {
			sourceChannels[c.ID] = channel.FromModel(c)
		}
	}

	for _, t := range shared {
		share, ok := byPost[xid.ID(t.ID)]
		if !ok {
			continue
		}

		if share.Subtitle != "" {
			t.Subtitle = opt.New(share.Subtitle)
		}
		if src, ok := sourceChannels[t.ChannelID]; ok {
			t.SharedFrom = opt.New(*src)
		}
		if share.Edges.Account != nil {
			pro, err := profile.MapRef(share.Edges.Account)
			if err != nil {
				return fault.Wrap(err, fctx.With(ctx))
			}
			t.SharedBy = opt.New(*pro)
		}
	}

	return nil
}

// effectivePinRankExpr returns the SQL expression for a thread's pin rank in the
// context of a specific channel feed. For threads that live in the channel it is
// the post's own pinned_rank; for threads featured via an interchannel share it is
// the share row's pinned_rank, so a thread can be pinned in a destination cohort
// without affecting its ordering in the source channel. When no channel is scoped
// (e.g. site-wide feeds) it falls back to the post's own pinned_rank.
func effectivePinRankExpr(s *sql.Selector, channelID opt.Optional[xid.ID]) string {
	cid, ok := channelID.Get()
	if !ok {
		return s.C(ent_post.FieldPinnedRank)
	}

	shares := sql.Table(ent_threadshare.Table)
	s.LeftJoin(shares).On(s.C(ent_post.FieldID), shares.C(ent_threadshare.FieldPostID))
	s.OnP(sql.And(
		sql.EQ(shares.C(ent_threadshare.FieldChannelID), cid),
		sql.IsNull(shares.C(ent_threadshare.FieldDeletedAt)),
	))

	return "COALESCE(" +
		shares.C(ent_threadshare.FieldPinnedRank) + ", " +
		s.C(ent_post.FieldPinnedRank) + ")"
}

func (d *Querier) GetMany(ctx context.Context, threadIDs []post.ID, accountID opt.Optional[account.AccountID]) ([]*thread.Thread, error) {
	ctx, span := d.ins.Instrument(ctx,
		kv.Int("thread_count", len(threadIDs)),
		kv.String("account_id", accountID.String()),
	)
	defer span.End()

	if len(threadIDs) == 0 {
		return []*thread.Thread{}, nil
	}

	idList := dt.Map(threadIDs, func(id post.ID) xid.ID { return xid.ID(id) })

	pool1 := pond.NewGroup()

	var replyStatsMap post.PostRepliesMap
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "replies_status")
		defer span.End()

		r, err := d.getRepliesStatus(ctx, idList, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		replyStatsMap = r
		return nil
	})

	var collectionsMap collection_item_status.CollectionStatusMap
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "collections_status")
		defer span.End()

		r, err := d.getCollectionsStatus(ctx, idList, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		collectionsMap = r
		return nil
	})

	var readStateMap post.ReadStateMap
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "read_status")
		defer span.End()

		r, err := d.getReadStatus(ctx, idList, accountID.String())
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
		readStateMap = r
		return nil
	})

	var tagsByThread map[xid.ID]tag_ref.Tags
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "thread_tags")
		defer span.End()

		tagsResult, err := d.db.Tag.Query().
			Where(ent_tag.HasPostsWith(ent_post.IDIn(idList...))).
			WithPosts(func(pq *ent.PostQuery) {
				pq.Where(ent_post.IDIn(idList...))
			}).
			All(ctx)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}

		tagsByThread = make(map[xid.ID]tag_ref.Tags)
		for _, tag := range tagsResult {
			for _, p := range tag.Edges.Posts {
				tagsByThread[p.ID] = append(tagsByThread[p.ID], tag_ref.Map(nil)(tag))
			}
		}

		return nil
	})

	var assetsByThread map[xid.ID][]*asset.Asset
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "thread_assets")
		defer span.End()

		assetsResult, err := d.db.Asset.Query().
			Where(ent_asset.HasPostsWith(ent_post.IDIn(idList...))).
			WithPosts(func(pq *ent.PostQuery) {
				pq.Where(ent_post.IDIn(idList...))
			}).
			All(ctx)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}

		assetsByThread = make(map[xid.ID][]*asset.Asset)
		for _, a := range assetsResult {
			for _, p := range a.Edges.Posts {
				assetsByThread[p.ID] = append(assetsByThread[p.ID], asset.Map(a))
			}
		}

		return nil
	})

	var threadResults []*ent.Post
	pool1.SubmitErr(func() error {
		ctx, span := d.ins.InstrumentNamed(ctx, "thread_roots")
		defer span.End()

		r, err := d.db.Post.Query().
			Where(
				ent_post.DeletedAtIsNil(),
				ent_post.RootPostIDIsNil(),
				ent_post.IDIn(idList...),
			).
			WithCategory().
			WithLink(func(lq *ent.LinkQuery) {
				lq.WithFaviconImage().WithPrimaryImage()
			}).
			All(ctx)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.Internal))
		}

		threadResults = r

		return nil
	})

	err := pool1.Wait()
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	postIDs := dt.Map(threadResults, func(p *ent.Post) xid.ID { return p.ID })
	accountIDs := dt.Map(threadResults, func(p *ent.Post) xid.ID { return p.AccountPosts })

	reactResult, err := d.db.React.Query().
		Where(ent_react.PostIDIn(postIDs...)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	likesMap, err := d.getLikesStatus(ctx, postIDs, accountID.String())
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	reacters := dt.Map(reactResult, func(r *ent.React) xid.ID { return r.AccountID })
	accountIDs = append(accountIDs, reacters...)
	accountIDs = lo.Uniq(accountIDs)

	var accountLookup account.Lookup
	accountEdges, err := d.db.Account.Query().
		Where(ent_account.IDIn(accountIDs...)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	accountLookup = account.NewAccountLookup(accountEdges)

	reacts, err := dt.MapErr(reactResult, reaction.Mapper(accountLookup))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	reactLookup := reaction.Reacts(reacts).Map()

	threadMapper := thread.Mapper(accountLookup, readStateMap, likesMap, collectionsMap, replyStatsMap, reactLookup)

	threads := make([]*thread.Thread, 0, len(threadResults))
	for _, threadResult := range threadResults {
		p, err := threadMapper(threadResult)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}

		p.Tags = tagsByThread[threadResult.ID]
		p.Assets = assetsByThread[threadResult.ID]
		p.Replies = pagination.NewPageResult(pagination.Parameters{}, 0, []*reply.Reply{})

		threads = append(threads, p)
	}

	return threads, nil
}
