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
