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

func TestRegisterPoolMetrics(t *testing.T) {
	a := assert.New(t)

	reader := metric.NewManualReader()
	otel.SetMeterProvider(metric.NewMeterProvider(metric.WithReader(reader)))

	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer database.Close()

	require.NoError(t, database.Ping())
	require.NoError(t, registerPoolMetrics(database))

	var collected metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &collected))

	names := map[string]bool{}
	for _, scope := range collected.ScopeMetrics {
		for _, m := range scope.Metrics {
			names[m.Name] = true
		}
	}

	a.True(names["db.client.connections.open"], "expected db.client.connections.open, got %v", names)
	a.True(names["db.client.connections.in_use"])
	a.True(names["db.client.connections.idle"])
	a.True(names["db.client.connections.wait_count"])
	a.True(names["db.client.connections.wait_duration"])
}
