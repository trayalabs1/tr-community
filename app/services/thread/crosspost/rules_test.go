package crosspost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func eligibleCandidate() candidate {
	return candidate{
		IsRoot:          true,
		Visibility:      "published",
		ChannelSlug:     "month-5-masters",
		CaseID:          "0123",
		SentimentStatus: "scored",
		SentimentTag:    "positive",
		Topic:           "RESULTS & PROGRESS",
	}
}

func TestEligible(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*candidate)
		want   bool
	}{
		{"baseline", func(c *candidate) {}, true},
		{"hair regrowth topic", func(c *candidate) { c.Topic = "HAIR REGROWTH" }, true},
		{"case id 1", func(c *candidate) { c.CaseID = "1999" }, true},
		{"month 6-8 channel", func(c *candidate) { c.ChannelSlug = "month-6-8-veterans" }, true},
		{"month 8+ channel", func(c *candidate) { c.ChannelSlug = "month-8-plus-legends" }, true},
		{"case id 2", func(c *candidate) { c.CaseID = "2123" }, false},
		{"missing case id", func(c *candidate) { c.CaseID = "" }, false},
		{"female cohort", func(c *candidate) { c.ChannelSlug = "month-5-elites" }, false},
		{"lead channel", func(c *candidate) { c.ChannelSlug = "traya-explorers" }, false},
		{"month 3 channel", func(c *candidate) { c.ChannelSlug = "month-3-warriors" }, false},
		{"other topic", func(c *candidate) { c.Topic = "HAIRFALL CONCERNS" }, false},
		{"neutral sentiment", func(c *candidate) { c.SentimentTag = "neutral" }, false},
		{"negative sentiment", func(c *candidate) { c.SentimentTag = "negative" }, false},
		{"unscored", func(c *candidate) { c.SentimentStatus = "unscored" }, false},
		{"in review", func(c *candidate) { c.Visibility = "review" }, false},
		{"reply", func(c *candidate) { c.IsRoot = false }, false},
		{"share row", func(c *candidate) { c.IsShare = true }, false},
		{"already a copy", func(c *candidate) { c.IsCopy = true }, false},
		{"admin author", func(c *candidate) { c.IsAdmin = true }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := eligibleCandidate()
			tt.mutate(&c)
			assert.Equal(t, tt.want, c.eligible())
		})
	}
}
