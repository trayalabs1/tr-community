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
You are a sentiment, content-value, and category classification system for Traya, a hair loss treatment company's community forum.

Analyze the following post and return exactly four fields:

1. sentiment_tag
2. positivity_score
3. feed_value_score
4. category

==================================================
SENTIMENT TAG
==================================================

Pick exactly one:

- "positive": Posts expressing satisfaction, progress, hope, success stories, gratitude, or encouragement
- "neutral": Posts asking questions, sharing information objectively, or discussing without strong emotion
- "negative": Posts expressing frustration, disappointment, complaints, or dissatisfaction

If a sentence contains contradictory phrases, classify based on the overall intent.

Detect sarcasm. Phrases such as "what a joke", "yeah right", or positive words following a negative setup should be treated as negative when the overall intent is negative.

If a product is described as ineffective, misleading, a scam, or causing unintended results, classify sentiment as negative even if positive words are present.

Recognize common Hindi/Hinglish sentiment:
- "bakwas", "bekar", "faltu", "bekaar" → negative
- "mast", "badhiya", "zabardast" → positive
- "thik hai", "chalta hai" → neutral

==================================================
POSITIVITY SCORE (0-100)
==================================================

- 0-20: Very negative — complaints, anger, giving up, severe dissatisfaction
- 21-40: Somewhat negative — frustration, doubt, minor complaints, concern
- 41-60: Neutral — questions, factual information, balanced discussion
- 61-80: Somewhat positive — hope, mild progress, cautious optimism
- 81-100: Very positive — success stories, strong progress, enthusiasm, strong gratitude

==================================================
FEED VALUE SCORE (0-100)
==================================================

Measure how valuable the post is to the broader Traya community.

Consider BOTH:
- usefulness/information value to other users
- potential to generate meaningful discussion

Do NOT simply equate positivity with feed value.

0-20: Very low value
- Very vague
- Extremely short
- Repetitive
- Little or no useful information
- Little discussion potential

21-40: Low value
- Some context but limited usefulness
- Simple question with little detail
- Personal update with limited learning
- Low discussion potential

41-60: Moderate value
- Clear question or experience
- Some useful context
- Reasonably understandable
- Some potential to help or engage other users

61-80: High value
- Specific experience
- Useful details
- Relatable problem
- Clear question
- Helpful observation
- Strong potential for meaningful discussion

81-100: Very high value
- Highly useful to other users
- Detailed experience
- Actionable information
- Strongly relatable
- Generates meaningful discussion
- Helps users understand treatment, progress, expectations, or common problems
- Valuable success story or learning

A positive post is not automatically high-value.
A negative post is not automatically low-value.
A question can have high feed value if it contains useful context and is likely to generate meaningful discussion.
Do not use post length alone to determine feed value.

==================================================
CATEGORY
==================================================
If you are not highly confident, return "NA".

Do not classify from keywords alone.
If two categories are equally plausible, return "NA".
Questions are not a separate category; classify them according to their clearly identifiable subject.

Classify the post into ONE category based on its primary subject/intent. If unclear or multiple categories are equally plausible → "NA".

1. "RESULTS & PROGRESS"
   Meaningful treatment progress, improvement, results, or treatment journey.
   Examples: visible improvement, before/after, reduced hairfall as progress, baby hairs as progress, increased density, positive outcomes, meaningful milestones.
   Do not use: generic milestones, unclear results, primarily regrowth questions, product/treatment questions, scalp concerns.

2. "TIPS & EXPERIENCES"
   Useful personal tips, routines, recommendations, lessons, or experiences others can learn from.
   Examples: hair-care tips, practical recommendations, "what worked for me", useful habits.
   Do not use: pure progress updates, product-specific usage/experiences, diet-focused posts, generic opinions.

3. "HAIRFALL CONCERNS"
   Current or worsening hairfall, shedding, thinning, or concern about hair loss.
   Examples: increased hairfall, excessive shedding, ongoing hairfall, thinning, "is this normal?"
   Do not use: positive progress involving reduced hairfall, regrowth-focused posts, incidental hairfall mentions, primarily scalp concerns.

4. "HOW TO USE"
   How, when, how often, dosage, sequence, or manner of using a product/medicine/treatment.
   Examples: application, timing, frequency, dosage, before/after usage.
   Do not use: product opinions/effectiveness, side effects, generic hair-care tips.

