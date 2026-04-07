package scorer

import (
	"context"
	"html/template"
	"log/slog"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/internal/infrastructure/ai"
)

var SentimentPrompt = template.Must(template.New("sentiment").Parse(`
You are a sentiment analysis system for Traya, a hair loss treatment company's community forum. Analyze the following post and classify it.

CLASSIFICATION RULES:

Sentiment Tag:
- "positive": Posts expressing satisfaction, progress, hope, success stories, gratitude, or encouragement
- "neutral": Posts asking questions, sharing information objectively, or discussing without strong emotion
- "negative": Posts expressing frustration, disappointment, complaints, or dissatisfaction

Positivity Score (0-100):
- 0-20: Very negative (complaints, anger, giving up)
- 21-40: Somewhat negative (frustration, doubt, minor complaints)
- 41-60: Neutral (questions, factual information, balanced discussion)
- 61-80: Somewhat positive (hope, mild progress, cautious optimism)
- 81-100: Very positive (success stories, strong progress, enthusiasm)

Primary Topic - Pick exactly one from this list:
- "usage_regimen": How to use minoxidil, oil, shampoo, tablets, timing, frequency
- "results_timeline": When results will show, how long it takes
- "hairfall_shedding": Hair fall increase, shedding, breakage, regrowth concerns
- "side_effects_safety": Itching, dandruff, acne, nausea, allergy, safety concerns
- "order_delivery_refill": Order delay, missing item, cancel, refund, refill
- "build_a_habit": App logging, streak, coins, redeem, habit tracking
- "diet_lifestyle_compatibility": Diet plan, smoking, travel, pregnancy, periods, other meds
- "product_effectiveness_trust": No improvement, waste of money, scam, doubt on effectiveness
- "doctor_support_call": Missed calls, support complaints, doctor/coach issues
- "progress_success_story": Visible progress, reduced hair fall, gratitude, encouragement, excitement, hopeful
- "other": Greeting, unclear, untitled, irrelevant

IMPORTANT INSTRUCTIONS:
- If a sentence contains contradictory phrases (e.g., negative phrase + positive phrase), check the overall intent.
- Detect sarcasm: phrases like "what a joke", "yeah right", or statements where positive words follow a negative setup should be treated as negative.
- If a product is described as ineffective, misleading, or causing unintended results, classify as "negative" even if positive words are present.
- Prioritize the user's intent over literal wording.
- Apply the same rules regardless of language (English, Hindi, Hinglish, or any other language).
- Recognize common Hindi/Hinglish sentiment words: "bakwas/bekar/faltu/bekaar" (negative), "mast/badhiya/zabardast" (positive), "thik hai/chalta hai" (neutral).

POST CONTENT:

Title: {{ .Title }}

Body:
{{ .Body }}

Analyze this post and provide your classification.
`))

type SentimentTag string

const (
	SentimentPositive SentimentTag = "positive"
	SentimentNeutral  SentimentTag = "neutral"
	SentimentNegative SentimentTag = "negative"
)

func (s SentimentTag) Weight() float64 {
	switch s {
	case SentimentPositive:
		return 100
	case SentimentNeutral:
		return 50
	case SentimentNegative:
		return 0
	default:
		return 0
	}
}

type AllowedTopic string

const (
	TopicUsageRegimen              AllowedTopic = "usage_regimen"
	TopicResultsTimeline           AllowedTopic = "results_timeline"
	TopicHairfallShedding          AllowedTopic = "hairfall_shedding"
	TopicSideEffectsSafety         AllowedTopic = "side_effects_safety"
	TopicOrderDeliveryRefill       AllowedTopic = "order_delivery_refill"
	TopicBuildAHabit               AllowedTopic = "build_a_habit"
	TopicDietLifestyleCompat       AllowedTopic = "diet_lifestyle_compatibility"
	TopicProductEffectivenessTrust AllowedTopic = "product_effectiveness_trust"
	TopicDoctorSupportCall         AllowedTopic = "doctor_support_call"
	TopicProgressSuccessStory      AllowedTopic = "progress_success_story"
	TopicOther                     AllowedTopic = "other"
)

