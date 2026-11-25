package summary_job

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/rueidis/rueidislock"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/ent"
)

func Build() fx.Option {
	return fx.Options(
		fx.Invoke(run),
	)
}

func run(lc fx.Lifecycle, logger *slog.Logger, cfg config.Config, db *ent.Client) {
	if cfg.SlackQueueSummaryWebhookURL == "" {
		return
	}

	loc, err := time.LoadLocation(cfg.SlackQueueSummaryTimezone)
	if err != nil {
		logger.Error("invalid SLACK_QUEUE_SUMMARY_TIMEZONE, queue summary job disabled", slog.String("error", err.Error()))
		return
	}

	var nextWake func(now time.Time) time.Time
	var lockKey func(now time.Time) string // nil disables locking entirely

	if cfg.SlackQueueSummaryInterval > 0 {
		logger.Warn("SLACK_QUEUE_SUMMARY_INTERVAL is set, running on a fixed interval instead of the daily schedule; do not use in production",
			slog.Duration("interval", cfg.SlackQueueSummaryInterval))

		interval := cfg.SlackQueueSummaryInterval
		nextWake = func(now time.Time) time.Time { return now.Add(interval) }
	} else {
		hour, minute, err := parseTimeOfDay(cfg.SlackQueueSummaryTime)
		if err != nil {
			logger.Error("invalid SLACK_QUEUE_SUMMARY_TIME, queue summary job disabled", slog.String("error", err.Error()))
			return
		}

		nextWake = func(now time.Time) time.Time { return nextFireTime(now, hour, minute, loc) }
		lockKey = func(now time.Time) string { return dailyLockKey(now, loc) }
	}

	var locker rueidislock.Locker
	if lockKey != nil {
		var err error
		locker, err = newLocker(cfg.RedisURL)
		if err != nil {
			logger.Error("failed to set up queue summary locker, queue summary job disabled", slog.String("error", err.Error()))
			return
		}
		if locker == nil {
			logger.Warn("no redis configured, queue summary job will run on every replica without dedup")
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}

	stop := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go loop(logger, client, cfg.SlackQueueSummaryWebhookURL, db, locker, nextWake, lockKey, loc, stop)
			return nil
		},
		OnStop: func(context.Context) error {
			close(stop)
			if locker != nil {
				locker.Close()
			}
			return nil
		},
	})
}

func loop(
	logger *slog.Logger,
	client *http.Client,
	webhookURL string,
	db *ent.Client,
	locker rueidislock.Locker,
	nextWake func(now time.Time) time.Time,
	lockKey func(now time.Time) string,
	loc *time.Location,
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
		logger.Info("queue summary job starting", slog.Time("fired_at", start))

		fireCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		release := func() {}
		if lockKey != nil {
			claimed, r, err := tryClaimDay(fireCtx, locker, lockKey(time.Now()))
			if err != nil {
				logger.Error("queue summary job finished: failed to claim lock", slog.String("error", err.Error()), slog.Duration("took", time.Since(start)))
				cancel()
				continue
			}
			if !claimed {
				logger.Info("queue summary job finished: another replica already claimed today's summary", slog.Duration("took", time.Since(start)))
				cancel()
				continue
			}
			release = r
		}

		computedAt := time.Now()
		counts, err := ComputeCounts(fireCtx, db, computedAt)
		if err != nil {
			logger.Error("queue summary job finished: failed to compute counts", slog.String("error", err.Error()), slog.Duration("took", time.Since(start)))
			release()
			cancel()
			continue
		}

		if err := postToSlack(fireCtx, client, webhookURL, formatMessage(counts, computedAt, loc)); err != nil {
			logger.Error("queue summary job finished: failed to post to slack",
				slog.String("error", err.Error()),
				slog.Int("pending_review", counts.PendingReview),
				slog.Int("pending_reply", counts.PendingReply),
				slog.Int("pending_reply_to_reply", counts.PendingReplyToReply),
				slog.Duration("took", time.Since(start)),
			)
		} else {
			logger.Info("queue summary job finished: posted to slack",
				slog.Int("pending_review", counts.PendingReview),
				slog.Int("pending_reply", counts.PendingReply),
				slog.Int("pending_reply_to_reply", counts.PendingReplyToReply),
				slog.Duration("took", time.Since(start)),
			)
		}
		release()
		cancel()
	}
}
