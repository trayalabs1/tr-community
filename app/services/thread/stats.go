package thread

import (
	"context"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"
)

const dailyUsersCacheTTL = 5 * time.Minute

type dailyUsersEntry struct {
	count     int
	expiresAt time.Time
}

type dailyUsersCache struct {
	mu      sync.Mutex
	entries map[xid.ID]dailyUsersEntry
}

var globalDailyUsersCache = &dailyUsersCache{
	entries: make(map[xid.ID]dailyUsersEntry),
}

func (s *service) DailyUsersCount(ctx context.Context, channelID xid.ID) (int, error) {
	now := time.Now()

	globalDailyUsersCache.mu.Lock()
	if entry, ok := globalDailyUsersCache.entries[channelID]; ok && now.Before(entry.expiresAt) {
		count := entry.count
		globalDailyUsersCache.mu.Unlock()
		return count, nil
	}
	globalDailyUsersCache.mu.Unlock()

	count, err := s.threadQuerier.CountUniqueAuthorsToday(ctx, channelID)
	if err != nil {
		return 0, fault.Wrap(err, fctx.With(ctx))
	}

	globalDailyUsersCache.mu.Lock()
	globalDailyUsersCache.entries[channelID] = dailyUsersEntry{
		count:     count,
		expiresAt: now.Add(dailyUsersCacheTTL),
	}
	globalDailyUsersCache.mu.Unlock()

	return count, nil
}
