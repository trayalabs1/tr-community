package thread

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/resources/post/thread_querier"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/authentication/session"
)

const (
	// feedSeqRows is the number of leading feed rows precomputed and cached as an
	// ordered ID sequence. Pages beyond this window hit the DB directly.
	feedSeqRows = 1000
	// feedSeqTTL is how long a cached feed sequence is served before rebuild.
	feedSeqTTL = 30 * time.Minute

	feedSeqKeyPrefix = "feed:seq:"
)

// isCacheableFeed reports whether the request targets the default unfiltered
// channel feed, which is the only variant precomputed and cached.
func isCacheableFeed(opts Params) bool {
	return opts.ChannelID.Ok() &&
		!opts.Query.Ok() &&
		!opts.AccountID.Ok() &&
		!opts.Tags.Ok() &&
		!opts.Categories.Ok() &&
		!opts.Visibility.Ok() &&
		!opts.CreatedBefore.Ok() &&
		!opts.CreatedAfter.Ok() &&
		!opts.UpdatedBefore.Ok() &&
		!opts.ExcludeBAH &&
		!opts.ExcludeFeedback &&
		!opts.BAHOnly &&
		len(opts.PostCategories) == 0 &&
		len(opts.Sentiments) == 0
}

func isModerator(ctx context.Context, accountID opt.Optional[account.AccountID]) bool {
	if !accountID.Ok() {
		return false
	}
	roles := session.GetRoles(ctx)
	return roles.Permissions().HasAny(rbac.PermissionManagePosts, rbac.PermissionAdministrator)
}

func feedSeqKey(channelID xid.ID, moderator bool) string {
	bucket := "member"
	if moderator {
		bucket = "mod"
	}
	return feedSeqKeyPrefix + channelID.String() + ":" + bucket
}

// cachedWindow computes the [lo:hi) slice of a cached sequence of length seqLen
// for the given page, and whether a following page exists. hasMore is true when
// the sequence extends past this window, or when the sequence fills the whole
// cached window (feedSeqRows) and the DB may hold further rows.
func cachedWindow(seqLen, page, size int) (lo, hi int, hasMore bool) {
	lo = page * size
	if lo > seqLen {
		lo = seqLen
	}
	hi = lo + size
	if hi > seqLen {
		hi = seqLen
	}
	hasMore = hi < seqLen || seqLen == feedSeqRows
	return lo, hi, hasMore
}

// listCachedInterleaved serves the interleaved feed from the cached ID sequence.
// The handled return is false when the request should fall through to the plain
// ranked query (page beyond the cached window, or on any cache failure).
func (s *service) listCachedInterleaved(
	ctx context.Context,
	channelID xid.ID,
	page int,
	size int,
	accountID opt.Optional[account.AccountID],
	moderator bool,
	q []thread_querier.Query,
) (*thread_querier.Result, bool, error) {
	lower := page * size
	if lower >= feedSeqRows {
		// Entirely beyond the cached window; let the plain query handle it.
		return nil, false, nil
	}

	seq, err := s.feedSequence(ctx, channelID, accountID, moderator, q)
	if err != nil {
		return nil, false, fault.Wrap(err, fctx.With(ctx))
	}

	lo, hi, hasMore := cachedWindow(len(seq), page, size)

	if lo >= hi {
		return &thread_querier.Result{
			PageSize:    size,
			Results:     0,
			CurrentPage: page,
			NextPage:    opt.NewSafe(page+1, hasMore),
			Threads:     []*thread.Thread{},
		}, true, nil
	}

	threads, err := s.hydrateInOrder(ctx, seq[lo:hi], accountID)
	if err != nil {
		return nil, false, fault.Wrap(err, fctx.With(ctx))
	}

	return &thread_querier.Result{
		PageSize:    size,
		Results:     len(threads),
		CurrentPage: page,
		NextPage:    opt.NewSafe(page+1, hasMore),
		Threads:     threads,
	}, true, nil
}

// feedSequence returns the cached ordered ID sequence, rebuilding it inline on a
// cache miss. When FEED_CACHE_DISABLE is set, the cache is bypassed entirely —
// every call rebuilds from the database and nothing is read from or written to
// cacheStore. Intended for local development, where a 30-minute-stale feed
// sequence makes it hard to see ranking changes take effect.
func (s *service) feedSequence(
	ctx context.Context,
	channelID xid.ID,
	accountID opt.Optional[account.AccountID],
	moderator bool,
	q []thread_querier.Query,
) ([]post.ID, error) {
	if s.feedCacheDisable {
		return s.buildFeedSequence(ctx, accountID, q)
	}

	key := feedSeqKey(channelID, moderator)

	if cached, err := s.cacheStore.Get(ctx, key); err == nil && cached != "" {
		var ids []string
		if json.Unmarshal([]byte(cached), &ids) == nil {
			return dt.Map(ids, func(s string) post.ID {
				id, _ := xid.FromString(s)
				return post.ID(id)
			}), nil
		}
		// Cached value is unparseable (corrupt or a stale format); drop it so the
		// rebuild below can repopulate a clean sequence.
		_ = s.cacheStore.Delete(ctx, key)
	}

	seq, err := s.buildFeedSequence(ctx, accountID, q)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	encoded, err := json.Marshal(dt.Map(seq, func(id post.ID) string { return id.String() }))
	if err == nil {
		_ = s.cacheStore.Set(ctx, key, string(encoded), feedSeqTTL)
	}

	return seq, nil
}

// buildFeedSequence computes the full interleaved feed over the leading
// feedSeqRows rows and returns the ordered post IDs.
func (s *service) buildFeedSequence(
	ctx context.Context,
	accountID opt.Optional[account.AccountID],
	q []thread_querier.Query,
) ([]post.ID, error) {
	combos := shareCombos()

	organicQ := append(append([]thread_querier.Query{}, q...), thread_querier.ExcludeShareCombos(combos))
	organic, err := s.threadQuerier.List(ctx, 0, feedSeqRows, accountID, organicQ...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to list organic threads"))
	}

	shareQ := append(append([]thread_querier.Query{}, q...), thread_querier.OnlyShareCombos(combos))
	shares, err := s.threadQuerier.ListShares(ctx, accountID, shareQ...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to list share threads"))
	}

	merged, _ := mergeFeed(organic.Threads, shares, 0, feedSeqRows)

	return dt.Map(merged, func(t *thread.Thread) post.ID { return t.ID }), nil
}

// hydrateInOrder loads the given threads and returns them in the requested ID
// order (GetMany does not preserve order).
func (s *service) hydrateInOrder(
	ctx context.Context,
	ids []post.ID,
	accountID opt.Optional[account.AccountID],
) ([]*thread.Thread, error) {
	threads, err := s.threadQuerier.GetMany(ctx, ids, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	byID := make(map[post.ID]*thread.Thread, len(threads))
	for _, t := range threads {
		byID[t.ID] = t
	}

	ordered := make([]*thread.Thread, 0, len(ids))
	for _, id := range ids {
		if t, ok := byID[id]; ok {
			ordered = append(ordered, t)
		}
	}

	return ordered, nil
}
