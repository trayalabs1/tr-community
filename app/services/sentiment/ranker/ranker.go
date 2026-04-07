package ranker

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

type Ranker struct {
	db *ent.Client
}

func New(db *ent.Client) *Ranker {
	return &Ranker{db: db}
}

type RecalculateResult struct {
	PostsUpdated int   `json:"posts_updated"`
	DurationMs   int64 `json:"duration_ms"`
}

func (r *Ranker) RecalculateBulk(ctx context.Context, channelID xid.ID) (*RecalculateResult, error) {
	start := time.Now()

	sentiments, err := r.db.PostSentiment.
		Query().
		Where(
			ent_post_sentiment.ScoringStatusEQ(ent_post_sentiment.ScoringStatusScored),
			ent_post_sentiment.HasPostWith(
				ent_post.ChannelIDEQ(channelID),
				ent_post.DeletedAtIsNil(),
			),
		).
		Select(
			ent_post_sentiment.FieldID,
			ent_post_sentiment.FieldSentimentTag,
			ent_post_sentiment.FieldPositivityScore,
			ent_post_sentiment.FieldPrimaryTopic,
		).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	updated := 0
	for _, s := range sentiments {
		if s.SentimentTag == nil || s.PositivityScore == nil || s.PrimaryTopic == nil {
			continue
		}

		rankScore := calculateRankScore(*s.SentimentTag, *s.PositivityScore, *s.PrimaryTopic)

		err := r.db.PostSentiment.
			UpdateOneID(s.ID).
			SetRankScore(rankScore).
			Exec(ctx)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
		updated++
	}

	return &RecalculateResult{
		PostsUpdated: updated,
		DurationMs:   time.Since(start).Milliseconds(),
	}, nil
}

func calculateRankScore(sentimentTag string, positivityScore int, primaryTopic string) float64 {
	sentiment := scorer.SentimentTag(sentimentTag)
	sentimentWeight := sentiment.Weight()
	topicBooster := scorer.AllowedTopic(primaryTopic).Booster(sentiment)
	return sentimentWeight + float64(positivityScore) + topicBooster
}
