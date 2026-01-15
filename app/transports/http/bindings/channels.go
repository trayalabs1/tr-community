package bindings

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/cachecontrol"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/channel_membership"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/post/category"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/thread_cache"
	"github.com/Southclaws/storyden/app/resources/profile/profile_querier"
	"github.com/Southclaws/storyden/app/resources/tag/tag_ref"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	category_svc "github.com/Southclaws/storyden/app/services/category"
	channel_svc "github.com/Southclaws/storyden/app/services/channel"
	membership_svc "github.com/Southclaws/storyden/app/services/channel_membership"
	reply_svc "github.com/Southclaws/storyden/app/services/reply"
	"github.com/Southclaws/storyden/app/services/reqinfo"
	thread_svc "github.com/Southclaws/storyden/app/services/thread"
	"github.com/Southclaws/storyden/app/services/thread_mark"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/deletable"
)

type Channels struct {
	channel_svc     channel_svc.Service
	membership_svc  membership_svc.Service
	category_svc    category_svc.Service
	category_repo   *category.Repository
	thread_cache    *thread_cache.Cache
	thread_svc      thread_svc.Service
	thread_mark_svc thread_mark.Service
	reply_svc       *reply_svc.Mutator
	accountQuery    *account_querier.Querier
	profileQuery    *profile_querier.Querier
}

func NewChannels(
	channel_svc channel_svc.Service,
	membership_svc membership_svc.Service,
	category_svc category_svc.Service,
	category_repo *category.Repository,
	thread_cache *thread_cache.Cache,
	thread_svc thread_svc.Service,
	thread_mark_svc thread_mark.Service,
	reply_svc *reply_svc.Mutator,
	accountQuery *account_querier.Querier,
	profileQuery *profile_querier.Querier,
) Channels {
	return Channels{
		channel_svc:     channel_svc,
		membership_svc:  membership_svc,
		category_svc:    category_svc,
		category_repo:   category_repo,
		thread_cache:    thread_cache,
		thread_svc:      thread_svc,
		thread_mark_svc: thread_mark_svc,
		reply_svc:       reply_svc,
		accountQuery:    accountQuery,
		profileQuery:    profileQuery,
	}
}

