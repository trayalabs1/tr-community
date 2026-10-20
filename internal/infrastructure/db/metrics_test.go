package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func openPinged(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	require.NoError(t, database.Ping())
	return database
}

func TestRegisterPoolMetrics(t *testing.T) {
	a := assert.New(t)

	reader := metric.NewManualReader()
	otel.SetMeterProvider(metric.NewMeterProvider(metric.WithReader(reader)))

	require.NoError(t, registerPoolMetrics(openPinged(t), "ent"))
	require.NoError(t, registerPoolMetrics(openPinged(t), "sqlx"))

	var collected metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &collected))

	openByPool := map[string]int64{}
	names := map[string]bool{}
	for _, scope := range collected.ScopeMetrics {
		for _, m := range scope.Metrics {
			names[m.Name] = true

			if m.Name != "db.client.connections.open" {
				continue
			}

			gauge, ok := m.Data.(metricdata.Gauge[int64])
			require.True(t, ok)
			for _, dp := range gauge.DataPoints {
				pool, _ := dp.Attributes.Value("db.client.connection.pool.name")
				openByPool[pool.AsString()] = dp.Value
			}
		}
	}

	a.True(names["db.client.connections.in_use"])
	a.True(names["db.client.connections.idle"])
	a.True(names["db.client.connections.wait_count"])
	a.True(names["db.client.connections.wait_duration"])

	a.Len(openByPool, 2)
	a.GreaterOrEqual(openByPool["ent"], int64(1))
	a.GreaterOrEqual(openByPool["sqlx"], int64(1))
}
