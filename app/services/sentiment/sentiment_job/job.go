package sentiment_job

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func runSentimentConsumer(
	ctx context.Context,
	lc fx.Lifecycle,
	logger *slog.Logger,
	bus *pubsub.Bus,
	sc *sentimentConsumer,
) {
	lc.Append(fx.StartHook(func(hctx context.Context) error {
		_, err := pubsub.SubscribeCommand(ctx, bus, "sentiment_job.score_post", func(ctx context.Context, cmd *message.CommandScorePostSentiment) error {
			if err := sc.scorePost(ctx, cmd.PostID); err != nil {
				logger.Error("failed to score post sentiment",
					slog.String("post_id", cmd.PostID.String()),
					slog.String("error", err.Error()),
				)
				return err
			}
			return nil
		})

		return err
	}))
}
