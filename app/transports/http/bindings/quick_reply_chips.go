package bindings

import (
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

var progressSuccessStoryChips = []string{
	"Amazing progress 🙌",
	"So happy for you",
	"Keep going 💪",
}

var buildAHabitChips = []string{
	"Great consistency 🔥",
	"Keep the streak going",
	"You're doing great 💪",
}

func quickReplyChipsFor(t *thread.Thread) *openapi.QuickReplyChips {
	// build_a_habit: identified by meta.post_category == "BAH"; no sentiment gate.
	if cat, _ := t.Meta["post_category"].(string); cat == "BAH" {
		return cloneChips(buildAHabitChips)
	}

	// progress_success_story: AI-classified, gated on positive sentiment.
	tag, ok := t.SentimentTag.Get()
	if !ok || tag != "positive" {
		return nil
	}
	topic, ok := t.PrimaryTopic.Get()
	if !ok || topic != "progress_success_story" {
		return nil
	}
	return cloneChips(progressSuccessStoryChips)
}

func cloneChips(in []string) *openapi.QuickReplyChips {
	out := make([]string, len(in))
	copy(out, in)
	return &openapi.QuickReplyChips{Candidates: out}
}
