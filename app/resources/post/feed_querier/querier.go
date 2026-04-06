package feed_querier

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/internal/ent"
	ent_link "github.com/Southclaws/storyden/internal/ent/link"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

type Querier struct {
	db *ent.Client
}

func New(db *ent.Client) *Querier {
	return &Querier{db: db}
}

type FeedResult struct {
	Threads     []*thread.Thread
	CurrentPage int
	PageSize    int
	TotalPages  int
	Results     int
}

func (q *Querier) GetFeed(ctx context.Context, channelID xid.ID, params pagination.Parameters) (*FeedResult, error) {
	pageSize := params.Limit()
	offset := params.Offset()

	baseQuery := q.db.Post.Query().
		Where(
			ent_post.ChannelIDEQ(channelID),
			ent_post.DeletedAtIsNil(),
			ent_post.VisibilityEQ(ent_post.VisibilityPublished),
			ent_post.RootPostIDIsNil(),
		).
		WithAuthor(func(q *ent.AccountQuery) {
			q.WithRoles()
		}).
		WithCategory().
		WithTags().
		WithAssets().
		WithLink(func(lq *ent.LinkQuery) {
			lq.WithAssets().Order(ent_link.ByCreatedAt(sql.OrderDesc()))
		}).
		WithSentiment()

	total, err := baseQuery.Clone().Count(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	posts, err := baseQuery.Clone().
		Modify(func(s *sql.Selector) {
			t := sql.Table(ent_post_sentiment.Table)
			s.LeftJoin(t).On(s.C(ent_post.FieldID), t.C(ent_post_sentiment.FieldPostID))
			s.OrderExpr(
				sql.Expr(s.C(ent_post.FieldPinnedRank)+" DESC"),
				sql.Expr("COALESCE("+t.C(ent_post_sentiment.FieldRankScore)+", -1) DESC"),
				sql.Expr(s.C(ent_post.FieldCreatedAt)+" DESC"),
			)
		}).
		Limit(pageSize).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	threads, err := mapThreads(posts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	totalPages := (total + pageSize - 1) / pageSize
	currentPage := (offset / pageSize) + 1

	return &FeedResult{
		Threads:     threads,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Results:     total,
	}, nil
}

func mapThreads(posts []*ent.Post) ([]*thread.Thread, error) {
	threads := make([]*thread.Thread, 0, len(posts))
	for _, p := range posts {
		t, err := thread.Map(p)
		if err != nil {
			return nil, err
		}
		threads = append(threads, t)
	}
	return threads, nil
}
