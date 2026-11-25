package bindings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithHowToUseTooltip(t *testing.T) {
	cutoff := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	before := cutoff.Add(-time.Second)
	after := cutoff.Add(time.Hour)

	assert.Equal(t, true, withHowToUseTooltip(nil, after, cutoff)[howToUseShowMetaKey])
	assert.Equal(t, true, withHowToUseTooltip(nil, cutoff, cutoff)[howToUseShowMetaKey])
	assert.Equal(t, false, withHowToUseTooltip(nil, before, cutoff)[howToUseShowMetaKey])
	assert.Equal(t, false, withHowToUseTooltip(nil, after, time.Time{})[howToUseShowMetaKey])
	assert.Equal(t, false, withHowToUseTooltip(map[string]any{howToUseSeenMetaKey: true}, after, cutoff)[howToUseShowMetaKey])

	original := map[string]any{"case_id": "x"}
	got := withHowToUseTooltip(original, after, cutoff)
	assert.Equal(t, "x", got["case_id"])
	assert.NotContains(t, original, howToUseShowMetaKey)
}