func (c Channels) ChannelList(ctx context.Context, request openapi.ChannelListRequestObject) (openapi.ChannelListResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channels, err := c.channel_svc.GetUserChannels(ctx, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelList200JSONResponse{
		ChannelListOKJSONResponse: openapi.ChannelListOKJSONResponse{
			Channels: dt.Map(channels, serialiseChannel),
		},
	}, nil
}

func serialiseChannel(c *channel.Channel) openapi.Channel {
	ch := openapi.Channel{
		Id:          openapi.Identifier(c.ID.String()),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Visibility:  openapi.ChannelVisibility(c.Visibility),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if coverImage, ok := c.CoverImage.Get(); ok {
		asset := serialiseAsset(coverImage)
		ch.CoverImage = &asset
	}

	if icon, ok := c.Icon.Get(); ok {
		asset := serialiseAsset(icon)
		ch.Icon = &asset
	}

	if c.Metadata != nil {
		metadata := openapi.Metadata(c.Metadata)
		ch.Meta = &metadata
	}

	return ch
}

func (c Channels) ChannelCreate(ctx context.Context, request openapi.ChannelCreateRequestObject) (openapi.ChannelCreateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	visibility := channel.VisibilityPublic
	if request.Body.Visibility != nil {
		visibility = channel.Visibility(*request.Body.Visibility)
	}

	partial := channel_svc.Partial{
		Name:        opt.New(request.Body.Name),
		Slug:        opt.New(request.Body.Slug),
		Description: opt.New(request.Body.Description),
		Visibility:  opt.New(visibility),
	}

	ch, err := c.channel_svc.Create(ctx, accountID, partial)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCreate200JSONResponse{
		ChannelGetOKJSONResponse: openapi.ChannelGetOKJSONResponse(serialiseChannel(ch)),
	}, nil
}

func (c Channels) ChannelGet(ctx context.Context, request openapi.ChannelGetRequestObject) (openapi.ChannelGetResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	ch, err := c.channel_svc.Get(ctx, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Check if user has access to this channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	return openapi.ChannelGet200JSONResponse{
		ChannelGetOKJSONResponse: openapi.ChannelGetOKJSONResponse(serialiseChannel(ch)),
	}, nil
}

func (c Channels) ChannelUpdate(ctx context.Context, request openapi.ChannelUpdateRequestObject) (openapi.ChannelUpdateResponseObject, error) {
	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	partial := channel_svc.Partial{}

	if request.Body.Name != nil {
		partial.Name = opt.New(*request.Body.Name)
	}
	if request.Body.Slug != nil {
		partial.Slug = opt.New(*request.Body.Slug)
	}
	if request.Body.Description != nil {
		partial.Description = opt.New(*request.Body.Description)
	}
	if request.Body.Visibility != nil {
		partial.Visibility = opt.New(channel.Visibility(*request.Body.Visibility))
	}

	// Handle cover image - deletable.NewMap converts nullable to deletable
	partial.CoverImageAssetID = deletable.NewMap(request.Body.CoverImage, func(s string) *xid.ID {
		id := xid.ID(openapi.ParseID(s))
		return &id
	})

	// Handle icon - deletable.NewMap converts nullable to deletable
	partial.IconAssetID = deletable.NewMap(request.Body.Icon, func(s string) *xid.ID {
		id := xid.ID(openapi.ParseID(s))
		return &id
	})

	ch, err := c.channel_svc.Update(ctx, channelID, partial)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelUpdate200JSONResponse{
		ChannelGetOKJSONResponse: openapi.ChannelGetOKJSONResponse(serialiseChannel(ch)),
	}, nil
}

func (c Channels) ChannelDelete(ctx context.Context, request openapi.ChannelDeleteRequestObject) (openapi.ChannelDeleteResponseObject, error) {
	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	err := c.channel_svc.Delete(ctx, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelDelete200Response{}, nil
}

func serialiseChannelMember(m *channel_membership.Membership) openapi.ChannelMember {
	return openapi.ChannelMember{
		Id:        openapi.Identifier(m.ID.String()),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.CreatedAt,
		ChannelId: openapi.Identifier(m.ChannelID.String()),
		Account:   serialiseProfileReferenceFromAccount(m.Account),
		Role:      openapi.ChannelMemberRole(m.Role),
	}
}

func (c Channels) ChannelMemberList(ctx context.Context, request openapi.ChannelMemberListRequestObject) (openapi.ChannelMemberListResponseObject, error) {
	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	members, err := c.membership_svc.ListMembers(ctx, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelMemberList200JSONResponse{
		ChannelMemberListOKJSONResponse: openapi.ChannelMemberListOKJSONResponse{
			Members: dt.Map(members, serialiseChannelMember),
		},
	}, nil
}

func (c Channels) ChannelMemberAdd(ctx context.Context, request openapi.ChannelMemberAddRequestObject) (openapi.ChannelMemberAddResponseObject, error) {
	requesterID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))
	accountID := account.AccountID(openapi.ParseID(request.Body.AccountId))

	role := channel_membership.RoleMember
	if request.Body.Role != "" {
		role = channel_membership.Role(request.Body.Role)
	}

	member, err := c.membership_svc.AddMember(ctx, requesterID, channelID, accountID, role)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelMemberAdd200JSONResponse{
		ChannelMemberOKJSONResponse: openapi.ChannelMemberOKJSONResponse(serialiseChannelMember(member)),
	}, nil
}

func (c Channels) ChannelMemberUpdateRole(ctx context.Context, request openapi.ChannelMemberUpdateRoleRequestObject) (openapi.ChannelMemberUpdateRoleResponseObject, error) {
	requesterID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))
	accountID := account.AccountID(openapi.ParseID(request.AccountID))
	role := channel_membership.Role(request.Body.Role)

	member, err := c.membership_svc.UpdateRole(ctx, requesterID, channelID, accountID, role)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelMemberUpdateRole200JSONResponse{
		ChannelMemberOKJSONResponse: openapi.ChannelMemberOKJSONResponse(serialiseChannelMember(member)),
	}, nil
}

func (c Channels) ChannelMemberRemove(ctx context.Context, request openapi.ChannelMemberRemoveRequestObject) (openapi.ChannelMemberRemoveResponseObject, error) {
	requesterID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))
	accountID := account.AccountID(openapi.ParseID(request.AccountID))

	err = c.membership_svc.RemoveMember(ctx, requesterID, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelMemberRemove200Response{}, nil
}

