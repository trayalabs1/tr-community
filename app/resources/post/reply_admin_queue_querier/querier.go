package reply_admin_queue_querier

import (
	"context"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue"
	"github.com/Southclaws/storyden/internal/ent"
	ent_raq "github.com/Southclaws/storyden/internal/ent/replyadminqueue"
)

type Querier struct {
	db *ent.Client
}

func New(db *ent.Client) *Querier {
	return &Querier{db: db}
}

type Filters struct {
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

func (q *Querier) List(
	ctx context.Context,
	page pagination.Parameters,
	filters Filters,
) (pagination.Result[*reply_admin_queue.Entry], error) {
	query := q.db.ReplyAdminQueue.Query().
		Order(ent.Desc(ent_raq.FieldCreatedAt))

	if filters.CreatedAfter != nil {
		query = query.Where(ent_raq.CreatedAtGTE(*filters.CreatedAfter))
	}
	if filters.CreatedBefore != nil {
		query = query.Where(ent_raq.CreatedAtLTE(*filters.CreatedBefore))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return pagination.Result[*reply_admin_queue.Entry]{}, fault.Wrap(err, fctx.With(ctx))
	}

	rows, err := query.
		Limit(page.Limit()).
		Offset(page.Offset()).
		All(ctx)
	if err != nil {
		return pagination.Result[*reply_admin_queue.Entry]{}, fault.Wrap(err, fctx.With(ctx))
	}

	mapped, err := dt.MapErr(rows, reply_admin_queue.Map)
	if err != nil {
		return pagination.Result[*reply_admin_queue.Entry]{}, fault.Wrap(err, fctx.With(ctx))
	}

	return pagination.NewPageResult(page, total, mapped), nil
}
