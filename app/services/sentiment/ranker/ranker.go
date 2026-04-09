package ranker

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/services/sentiment/engagement"
	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/ent/predicate"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

type Ranker struct {
	db  *ent.Client
	bus *pubsub.Bus
}

func New(db *ent.Client, bus *pubsub.Bus) *Ranker {
	return &Ranker{db: db, bus: bus}
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
			ent_post_sentiment.FieldPostID,
			ent_post_sentiment.FieldSentimentTag,
			ent_post_sentiment.FieldPositivityScore,
			ent_post_sentiment.FieldPrimaryTopic,
		).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	postIDs := make([]xid.ID, 0, len(sentiments))
	for _, s := range sentiments {
		if s.SentimentTag != nil && s.PositivityScore != nil && s.PrimaryTopic != nil {
			postIDs = append(postIDs, s.PostID)
		}
	}

	engagementMap, err := engagement.GetBulk(ctx, r.db, postIDs)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	updated := 0
	for _, s := range sentiments {
		if s.SentimentTag == nil || s.PositivityScore == nil || s.PrimaryTopic == nil {
			continue
		}

		eng := engagementMap[s.PostID]
		rankScore := calculateRankScore(*s.SentimentTag, *s.PositivityScore, *s.PrimaryTopic, eng.Likes, eng.Replies)

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

type ScoreUnscoredParams struct {
	ChannelID     xid.ID
	IncludeFailed bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         *int
}

type ScoreUnscoredResult struct {
	PostsEnqueued int   `json:"posts_enqueued"`
	DurationMs    int64 `json:"duration_ms"`
}

func (r *Ranker) ScoreUnscored(ctx context.Context, params ScoreUnscoredParams) (*ScoreUnscoredResult, error) {
	start := time.Now()

	predicates := []predicate.Post{
		ent_post.DeletedAtIsNil(),
		ent_post.RootPostIDIsNil(),
		ent_post.VisibilityEQ(ent_post.VisibilityPublished),
		ent_post.ChannelIDEQ(params.ChannelID),
	}

	if params.CreatedAfter != nil {
		predicates = append(predicates, ent_post.CreatedAtGTE(*params.CreatedAfter))
	}
	if params.CreatedBefore != nil {
		predicates = append(predicates, ent_post.CreatedAtLTE(*params.CreatedBefore))
	}

	if params.IncludeFailed {
		predicates = append(predicates, ent_post.Not(ent_post.HasSentimentWith(
			ent_post_sentiment.ScoringStatusEQ(ent_post_sentiment.ScoringStatusScored),
		)))
	} else {
		predicates = append(predicates, ent_post.Not(ent_post.HasSentiment()))
	}

	query := r.db.Post.
		Query().
		Where(predicates...)

	if params.Limit != nil {
		query = query.Limit(*params.Limit)
	}

	posts, err := query.
		Select(ent_post.FieldID).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	for _, p := range posts {
		err := r.bus.SendCommand(ctx, &message.CommandScorePostSentiment{
			PostID: post.ID(p.ID),
		})
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	return &ScoreUnscoredResult{
		PostsEnqueued: len(posts),
		DurationMs:    time.Since(start).Milliseconds(),
	}, nil
}

func calculateRankScore(sentimentTag string, positivityScore int, primaryTopic string, likes, replies int) float64 {
	sentiment := scorer.SentimentTag(sentimentTag)
	sentimentWeight := sentiment.Weight()
	topicBooster := scorer.AllowedTopic(primaryTopic).Booster(sentiment)
	engagementBooster := CalculateEngagementBooster(likes, replies)
	return sentimentWeight + float64(positivityScore) + topicBooster + engagementBooster
}

func CalculateEngagementBooster(likes, replies int) float64 {
	likeBoost := float64(min(likes*2, 20))
	replyBoost := float64(min(replies, 10))
	return likeBoost + replyBoost
}