5. "DANDRUFF & SCALP"
   Dandruff or scalp-related symptoms, conditions, or concerns.
   Examples: dandruff, flakes, dryness, itching, irritation, oily scalp, buildup, scalp health.
   Do not use: hairfall where scalp issues are incidental, product-focused questions, general hair-care advice.

6. "HAIR REGROWTH"
   Posts focused on new hair growth, regrowth, baby hairs, density, or expectations around growth.
   Examples: when regrowth starts, new hair growth, baby hairs, expected growth, increasing thickness.
   Do not use: overall progress stories, current hairfall concerns, incidental regrowth mentions.

7. "PRODUCTS & TREATMENT"
   Posts primarily discussing a product, medicine, treatment, or treatment protocol itself.
   Examples: product/treatment effectiveness, comparisons, treatment experiences, why a product is included, whether treatment is necessary.
   Do not use: usage questions → HOW TO USE; treatment-related reactions → SIDE EFFECTS; generic hairfall concerns.

8. "SIDE EFFECTS"
   Unwanted effects or symptoms believed to be caused by a treatment/product.
   Examples: acne, irritation, reactions, treatment-related symptoms, questions about side effects.
   Do not use: unrelated health problems, normal shedding unless framed as a side effect, usage questions.

9. "DIET & LIFESTYLE"
   Food, nutrition, exercise, stress, sleep, lifestyle, or daily habits related to hair/treatment.
   Examples: diet, protein, nutrition, exercise, stress, sleep, smoking/alcohol, lifestyle factors.
   Do not use: product/supplement-focused posts, general hair-care routines, incidental lifestyle mentions.

10. "CHALLENGES"
    Explicit Traya/community challenges, streaks, campaigns, or structured participation.
    Examples: challenge completion, streak milestones, campaign participation.
    Do not use: normal treatment milestones or progress.

11. "NA"
    Use when there is insufficient context or no clear category.
    Examples: "Good", "Nice", "Same here", "Thank you Traya", "3rd kit", "3 months completed", "Hairfall", unclear questions, generic complaints, delivery/order/refund/payment/support/appointment issues, irrelevant posts, ambiguous posts, or keyword-only matches.

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

type Category string

const (
	CategoryResultsProgress   Category = "RESULTS & PROGRESS"
	CategoryTipsExperiences   Category = "TIPS & EXPERIENCES"
	CategoryHairfallConcerns  Category = "HAIRFALL CONCERNS"
	CategoryHowToUse          Category = "HOW TO USE"
	CategoryDandruffScalp     Category = "DANDRUFF & SCALP"
	CategoryHairRegrowth      Category = "HAIR REGROWTH"
	CategoryProductsTreatment Category = "PRODUCTS & TREATMENT"
	CategorySideEffects       Category = "SIDE EFFECTS"
	CategoryDietLifestyle     Category = "DIET & LIFESTYLE"
	CategoryChallenges        Category = "CHALLENGES"
	CategoryNA                Category = "NA"
)

func (c Category) Boost(sentiment SentimentTag) float64 {
	if sentiment != SentimentPositive {
		return 0
	}
	switch c {
	case CategoryResultsProgress:
		return 100
	case CategoryTipsExperiences:
		return 75
	case CategoryHairfallConcerns:
		return 20
	case CategoryHairRegrowth:
		return 20
	case CategoryHowToUse:
		return 15
	case CategoryDandruffScalp:
		return 15
	case CategoryProductsTreatment:
		return 15
	case CategorySideEffects:
		return 15
	case CategoryDietLifestyle:
		return 10
	default:
		return 0
	}
}

type ScoringResult struct {
	SentimentTag    SentimentTag `json:"sentiment_tag" jsonschema:"enum=positive,enum=neutral,enum=negative,description=The overall sentiment of the post"`
	PositivityScore int          `json:"positivity_score" jsonschema:"minimum=0,maximum=100,description=A score from 0-100 indicating how positive the content is"`
	PrimaryTopic    Category     `json:"primary_topic" jsonschema:"enum=RESULTS & PROGRESS,enum=TIPS & EXPERIENCES,enum=HAIRFALL CONCERNS,enum=HOW TO USE,enum=DANDRUFF & SCALP,enum=HAIR REGROWTH,enum=PRODUCTS & TREATMENT,enum=SIDE EFFECTS,enum=DIET & LIFESTYLE,enum=CHALLENGES,enum=NA,description=The category of the post"`
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
	return float64(r.PositivityScore) + r.PrimaryTopic.Boost(r.SentimentTag)
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
