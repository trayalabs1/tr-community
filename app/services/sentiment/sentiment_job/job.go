package sentiment_job

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

const (
	batchSize    = 10
	batchTimeout = 2 * time.Second
)

func runSentimentConsumer(
	ctx context.Context,
	lc fx.Lifecycle,
	logger *slog.Logger,
	bus *pubsub.Bus,
	sc *sentimentConsumer,
) {
	batchCh := make(chan post.ID, batchSize*2)

	lc.Append(fx.StartHook(func(hctx context.Context) error {
		logger.Info("ai scoring: subscribing to CommandScorePostSentiment",
			slog.String("handler", "sentiment_job.score_post"),
		)
		_, err := pubsub.SubscribeCommand(ctx, bus, "sentiment_job.score_post", func(ctx context.Context, cmd *message.CommandScorePostSentiment) error {
			logger.Info("ai scoring: command received from bus, queuing for batch",
				slog.String("post_id", cmd.PostID.String()),
				slog.Int("queue_depth", len(batchCh)),
			)
			select {
			case batchCh <- cmd.PostID:
				logger.Info("ai scoring: command queued",
					slog.String("post_id", cmd.PostID.String()),
				)
			default:
				logger.Warn("ai scoring: batch channel full, blocking enqueue",
					slog.String("post_id", cmd.PostID.String()),
					slog.Int("capacity", cap(batchCh)),
				)
				batchCh <- cmd.PostID
			}
			return nil
		})
		if err != nil {
			logger.Error("ai scoring: failed to subscribe to CommandScorePostSentiment",
				slog.String("error", err.Error()),
			)
			return err
		}

		logger.Info("ai scoring: starting batch processor",
			slog.Int("batch_size", batchSize),
			slog.Duration("batch_timeout", batchTimeout),
		)
		go runBatchProcessor(ctx, logger, sc, batchCh)

		return nil
	}))
}

func runBatchProcessor(ctx context.Context, logger *slog.Logger, sc *sentimentConsumer, batchCh <-chan post.ID) {
	var batch []post.ID
	var timer *time.Timer

	flush := func(reason string) {
		if len(batch) == 0 {
			return
		}
		logger.Info("ai scoring: flushing batch to scorer",
			slog.String("reason", reason),
			slog.Int("size", len(batch)),
		)
		sc.scoreBatch(ctx, batch)
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("ai scoring: batch processor shutting down, flushing remaining",
				slog.Int("remaining", len(batch)),
			)
			flush("shutdown")
			return

		case id := <-batchCh:
			batch = append(batch, id)
			logger.Info("ai scoring: post added to batch",
				slog.String("post_id", id.String()),
				slog.Int("batch_size_now", len(batch)),
			)

			if timer == nil {
				timer = time.NewTimer(batchTimeout)
			}

			if len(batch) >= batchSize {
				if timer != nil {
					timer.Stop()
					timer = nil
				}
				flush("size_threshold")
			}

		case <-func() <-chan time.Time {
			if timer != nil {
				return timer.C
			}
			return nil
		}():
			timer = nil
			flush("timeout")
		}
	}
}
