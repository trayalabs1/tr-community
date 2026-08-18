package summary_job

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

// dailyLockKey returns a lock key unique to the calendar day of `at` in loc,
// so that at most one replica sends the summary for any given day.
func dailyLockKey(at time.Time, loc *time.Location) string {
	return fmt.Sprintf("queue-summary:%s", at.In(loc).Format("2006-01-02"))
}

// newLocker builds a dedicated distributed locker with its own Redis
// connection, independent of any Redis client used elsewhere in the app.
// Returns nil if redisURL is empty, meaning dedup across replicas is
// unavailable and every replica will run the job independently.
func newLocker(redisURL url.URL) (rueidislock.Locker, error) {
	if redisURL.String() == "" {
		return nil, nil
	}

	password, _ := redisURL.User.Password()

	return rueidislock.NewLocker(rueidislock.LockerOption{
		KeyPrefix:   "storyden-queue-summary-lock",
		KeyValidity: time.Hour,
		ClientOption: rueidis.ClientOption{
			InitAddress:      []string{redisURL.Host},
			Username:         redisURL.User.Username(),
			Password:         password,
			DisableCache:     true,
			ConnWriteTimeout: 5 * time.Second,
		},
	})
}

// tryClaimDay attempts to acquire the distributed lock for the given key. It
// returns true if this replica should proceed (either it holds the lock, or
// no locker is configured and dedup is skipped entirely).
func tryClaimDay(ctx context.Context, locker rueidislock.Locker, key string) (claimed bool, release func(), err error) {
	if locker == nil {
		return true, func() {}, nil
	}

	_, cancel, err := locker.TryWithContext(ctx, key)
	if err != nil {
		if errors.Is(err, rueidislock.ErrNotLocked) {
			return false, func() {}, nil
		}
		return false, nil, err
	}

	return true, cancel, nil
}
