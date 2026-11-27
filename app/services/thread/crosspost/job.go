package crosspost

import (
	"context"

	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func runCrossposter(ctx context.Context, lc fx.Lifecycle, bus *pubsub.Bus, c *Crossposter) {
	lc.Append(fx.StartHook(func(hctx context.Context) error {
		if _, err := pubsub.Subscribe(ctx, bus, "crosspost.published", func(ctx context.Context, evt *message.EventThreadPublished) error {
			return c.OnPublished(ctx, xid.ID(evt.ID))
		}); err != nil {
			return err
		}

		if _, err := pubsub.Subscribe(ctx, bus, "crosspost.scored", func(ctx context.Context, evt *message.EventThreadSentimentScored) error {
			return c.OnScored(ctx, xid.ID(evt.ID))
		}); err != nil {
			return err
		}

		if _, err := pubsub.Subscribe(ctx, bus, "crosspost.unpublished", func(ctx context.Context, evt *message.EventThreadUnpublished) error {
			return c.OnUnpublished(ctx, xid.ID(evt.ID))
		}); err != nil {
			return err
		}

		if _, err := pubsub.Subscribe(ctx, bus, "crosspost.submitted_for_review", func(ctx context.Context, evt *message.EventThreadSubmittedForReview) error {
			return c.OnUnpublished(ctx, xid.ID(evt.ID))
		}); err != nil {
			return err
		}

		_, err := pubsub.Subscribe(ctx, bus, "crosspost.deleted", func(ctx context.Context, evt *message.EventThreadDeleted) error {
			return c.OnDeleted(ctx, xid.ID(evt.ID))
		})
		return err
	}))
}
