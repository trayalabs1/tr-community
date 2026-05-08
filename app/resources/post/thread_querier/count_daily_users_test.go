package thread_querier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIstDayWindow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		now       time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "midday IST",
			now:       time.Date(2026, 5, 6, 6, 30, 0, 0, time.UTC), // 12:00 IST
			wantStart: time.Date(2026, 5, 6, 0, 0, 0, 0, istLocation),
			wantEnd:   time.Date(2026, 5, 7, 0, 0, 0, 0, istLocation),
		},
		{
			name:      "just before IST midnight",
			now:       time.Date(2026, 5, 6, 18, 29, 0, 0, time.UTC), // 23:59 IST same day
			wantStart: time.Date(2026, 5, 6, 0, 0, 0, 0, istLocation),
			wantEnd:   time.Date(2026, 5, 7, 0, 0, 0, 0, istLocation),
		},
		{
			name:      "just after IST midnight rolls to next day",
			now:       time.Date(2026, 5, 6, 18, 31, 0, 0, time.UTC), // 00:01 IST next day
			wantStart: time.Date(2026, 5, 7, 0, 0, 0, 0, istLocation),
			wantEnd:   time.Date(2026, 5, 8, 0, 0, 0, 0, istLocation),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := istDayWindow(tc.now)
			assert.True(t, tc.wantStart.Equal(start), "start: want %s got %s", tc.wantStart, start)
			assert.True(t, tc.wantEnd.Equal(end), "end: want %s got %s", tc.wantEnd, end)
			assert.Equal(t, 24*time.Hour, end.Sub(start))
		})
	}
}
