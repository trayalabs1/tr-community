package summary_job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimeOfDay(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		hour, minute, err := parseTimeOfDay("23:55")
		require.NoError(t, err)
		assert.Equal(t, 23, hour)
		assert.Equal(t, 55, minute)
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, err := parseTimeOfDay("2355")
		assert.Error(t, err)
	})

	t.Run("invalid hour", func(t *testing.T) {
		_, _, err := parseTimeOfDay("24:00")
		assert.Error(t, err)
	})

	t.Run("invalid minute", func(t *testing.T) {
		_, _, err := parseTimeOfDay("10:60")
		assert.Error(t, err)
	})
}

func TestNextFireTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	t.Run("later today", func(t *testing.T) {
		now := time.Date(2026, 3, 5, 10, 0, 0, 0, loc)
		next := nextFireTime(now, 23, 55, loc)
		assert.Equal(t, time.Date(2026, 3, 5, 23, 55, 0, 0, loc), next)
	})

	t.Run("already past today, rolls to tomorrow", func(t *testing.T) {
		now := time.Date(2026, 3, 5, 23, 56, 0, 0, loc)
		next := nextFireTime(now, 23, 55, loc)
		assert.Equal(t, time.Date(2026, 3, 6, 23, 55, 0, 0, loc), next)
	})

	t.Run("exactly at fire time, rolls to tomorrow", func(t *testing.T) {
		now := time.Date(2026, 3, 5, 23, 55, 0, 0, loc)
		next := nextFireTime(now, 23, 55, loc)
		assert.Equal(t, time.Date(2026, 3, 6, 23, 55, 0, 0, loc), next)
	})

	t.Run("normalises input timezone to target location", func(t *testing.T) {
		utc := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC) // 17:30 IST
		next := nextFireTime(utc, 23, 55, loc)
		assert.Equal(t, time.Date(2026, 3, 5, 23, 55, 0, 0, loc), next)
	})
}
