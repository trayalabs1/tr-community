package thread

import (
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
)

func org(slug string) *thread.Thread {
	return &thread.Thread{Slug: slug, Post: post.Post{Slug: slug}}
}

func share(slug, category, typ string) *thread.Thread {
	meta := map[string]any{"post_category": category}
	if typ != "" {
		meta["type"] = typ
	}
	return &thread.Thread{Slug: slug, Post: post.Post{Slug: slug, Meta: meta}}
}

func slugs(threads []*thread.Thread) []string {
	out := make([]string, len(threads))
	for i, t := range threads {
		out[i] = t.Slug
	}
	return out
}

func orgSeq(n int) []*thread.Thread {
	out := make([]*thread.Thread, n)
	for i := 0; i < n; i++ {
		out[i] = org("o" + string(rune('A'+i)))
	}
	return out
}

// oneOfEach returns one share for each of the 5 rotation steps, in rotation
// order: feedback/story, feedback/progress, BAH/21, tip, BAH/7.
func oneOfEach() []*thread.Thread {
	return []*thread.Thread{
		share("s-story", "feedback", "story"),
		share("s-prog", "feedback", "progress"),
		share("s-bah21", "BAH", "21"),
		share("s-tip", "tip", ""),
		share("s-bah7", "BAH", "7"),
	}
}

func TestMergeFeed_SharesAtEveryFifthSlot(t *testing.T) {
	organic := orgSeq(20)
	shares := oneOfEach()

	window, hasMore := mergeFeed(organic, shares, 0, 25)
	got := slugs(window)

	// positions 5,10,15,20,25 (indices 4,9,14,19,24) follow the rotation order.
	assert.Equal(t, "s-story", got[4], "slot 5 -> feedback/story")
	assert.Equal(t, "s-prog", got[9], "slot 10 -> feedback/progress")
	assert.Equal(t, "s-bah21", got[14], "slot 15 -> BAH/21")
	assert.Equal(t, "s-tip", got[19], "slot 20 -> tip")
	assert.Equal(t, "s-bah7", got[24], "slot 25 -> BAH/7")

	// organic fills all other slots in order.
	assert.Equal(t, []string{"oA", "oB", "oC", "oD"}, got[0:4])
	assert.Equal(t, []string{"oE", "oF", "oG", "oH"}, got[5:9])
	assert.False(t, hasMore)
}

func TestMergeFeed_BorrowsForwardWhenStepEmpty(t *testing.T) {
	organic := orgSeq(30)
	// story has 2 posts (s1 newest, s2), progress has 1 (p1), all other steps
	// empty. Shares are pre-sorted created_at DESC within their type.
	shares := []*thread.Thread{
		share("s1", "feedback", "story"),
		share("s2", "feedback", "story"),
		share("p1", "feedback", "progress"),
	}

	window, _ := mergeFeed(organic, shares, 0, 30)
	got := slugs(window)

	assert.Equal(t, "s1", got[4], "slot 5 -> story s1")
	assert.Equal(t, "p1", got[9], "slot 10 -> progress p1")
	// BAH21 step is empty; the rotation walks forward and story's remaining s2 is
	// borrowed into slot 15.
	assert.Equal(t, "s2", got[14], "slot 15 -> story s2 (borrowed forward)")
	// All shares consumed; remaining share slots fall back to organic.
	shareCats := map[string]bool{"feedback": true, "BAH": true, "tip": true}
	assert.False(t, shareCats[metaString(window[19], "post_category")], "slot 20 -> organic")
	assert.False(t, shareCats[metaString(window[24], "post_category")], "slot 25 -> organic")
	assert.False(t, shareCats[metaString(window[29], "post_category")], "slot 30 -> organic")
}

func TestMergeFeed_TypeSpecificMatching(t *testing.T) {
	organic := orgSeq(20)
	// BAH steps are type-specific (21 then 7). A BAH post with an off-list type
	// (99) is not a share at all and should never occupy a share slot.
	shares := []*thread.Thread{
		share("s-story", "feedback", "story"),
		share("s-prog", "feedback", "progress"),
		share("s-bah21", "BAH", "21"),
		share("s-tip", "tip", ""),
		share("s-bah7", "BAH", "7"),
	}

	window, _ := mergeFeed(organic, shares, 0, 25)
	got := slugs(window)

	assert.Equal(t, "s-bah21", got[14], "slot 15 -> BAH/21 specifically")
	assert.Equal(t, "s-bah7", got[24], "slot 25 -> BAH/7 specifically")
}

