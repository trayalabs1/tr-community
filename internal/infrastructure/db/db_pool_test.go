package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/tracing"
)

func TestNewSQLAppliesPoolSettings(t *testing.T) {
	cfg := config.Config{
		DatabaseURL:             "sqlite://" + t.TempDir() + "/data.db",
		DatabaseMaxOpenConns:    17,
		DatabaseMaxIdleConns:    17,
		DatabaseConnMaxLifetime: 30 * time.Minute,
		DatabaseConnMaxIdleTime: 5 * time.Minute,
	}

	d, x, err := newSQL(fxtest.NewLifecycle(t), tracing.NewNoop(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close(); _ = x.Close() })

	stats := d.Stats()
	require.Equal(t, 17, stats.MaxOpenConnections)

	xstats := x.DB.Stats()
	require.Equal(t, 17, xstats.MaxOpenConnections)
}
