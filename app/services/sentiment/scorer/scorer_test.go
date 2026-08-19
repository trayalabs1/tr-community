package scorer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategoryBoost(t *testing.T) {
	tests := []struct {
		name      string
		category  Category
		sentiment SentimentTag
		want      float64
	}{
		{"results_progress positive", CategoryResultsProgress, SentimentPositive, 100},
		{"tips_experiences positive", CategoryTipsExperiences, SentimentPositive, 75},
		{"hairfall_concerns positive", CategoryHairfallConcerns, SentimentPositive, 20},
		{"hair_regrowth positive", CategoryHairRegrowth, SentimentPositive, 20},
		{"how_to_use positive", CategoryHowToUse, SentimentPositive, 15},
		{"dandruff_scalp positive", CategoryDandruffScalp, SentimentPositive, 15},
		{"products_treatment positive", CategoryProductsTreatment, SentimentPositive, 15},
		{"side_effects positive", CategorySideEffects, SentimentPositive, 15},
		{"diet_lifestyle positive", CategoryDietLifestyle, SentimentPositive, 10},
		{"challenges positive", CategoryChallenges, SentimentPositive, 0},
		{"na positive", CategoryNA, SentimentPositive, 0},
		{"results_progress neutral", CategoryResultsProgress, SentimentNeutral, 0},
		{"results_progress negative", CategoryResultsProgress, SentimentNegative, 0},
		{"unknown category positive", Category("bogus"), SentimentPositive, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.category.Boost(tt.sentiment))
		})
	}
}

func TestQualityScore(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   float64
	}{
		{"empty", 0, 25},
		{"exactly 100", 100, 25},
		{"101 chars", 101, 50},
		{"exactly 300", 300, 50},
		{"301 chars", 301, 75},
		{"exactly 500", 500, 75},
		{"501 chars", 501, 100},
		{"very long", 5000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, qualityScore(tt.length))
		})
	}
}

func TestSentimentMultiplier(t *testing.T) {
	tests := []struct {
		name string
		tag  SentimentTag
		want float64
	}{
		{"negative", SentimentNegative, 0.1},
		{"neutral", SentimentNeutral, 1.0},
		{"positive", SentimentPositive, 1.0},
		{"unknown", SentimentTag("bogus"), 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SentimentMultiplier(tt.tag))
		})
	}
}

func TestScoringResultCalculateRankScore(t *testing.T) {
	tests := []struct {
		name       string
		result     ScoringResult
		bodyLength int
		want       float64
	}{
		{
			name: "positive results_progress, short body",
			result: ScoringResult{
				SentimentTag:    SentimentPositive,
				PositivityScore: 90,
				FeedValueScore:  80,
				Category:        CategoryResultsProgress,
			},
			bodyLength: 50,
			// 1.0*90 + 1.5*80 + 1.0*25(quality) + 1.0*100(category boost) = 90+120+25+100 = 335
			want: 335,
		},
		{
			name: "negative hairfall_concerns, long body, no boost applies",
			result: ScoringResult{
				SentimentTag:    SentimentNegative,
				PositivityScore: 10,
				FeedValueScore:  60,
				Category:        CategoryHairfallConcerns,
			},
			bodyLength: 600,
			// 1.0*10 + 1.5*60 + 1.0*100(quality) + 1.0*0(boost, non-positive) = 10+90+100+0 = 200
			want: 200,
		},
		{
			name: "neutral NA category",
			result: ScoringResult{
				SentimentTag:    SentimentNeutral,
				PositivityScore: 50,
				FeedValueScore:  40,
				Category:        CategoryNA,
			},
			bodyLength: 200,
			// 1.0*50 + 1.5*40 + 1.0*50(quality) + 1.0*0(boost) = 50+60+50+0 = 160
			want: 160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.CalculateRankScore(tt.bodyLength))
		})
	}
}
