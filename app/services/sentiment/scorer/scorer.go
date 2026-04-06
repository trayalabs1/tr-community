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
You are a sentiment analysis system for a hair loss treatment community forum. Analyze the following post and classify it.

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

Primary Topic - Choose the most relevant:
- "results": Progress updates, before/after, visible changes
- "product_review": Reviews or opinions about specific products
- "treatment_plan": Discussion of treatment regimens, protocols, routines
- "question": Asking for advice, help, or information
- "hair_loss": General discussion about hair loss causes, patterns
- "lifestyle": Diet, exercise, stress, sleep related to hair health
- "side_effects": Discussion of medication or treatment side effects
- "other": Does not fit other categories

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
	TopicHairLoss      AllowedTopic = "hair_loss"
	TopicProductReview AllowedTopic = "product_review"
	TopicTreatmentPlan AllowedTopic = "treatment_plan"
	TopicResults       AllowedTopic = "results"
	TopicSideEffects   AllowedTopic = "side_effects"
	TopicLifestyle     AllowedTopic = "lifestyle"
	TopicQuestion      AllowedTopic = "question"
	TopicOther         AllowedTopic = "other"
)

func (t AllowedTopic) Booster() float64 {
	switch t {
	case TopicResults:
		return 20
	case TopicProductReview:
		return 15
	case TopicTreatmentPlan:
		return 10
	case TopicQuestion:
		return 10
	case TopicHairLoss:
		return 5
	case TopicLifestyle:
		return 5
	case TopicSideEffects:
		return 0
	case TopicOther:
		return 0
	default:
		return 0
	}
}

type ScoringResult struct {
	SentimentTag    SentimentTag `json:"sentiment_tag" jsonschema:"enum=positive,enum=neutral,enum=negative,description=The overall sentiment of the post"`
	PositivityScore int          `json:"positivity_score" jsonschema:"minimum=0,maximum=100,description=A score from 0-100 indicating how positive the content is"`
	PrimaryTopic    AllowedTopic `json:"primary_topic" jsonschema:"enum=hair_loss,enum=product_review,enum=treatment_plan,enum=results,enum=side_effects,enum=lifestyle,enum=question,enum=other,description=The primary topic of the post"`
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
	return r.SentimentTag.Weight() + float64(r.PositivityScore) + r.PrimaryTopic.Booster()
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
