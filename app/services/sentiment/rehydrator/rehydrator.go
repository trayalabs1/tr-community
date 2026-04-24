package rehydrator

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/services/sentiment/postfilter"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func Build() fx.Option {
	return fx.Invoke(runRehydrator)
}

func runRehydrator(
	lc fx.Lifecycle,
	logger *slog.Logger,
	db *ent.Client,
	bus *pubsub.Bus,
) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		posts, err := db.Post.
			Query().
			Where(
				ent_post.DeletedAtIsNil(),
				ent_post.RootPostIDIsNil(),
				ent_post.VisibilityEQ(ent_post.VisibilityPublished),
				ent_post.Not(ent_post.HasSentimentWith(
					ent_post_sentiment.ScoringStatusEQ(ent_post_sentiment.ScoringStatusScored),
				)),
				postfilter.NotBAHPost(),
			).
			Select(ent_post.FieldID).
			All(ctx)
		if err != nil {
			logger.Error("failed to query unscored posts for rehydration",
				slog.String("error", err.Error()),
			)
			return nil
		}

		if len(posts) == 0 {
			return nil
		}

		logger.Info("rehydrating unscored posts",
			slog.Int("count", len(posts)),
		)

		for _, p := range posts {
			err := bus.SendCommand(ctx, &message.CommandScorePostSentiment{
				PostID: post.ID(p.ID),
			})
			if err != nil {
				logger.Error("failed to enqueue post for scoring during rehydration",
					slog.String("post_id", p.ID.String()),
					slog.String("error", err.Error()),
				)
			}
		}

		return nil
	}))
}
