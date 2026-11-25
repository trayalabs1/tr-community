package session_cache

import (
	"context"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

type Invalidator struct {
	loader account_querier.SessionAccountLoader
	bus    *pubsub.Bus
}

func newInvalidator(loader account_querier.SessionAccountLoader, bus *pubsub.Bus) *Invalidator {
	return &Invalidator{loader: loader, bus: bus}
}

func (i *Invalidator) subscribe(ctx context.Context) error {
	if i.bus == nil {
		return nil
	}

	if _, err := pubsub.Subscribe(ctx, i.bus, "session_account.updated", func(ctx context.Context, evt *message.EventAccountUpdated) error {
		return i.loader.Invalidate(ctx, evt.ID)
	}); err != nil {
		return err
	}

	if _, err := pubsub.Subscribe(ctx, i.bus, "session_account.suspended", func(ctx context.Context, evt *message.EventAccountSuspended) error {
		return i.loader.Invalidate(ctx, evt.ID)
	}); err != nil {
		return err
	}

	if _, err := pubsub.Subscribe(ctx, i.bus, "session_account.unsuspended", func(ctx context.Context, evt *message.EventAccountUnsuspended) error {
		return i.loader.Invalidate(ctx, evt.ID)
	}); err != nil {
		return err
	}

	return nil
}

func runOnBoot(ctx context.Context, lc fx.Lifecycle, loader account_querier.SessionAccountLoader, bus *pubsub.Bus) {
	i := newInvalidator(loader, bus)

	lc.Append(fx.StartHook(func(context.Context) error {
		return i.subscribe(ctx)
	}))
}

func Build() fx.Option {
	return fx.Options(
		fx.Invoke(runOnBoot),
	)
}
