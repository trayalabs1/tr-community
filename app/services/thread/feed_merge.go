package thread

import (
	"strconv"

	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/resources/post/thread_querier"
)

const shareSlotInterval = 5

type rotationStep struct {
	category string
	// typ, when non-empty, must match the post's metadata "type". An empty typ
	// matches any post in the category.
	typ string
}

var rotation = []rotationStep{
	{category: "feedback", typ: "story"},
	{category: "feedback", typ: "progress"},
	{category: "BAH", typ: "21"},
	{category: "tip"},
	{category: "BAH", typ: "7"},
}

// shareCombos projects the rotation into the querier's combo type so the organic
// and share queries scope on the exact same set the merge rotates over.
func shareCombos() []thread_querier.ShareCombo {
	combos := make([]thread_querier.ShareCombo, len(rotation))
	for i, step := range rotation {
		combos[i] = thread_querier.ShareCombo{Category: step.category, Type: step.typ}
	}
	return combos
}

func metaString(t *thread.Thread, key string) string {
	if t.Meta == nil {
		return ""
	}
	switch v := t.Meta[key].(type) {
	case string:
		return v
	case float64:
		// JSON numbers (e.g. BAH streak counts) decode as float64; render as the
		// integer text Postgres yields from metadata->>'type'.
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

func matchesStep(t *thread.Thread, step rotationStep) bool {
	if metaString(t, "post_category") != step.category {
		return false
	}
	if step.typ == "" {
		return true
	}
	return metaString(t, "type") == step.typ
}

// mergeFeed interleaves the organic and share queues into a single feed and
// returns the window for the requested page. Positions 5, 10, 15, ... (1-indexed)
// are reserved for share posts drawn from the rotation. At each share slot the
// rotation is walked from the current pointer and the first type with an
// available share is placed, so types with no eligible share collapse out rather
// than yielding the slot to organic; the slot only becomes organic when every
// type is dry. When the organic queue is exhausted before the share queue,
// remaining shares are appended sequentially in rotation order.
//
// hasMore reports whether at least one item exists beyond the returned window.
func mergeFeed(organic []*thread.Thread, shares []*thread.Thread, page, size int) ([]*thread.Thread, bool) {
	upper := (page + 1) * size

	merged := make([]*thread.Thread, 0, upper+1)
	used := make([]bool, len(shares))
	organicIdx := 0
	rotationPtr := 0

	// Build enough of the merged feed to fill the requested window plus one
	// extra item to detect a following page. Stop early once both queues drain.
	for len(merged) <= upper {
		organicRemaining := organicIdx < len(organic)
		sharesRemaining := usedCount(used) < len(shares)
		if !organicRemaining && !sharesRemaining {
			break
		}

		pos := len(merged) + 1
		isShareSlot := pos%shareSlotInterval == 0

		if isShareSlot && organicRemaining {
			// Walk the rotation from the current pointer and place the first type
			// that has an available share, resuming the pointer after it. Empty
			// types collapse out rather than yielding the slot to organic.
			if idx := nextRemainingShare(shares, used, &rotationPtr); idx >= 0 {
				used[idx] = true
				merged = append(merged, shares[idx])
				continue
			}

			// Whole rotation is dry: fill from organic.
			merged = append(merged, organic[organicIdx])
			organicIdx++
			continue
		}

		if organicRemaining {
			merged = append(merged, organic[organicIdx])
			organicIdx++
			continue
		}

		// Organic exhausted: append remaining shares sequentially in rotation
		// order, cycling through steps until every share is placed.
		if idx := nextRemainingShare(shares, used, &rotationPtr); idx >= 0 {
			used[idx] = true
			merged = append(merged, shares[idx])
			continue
		}

		break
	}

	lower := page * size
	if lower >= len(merged) {
		return []*thread.Thread{}, false
	}

	end := lower + size
	hasMore := end < len(merged)
	if end > len(merged) {
		end = len(merged)
	}

	return merged[lower:end], hasMore
}

func usedCount(used []bool) int {
	n := 0
	for _, u := range used {
		if u {
			n++
		}
	}
	return n
}

// nextShareForStep returns the index of the most-recent unused share matching
// the step, or -1. Shares are pre-sorted created_at DESC, so the first match wins.
func nextShareForStep(shares []*thread.Thread, used []bool, step rotationStep) int {
	for i, s := range shares {
		if used[i] {
			continue
		}
		if matchesStep(s, step) {
			return i
		}
	}
	return -1
}

// nextRemainingShare drains the share queue once organic is exhausted. It walks
// the rotation to preserve rotation order; if the current step has no unused
// share it advances until it finds one, giving up after a full cycle with no hit.
func nextRemainingShare(shares []*thread.Thread, used []bool, rotationPtr *int) int {
	for attempts := 0; attempts < len(rotation); attempts++ {
		step := rotation[*rotationPtr%len(rotation)]
		*rotationPtr++
		if idx := nextShareForStep(shares, used, step); idx >= 0 {
			return idx
		}
	}

	// No step matched a remaining share (e.g. a category not covered by the
	// rotation); append whatever is left in created_at order.
	for i := range shares {
		if !used[i] {
			return i
		}
	}
	return -1
}
