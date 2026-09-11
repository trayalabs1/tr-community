package db

import (
	"context"
	"database/sql"

	"github.com/Southclaws/fault"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func registerPoolMetrics(database *sql.DB) error {
	meter := otel.Meter("github.com/Southclaws/storyden/internal/infrastructure/db")

	open, err := meter.Int64ObservableGauge("db.client.connections.open",
		metric.WithDescription("Established connections, both in use and idle"),
	)
	if err != nil {
		return fault.Wrap(err)
	}

	inUse, err := meter.Int64ObservableGauge("db.client.connections.in_use",
		metric.WithDescription("Connections currently in use"),
	)
	if err != nil {
		return fault.Wrap(err)
	}

	idle, err := meter.Int64ObservableGauge("db.client.connections.idle",
		metric.WithDescription("Idle connections in the pool"),
	)
	if err != nil {
		return fault.Wrap(err)
	}

	waitCount, err := meter.Int64ObservableCounter("db.client.connections.wait_count",
		metric.WithDescription("Total number of connections waited for"),
	)
	if err != nil {
		return fault.Wrap(err)
	}

	waitDuration, err := meter.Int64ObservableCounter("db.client.connections.wait_duration",
		metric.WithDescription("Total time blocked waiting for a new connection"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return fault.Wrap(err)
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		stats := database.Stats()

		observer.ObserveInt64(open, int64(stats.OpenConnections))
		observer.ObserveInt64(inUse, int64(stats.InUse))
		observer.ObserveInt64(idle, int64(stats.Idle))
		observer.ObserveInt64(waitCount, stats.WaitCount)
		observer.ObserveInt64(waitDuration, stats.WaitDuration.Milliseconds())

		return nil
	}, open, inUse, idle, waitCount, waitDuration)
	if err != nil {
		return fault.Wrap(err)
	}

	return nil
}
