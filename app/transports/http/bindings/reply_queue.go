package bindings

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/admin/reply_queue_manager"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type ReplyQueue struct {
	mgr *reply_queue_manager.Manager
}

func NewReplyQueue(mgr *reply_queue_manager.Manager) ReplyQueue {
	return ReplyQueue{mgr: mgr}
}

func (h *ReplyQueue) AdminReplyQueueList(ctx context.Context, request openapi.AdminReplyQueueListRequestObject) (openapi.AdminReplyQueueListResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	pp := deserialisePageParams(request.Params.Page, 50)

	opts := reply_queue_manager.ListOpts{
		CreatedAfter:  opt.NewPtr(request.Params.CreatedAfter),
		CreatedBefore: opt.NewPtr(request.Params.CreatedBefore),
	}

	result, err := h.mgr.List(ctx, pp, opts)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	items := dt.Map(result.Items, serialiseReplyQueueEntry)

	return openapi.AdminReplyQueueList200JSONResponse{
		AdminReplyQueueListOKJSONResponse: openapi.AdminReplyQueueListOKJSONResponse{
			ReplyQueueEntries: items,
			CurrentPage:       result.CurrentPage,
			NextPage:          result.NextPage.Ptr(),
			TotalPages:        result.TotalPages,
			Results:           result.Results,
			PageSize:          result.Size,
		},
	}, nil
}

func (h *ReplyQueue) AdminReplyQueueDismiss(ctx context.Context, request openapi.AdminReplyQueueDismissRequestObject) (openapi.AdminReplyQueueDismissResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	id, err := xid.FromString(request.ReplyQueueId)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if err := h.mgr.Dismiss(ctx, reply_admin_queue.ID(id)); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.AdminReplyQueueDismiss204Response{}, nil
}

func serialiseReplyQueueEntry(e *reply_admin_queue.Entry) openapi.ReplyQueueEntry {
	return openapi.ReplyQueueEntry{
		Id:             e.ID.String(),
		ReplyId:        e.ReplyID.String(),
		ThreadId:       e.ThreadID.String(),
		ChannelId:      e.ChannelID.String(),
		ContentSnippet: e.ContentSnippet,
		CreatedAt:      e.CreatedAt,
	}
}
