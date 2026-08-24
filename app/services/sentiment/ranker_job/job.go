package ranker_job

import (
	"context"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/settings"
	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

func Build() fx.Option {
	return fx.Options(
		fx.Invoke(run),
	)
}

func run(lc fx.Lifecycle, logger *slog.Logger, cfg config.Config, db *ent.Client, settingsRepo *settings.SettingsRepository) {
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
			go loop(logger, db, settingsRepo, nextWake, stop)
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
	settingsRepo *settings.SettingsRepository,
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

		updated, err := ApplyDailyIncrements(fireCtx, db, settingsRepo, since)
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
func ApplyDailyIncrements(ctx context.Context, db *ent.Client, settingsRepo *settings.SettingsRepository, since time.Time) (int, error) {
	weights, err := LoadEngagementWeights(ctx, settingsRepo)
	if err != nil {
		weights = DefaultEngagementWeights()
	}

	deltas, err := GetDailyIncrements(ctx, db, since, weights)
	if err != nil {
		return 0, err
	}

	ids := make([]xid.ID, 0, len(deltas))
	for postID, delta := range deltas {
		if delta == 0 {
			continue
		}
		ids = append(ids, postID)
	}

	if len(ids) == 0 {
		return 0, nil
	}

	err = bulkAddRankScore(ctx, db, deltas, ids)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// bulkAddRankScore applies every post's rank_score delta in a single
// UPDATE statement instead of one round-trip per post, using a SQL CASE
// expression to vary the increment per row:
//
//	UPDATE post_sentiments
//	SET rank_score = rank_score + CASE post_id
//		WHEN $1 THEN $2
//		WHEN $3 THEN $4
//		...
//	END
//	WHERE post_id IN ($1, $3, ...)
func bulkAddRankScore(ctx context.Context, db *ent.Client, deltas map[xid.ID]float64, ids []xid.ID) error {
	_, err := db.PostSentiment.
		Update().
		Where(ent_post_sentiment.PostIDIn(ids...)).
		Modify(func(u *sql.UpdateBuilder) {
			u.Set(ent_post_sentiment.FieldRankScore, sql.ExprFunc(func(b *sql.Builder) {
				b.Ident(ent_post_sentiment.FieldRankScore)
				b.WriteString(" + (CASE ")
				b.Ident(ent_post_sentiment.FieldPostID)
				for _, id := range ids {
					b.WriteString(" WHEN ")
					b.Arg(id)
					b.WriteString(" THEN ")
					b.Arg(deltas[id])
				}
				b.WriteString(" END)")
			}))
		}).
		Save(ctx)

	return err
}
