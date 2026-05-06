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
		_, err := pubsub.SubscribeCommand(ctx, bus, "sentiment_job.score_post", func(ctx context.Context, cmd *message.CommandScorePostSentiment) error {
			select {
			case batchCh <- cmd.PostID:
			default:
				logger.Warn("sentiment scoring batch channel full, blocking enqueue",
					slog.String("post_id", cmd.PostID.String()),
					slog.Int("capacity", cap(batchCh)),
				)
				batchCh <- cmd.PostID
			}
			return nil
		})
		if err != nil {
			return err
		}

		go runBatchProcessor(ctx, logger, sc, batchCh)

		return nil
	}))
}

func runBatchProcessor(ctx context.Context, logger *slog.Logger, sc *sentimentConsumer, batchCh <-chan post.ID) {
	var batch []post.ID
	var timer *time.Timer

	flush := func() {
		if len(batch) == 0 {
			return
		}
		sc.scoreBatch(ctx, batch)
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return

		case id := <-batchCh:
			batch = append(batch, id)

			if timer == nil {
				timer = time.NewTimer(batchTimeout)
			}

			if len(batch) >= batchSize {
				if timer != nil {
					timer.Stop()
					timer = nil
				}
				flush()
			}

		case <-func() <-chan time.Time {
			if timer != nil {
				return timer.C
			}
			return nil
		}():
			timer = nil
			flush()
		}
	}
}