func (c Channels) ChannelJoin(ctx context.Context, request openapi.ChannelJoinRequestObject) (openapi.ChannelJoinResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	member, err := c.membership_svc.JoinPublicChannel(ctx, accountID, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelJoin200JSONResponse{
		ChannelMemberOKJSONResponse: openapi.ChannelMemberOKJSONResponse(serialiseChannelMember(member)),
	}, nil
}

func (c Channels) ChannelLeave(ctx context.Context, request openapi.ChannelLeaveRequestObject) (openapi.ChannelLeaveResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	err = c.membership_svc.LeaveChannel(ctx, accountID, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelLeave200Response{}, nil
}

// Channel-scoped category operations

func (c Channels) ChannelCategoryList(ctx context.Context, request openapi.ChannelCategoryListRequestObject) (openapi.ChannelCategoryListResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check channel membership
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	// Get categories for this channel
	categories, err := c.category_repo.GetCategoriesByChannel(ctx, xid.ID(channelID))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCategoryList200JSONResponse{
		CategoryListOKJSONResponse: openapi.CategoryListOKJSONResponse{
			Categories: dt.Map(categories, serialiseCategory),
		},
	}, nil
}

func (c Channels) ChannelCategoryCreate(ctx context.Context, request openapi.ChannelCategoryCreateRequestObject) (openapi.ChannelCategoryCreateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	// Check if user can manage categories (owner/admin)
	canManage, err := c.membership_svc.CheckPermission(ctx, channelID, accountID, membership_svc.PermissionManageSettings)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !canManage {
		return nil, fault.Wrap(fault.New("permission denied"), fctx.With(ctx))
	}

	parentID := opt.NewEmpty[category.CategoryID]()
	if request.Body.Parent != nil {
		parentID = opt.New(category.CategoryID(openapi.ParseID(*request.Body.Parent)))
	}

	coverImageAssetID := deletable.Value[*xid.ID]{}
	if request.Body.CoverImageAssetId != nil {
		xidValue := openapi.ParseID(*request.Body.CoverImageAssetId)
		coverImageAssetID = deletable.Skip(opt.New(&xidValue))
	}

	cat, err := c.category_svc.Create(ctx, category_svc.Partial{
		Name:              opt.New(request.Body.Name),
		Slug:              opt.NewPtr(request.Body.Slug),
		Description:       opt.New(request.Body.Description),
		Colour:            opt.New(request.Body.Colour),
		Parent:            parentID,
		ChannelID:         opt.New(channelID),
		CoverImageAssetID: coverImageAssetID,
		Meta:              opt.NewPtr((*map[string]any)(request.Body.Meta)),
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCategoryCreate200JSONResponse{
		CategoryCreateOKJSONResponse: openapi.CategoryCreateOKJSONResponse(serialiseCategory(cat)),
	}, nil
}

func (c Channels) ChannelCategoryGet(ctx context.Context, request openapi.ChannelCategoryGetRequestObject) (openapi.ChannelCategoryGetResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check channel membership
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	cat, err := c.category_repo.Get(ctx, request.CategorySlug)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Verify category belongs to this channel
	if cat.ChannelID != xid.ID(channelID) {
		return nil, fault.Wrap(fault.New("category not found in this channel"), fctx.With(ctx))
	}

	return openapi.ChannelCategoryGet200JSONResponse{
		CategoryGetOKJSONResponse: openapi.CategoryGetOKJSONResponse{
			Body: serialiseCategory(cat),
			Headers: openapi.CategoryGetOKResponseHeaders{
				CacheControl: "public, no-cache",
			},
		},
	}, nil
}

func (c Channels) ChannelCategoryUpdate(ctx context.Context, request openapi.ChannelCategoryUpdateRequestObject) (openapi.ChannelCategoryUpdateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	// Check if user can manage categories (owner/admin)
	canManage, err := c.membership_svc.CheckPermission(ctx, channelID, accountID, membership_svc.PermissionManageSettings)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !canManage {
		return nil, fault.Wrap(fault.New("permission denied"), fctx.With(ctx))
	}

	// Get existing category to verify it belongs to this channel
	existingCat, err := c.category_repo.Get(ctx, request.CategorySlug)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if existingCat.ChannelID != channelID {
		return nil, fault.Wrap(fault.New("category not found in this channel"), fctx.With(ctx))
	}

	coverImageAssetID := deletable.NewMap(request.Body.CoverImageAssetId, func(id openapi.NullableIdentifier) *xid.ID {
		xidValue := openapi.ParseID(openapi.Identifier(id))
		return &xidValue
	})

	cat, err := c.category_svc.Update(ctx, request.CategorySlug, category_svc.Partial{
		Name:              opt.NewPtr(request.Body.Name),
		Slug:              opt.NewPtr(request.Body.Slug),
		Description:       opt.NewPtr(request.Body.Description),
		Colour:            opt.NewPtr(request.Body.Colour),
		CoverImageAssetID: coverImageAssetID,
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCategoryUpdate200JSONResponse{
		CategoryUpdateOKJSONResponse: openapi.CategoryUpdateOKJSONResponse(serialiseCategory(cat)),
	}, nil
}

func (c Channels) ChannelCategoryDelete(ctx context.Context, request openapi.ChannelCategoryDeleteRequestObject) (openapi.ChannelCategoryDeleteResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	// Check if user can manage categories (owner/admin)
	canManage, err := c.membership_svc.CheckPermission(ctx, channelID, accountID, membership_svc.PermissionManageSettings)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !canManage {
		return nil, fault.Wrap(fault.New("permission denied"), fctx.With(ctx))
	}

	// Get existing category to verify it belongs to this channel
	existingCat, err := c.category_repo.Get(ctx, request.CategorySlug)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if existingCat.ChannelID != channelID {
		return nil, fault.Wrap(fault.New("category not found in this channel"), fctx.With(ctx))
	}

	moveToID := category.CategoryID(openapi.ParseID(request.Body.MoveTo))

	cat, err := c.category_svc.Delete(ctx, request.CategorySlug, moveToID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCategoryDelete200JSONResponse{
		CategoryDeleteOKJSONResponse: openapi.CategoryDeleteOKJSONResponse(serialiseCategory(cat)),
	}, nil
}

func (c Channels) ChannelCategoryUpdatePosition(ctx context.Context, request openapi.ChannelCategoryUpdatePositionRequestObject) (openapi.ChannelCategoryUpdatePositionResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := xid.ID(openapi.ParseID(request.ChannelID))

	// Check if user can manage categories (owner/admin)
	canManage, err := c.membership_svc.CheckPermission(ctx, channelID, accountID, membership_svc.PermissionManageSettings)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !canManage {
		return nil, fault.Wrap(fault.New("permission denied"), fctx.With(ctx))
	}

	// Get existing category to verify it belongs to this channel
	existingCat, err := c.category_repo.Get(ctx, request.CategorySlug)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if existingCat.ChannelID != channelID {
		return nil, fault.Wrap(fault.New("category not found in this channel"), fctx.With(ctx))
	}

	parent, err := deletable.NewMapErr(request.Body.Parent, func(in openapi.Identifier) (category.CategoryID, error) {
		return category.CategoryID(openapi.ParseID(in)), nil
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	before := opt.NewEmpty[category.CategoryID]()
	if request.Body.Before != nil {
		before = opt.New(category.CategoryID(openapi.ParseID(*request.Body.Before)))
	}

	after := opt.NewEmpty[category.CategoryID]()
	if request.Body.After != nil {
		after = opt.New(category.CategoryID(openapi.ParseID(*request.Body.After)))
	}

	move := category_svc.Move{
		Parent: parent,
		Before: before,
		After:  after,
	}

	cats, err := c.category_svc.Move(ctx, request.CategorySlug, move)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelCategoryUpdatePosition200JSONResponse{
		CategoryListOKJSONResponse: openapi.CategoryListOKJSONResponse{
			Categories: dt.Map(cats, serialiseCategory),
		},
	}, nil
}

// Channel-scoped thread handlers

func (c Channels) ChannelThreadCreate(ctx context.Context, request openapi.ChannelThreadCreateRequestObject) (openapi.ChannelThreadCreateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	status, err := opt.MapErr(opt.NewPtr(request.Body.Visibility), deserialiseThreadStatus)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	var meta map[string]any
	if request.Body.Meta != nil {
		meta = *request.Body.Meta
	}

	tags := opt.Map(opt.NewPtr(request.Body.Tags), func(tags []string) tag_ref.Names {
		return dt.Map(tags, deserialiseTagName)
	})

	richContent, err := opt.MapErr(opt.NewPtr(request.Body.Body), datagraph.NewRichText)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	category := opt.NewPtrMap(request.Body.Category, func(cat openapi.Identifier) xid.ID {
		return openapi.ParseID(cat)
	})

	url, err := opt.MapErr(opt.NewPtr(request.Body.Url), func(s string) (url.URL, error) {
		u, err := url.Parse(s)
		if err != nil {
			return url.URL{}, err
		}
		return *u, nil
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	thread, err := c.thread_svc.Create(ctx,
		request.Body.Title,
		accountID,
		meta,
		thread_svc.Partial{
			Content:    richContent,
			Category:   category,
			Tags:       tags,
			Visibility: status,
			URL:        url,
			ChannelID:  opt.New(xid.ID(channelID)),
		},
	)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelThreadCreate200JSONResponse{
		ThreadCreateOKJSONResponse: openapi.ThreadCreateOKJSONResponse(serialiseThread(thread)),
	}, nil
}

func (c Channels) ChannelThreadList(ctx context.Context, request openapi.ChannelThreadListRequestObject) (openapi.ChannelThreadListResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	pageSize := 50

	page := opt.NewPtrMap(request.Params.Page, func(s string) int {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0
		}

		return max(1, int(v))
	}).Or(1)

	query := opt.NewPtr(request.Params.Q)

	author, err := openapi.OptionalID(ctx, c.profileQuery, request.Params.Author)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	visibilities, err := opt.MapErr(opt.NewPtr(request.Params.Visibility), deserialiseVisibilityList)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	tags := opt.NewPtrMap(request.Params.Tags, func(t []openapi.Identifier) []xid.ID {
		return dt.Map(t, func(i openapi.Identifier) xid.ID {
			return openapi.ParseID(i)
		})
	})

	cats := deserialiseCategorySlugQueryParam(request.Params.Categories)

	page = max(0, page-1)
	result, err := c.thread_svc.List(ctx, page, pageSize, thread_svc.Params{
		Query:      query,
		AccountID:  author,
		Visibility: visibilities,
		Tags:       tags,
		Categories: cats,
		ChannelID:  opt.New(xid.ID(channelID)),
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	page = result.CurrentPage + 1
	nextPage := opt.Map(result.NextPage, func(i int) int { return i + 1 })

	return openapi.ChannelThreadList200JSONResponse{
		ThreadListOKJSONResponse: openapi.ThreadListOKJSONResponse{
			Body: openapi.ThreadListResult{
				CurrentPage: page,
				NextPage:    nextPage.Ptr(),
				PageSize:    result.PageSize,
				Results:     result.Results,
				Threads:     dt.Map(result.Threads, serialiseThreadReference),
				TotalPages:  result.TotalPages,
			},
			Headers: openapi.ThreadListOKResponseHeaders{
				CacheControl: "no-store",
			},
		},
	}, nil
}

func (c Channels) ChannelThreadGet(ctx context.Context, request openapi.ChannelThreadGetRequestObject) (openapi.ChannelThreadGetResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	postID, err := c.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	etag, notModified := c.thread_cache.Check(ctx, reqinfo.GetCacheQuery(ctx), xid.ID(postID))
	if notModified {
		return openapi.ChannelThreadGet304Response{
			Headers: openapi.NotModifiedResponseHeaders{
				CacheControl: getAuthStateCacheControl(ctx, "no-cache"),
				LastModified: etag.Time.Format(time.RFC1123),
				ETag:         etag.String(),
			},
		}, nil
	}

	pp := deserialisePageParams(request.Params.Page, reply.RepliesPerPage)

	thread, err := c.thread_svc.Get(ctx, postID, pp)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Verify the thread belongs to this channel
	if thread.ChannelID != xid.ID(channelID) {
		return nil, fault.Wrap(fault.New("thread not found in this channel"), fctx.With(ctx))
	}

	if etag == nil {
		c.thread_cache.Store(ctx, xid.ID(thread.ID), thread.UpdatedAt)
		etag = cachecontrol.NewETag(thread.UpdatedAt)
	}

	return openapi.ChannelThreadGet200JSONResponse{
		ThreadGetJSONResponse: openapi.ThreadGetJSONResponse{
			Body: serialiseThread(thread),
			Headers: openapi.ThreadGetResponseHeaders{
				CacheControl: getAuthStateCacheControl(ctx, "no-cache"),
				LastModified: etag.Time.Format(time.RFC1123),
				ETag:         etag.String(),
			},
		},
	}, nil
}

func (c Channels) ChannelThreadUpdate(ctx context.Context, request openapi.ChannelThreadUpdateRequestObject) (openapi.ChannelThreadUpdateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	postID, err := c.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Get the thread first to verify it belongs to this channel
	pp := deserialisePageParams(nil, reply.RepliesPerPage)
	existingThread, err := c.thread_svc.Get(ctx, postID, pp)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Verify the thread belongs to this channel
	if existingThread.ChannelID != xid.ID(channelID) {
		return nil, fault.Wrap(fault.New("thread not found in this channel"), fctx.With(ctx))
	}

	tags := opt.Map(opt.NewPtr(request.Body.Tags), func(tags []string) tag_ref.Names {
		return dt.Map(tags, deserialiseTagName)
	})

	Visibility, err := opt.MapErr(opt.NewPtr(request.Body.Visibility), deserialiseThreadStatus)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	richContent, err := opt.MapErr(opt.NewPtr(request.Body.Body), datagraph.NewRichText)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	thread, err := c.thread_svc.Update(ctx, postID, thread_svc.Partial{
		Title:      opt.NewPtr(request.Body.Title),
		Content:    richContent,
		Tags:       tags,
		Category:   opt.NewPtrMap(request.Body.Category, deserialiseID),
		Visibility: Visibility,
	})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelThreadUpdate200JSONResponse{
		ThreadUpdateOKJSONResponse: openapi.ThreadUpdateOKJSONResponse(serialiseThread(thread)),
	}, nil
}

func (c Channels) ChannelThreadDelete(ctx context.Context, request openapi.ChannelThreadDeleteRequestObject) (openapi.ChannelThreadDeleteResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	postID, err := c.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Get the thread first to verify it belongs to this channel
	pp := deserialisePageParams(nil, reply.RepliesPerPage)
	existingThread, err := c.thread_svc.Get(ctx, postID, pp)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Verify the thread belongs to this channel
	if existingThread.ChannelID != xid.ID(channelID) {
		return nil, fault.Wrap(fault.New("thread not found in this channel"), fctx.With(ctx))
	}

	err = c.thread_svc.Delete(ctx, postID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelThreadDelete200Response{}, nil
}

func (c Channels) ChannelReplyCreate(ctx context.Context, request openapi.ChannelReplyCreateRequestObject) (openapi.ChannelReplyCreateResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channelID := channel.ChannelID(openapi.ParseID(request.ChannelID))

	// Check if user is a member of the channel
	hasAccess, err := c.channel_svc.CheckAccess(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasAccess {
		return nil, fault.Wrap(fault.New("access denied"), fctx.With(ctx))
	}

	postID, err := c.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Get the thread first to verify it belongs to this channel
	pp := deserialisePageParams(nil, reply.RepliesPerPage)
	existingThread, err := c.thread_svc.Get(ctx, postID, pp)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Verify the thread belongs to this channel
	if existingThread.ChannelID != xid.ID(channelID) {
		return nil, fault.Wrap(fault.New("thread not found in this channel"), fctx.With(ctx))
	}

	richContent, err := datagraph.NewRichText(request.Body.Body)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	partial := reply_svc.Partial{
		Content: opt.New(richContent),
		ReplyTo: opt.Map(opt.NewPtr(request.Body.ReplyTo), deserialisePostID),
		Meta:    opt.NewPtr((*map[string]any)(request.Body.Meta)),
	}

	post, err := c.reply_svc.Create(ctx,
		accountID,
		postID,
		partial,
	)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ChannelReplyCreate200JSONResponse{
		ReplyCreateOKJSONResponse: openapi.ReplyCreateOKJSONResponse(serialiseReplyPtr(post)),
	}, nil
}
