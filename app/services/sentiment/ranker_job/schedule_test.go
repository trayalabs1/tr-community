package ranker_job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimeOfDay(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		hour, minute, err := parseTimeOfDay("06:30")
		require.NoError(t, err)
		assert.Equal(t, 6, hour)
		assert.Equal(t, 30, minute)
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, err := parseTimeOfDay("0630")
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
		now := time.Date(2026, 3, 5, 5, 0, 0, 0, loc)
		next := nextFireTime(now, 6, 30, loc)
		assert.Equal(t, time.Date(2026, 3, 5, 6, 30, 0, 0, loc), next)
	})

	t.Run("already past today, rolls to tomorrow", func(t *testing.T) {
		now := time.Date(2026, 3, 5, 6, 31, 0, 0, loc)
		next := nextFireTime(now, 6, 30, loc)
		assert.Equal(t, time.Date(2026, 3, 6, 6, 30, 0, 0, loc), next)
	})

	t.Run("exactly at fire time, rolls to tomorrow", func(t *testing.T) {
		now := time.Date(2026, 3, 5, 6, 30, 0, 0, loc)
		next := nextFireTime(now, 6, 30, loc)
		assert.Equal(t, time.Date(2026, 3, 6, 6, 30, 0, 0, loc), next)
	})

	t.Run("normalises input timezone to target location", func(t *testing.T) {
		utc := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC) // 05:30 IST
		next := nextFireTime(utc, 6, 30, loc)
		assert.Equal(t, time.Date(2026, 3, 5, 6, 30, 0, 0, loc), next)
	})
}
