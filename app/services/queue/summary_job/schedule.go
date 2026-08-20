package summary_job

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseTimeOfDay parses an "HH:MM" 24-hour string into hour/minute components.
func parseTimeOfDay(s string) (hour, minute int, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time of day %q, expected HH:MM", s)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour in time of day %q", s)
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute in time of day %q", s)
	}

	return hour, minute, nil
}

// nextFireTime returns the next moment at or after `now` that matches the
// given hour:minute in loc. If `now` is already past today's occurrence, it
// returns tomorrow's occurrence instead.
func nextFireTime(now time.Time, hour, minute int, loc *time.Location) time.Time {
	local := now.In(loc)

	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}
