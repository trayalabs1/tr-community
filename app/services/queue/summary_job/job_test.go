package summary_job

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/go-sqlite"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/ent/enttest"
)

func init() {
	sqlite.RegisterAsSQLITE3()
}

func newTestEntClient(t *testing.T) (*ent.Client, func()) {
	t.Helper()

	dsn := "file:" + xid.New().String() + "?mode=memory&cache=shared&_pragma=foreign_keys(1)"
	c := enttest.Open(t, "sqlite3", dsn)

	return c, func() { c.Close() }
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLoopFixedIntervalWithoutLocking(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db, cleanup := newTestEntClient(t)
	defer cleanup()

	logger := testLogger()

	stop := make(chan struct{})
	nextWake := func(now time.Time) time.Time { return now.Add(5 * time.Millisecond) }

	go loop(logger, server.Client(), server.URL, db, nil, nextWake, nil, time.UTC, stop)

	require.Eventually(t, func() bool {
		return posts.Load() >= 3
	}, time.Second, 5*time.Millisecond, "loop should fire repeatedly on the fixed interval with no locking")

	close(stop)
}

func TestLoopSkipsWhenLockNotClaimed(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db, cleanup := newTestEntClient(t)
	defer cleanup()

	logger := testLogger()

	locker := newFakeLocker()
	locker.locked["always-locked"] = true // pre-claimed by "another replica"

	stop := make(chan struct{})
	nextWake := func(now time.Time) time.Time { return now.Add(5 * time.Millisecond) }
	lockKey := func(now time.Time) string { return "always-locked" }

	go loop(logger, server.Client(), server.URL, db, locker, nextWake, lockKey, time.UTC, stop)

	time.Sleep(50 * time.Millisecond)
	close(stop)

	assert.Equal(t, int32(0), posts.Load(), "loop must not post when the lock is held elsewhere")
}
