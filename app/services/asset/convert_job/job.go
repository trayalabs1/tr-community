package convert_job

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func runConvertConsumer(
	ctx context.Context,
	lc fx.Lifecycle,
	logger *slog.Logger,
	bus *pubsub.Bus,
	cc *Consumer,
) {
	lc.Append(fx.StartHook(func(hctx context.Context) error {
		_, err := pubsub.SubscribeCommand(ctx, bus, "convert_job.convert_asset", func(ctx context.Context, cmd *message.CommandConvertAsset) error {
			if err := cc.Convert(ctx, cmd.AssetID); err != nil {
				logger.Error("failed to convert asset", slog.String("error", err.Error()), slog.String("asset_id", cmd.AssetID.String()))
				return err
			}
			return nil
		})

		return err
	}))
}
