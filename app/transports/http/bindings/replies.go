package bindings

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"

	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/services/reply"
	reply_service "github.com/Southclaws/storyden/app/services/reply"
	"github.com/Southclaws/storyden/app/services/thread_mark"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type Replies struct {
	replyMutator    *reply.Mutator
	thread_mark_svc thread_mark.Service
}

func NewReplies(
	replyMutator *reply.Mutator,
	thread_mark_svc thread_mark.Service,
) Replies {
	return Replies{
		replyMutator:    replyMutator,
		thread_mark_svc: thread_mark_svc,
	}
}

func (p *Replies) ReplyCreate(ctx context.Context, request openapi.ReplyCreateRequestObject) (openapi.ReplyCreateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	postID, err := p.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	richContent, err := datagraph.NewRichText(request.Body.Body)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	partial := reply_service.Partial{
		Content: opt.New(richContent),
		ReplyTo: opt.Map(opt.NewPtr(request.Body.ReplyTo), deserialisePostID),
		Meta:    opt.NewPtr((*map[string]any)(request.Body.Meta)),
	}

	post, err := p.replyMutator.Create(ctx,
		accountID,
		postID,
		partial,
	)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ReplyCreate200JSONResponse{
		ReplyCreateOKJSONResponse: openapi.ReplyCreateOKJSONResponse(serialiseReplyPtr(post)),
	}, nil
}

func (p *Replies) ReplyCreateMany(ctx context.Context, request openapi.ReplyCreateManyRequestObject) (openapi.ReplyCreateManyResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if request.Body == nil {
		return openapi.ReplyCreateMany200JSONResponse{
			ReplyCreateManyOKJSONResponse: openapi.ReplyCreateManyOKJSONResponse{Requested: 0, Created: 0},
		}, nil
	}

	requested := len(request.Body.Items)
	items := make([]reply_service.BulkItem, 0, requested)
	for _, item := range request.Body.Items {
		postID, err := p.thread_mark_svc.Lookup(ctx, string(item.ThreadMark))
		if err != nil {
			continue
		}

		richContent, err := datagraph.NewRichText(item.Body)
		if err != nil {
			continue
		}

		items = append(items, reply_service.BulkItem{
			ParentID: postID,
			Partial:  reply_service.Partial{Content: opt.New(richContent)},
		})
	}

	created, err := p.replyMutator.CreateMany(ctx, accountID, items)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ReplyCreateMany200JSONResponse{
		ReplyCreateManyOKJSONResponse: openapi.ReplyCreateManyOKJSONResponse{
			Requested: requested,
			Created:   created,
		},
	}, nil
}
