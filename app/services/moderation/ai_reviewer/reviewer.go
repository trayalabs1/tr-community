package ai_reviewer

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread_writer"
	"github.com/Southclaws/storyden/app/resources/visibility"
	"github.com/Southclaws/storyden/internal/infrastructure/ai"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

const (
	reviewModel   = "gpt-4.1-mini"
	reviewTimeout = 10 * time.Second
)

const systemPrompt = `You are a content moderator for Traya's in-app customer community. Traya is a personalized hair loss treatment brand combining Ayurveda, dermatology, and nutrition.

Classify each post as PASS or FAIL.

FAIL if the post does ANY of these:
- Expresses negative sentiment about Traya (calls it scam, fraud, fake, waste, worst, bakwas, bekar, faltu, etc.)
- Discourages others from buying/using Traya ("don't buy", "don't use", "mat karo", "stay away")
- Praises or recommends a competitor brand by name (e.g., Cure Skin, Man Matters, Vedix, any other brand)
- Claims Traya is dangerous, harmful, or causes health damage
- Demands refund in angry/accusatory tone
- Calls Traya's marketing fake or reviews paid/fraudulent
- Uses abusive language or personal attacks toward Traya staff

PASS if the post does ANY of these:
- Asks product usage questions (how/when to apply oil, minoxidil, shampoo, tablets, etc.)
- Asks about treatment routine, diet, or lifestyle during treatment
- Shares progress updates, even if results are slow (as long as not bashing Traya)
- Reports a side effect while genuinely seeking help (shedding, itching, dryness)
- Asks about order/delivery/kit logistics neutrally
- Shares motivation, tips, or positive experiences
- Asks general hair care questions
- Expresses concern about early hair fall increase while seeking reassurance (this is normal during treatment)

EDGE CASES — apply these rules:
- Damaged product complaints without bashing Traya → PASS
- "No results yet" framed as a question seeking help → PASS
- "No results" framed as a verdict condemning Traya → FAIL
- Neutral questions doubting efficacy ("does this really work?") → PASS
- Apply same rules regardless of language

Respond in this exact format only:
PASS or FAIL | <reason in under 10 words>

User post:
`

func Build() fx.Option {
	return fx.Invoke(runAIReviewConsumer)
}

type aiReviewer struct {
	logger       *slog.Logger
	prompter     ai.Prompter
	threadWriter *thread_writer.Writer
	bus          *pubsub.Bus
}

func runAIReviewConsumer(
	ctx context.Context,
	lc fx.Lifecycle,
	logger *slog.Logger,
	prompter ai.Prompter,
	threadWriter *thread_writer.Writer,
	bus *pubsub.Bus,
) {
	r := &aiReviewer{
		logger:       logger,
		prompter:     prompter,
		threadWriter: threadWriter,
		bus:          bus,
	}

	lc.Append(fx.StartHook(func(hctx context.Context) error {
		_, err := pubsub.Subscribe(ctx, bus, "ai_reviewer.review_thread", r.handleThreadSubmittedForReview)
		return err
	}))
}

func (r *aiReviewer) handleThreadSubmittedForReview(ctx context.Context, evt *message.EventThreadSubmittedForReview) error {
	reviewCtx, cancel := context.WithTimeout(ctx, reviewTimeout)
	defer cancel()

	prompt := systemPrompt + evt.Title + "\n" + evt.Body

	result, err := r.prompter.PromptWithModel(reviewCtx, reviewModel, prompt)
	if err != nil {
		r.logger.Error("ai review failed, leaving post in manual review",
			slog.String("thread_id", evt.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if result == nil {
		r.logger.Warn("ai reviewer disabled, leaving post in manual review",
			slog.String("thread_id", evt.ID.String()),
		)
		return nil
	}

	decision, reason := parseDecision(result.Answer)

	r.logger.Info("ai review completed",
		slog.String("thread_id", evt.ID.String()),
		slog.String("decision", decision),
		slog.String("reason", reason),
	)

	if decision != "PASS" {
		return nil
	}

	thr, err := r.threadWriter.Update(ctx, post.ID(evt.ID), thread_writer.WithVisibility(visibility.VisibilityPublished))
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	r.bus.Publish(ctx, &message.EventThreadPublished{ID: thr.ID})
	return nil
}

func parseDecision(response string) (decision, reason string) {
	trimmed := strings.TrimSpace(response)
	upper := strings.ToUpper(trimmed)

	if strings.HasPrefix(upper, "PASS") {
		parts := strings.SplitN(trimmed, "|", 2)
		if len(parts) == 2 {
			return "PASS", strings.TrimSpace(parts[1])
		}
		return "PASS", ""
	}

	if strings.HasPrefix(upper, "FAIL") {
		parts := strings.SplitN(trimmed, "|", 2)
		if len(parts) == 2 {
			return "FAIL", strings.TrimSpace(parts[1])
		}
		return "FAIL", ""
	}

	return "UNKNOWN", trimmed
}
