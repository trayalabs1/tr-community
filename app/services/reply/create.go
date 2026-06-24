package reply

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/reply_writer"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/resources/visibility"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/services/moderation/checker"
)

type BulkItem struct {
	ParentID post.ID
	Partial  Partial
}

// CreateMany creates replies across many threads using a single bulk INSERT at
// the DB layer. Unlike Create, this path does NOT run per-reply moderation or
// admin-queue branching — it's intended for admin-authored bulk replies, which
// publish as PUBLISHED directly. Cache invalidation and reply-created events
// (search indexing + notifications) are still emitted per reply.
func (s *Mutator) CreateMany(
	ctx context.Context,
	authorID account.AccountID,
	items []BulkItem,
) (int, error) {
	writerItems := make([]reply_writer.BulkItem, len(items))
	for i, item := range items {
		opts := item.Partial.Opts()
		opts = append(opts, reply_writer.WithVisibility(visibility.VisibilityPublished))
		writerItems[i] = reply_writer.BulkItem{
			ParentID: item.ParentID,
			Opts:     opts,
		}
	}

	created, err := s.replyWriter.CreateBulk(ctx, authorID, writerItems)
	if err != nil {
		return 0, fault.Wrap(err, fctx.With(ctx))
	}

	invalidated := map[post.ID]struct{}{}
	for _, c := range created {
		if _, ok := invalidated[c.ParentID]; !ok {
			invalidated[c.ParentID] = struct{}{}
			if err := s.cache.Invalidate(ctx, xid.ID(c.ParentID)); err != nil {
				return 0, fault.Wrap(err, fctx.With(ctx))
			}
		}

		s.bus.Publish(ctx, &message.EventThreadReplyCreated{
			ThreadID:       c.ParentID,
			ReplyID:        c.ReplyID,
			ThreadAuthorID: c.ThreadAuthorID,
			ReplyAuthorID:  authorID,
		})
	}

	return len(created), nil
}

func (s *Mutator) Create(
	ctx context.Context,
	authorID account.AccountID,
	parentID post.ID,
	partial Partial,
) (*reply.Reply, error) {
	opts := partial.Opts()
	opts = append(opts, reply_writer.WithVisibility(visibility.VisibilityPublished))

	p, err := s.replyWriter.Create(ctx, authorID, parentID, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to create reply post in thread"))
	}

	wasMovedToReview := false
	if content, ok := partial.Content.Get(); ok {
		result, err := s.cpm.CheckContent(ctx, xid.ID(p.ID), datagraph.KindReply, "", content)
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}

		if result.Action == checker.ActionReport {
			updatedReply, err := s.replyWriter.Update(ctx, p.ID, reply_writer.WithVisibility(visibility.VisibilityReview))
			if err != nil {
				return nil, fault.Wrap(err, fctx.With(ctx))
			}
			p = updatedReply
			wasMovedToReview = true
		}
	}

	pref, err := s.replyQuerier.Probe(ctx, p.ID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if err := s.cache.Invalidate(ctx, xid.ID(pref.RootPostID)); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	replyToAuthorID := opt.Map(p.ReplyTo, func(r reply.Reply) account.AccountID {
		return r.Author.ID
	})
	replyToReplyID := opt.Map(p.ReplyTo, func(r reply.Reply) post.ID {
		return r.ID
	})

	snippet := ""
	if content, ok := partial.Content.Get(); ok {
		runes := []rune(content.Plaintext())
		if len(runes) > 100 {
			runes = runes[:100]
		}
		snippet = string(runes)
	}

	// Only emit created event (which triggers indexing) if reply is published
	if !wasMovedToReview {
		s.bus.Publish(ctx, &message.EventThreadReplyCreated{
			ThreadID:        p.RootPostID,
			ReplyID:         p.ID,
			ThreadAuthorID:  p.RootAuthor.ID,
			ReplyAuthorID:   authorID,
			ReplyToAuthorID: replyToAuthorID,
			ReplyToTargetID: replyToReplyID,
		})

		if !session.GetRoles(ctx).Permissions().HasAny(rbac.PermissionAdministrator) {
			s.bus.Publish(ctx, &message.EventReplyRequiresAdminAttention{
				ThreadID: p.RootPostID,
				ReplyID:  p.ID,
				Snippet:  snippet,
			})
		} else {
			s.bus.Publish(ctx, &message.EventAdminReplied{
				AdminAccountID: authorID,
				AdminPostID:    p.ID,
				AdminPostTime:  p.CreatedAt,
				ParentPostID:   parentID,
			})
		}
	}

	return p, nil
}
