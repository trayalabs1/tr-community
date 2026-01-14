package username

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/redis/rueidis"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
)

const (
	usernameSetKey     = "usernames:set"
	usernameLockPrefix = "username:lock:"
	lockTTL            = 10 * time.Second
)

var (
	ErrUsernameTaken = fault.New("username already taken", ftag.With(ftag.AlreadyExists))
	ErrLockFailed    = fault.New("failed to acquire lock")
	ErrAlreadySet    = fault.New("username already set", ftag.With(ftag.InvalidArgument))
	ErrInvalidFormat = fault.New("invalid username format", ftag.With(ftag.InvalidArgument))
)

type Service struct {
	redis         rueidis.Client
	accountWriter *account_writer.Writer
	accountQuery  *account_querier.Querier
}

func New(
	redis rueidis.Client,
	accountWriter *account_writer.Writer,
	accountQuery *account_querier.Querier,
) *Service {
	return &Service{
		redis:         redis,
		accountWriter: accountWriter,
		accountQuery:  accountQuery,
	}
}

// CheckAvailability checks if username is available (fast Redis check + DB fallback)
func (s *Service) CheckAvailability(ctx context.Context, username string) (bool, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	// Validate format first
	if err := validateUsername(username); err != nil {
		return false, err
	}

	// Check Redis set first (fast path)
	if s.redis != nil {
		exists, err := s.checkRedisSet(ctx, username)
		if err != nil {
			// Log error but continue to DB check
			// Don't fail the check if Redis is down
		} else {
			if exists {
				return false, nil // Definitely taken
			}
		}
	}

	// Double-check in DB (Redis might be stale or miss)
	_, exists, err := s.accountQuery.LookupByHandle(ctx, username)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx))
	}

	// If exists in DB but not in Redis, add to Redis for future checks
	if exists && s.redis != nil {
		_ = s.addToRedisSet(ctx, username)
	}

	return !exists, nil
}

// SetUsername sets the username for an account with distributed locking
func (s *Service) SetUsername(ctx context.Context, accountID account.AccountID, username string) (*account.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	// Validate format
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	// Get current account
	acc, err := s.accountQuery.GetByID(ctx, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Check if already has permanent username
	if !strings.HasPrefix(acc.Handle, "temp_") {
		return nil, fault.Wrap(ErrAlreadySet, fctx.With(ctx))
	}

	// Acquire distributed lock (only if Redis is available)
	var lockKey string
	if s.redis != nil {
		lockKey = usernameLockPrefix + username
		locked, err := s.acquireLock(ctx, lockKey, lockTTL)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
		if !locked {
			return nil, fault.Wrap(ErrLockFailed,
				fctx.With(ctx),
				fmsg.WithDesc("username unavailable", "This username is being claimed by another user. Please try again."))
		}

		// Ensure lock is released
		defer s.releaseLock(ctx, lockKey)
	}

	// Double-check availability (someone might have claimed it before we got the lock)
	available, err := s.CheckAvailability(ctx, username)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !available {
		return nil, fault.Wrap(ErrUsernameTaken,
			fctx.With(ctx),
			fmsg.WithDesc("username taken", "This username was just taken. Please try another."))
	}

	// Update handle in database
	updated, err := s.accountWriter.Update(ctx, accountID,
		account_writer.SetHandle(username))
	if err != nil {
		// Handle unique constraint violation (race condition that bypassed our checks)
		if ftag.Get(err) == ftag.AlreadyExists {
			return nil, fault.Wrap(ErrUsernameTaken,
				fctx.With(ctx),
				fmsg.WithDesc("username taken", "This username is already taken. Please try another."))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Success! Add to Redis set for future checks
	if s.redis != nil {
		_ = s.addToRedisSet(ctx, username)
	}

	return &updated.Account, nil
}

// Redis operations

func (s *Service) checkRedisSet(ctx context.Context, username string) (bool, error) {
	cmd := s.redis.B().Sismember().Key(usernameSetKey).Member(username).Build()
	resp := s.redis.Do(ctx, cmd)
	return resp.AsBool()
}

func (s *Service) addToRedisSet(ctx context.Context, username string) error {
	cmd := s.redis.B().Sadd().Key(usernameSetKey).Member(username).Build()
	return s.redis.Do(ctx, cmd).Error()
}

func (s *Service) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use SET with NX (only set if not exists) and EX (expiry)
	// This is the standard distributed lock pattern
	cmd := s.redis.B().Set().
		Key(key).
		Value("1").
		Nx().
		ExSeconds(int64(ttl.Seconds())).
		Build()

	resp := s.redis.Do(ctx, cmd)
	if resp.Error() != nil {
		return false, resp.Error()
	}

	// If response is "OK", lock acquired successfully
	// If response is nil/empty, key already exists (lock held by someone else)
	str, err := resp.ToString()
	if err != nil {
		// nil response means key already exists
		return false, nil
	}
	return str == "OK", nil
}

func (s *Service) releaseLock(ctx context.Context, key string) error {
	cmd := s.redis.B().Del().Key(key).Build()
	return s.redis.Do(ctx, cmd).Error()
}

// Validation

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var blacklist = map[string]bool{
	"admin": true, "root": true, "system": true, "moderator": true,
	"traya": true, "support": true, "help": true, "api": true,
	"null": true, "undefined": true, "www": true,
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return fault.Wrap(ErrInvalidFormat,
			fmsg.With("username must be 3-20 characters"))
	}

	if !usernameRegex.MatchString(username) {
		return fault.Wrap(ErrInvalidFormat,
			fmsg.With("username can only contain letters, numbers, underscores, and hyphens"))
	}

	if blacklist[strings.ToLower(username)] {
		return fault.Wrap(ErrInvalidFormat,
			fmsg.With("this username is reserved"))
	}

	return nil
}
