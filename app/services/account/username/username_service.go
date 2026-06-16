package username

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/redis/rueidis"
	"github.com/rs/xid"

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
	logger        *slog.Logger
	redis         rueidis.Client
	accountWriter *account_writer.Writer
	accountQuery  *account_querier.Querier
}

func New(
	logger *slog.Logger,
	redis rueidis.Client,
	accountWriter *account_writer.Writer,
	accountQuery *account_querier.Querier,
) *Service {
	return &Service{
		logger:        logger,
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

// RegenerateResult summarises one batch of a temporary-handle regeneration run.
type RegenerateResult struct {
	// Processed is the number of temp accounts examined in this batch.
	Processed int
	Updated   int
	Skipped   int
	Failed    int
	// NextCursor is the ID to pass as the cursor for the next batch. It is empty
	// when fewer rows than the requested batch size were returned, signalling
	// that the table has been fully drained.
	NextCursor string
}

const (
	handleGenerationAttempts = 5
	defaultBatchSize         = 500
	maxBatchSize             = 5000
)

// RegenerateTempHandles processes a single bounded batch of accounts that still
// carry a temporary "temp_" handle, replacing each with a freshly generated
// unique handle derived from the account's display name.
//
// It uses keyset pagination: pass an empty cursor to start, then feed the
// returned NextCursor back in for each subsequent batch until NextCursor is
// empty. Because every successfully updated row stops matching the "temp_"
// prefix, the unprocessed set shrinks monotonically and the run is idempotent
// and safely resumable.
func (s *Service) RegenerateTempHandles(ctx context.Context, cursor string, batchSize int) (*RegenerateResult, error) {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	if batchSize > maxBatchSize {
		batchSize = maxBatchSize
	}

	after, err := parseCursor(cursor)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	accounts, err := s.accountQuery.ListTempAccounts(ctx, after, batchSize)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result := &RegenerateResult{Processed: len(accounts)}

	// Only accounts whose name yields a usable handle prefix are candidates for
	// regeneration; the rest are skipped up front.
	pending := make([]account_querier.TempAccount, 0, len(accounts))
	for _, acc := range accounts {
		if handlePrefixFromName(acc.Name) == "" {
			result.Skipped++
			continue
		}
		pending = append(pending, acc)
	}

	if len(pending) > 0 {
		if err := s.assignHandles(ctx, pending, result); err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	// A full batch means there may be more rows; advance the cursor. A short
	// batch means we have reached the end.
	if len(accounts) == batchSize {
		result.NextCursor = accounts[len(accounts)-1].ID.String()
	}

	return result, nil
}

// assignHandles generates unique handles for the pending accounts and writes
// them in one CASE update. Candidates are checked against existing handles in a
// single query and de-duplicated within the batch. On a unique-constraint
// collision from a concurrent write, the whole assignment is retried with fresh
// candidates for the rows that have not yet been committed.
func (s *Service) assignHandles(ctx context.Context, pending []account_querier.TempAccount, result *RegenerateResult) error {
	remaining := pending

	for attempt := 0; attempt < handleGenerationAttempts && len(remaining) > 0; attempt++ {
		candidates := make([]string, 0, len(remaining))
		for _, acc := range remaining {
			candidates = append(candidates, generateHandle(acc.Name))
		}

		taken, err := s.accountQuery.ExistingHandles(ctx, candidates)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}

		assigned := make(map[account.AccountID]string, len(remaining))
		seen := make(map[string]bool, len(remaining))
		var retry []account_querier.TempAccount

		for i, acc := range remaining {
			h := candidates[i]
			if taken[h] || seen[h] {
				retry = append(retry, acc) // collides with DB or another row in this batch
				continue
			}
			seen[h] = true
			assigned[account.AccountID(acc.ID)] = h
		}

		if len(assigned) > 0 {
			err := s.accountWriter.BulkUpdateHandles(ctx, assigned)
			if err != nil {
				if ftag.Get(err) == ftag.AlreadyExists {
					// A concurrent write claimed one of these handles after our
					// existence check. The CASE update is atomic so nothing was
					// committed; retry every still-pending row next round.
					s.logger.Warn("bulk handle update hit a conflict, retrying",
						slog.Int("count", len(assigned)),
						slog.String("error", err.Error()))
					remaining = dedupeAccounts(append(remaining, retry...))
					continue
				}
				return fault.Wrap(err, fctx.With(ctx))
			}

			result.Updated += len(assigned)
			if s.redis != nil {
				for _, h := range assigned {
					_ = s.addToRedisSet(ctx, h)
				}
			}
		}

		remaining = retry
	}

	// Anything still unresolved after all attempts could not be given a unique
	// handle and is reported as failed.
	result.Failed += len(remaining)
	for _, acc := range remaining {
		s.logger.Warn("could not assign unique handle after retries",
			slog.String("account_id", acc.ID.String()),
			slog.String("old_handle", acc.Handle))
	}

	return nil
}

func dedupeAccounts(in []account_querier.TempAccount) []account_querier.TempAccount {
	seen := make(map[xid.ID]bool, len(in))
	out := in[:0]
	for _, acc := range in {
		if seen[acc.ID] {
			continue
		}
		seen[acc.ID] = true
		out = append(out, acc)
	}
	return out
}

func parseCursor(cursor string) (xid.ID, error) {
	if cursor == "" {
		return xid.NilID(), nil
	}
	return xid.FromString(cursor)
}

// generateHandle derives a handle candidate from a display name: a lowercase
// alphanumeric prefix (max 4 chars) plus four random digits. Callers must check
// uniqueness; the caller has already ensured the prefix is non-empty.
func generateHandle(name string) string {
	return fmt.Sprintf("%s%04d", handlePrefixFromName(name), rand.Intn(10000))
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]`)

func handlePrefixFromName(name string) string {
	prefix := nonAlphanumeric.ReplaceAllString(strings.ToLower(name), "")
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	return prefix
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
