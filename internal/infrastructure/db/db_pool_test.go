package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Southclaws/storyden/internal/config"
)

func TestNewSQLAppliesPoolSettings(t *testing.T) {
	cfg := config.Config{
		DatabaseURL:             "sqlite://" + t.TempDir() + "/data.db",
		DatabaseMaxOpenConns:    17,
		DatabaseMaxIdleConns:    17,
		DatabaseConnMaxLifetime: 30 * time.Minute,
		DatabaseConnMaxIdleTime: 5 * time.Minute,
	}

	d, x, err := newSQL(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close(); _ = x.Close() })

	stats := d.Stats()
	require.Equal(t, 17, stats.MaxOpenConnections)

	xstats := x.DB.Stats()
	require.Equal(t, 17, xstats.MaxOpenConnections)
}