func TestMergeFeed_MissingTypeCompacts(t *testing.T) {
	organic := orgSeq(20)
	// No feedback/story share. Compaction: slot 5 places the next available step
	// (feedback/progress); the pointer resumes after it, so slot 10 gets BAH/21.
	shares := []*thread.Thread{
		share("s-prog", "feedback", "progress"),
		share("s-bah21", "BAH", "21"),
		share("s-tip", "tip", ""),
	}

	window, _ := mergeFeed(organic, shares, 0, 20)
	got := slugs(window)

	assert.Equal(t, "s-prog", got[4], "slot 5 -> feedback/progress (story empty, compacted)")
	assert.Equal(t, "s-bah21", got[9], "slot 10 -> BAH/21")
	assert.Equal(t, "s-tip", got[14], "slot 15 -> tip")
	// All shares consumed; slot 20 is dry -> organic fills it.
	assert.NotContains(t, []string{"feedback", "BAH", "tip"}, metaString(window[19], "post_category"), "slot 20 -> organic")
}

func TestMergeFeed_AllTypesDryFillsOrganic(t *testing.T) {
	organic := orgSeq(14)
	shares := []*thread.Thread{
		share("s-tip", "tip", ""),
	}

	window, _ := mergeFeed(organic, shares, 0, 15)
	got := slugs(window)

	assert.Equal(t, "s-tip", got[4], "slot 5 -> tip")
	assert.Equal(t, []string{"oE", "oF", "oG", "oH", "oI"}, got[5:10], "slots after shares drain -> organic")
}

func TestMergeFeed_OrganicExhaustedAppendsRemainingShares(t *testing.T) {
	organic := orgSeq(3)
	shares := []*thread.Thread{
		share("s-story", "feedback", "story"),
		share("s-bah21", "BAH", "21"),
		share("s-tip", "tip", ""),
	}

	window, hasMore := mergeFeed(organic, shares, 0, 50)
	got := slugs(window)

	// 3 organic then all shares appended sequentially in rotation order, no gaps.
	assert.Equal(t, []string{"oA", "oB", "oC", "s-story", "s-bah21", "s-tip"}, got)
	assert.False(t, hasMore)
}

func TestMergeFeed_SharesExhaustedPureOrganic(t *testing.T) {
	organic := orgSeq(12)
	shares := []*thread.Thread{
		share("s-tip", "tip", ""),
	}

	window, _ := mergeFeed(organic, shares, 0, 15)
	got := slugs(window)

	assert.Equal(t, "s-tip", got[4], "slot 5 -> only share")
	assert.Equal(t, []string{"oE", "oF", "oG", "oH", "oI"}, got[5:10])
	assert.Len(t, got, 13, "12 organic + 1 share")
}

func TestMergeFeed_EmptyShares(t *testing.T) {
	organic := orgSeq(7)
	window, _ := mergeFeed(organic, nil, 0, 15)
	assert.Equal(t, slugs(organic), slugs(window))
}

func TestMergeFeed_EmptyOrganic(t *testing.T) {
	shares := []*thread.Thread{
		share("s-story", "feedback", "story"),
		share("s-bah21", "BAH", "21"),
	}
	window, hasMore := mergeFeed(nil, shares, 0, 15)
	assert.Equal(t, []string{"s-story", "s-bah21"}, slugs(window))
	assert.False(t, hasMore)
}

