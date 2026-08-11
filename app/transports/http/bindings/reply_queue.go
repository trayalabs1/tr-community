package bindings

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/post/reply_admin_queue"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/admin/admin_reply_manager"
	"github.com/Southclaws/storyden/app/services/admin/reply_queue_manager"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type ReplyQueue struct {
	mgr          *reply_queue_manager.Manager
	adminReplies *admin_reply_manager.Manager
}

func NewReplyQueue(
	mgr *reply_queue_manager.Manager,
	adminReplies *admin_reply_manager.Manager,
) ReplyQueue {
	return ReplyQueue{mgr: mgr, adminReplies: adminReplies}
}

func (h *ReplyQueue) AdminAccountList(ctx context.Context, request openapi.AdminAccountListRequestObject) (openapi.AdminAccountListResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	admins, err := h.adminReplies.Admins(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.AdminAccountList200JSONResponse{
		AdminAccountListOKJSONResponse: openapi.AdminAccountListOKJSONResponse{
			Admins: dt.Map(admins, func(a *account.Account) openapi.ProfileReference {
				return serialiseProfileReferenceFromAccount(*a)
			}),
		},
	}, nil
}

func (h *ReplyQueue) AdminReplyList(ctx context.Context, request openapi.AdminReplyListRequestObject) (openapi.AdminReplyListResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	pp := deserialisePageParams(request.Params.Page, 50)

	repliedBy, err := opt.MapErr(opt.NewPtr(request.Params.RepliedBy), func(id openapi.Identifier) (account.AccountID, error) {
		parsed, err := xid.FromString(string(id))
		if err != nil {
			return account.AccountID{}, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
		}
		return account.AccountID(parsed), nil
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result, err := h.adminReplies.List(ctx, pp, admin_reply_manager.ListOpts{
		RepliedBy:     repliedBy,
		CreatedAfter:  opt.NewPtr(request.Params.CreatedAfter),
		CreatedBefore: opt.NewPtr(request.Params.CreatedBefore),
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.AdminReplyList200JSONResponse{
		AdminReplyListOKJSONResponse: openapi.AdminReplyListOKJSONResponse{
			Replies:     dt.Map(result.Items, serialiseReplyPtr),
			CurrentPage: result.CurrentPage,
			NextPage:    result.NextPage.Ptr(),
			TotalPages:  result.TotalPages,
			Results:     result.Results,
			PageSize:    result.Size,
		},
	}, nil
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
