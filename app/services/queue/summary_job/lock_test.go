package summary_job

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyLockKey(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)

	key := dailyLockKey(time.Date(2026, 3, 5, 23, 55, 0, 0, loc), loc)
	assert.Equal(t, "queue-summary:2026-03-05", key)

	// A moment that's a different UTC day but the same IST day must
	// produce the same key, since the lock should dedup by local day.
	utc := time.Date(2026, 3, 5, 20, 0, 0, 0, time.UTC) // 01:30 IST on 2026-03-06
	assert.Equal(t, "queue-summary:2026-03-06", dailyLockKey(utc, loc))
}

// fakeLocker is a minimal in-memory stand-in for rueidislock.Locker, used to
// test tryClaimDay's contract without a real Redis instance. It models
// exactly the semantics tryClaimDay depends on: TryWithContext fails with
// ErrNotLocked if another caller already holds the named lock.
type fakeLocker struct {
	mu     sync.Mutex
	locked map[string]bool
}

func newFakeLocker() *fakeLocker {
	return &fakeLocker{locked: map[string]bool{}}
}

func (f *fakeLocker) TryWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.locked[name] {
		return nil, nil, rueidislock.ErrNotLocked
	}
	f.locked[name] = true

	release := func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		delete(f.locked, name)
	}

	return ctx, release, nil
}

func (f *fakeLocker) WithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	panic("not used")
}

func (f *fakeLocker) ForceWithContext(ctx context.Context, name string) (context.Context, context.CancelFunc, error) {
	panic("not used")
}

func (f *fakeLocker) Client() rueidis.Client { panic("not used") }

func (f *fakeLocker) Close() {}

func TestTryClaimDay(t *testing.T) {
	t.Run("nil locker always claims", func(t *testing.T) {
		claimed, release, err := tryClaimDay(context.Background(), nil, "key")
		require.NoError(t, err)
		assert.True(t, claimed)
		release() // must not panic
	})

	t.Run("only one of many concurrent callers claims the same key", func(t *testing.T) {
		locker := newFakeLocker()

		const n = 20
		var wg sync.WaitGroup
		var start sync.WaitGroup
		claims := make([]bool, n)
		releases := make([]func(), n)
		start.Add(1)

		for i := range n {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				start.Wait() // maximise overlap so all n attempts race the same window
				claimed, release, err := tryClaimDay(context.Background(), locker, "shared-key")
				require.NoError(t, err)
				claims[i] = claimed
				releases[i] = release
			}(i)
		}
		start.Done()
		wg.Wait()

		claimedCount := 0
		for i, c := range claims {
			if c {
				claimedCount++
				releases[i]()
			}
		}
		assert.Equal(t, 1, claimedCount, "exactly one caller should claim the lock")
	})

	t.Run("different keys can be claimed independently", func(t *testing.T) {
		locker := newFakeLocker()

		claimedA, releaseA, err := tryClaimDay(context.Background(), locker, "day-a")
		require.NoError(t, err)
		assert.True(t, claimedA)
		defer releaseA()

		claimedB, releaseB, err := tryClaimDay(context.Background(), locker, "day-b")
		require.NoError(t, err)
		assert.True(t, claimedB)
		defer releaseB()
	})

	t.Run("releasing allows a later claim of the same key", func(t *testing.T) {
		locker := newFakeLocker()

		claimed, release, err := tryClaimDay(context.Background(), locker, "key")
		require.NoError(t, err)
		require.True(t, claimed)
		release()

		claimedAgain, release2, err := tryClaimDay(context.Background(), locker, "key")
		require.NoError(t, err)
		assert.True(t, claimedAgain)
		release2()
	})
}
