package ranker_job

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

func Build() fx.Option {
	return fx.Options(
		fx.Invoke(run),
	)
}

func run(lc fx.Lifecycle, logger *slog.Logger, cfg config.Config, db *ent.Client) {
	loc, err := time.LoadLocation(cfg.RankerJobTimezone)
	if err != nil {
		logger.Error("invalid RANKER_JOB_TIMEZONE, rank_score engagement job disabled", slog.String("error", err.Error()))
		return
	}

	var nextWake func(now time.Time) time.Time

	if cfg.RankerJobInterval > 0 {
		logger.Warn("RANKER_JOB_INTERVAL is set, running on a fixed interval instead of the daily schedule; do not use in production",
			slog.Duration("interval", cfg.RankerJobInterval))

		interval := cfg.RankerJobInterval
		nextWake = func(now time.Time) time.Time { return now.Add(interval) }
	} else {
		hour, minute, err := parseTimeOfDay(cfg.RankerJobTime)
		if err != nil {
			logger.Error("invalid RANKER_JOB_TIME, rank_score engagement job disabled", slog.String("error", err.Error()))
			return
		}

		nextWake = func(now time.Time) time.Time { return nextFireTime(now, hour, minute, loc) }
	}

	stop := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go loop(logger, db, nextWake, stop)
			return nil
		},
		OnStop: func(context.Context) error {
			close(stop)
			return nil
		},
	})
}

func loop(
	logger *slog.Logger,
	db *ent.Client,
	nextWake func(now time.Time) time.Time,
	stop <-chan struct{},
) {
	for {
		next := nextWake(time.Now())

		timer := time.NewTimer(time.Until(next))
		select {
		case <-timer.C:
		case <-stop:
			timer.Stop()
			return
		}

		start := time.Now()
		logger.Info("rank_score engagement job starting", slog.Time("fired_at", start))

		fireCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		since := start.Add(-24 * time.Hour)

		updated, err := ApplyDailyIncrements(fireCtx, db, since)
		if err != nil {
			logger.Error("rank_score engagement job finished: failed to apply increments",
				slog.String("error", err.Error()),
				slog.Duration("took", time.Since(start)),
			)
		} else {
			logger.Info("rank_score engagement job finished",
				slog.Int("posts_updated", updated),
				slog.Duration("took", time.Since(start)),
			)
		}

		cancel()
	}
}

// ApplyDailyIncrements computes each post's engagement_bonus delta from
// likes/replies created since `since` and adds it onto that post's stored
// rank_score. Returns the number of posts updated.
func ApplyDailyIncrements(ctx context.Context, db *ent.Client, since time.Time) (int, error) {
	deltas, err := GetDailyIncrements(ctx, db, since)
	if err != nil {
		return 0, err
	}

	for postID, delta := range deltas {
		if delta == 0 {
			continue
		}

		_, err := db.PostSentiment.
			Update().
			Where(ent_post_sentiment.PostIDEQ(postID)).
			AddRankScore(delta).
			Save(ctx)
		if err != nil {
			return 0, err
		}
	}

	return len(deltas), nil
}
