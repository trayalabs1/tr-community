package ranker

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/services/sentiment/postfilter"
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
		postfilter.NotPrescoredPost(),
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
