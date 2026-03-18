package reply_queue_manager

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/opt"

	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue_querier"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue_writer"
)

type Manager struct {
	querier *reply_admin_queue_querier.Querier
	writer  *reply_admin_queue_writer.Writer
}

func New(
	querier *reply_admin_queue_querier.Querier,
	writer *reply_admin_queue_writer.Writer,
) *Manager {
	return &Manager{querier: querier, writer: writer}
}

type ListOpts struct {
	CreatedAfter  opt.Optional[time.Time]
	CreatedBefore opt.Optional[time.Time]
}

func (m *Manager) List(
	ctx context.Context,
	page pagination.Parameters,
	opts ListOpts,
) (pagination.Result[*reply_admin_queue.Entry], error) {
	filters := reply_admin_queue_querier.Filters{}
	opts.CreatedAfter.Call(func(t time.Time) { filters.CreatedAfter = &t })
	opts.CreatedBefore.Call(func(t time.Time) { filters.CreatedBefore = &t })

	result, err := m.querier.List(ctx, page, filters)
	if err != nil {
		return pagination.Result[*reply_admin_queue.Entry]{}, fault.Wrap(err, fctx.With(ctx))
	}

	return result, nil
}

func (m *Manager) Dismiss(ctx context.Context, id reply_admin_queue.ID) error {
	return fault.Wrap(m.writer.Dismiss(ctx, id), fctx.With(ctx))
}
