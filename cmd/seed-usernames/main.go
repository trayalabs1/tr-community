package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/services/account/username"
	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/infrastructure/db"
	"github.com/Southclaws/storyden/internal/infrastructure/redis"
)

func main() {
	app := fx.New(
		fx.NopLogger,
		config.Build(),
		db.Build(),
		redis.Build(),
		username.Build(),

		fx.Invoke(func(seeder *username.Seeder) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			log.Println("Starting manual username seeding...")
			startTime := time.Now()

			// Check current status first
			redisCount, _ := seeder.GetUsernameCount(ctx)
			log.Printf("Current Redis count: %d\n", redisCount)

			// Force reseed
			log.Println("Clearing existing cache...")
			if err := seeder.ClearUsernameCache(ctx); err != nil {
				log.Fatalf("Failed to clear cache: %v", err)
			}

			log.Println("Seeding usernames from database...")
			if err := seeder.SeedUsernamesFromDB(ctx); err != nil {
				log.Fatalf("Failed to seed usernames: %v", err)
			}

			newCount, err := seeder.GetUsernameCount(ctx)
			if err != nil {
				log.Fatalf("Failed to get count: %v", err)
			}

			elapsed := time.Since(startTime)
			fmt.Printf("✓ Successfully seeded %d usernames to Redis in %v\n", newCount, elapsed)
			os.Exit(0)
		}),
	)

	app.Run()
}
