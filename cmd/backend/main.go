package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/dig"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources"
	"github.com/Southclaws/storyden/app/services"
	"github.com/Southclaws/storyden/app/services/account/username"
	transport "github.com/Southclaws/storyden/app/transports"
	"github.com/Southclaws/storyden/internal/boot_time"
	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/infrastructure"
)

// startupSeeding runs username cache seeding in a goroutine on server startup
func startupSeeding(
	lc fx.Lifecycle,
	logger *slog.Logger,
	seeder *username.Seeder,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run seeding in goroutine to not block server startup
			go func() {
				// Create a background context with timeout
				seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				logger.Info("starting username cache seeding")

				if err := seeder.SeedIfNeeded(seedCtx); err != nil {
					// Log error but don't fail server startup
					logger.Error("failed to seed username cache",
						slog.String("error", err.Error()))
					logger.Warn("username checks will fall back to database queries")
				} else {
					logger.Info("username cache seeding completed successfully")
				}
			}()

			// Don't block startup
			return nil
		},
	})
}

// Start starts the application and blocks until fatal error
// The server will shut down if the root context is cancelled
// nolint:errcheck
func Start(ctx context.Context) {
	app := fx.New(
		fx.NopLogger,

		fx.Provide(func() context.Context { return ctx }),

		config.Build(),
		infrastructure.Build(),
		resources.Build(),
		services.Build(),
		transport.Build(),

		// Add startup hook for username seeding
		fx.Invoke(startupSeeding),
	)

	err := app.Start(ctx)
	if err != nil {
		// Get the underlying error, without all the fx details.
		underlying := dig.RootCause(err)

		fmt.Println(underlying)

		os.Exit(1)
	}

	// Wait for context cancellation from the caller (interrupt signals etc.)
	<-ctx.Done()

	// Graceful shutdown time is 30 seconds. This context is passed to fx's stop
	// API which is then used to run all the OnStop hooks with a 30 sec timeout.
	ctx, cf := context.WithTimeout(context.Background(), time.Second*30)
	defer cf()

	if err := app.Stop(ctx); err != nil {
		slog.Error("fatal error occurred", slog.String("error", err.Error()))
	}
}

func main() {
	boot_time.StartedAt = time.Now()

	godotenv.Load()

	ctx, cf := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cf()

	Start(ctx)
}
