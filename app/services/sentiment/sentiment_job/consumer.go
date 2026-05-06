package sentiment_job

import (
	"context"
	"log/slog"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"
	"golang.org/x/sync/errgroup"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

type sentimentConsumer struct {
	logger *slog.Logger
	db     *ent.Client
	scorer *scorer.Scorer
}

func newSentimentConsumer(
	logger *slog.Logger,
	db *ent.Client,
	scorer *scorer.Scorer,
) *sentimentConsumer {
	return &sentimentConsumer{
		logger: logger,
		db:     db,
		scorer: scorer,
	}
}

func (c *sentimentConsumer) scorePost(ctx context.Context, postID post.ID) error {
	c.logger.Info("ai scoring: scorePost started",
		slog.String("post_id", postID.String()),
	)

	p, err := c.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(postID))).
		Select(ent_post.FieldTitle, ent_post.FieldBody).
		Only(ctx)
	if err != nil {
		c.logger.Error("ai scoring: failed to load post for scoring",
			slog.String("post_id", postID.String()),
			slog.String("error", err.Error()),
		)
		return fault.Wrap(err, fctx.With(ctx))
	}

	title := ""
	if p.Title != "" {
		title = p.Title
	}

	result, err := c.scorer.Score(ctx, scorer.ScoreInput{
		Title: title,
		Body:  p.Body,
	})
	if err != nil {
		if markErr := c.markFailed(ctx, postID); markErr != nil {
			return fault.Wrap(markErr, fctx.With(ctx))
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	rankScore := result.CalculateRankScore()

	err = c.db.PostSentiment.
		Create().
		SetPostID(xid.ID(postID)).
		SetSentimentTag(string(result.SentimentTag)).
		SetPositivityScore(result.PositivityScore).
		SetPrimaryTopic(string(result.PrimaryTopic)).
		SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
		SetRankScore(rankScore).
		OnConflictColumns(ent_post_sentiment.FieldPostID).
		UpdateNewValues().
		Exec(ctx)
	if err != nil {
		c.logger.Error("ai scoring: failed to persist sentiment",
			slog.String("post_id", postID.String()),
			slog.String("error", err.Error()),
		)
		return fault.Wrap(err, fctx.With(ctx))
	}

	c.logger.Info("ai scoring: sentiment persisted",
		slog.String("post_id", postID.String()),
		slog.String("sentiment_tag", string(result.SentimentTag)),
		slog.Int("positivity_score", result.PositivityScore),
		slog.Float64("rank_score", rankScore),
	)

	return nil
}

func (c *sentimentConsumer) markFailed(ctx context.Context, postID post.ID) error {
	return c.db.PostSentiment.
		Create().
		SetPostID(xid.ID(postID)).
		SetScoringStatus(ent_post_sentiment.ScoringStatusFailed).
		OnConflictColumns(ent_post_sentiment.FieldPostID).
		UpdateScoringStatus().
		Exec(ctx)
}

func (c *sentimentConsumer) scoreBatch(ctx context.Context, postIDs []post.ID) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for _, id := range postIDs {
		g.Go(func() error {
			if err := c.scorePost(ctx, id); err != nil {
				c.logger.Error("failed to score post in batch",
					slog.String("post_id", id.String()),
					slog.String("error", err.Error()),
				)
			}
			return nil
		})
	}

	g.Wait()
}
