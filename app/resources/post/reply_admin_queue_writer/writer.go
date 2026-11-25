package reply_admin_queue_writer

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue"
	"github.com/Southclaws/storyden/internal/ent"
	ent_raq "github.com/Southclaws/storyden/internal/ent/replyadminqueue"
)

type Writer struct {
	db *ent.Client
}

func New(db *ent.Client) *Writer {
	return &Writer{db: db}
}

func (w *Writer) Enqueue(
	ctx context.Context,
	replyID post.ID,
	threadID post.ID,
	channelID xid.ID,
	snippet string,
) (*reply_admin_queue.Entry, error) {
	row, err := w.db.ReplyAdminQueue.Create().
		SetReplyID(xid.ID(replyID)).
		SetThreadID(xid.ID(threadID)).
		SetChannelID(channelID).
		SetContentSnippet(snippet).
		Save(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return reply_admin_queue.Map(row)
}

func (w *Writer) Dismiss(ctx context.Context, id reply_admin_queue.ID) error {
	err := w.db.ReplyAdminQueue.DeleteOneID(xid.ID(id)).Exec(ctx)
	return fault.Wrap(err, fctx.With(ctx))
}

func (w *Writer) DismissByReplyID(ctx context.Context, replyID post.ID) error {
	_, err := w.db.ReplyAdminQueue.Delete().
		Where(ent_raq.ReplyID(xid.ID(replyID))).
		Exec(ctx)
	return fault.Wrap(err, fctx.With(ctx))
}
