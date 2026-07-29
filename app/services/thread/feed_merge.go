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

	sq := newShareQueues(shares)

	merged := make([]*thread.Thread, 0, upper+1)
	organicIdx := 0
	rotationPtr := 0

	// Build enough of the merged feed to fill the requested window plus one
	// extra item to detect a following page. Stop early once both queues drain.
	for len(merged) <= upper {
		organicRemaining := organicIdx < len(organic)
		if !organicRemaining && sq.remaining == 0 {
			break
		}

		pos := len(merged) + 1
		isShareSlot := pos%shareSlotInterval == 0

		if isShareSlot && organicRemaining {
			// Walk the rotation from the current pointer and place the first step
			// that still has a share, resuming the pointer after it. Empty steps
			// collapse out rather than yielding the slot to organic.
			if s := sq.next(&rotationPtr); s != nil {
				merged = append(merged, s)
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
		if s := sq.next(&rotationPtr); s != nil {
			merged = append(merged, s)
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

// shareQueues buckets shares by rotation step for O(1) selection. Each share is
// assigned to the first rotation step it matches (mirroring first-match-wins);
// any share matching no step lands in the fallback queue. Input order is
// preserved within each queue, so heads remain most-recent (created_at DESC).
type shareQueues struct {
	byStep    [][]*thread.Thread
	stepHead  []int
	fallback  []*thread.Thread
	fallHead  int
	remaining int
}

func newShareQueues(shares []*thread.Thread) *shareQueues {
	sq := &shareQueues{
		byStep:   make([][]*thread.Thread, len(rotation)),
		stepHead: make([]int, len(rotation)),
	}

	for _, s := range shares {
		assigned := false
		for i, step := range rotation {
			if matchesStep(s, step) {
				sq.byStep[i] = append(sq.byStep[i], s)
				assigned = true
				break
			}
		}
		if !assigned {
			sq.fallback = append(sq.fallback, s)
		}
	}

	sq.remaining = len(shares)
	return sq
}

// next returns the most-recent unused share by walking the rotation from
// rotationPtr, resuming the pointer just after the step it drew from. When every
// step queue is empty it drains the fallback queue; nil when all are exhausted.
func (sq *shareQueues) next(rotationPtr *int) *thread.Thread {
	for attempts := 0; attempts < len(rotation); attempts++ {
		i := *rotationPtr % len(rotation)
		(*rotationPtr)++
		if sq.stepHead[i] < len(sq.byStep[i]) {
			s := sq.byStep[i][sq.stepHead[i]]
			sq.stepHead[i]++
			sq.remaining--
			return s
		}
	}

	if sq.fallHead < len(sq.fallback) {
		s := sq.fallback[sq.fallHead]
		sq.fallHead++
		sq.remaining--
		return s
	}

	return nil
}
