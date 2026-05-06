package thread

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/resources/post/thread_writer"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/resources/tag/tag_ref"
	"github.com/Southclaws/storyden/app/resources/visibility"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/services/link/fetcher"
	"github.com/Southclaws/storyden/app/services/moderation/checker"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
)

const bahCooldownWindow = 12 * time.Hour

const (
	bahSentimentTag      = "neutral"
	bahRankScoreMin      = 95
	bahRankScoreRangeLen = 11 // 95..105 inclusive => 11 values
)

func (s *service) Create(ctx context.Context,
	title string,
	authorID account.AccountID,
	meta map[string]any,
	partial Partial,
) (*thread.Thread, error) {
	if err := authoriseMutation(ctx, partial); err != nil {
		return nil, err
	}

	opts := partial.Opts()
	opts = append(opts,
		thread_writer.WithMeta(meta),
	)

	// Small hack: default to zero-value of content, which is actually not zero
	// it's <body></body>. Why? who knows... oh, me, yes I should know. I don't.
	if !partial.Content.Ok() {
		c, _ := datagraph.NewRichText("")
		opts = append(opts, thread_writer.WithContent(c))
	}

	if u, ok := partial.URL.Get(); ok {
		ln, err := s.fetcher.Fetch(ctx, u, fetcher.Options{})
		if err == nil {
			opts = append(opts, thread_writer.WithLink(xid.ID(ln.ID)))
		}
	}

	if tags, ok := partial.Tags.Get(); ok {
		newTags, err := s.tagWriter.Add(ctx, tags...)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}

		tagIDs := dt.Map(newTags, func(t *tag_ref.Tag) tag_ref.ID { return t.ID })

		opts = append(opts, thread_writer.WithTagsAdd(tagIDs...))
	}

	thr, err := s.threadWriter.Create(ctx,
		title,
		authorID,
		opts...,
	)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to create thread"))
	}

	if meta["post_category"] == "BAH" {
		if err := s.assignBAHSentiment(ctx, thr.ID); err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	if meta["post_category"] == "BAH" &&
		(thr.Visibility == visibility.VisibilityPublished || thr.Visibility == visibility.VisibilityReview) {
		recent, err := s.threadQuerier.HasRecentChannelBAH(
			ctx,
			thr.ChannelID,
			time.Now().Add(-bahCooldownWindow),
			xid.ID(thr.ID),
		)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}

		switch {
		case thr.Visibility == visibility.VisibilityPublished && recent:
			thr, err = s.threadWriter.Update(ctx, thr.ID, thread_writer.WithVisibility(visibility.VisibilityReview))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}

			s.bus.Publish(ctx, &message.EventThreadSubmittedForReview{
				ID:    thr.ID,
				Title: thr.Title,
				Body:  thr.Content.Plaintext(),
			})

		case thr.Visibility == visibility.VisibilityReview && !recent:
			thr, err = s.threadWriter.Update(ctx, thr.ID, thread_writer.WithVisibility(visibility.VisibilityPublished))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}
		}
	}

	if content, ok := partial.Content.Get(); ok {
		result, err := s.cpm.CheckContent(ctx, xid.ID(thr.ID), datagraph.KindThread, title, content)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}

		if result.Action == checker.ActionReport {
			thr, err = s.threadWriter.Update(ctx, thr.ID, thread_writer.WithVisibility(visibility.VisibilityReview))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}
		}

		if thr.Visibility == visibility.VisibilityReview {
			s.bus.Publish(ctx, &message.EventThreadSubmittedForReview{
				ID:    thr.ID,
				Title: thr.Title,
				Body:  thr.Content.Plaintext(),
			})
		}
	}

	if err := s.cache.Invalidate(ctx, xid.ID(thr.ID)); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	postCategory, _ := meta["post_category"].(string)

	if thr.Visibility == visibility.VisibilityPublished {
		s.bus.Publish(ctx, &message.EventThreadPublished{
			ID: thr.ID,
		})

		if postCategory != "BAH" {
			s.logger.Info("ai scoring: dispatching CommandScorePostSentiment from thread create",
				slog.String("post_id", thr.ID.String()),
				slog.String("post_category", postCategory),
				slog.String("visibility", thr.Visibility.String()),
			)
			if err := s.bus.SendCommand(ctx, &message.CommandScorePostSentiment{
				PostID: thr.ID,
			}); err != nil {
				s.logger.Error("ai scoring: failed to dispatch CommandScorePostSentiment from thread create",
					slog.String("post_id", thr.ID.String()),
					slog.String("error", err.Error()),
				)
			}
		} else {
			s.logger.Info("ai scoring: skipped for BAH post on create",
				slog.String("post_id", thr.ID.String()),
			)
		}
	} else {
		s.logger.Info("ai scoring: skipped, thread not published on create",
			slog.String("post_id", thr.ID.String()),
			slog.String("visibility", thr.Visibility.String()),
			slog.String("post_category", postCategory),
		)
	}

	// TODO: Do this using event consumer.
	s.mentioner.Send(ctx, authorID, *datagraph.NewRef(thr), thr.Content.References()...)

	return thr, nil
}

func (s *service) assignBAHSentiment(ctx context.Context, postID post.ID) error {
	rankScore := float64(rand.IntN(bahRankScoreRangeLen) + bahRankScoreMin)

	err := s.db.PostSentiment.
		Create().
		SetPostID(xid.ID(postID)).
		SetSentimentTag(bahSentimentTag).
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

func authoriseMutation(ctx context.Context, partial Partial) error {
	if partial.Pinned.Ok() {
		err := session.Authorise(ctx, nil, rbac.PermissionManagePosts)
		if err != nil {
			return fault.Wrap(err,
				fctx.With(ctx),
				fmsg.WithDesc("pinned state", "You do not have permission to create a pinned thread."),
			)
		}
	}

	return nil
}
