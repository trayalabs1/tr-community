package thread

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"strconv"
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
const feedbackCooldownWindow = 12 * time.Hour

const (
	prescoredSentimentTag      = "neutral"
	prescoredRankScoreMin      = 95
	prescoredRankScoreRangeLen = 11 // 95..105 inclusive => 11 values
)

// isPrescoredCategory reports whether a post category bypasses AI sentiment
// scoring and instead receives a fixed neutral tag + 95–105 rank score at
// creation. New prescored categories should be added here and to the
// rehydrator/ranker/AI-reviewer filters.
func isPrescoredCategory(category string) bool {
	switch category {
	case "BAH", "feedback":
		return true
	}
	return false
}

// metaTypeString normalises meta["type"] to the text form Postgres yields from
// metadata->>'type'. BAH sends the streak count as a JSON number (float64 after
// unmarshal); feedback sends a string.
func metaTypeString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return ""
	}
}

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

	postCategoryAtCreate, _ := meta["post_category"].(string)

	if isPrescoredCategory(postCategoryAtCreate) {
		if err := s.assignPrescoredSentiment(ctx, thr.ID); err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	if postCategoryAtCreate == "BAH" &&
		(thr.Visibility == visibility.VisibilityPublished || thr.Visibility == visibility.VisibilityReview) {
		recent, err := s.threadQuerier.HasRecentChannelPrescored(
			ctx,
			thr.ChannelID,
			time.Now().Add(-bahCooldownWindow),
			xid.ID(thr.ID),
			"BAH",
			metaTypeString(meta["type"]),
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

	if postCategoryAtCreate == "feedback" &&
		(thr.Visibility == visibility.VisibilityPublished || thr.Visibility == visibility.VisibilityReview) {
		postType, _ := meta["type"].(string)
		recent, err := s.threadQuerier.HasRecentChannelPrescored(
			ctx,
			thr.ChannelID,
			time.Now().Add(-feedbackCooldownWindow),
			xid.ID(thr.ID),
			"feedback",
			postType,
		)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
		s.logger.Info("feedback cooldown check",
			slog.String("thread_id", thr.ID.String()),
			slog.String("channel_id", thr.ChannelID.String()),
			slog.String("post_type", postType),
			slog.String("visibility_before", thr.Visibility.String()),
			slog.Bool("recent", recent),
		)

		switch {
		case thr.Visibility == visibility.VisibilityPublished && recent:
			// Demote to review when another feedback/<type> is within window.
			// Skip EventThreadSubmittedForReview — feedback is prescored at
			// creation and must not be pushed to the AI reviewer.
			thr, err = s.threadWriter.Update(ctx, thr.ID, thread_writer.WithVisibility(visibility.VisibilityReview))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}

		case thr.Visibility == visibility.VisibilityReview && !recent:
			// First feedback/<type> in window — auto-promote to published.
			thr, err = s.threadWriter.Update(ctx, thr.ID, thread_writer.WithVisibility(visibility.VisibilityPublished))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}

		case thr.Visibility == visibility.VisibilityReview && recent:
			// Within window — leave in review. No event, no promotion.
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

		if thr.Visibility == visibility.VisibilityReview && !isPrescoredCategory(postCategoryAtCreate) {
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

	if thr.Visibility == visibility.VisibilityPublished {
		s.bus.Publish(ctx, &message.EventThreadPublished{
			ID: thr.ID,
		})
	}

	if !isPrescoredCategory(postCategoryAtCreate) {
		if err := s.bus.SendCommand(ctx, &message.CommandScorePostSentiment{
			PostID: thr.ID,
		}); err != nil {
			s.logger.Error("failed to dispatch CommandScorePostSentiment from thread create",
				slog.String("post_id", thr.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	// TODO: Do this using event consumer.
	s.mentioner.Send(ctx, authorID, *datagraph.NewRef(thr), thr.Content.References()...)

	return thr, nil
}

func (s *service) assignPrescoredSentiment(ctx context.Context, postID post.ID) error {
	rankScore := float64(rand.IntN(prescoredRankScoreRangeLen) + prescoredRankScoreMin)

	err := s.db.PostSentiment.
		Create().
		SetPostID(xid.ID(postID)).
		SetSentimentTag(prescoredSentimentTag).
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
