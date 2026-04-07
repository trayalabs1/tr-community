package sentiment_job

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

type sentimentConsumer struct {
	db     *ent.Client
	scorer *scorer.Scorer
}

func newSentimentConsumer(
	db *ent.Client,
	scorer *scorer.Scorer,
) *sentimentConsumer {
	return &sentimentConsumer{
		db:     db,
		scorer: scorer,
	}
}

func (c *sentimentConsumer) scorePost(ctx context.Context, postID post.ID) error {
	p, err := c.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(postID))).
		Select(ent_post.FieldTitle, ent_post.FieldBody).
		Only(ctx)
	if err != nil {
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
		return fault.Wrap(err, fctx.With(ctx))
	}

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
