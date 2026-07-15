package bindings

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/thread_share"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

func (i *Threads) ThreadShareCreate(ctx context.Context, request openapi.ThreadShareCreateRequestObject) (openapi.ThreadShareCreateResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.PermissionDenied))
	}

	postID, err := i.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelIDs := dt.Map(request.Body.Channels, func(id openapi.Identifier) xid.ID {
		return openapi.ParseID(id)
	})

	subtitle := ""
	if request.Body.Subtitle != nil {
		subtitle = *request.Body.Subtitle
	}

	if _, err := i.thread_share_svc.Share(ctx, postID, channelIDs, subtitle); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	thr, err := i.thread_svc.Get(ctx, postID, deserialisePageParams(nil, reply.RepliesPerPage))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadShareCreate200JSONResponse{
		ThreadShareCreateOKJSONResponse: openapi.ThreadShareCreateOKJSONResponse(serialiseThread(thr)),
	}, nil
}

func (i *Threads) ThreadShareList(ctx context.Context, request openapi.ThreadShareListRequestObject) (openapi.ThreadShareListResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.PermissionDenied))
	}

	postID, err := i.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	shares, err := i.thread_share_svc.ListForThread(ctx, postID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadShareList200JSONResponse{
		ThreadShareListOKJSONResponse: openapi.ThreadShareListOKJSONResponse{
			Shares: dt.Map(shares, serialiseThreadShare),
		},
	}, nil
}

func (i *Threads) ThreadShareDelete(ctx context.Context, request openapi.ThreadShareDeleteRequestObject) (openapi.ThreadShareDeleteResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.PermissionDenied))
	}

	postID, err := i.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := openapi.ParseID(request.ChannelID)

	if err := i.thread_share_svc.Unshare(ctx, postID, channelID); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadShareDelete200Response{}, nil
}

func (c Channels) ThreadSharePin(ctx context.Context, request openapi.ThreadSharePinRequestObject) (openapi.ThreadSharePinResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.PermissionDenied))
	}

	postID, err := c.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := openapi.ParseID(request.ChannelID)

	share, err := c.thread_share_svc.SetPin(ctx, postID, channelID, request.Body.Pinned)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadSharePin200JSONResponse{
		ThreadSharePinOKJSONResponse: openapi.ThreadSharePinOKJSONResponse(serialiseThreadShare(share)),
	}, nil
}

func serialiseThreadShare(s *thread_share.Share) openapi.ThreadShare {
	var subtitle *string
	if s.Subtitle != "" {
		subtitle = &s.Subtitle
	}

	return openapi.ThreadShare{
		Channel:  serialiseChannelReference(s.Channel),
		Pinned:   s.Pinned,
		Subtitle: subtitle,
	}
}