func TestMergeFeed_BAHNumericType(t *testing.T) {
	organic := orgSeq(20)
	// BAH type arrives as a JSON number (float64 after decode), matching the "21"
	// and "7" string steps via metaString's numeric handling.
	bah21 := &thread.Thread{Slug: "s-bah21", Post: post.Post{Slug: "s-bah21", Meta: map[string]any{"post_category": "BAH", "type": float64(21)}}}
	bah7 := &thread.Thread{Slug: "s-bah7", Post: post.Post{Slug: "s-bah7", Meta: map[string]any{"post_category": "BAH", "type": float64(7)}}}
	shares := []*thread.Thread{bah21, bah7}

	window, _ := mergeFeed(organic, shares, 0, 15)
	got := slugs(window)

	// story/progress empty -> compaction places BAH/21 at slot 5, BAH/7 at slot 10.
	assert.Equal(t, "s-bah21", got[4], "slot 5 -> BAH numeric 21")
	assert.Equal(t, "s-bah7", got[9], "slot 10 -> BAH numeric 7")
}

func TestMergeFeed_PaginationContinuity(t *testing.T) {
	organic := orgSeq(60)
	shares := oneOfEach()

	page0, more0 := mergeFeed(organic, shares, 0, 50)
	page1, _ := mergeFeed(organic, shares, 1, 50)

	assert.Len(t, page0, 50)
	assert.True(t, more0)

	full, _ := mergeFeed(organic, shares, 0, 100)
	assert.Equal(t, slugs(full)[0:50], slugs(page0))
	assert.Equal(t, slugs(full)[50:], slugs(page1))
}

func TestCachedWindow(t *testing.T) {
	// Full page within a short sequence.
	lo, hi, more := cachedWindow(120, 0, 50)
	assert.Equal(t, 0, lo)
	assert.Equal(t, 50, hi)
	assert.True(t, more)

	// Last partial page of a short sequence.
	lo, hi, more = cachedWindow(120, 2, 50)
	assert.Equal(t, 100, lo)
	assert.Equal(t, 120, hi)
	assert.False(t, more, "120 rows, page 2 is the tail")

	// Page past the end of a short sequence.
	lo, hi, more = cachedWindow(120, 5, 50)
	assert.Equal(t, 120, lo)
	assert.Equal(t, 120, hi)
	assert.False(t, more)

	// Full cached window (feedSeqRows) always reports hasMore so the DB path
	// picks up beyond row 1000.
	lo, hi, more = cachedWindow(feedSeqRows, (feedSeqRows/50)-1, 50)
	assert.Equal(t, feedSeqRows-50, lo)
	assert.Equal(t, feedSeqRows, hi)
	assert.True(t, more, "sequence filled to 1000 -> DB may have more")
}

func TestIsCacheableFeed(t *testing.T) {
	base := Params{ChannelID: opt.New(xid.NilID()), InterleaveShares: true}
	assert.True(t, isCacheableFeed(base), "default unfiltered channel feed is cacheable")

	assert.False(t, isCacheableFeed(Params{}), "no channel id -> not cacheable")

	withQuery := base
	withQuery.Query = opt.New("hair")
	assert.False(t, isCacheableFeed(withQuery), "search query -> not cacheable")

	withAuthor := base
	withAuthor.AccountID = opt.New(account.AccountID(xid.NilID()))
	assert.False(t, isCacheableFeed(withAuthor), "author filter -> not cacheable")

	withExclude := base
	withExclude.ExcludeBAH = true
	assert.False(t, isCacheableFeed(withExclude), "exclude_bah -> not cacheable")

	withCats := base
	withCats.PostCategories = []string{"BAH"}
	assert.False(t, isCacheableFeed(withCats), "post categories -> not cacheable")
}

func TestFeedSeqKey(t *testing.T) {
	id := xid.NilID()
	assert.NotEqual(t, feedSeqKey(id, true), feedSeqKey(id, false), "mod and member buckets differ")
	assert.Contains(t, feedSeqKey(id, true), "mod")
	assert.Contains(t, feedSeqKey(id, false), "member")
}

func TestMergeFeed_HasMoreDetection(t *testing.T) {
	organic := orgSeq(50)
	window, hasMore := mergeFeed(organic, nil, 0, 50)
	assert.Len(t, window, 50)
	assert.False(t, hasMore, "exactly one page, no more")

	organic = orgSeq(51)
	window, hasMore = mergeFeed(organic, nil, 0, 50)
	assert.Len(t, window, 50)
	assert.True(t, hasMore, "51 items -> second page exists")
}
