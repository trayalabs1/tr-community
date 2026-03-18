package reply_admin_queue_consumer

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue_writer"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func Build() fx.Option {
	return fx.Invoke(func(
		ctx context.Context,
		lc fx.Lifecycle,
		bus *pubsub.Bus,
		logger *slog.Logger,
		postQuerier *post_querier.Querier,
		writer *reply_admin_queue_writer.Writer,
	) {
		consumer := func(_ context.Context) error {
			_, err := pubsub.Subscribe(ctx, bus, "reply_admin_queue.reply_created", func(ctx context.Context, evt *message.EventReplyRequiresAdminAttention) error {
				_, channelID, err := postQuerier.AuthorAndChannelID(ctx, evt.ThreadID)
				if err != nil {
					logger.WarnContext(ctx, "reply_admin_queue: failed to resolve channel_id, proceeding with nil",
						slog.String("thread_id", evt.ThreadID.String()),
						slog.String("error", err.Error()),
					)
				}

				_, err = writer.Enqueue(ctx, evt.ReplyID, evt.ThreadID, channelID, evt.Snippet)
				return err
			})
			return err
		}

		lc.Append(fx.StartHook(consumer))
	})
}