func (t AllowedTopic) Booster(sentiment SentimentTag) float64 {
	switch t {
	case TopicProgressSuccessStory:
		switch sentiment {
		case SentimentPositive:
			return 20
		case SentimentNeutral:
			return 5
		case SentimentNegative:
			return -20
		}
	case TopicProductEffectivenessTrust:
		switch sentiment {
		case SentimentPositive:
			return 10
		case SentimentNeutral:
			return 5
		case SentimentNegative:
			return -10
		}
	case TopicHairfallShedding:
		switch sentiment {
		case SentimentPositive:
			return 0
		case SentimentNeutral:
			return 0
		case SentimentNegative:
			return -5
		}
	case TopicSideEffectsSafety:
		switch sentiment {
		case SentimentPositive:
			return 0
		case SentimentNeutral:
			return 0
		case SentimentNegative:
			return -10
		}
	case TopicOrderDeliveryRefill:
		switch sentiment {
		case SentimentPositive:
			return 0
		case SentimentNeutral:
			return -5
		case SentimentNegative:
			return -10
		}
	case TopicBuildAHabit:
		switch sentiment {
		case SentimentPositive:
			return 15
		case SentimentNeutral:
			return 10
		case SentimentNegative:
			return 0
		}
	}
	return 0
}

type ScoringResult struct {
	SentimentTag    SentimentTag `json:"sentiment_tag" jsonschema:"enum=positive,enum=neutral,enum=negative,description=The overall sentiment of the post"`
	PositivityScore int          `json:"positivity_score" jsonschema:"minimum=0,maximum=100,description=A score from 0-100 indicating how positive the content is"`
	PrimaryTopic    AllowedTopic `json:"primary_topic" jsonschema:"enum=usage_regimen,enum=results_timeline,enum=hairfall_shedding,enum=side_effects_safety,enum=order_delivery_refill,enum=build_a_habit,enum=diet_lifestyle_compatibility,enum=product_effectiveness_trust,enum=doctor_support_call,enum=progress_success_story,enum=other,description=The primary topic of the post"`
}

func (r *ScoringResult) Validate() {
	if r.PositivityScore < 0 {
		r.PositivityScore = 0
	}
	if r.PositivityScore > 100 {
		r.PositivityScore = 100
	}
}

func (r *ScoringResult) CalculateRankScore() float64 {
	return r.SentimentTag.Weight() + float64(r.PositivityScore) + r.PrimaryTopic.Booster(r.SentimentTag)
}

type Scorer struct {
	logger   *slog.Logger
	prompter ai.Prompter
}

func New(logger *slog.Logger, prompter ai.Prompter) *Scorer {
	return &Scorer{
		logger:   logger,
		prompter: prompter,
	}
}

type ScoreInput struct {
	Title string
	Body  string
}

func (s *Scorer) Score(ctx context.Context, input ScoreInput) (*ScoringResult, error) {
	if s.prompter == nil {
		return nil, fault.New("AI prompter not configured")
	}

	prompt := strings.Builder{}
	err := SentimentPrompt.Execute(&prompt, map[string]any{
		"Title": input.Title,
		"Body":  input.Body,
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	promptStr := prompt.String()

	s.logger.Info("sentiment scoring request",
		slog.String("title", input.Title),
		slog.String("prompt", promptStr),
	)

	result, err := ai.PromptObject(
		ctx,
		s.prompter,
		"Sentiment analysis result for a community forum post",
		promptStr,
		ScoringResult{},
	)
	if err != nil {
		s.logger.Error("sentiment scoring failed",
			slog.String("title", input.Title),
			slog.String("error", err.Error()),
		)
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result.Validate()

	s.logger.Info("sentiment scoring response",
		slog.String("title", input.Title),
		slog.String("sentiment_tag", string(result.SentimentTag)),
		slog.Int("positivity_score", result.PositivityScore),
		slog.String("primary_topic", string(result.PrimaryTopic)),
		slog.Float64("rank_score", result.CalculateRankScore()),
	)

	return result, nil
}
