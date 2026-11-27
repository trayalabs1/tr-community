package crosspost

import (
	"strings"

	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
)

const (
	TargetChannelSlug = "traya-explorers"

	visibilityPublished = "published"
	sentimentScored     = "scored"
)

var sourceChannelSlugs = map[string]struct{}{
	"month-5-masters":      {},
	"month-6-8-veterans":   {},
	"month-8-plus-legends": {},
}

var allowedTopics = map[string]struct{}{
	string(scorer.CategoryResultsProgress): {},
	string(scorer.CategoryHairRegrowth):    {},
}

var allowedCaseIDPrefixes = []string{"0", "1"}

type candidate struct {
	IsAdmin         bool
	IsRoot          bool
	IsShare         bool
	IsCopy          bool
	Visibility      string
	ChannelSlug     string
	CaseID          string
	SentimentStatus string
	SentimentTag    string
	Topic           string
}

func (c candidate) eligible() bool {
	if c.IsAdmin || !c.IsRoot || c.IsShare || c.IsCopy {
		return false
	}
	if c.Visibility != visibilityPublished {
		return false
	}
	if _, ok := sourceChannelSlugs[c.ChannelSlug]; !ok {
		return false
	}
	if !hasAllowedCaseID(c.CaseID) {
		return false
	}
	if c.SentimentStatus != sentimentScored || c.SentimentTag != string(scorer.SentimentPositive) {
		return false
	}
	_, ok := allowedTopics[c.Topic]
	return ok
}

func hasAllowedCaseID(caseID string) bool {
	for _, prefix := range allowedCaseIDPrefixes {
		if strings.HasPrefix(caseID, prefix) {
			return true
		}
	}
	return false
}
