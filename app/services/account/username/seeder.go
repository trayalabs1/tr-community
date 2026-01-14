package username

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/redis/rueidis"

	"github.com/Southclaws/storyden/app/resources/account/account_querier"
)

type Seeder struct {
	logger       *slog.Logger
	redis        rueidis.Client
	accountQuery *account_querier.Querier
}

func NewSeeder(
	logger *slog.Logger,
	redis rueidis.Client,
	accountQuery *account_querier.Querier,
) *Seeder {
	return &Seeder{
		logger:       logger,
		redis:        redis,
		accountQuery: accountQuery,
	}
}

// SeedIfNeeded checks if seeding is needed and performs it if necessary
// This is the main method called on startup
func (s *Seeder) SeedIfNeeded(ctx context.Context) error {
	if s.redis == nil {
		s.logger.Warn("Redis not configured, skipping username seeding")
		return nil
	}

	startTime := time.Now()

	// Get counts from both sources
	redisCount, err := s.GetUsernameCount(ctx)
	if err != nil {
		s.logger.Error("failed to get Redis count, will reseed",
			slog.String("error", err.Error()))
		redisCount = 0
	}

	dbCount, err := s.accountQuery.CountNonTempAccounts(ctx)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx),
			fmsg.With("failed to count accounts in database"))
	}

	s.logger.Info("checking username seeding status",
		slog.Int64("redis_count", redisCount),
		slog.Int("db_count", dbCount))

	// If counts match, no seeding needed
	if redisCount == int64(dbCount) {
		s.logger.Info("username cache is up to date, skipping seeding")
		return nil
	}

	// Counts differ, perform seeding
	s.logger.Info("username counts differ, starting seeding",
		slog.Int64("redis_count", redisCount),
		slog.Int("db_count", dbCount))

	if err := s.SeedUsernamesFromDB(ctx); err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	elapsed := time.Since(startTime)
	s.logger.Info("username seeding completed",
		slog.Int("total_usernames", dbCount),
		slog.Duration("elapsed", elapsed))

	return nil
}

// SeedUsernamesFromDB populates Redis set with all existing usernames from database
// Optimized to fetch only handles (not full account objects)
func (s *Seeder) SeedUsernamesFromDB(ctx context.Context) error {
	if s.redis == nil {
		return fault.New("Redis not configured")
	}

	// Fetch only handles from database (memory efficient)
	handles, err := s.accountQuery.GetAllHandles(ctx)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx),
			fmsg.With("failed to fetch handles from database"))
	}

	if len(handles) == 0 {
		s.logger.Info("no usernames to seed")
		return nil
	}

	s.logger.Info("fetched handles from database",
		slog.Int("count", len(handles)))

	// Clear existing set to ensure consistency
	if err := s.ClearUsernameCache(ctx); err != nil {
		s.logger.Warn("failed to clear existing cache",
			slog.String("error", err.Error()))
		// Continue anyway
	}

	// Add all usernames to Redis set in batches
	batchSize := 1000
	batchesCount := (len(handles) + batchSize - 1) / batchSize

	for i := 0; i < len(handles); i += batchSize {
		end := i + batchSize
		if end > len(handles) {
			end = len(handles)
		}

		batch := handles[i:end]
		batchNum := (i / batchSize) + 1

		// Build SADD command with multiple members
		cmd := s.redis.B().Sadd().Key(usernameSetKey).Member(batch...).Build()
		if err := s.redis.Do(ctx, cmd).Error(); err != nil {
			return fault.Wrap(err, fctx.With(ctx),
				fmsg.WithDesc("batch failed",
					fmt.Sprintf("failed to add batch %d/%d to Redis", batchNum, batchesCount)))
		}

		s.logger.Debug("seeded username batch",
			slog.Int("batch", batchNum),
			slog.Int("total_batches", batchesCount),
			slog.Int("batch_size", len(batch)))
	}

	return nil
}

// GetUsernameCount returns the number of usernames in Redis set
func (s *Seeder) GetUsernameCount(ctx context.Context) (int64, error) {
	cmd := s.redis.B().Scard().Key(usernameSetKey).Build()
	return s.redis.Do(ctx, cmd).AsInt64()
}

// ClearUsernameCache clears the Redis username set
func (s *Seeder) ClearUsernameCache(ctx context.Context) error {
	cmd := s.redis.B().Del().Key(usernameSetKey).Build()
	return s.redis.Do(ctx, cmd).Error()
}
