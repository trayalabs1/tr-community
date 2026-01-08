package channel_membership

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/channel_membership"
)

var (
	errUnauthorized = fault.New("unauthorized", ftag.With(ftag.PermissionDenied))
)

type Service interface {
	AddMember(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID, role channel_membership.Role) (*channel_membership.Membership, error)
	BulkAddMembers(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountIDs []account.AccountID, role channel_membership.Role) ([]*channel_membership.Membership, error)
	RemoveMember(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID) error
	UpdateRole(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID, role channel_membership.Role) (*channel_membership.Membership, error)
	ListMembers(ctx context.Context, channelID xid.ID) ([]*channel_membership.Membership, error)
	GetMembership(ctx context.Context, channelID xid.ID, accountID account.AccountID) (*channel_membership.Membership, error)
	JoinPublicChannel(ctx context.Context, accountID account.AccountID, channelID xid.ID) (*channel_membership.Membership, error)
	LeaveChannel(ctx context.Context, accountID account.AccountID, channelID xid.ID) error
	CheckPermission(ctx context.Context, channelID xid.ID, accountID account.AccountID, permission Permission) (bool, error)
}

type Permission string

const (
	PermissionManageMembers   Permission = "manage_members"
	PermissionManageSettings  Permission = "manage_settings"
	PermissionModerateContent Permission = "moderate_content"
	PermissionDeleteChannel   Permission = "delete_channel"
)

func Build() fx.Option {
	return fx.Provide(New)
}

type service struct {
	membership_repo *channel_membership.Repository
}

func New(membership_repo *channel_membership.Repository) Service {
	return &service{
		membership_repo: membership_repo,
	}
}

func (s *service) checkPermission(ctx context.Context, channelID xid.ID, accountID account.AccountID, permission Permission) (bool, error) {
	membership, err := s.membership_repo.GetByChannelAndAccount(ctx, channelID, accountID)
	if err != nil {
		return false, nil
	}

	switch permission {
	case PermissionManageMembers:
		return membership.Role.CanManageMembers(), nil
	case PermissionManageSettings:
		return membership.Role.CanManageSettings(), nil
	case PermissionModerateContent:
		return membership.Role.CanModerateContent(), nil
	case PermissionDeleteChannel:
		return membership.Role.CanDeleteChannel(), nil
	default:
		return false, nil
	}
}

func (s *service) AddMember(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID, role channel_membership.Role) (*channel_membership.Membership, error) {
	hasPermission, err := s.checkPermission(ctx, channelID, requesterID, PermissionManageMembers)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasPermission {
		return nil, fault.Wrap(errUnauthorized, fctx.With(ctx))
	}

	membership, err := s.membership_repo.Add(ctx, channelID, accountID, role)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return membership, nil
}

func (s *service) BulkAddMembers(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountIDs []account.AccountID, role channel_membership.Role) ([]*channel_membership.Membership, error) {
	hasPermission, err := s.checkPermission(ctx, channelID, requesterID, PermissionManageMembers)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasPermission {
		return nil, fault.Wrap(errUnauthorized, fctx.With(ctx))
	}

	memberships, err := s.membership_repo.BulkAdd(ctx, channelID, accountIDs, role)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return memberships, nil
}

func (s *service) RemoveMember(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID) error {
	hasPermission, err := s.checkPermission(ctx, channelID, requesterID, PermissionManageMembers)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}
	if !hasPermission {
		return fault.Wrap(errUnauthorized, fctx.With(ctx))
	}

	err = s.membership_repo.Remove(ctx, channelID, accountID)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (s *service) UpdateRole(ctx context.Context, requesterID account.AccountID, channelID xid.ID, accountID account.AccountID, role channel_membership.Role) (*channel_membership.Membership, error) {
	hasPermission, err := s.checkPermission(ctx, channelID, requesterID, PermissionManageMembers)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	if !hasPermission {
		return nil, fault.Wrap(errUnauthorized, fctx.With(ctx))
	}

	membership, err := s.membership_repo.UpdateRole(ctx, channelID, accountID, role)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return membership, nil
}

func (s *service) ListMembers(ctx context.Context, channelID xid.ID) ([]*channel_membership.Membership, error) {
	memberships, err := s.membership_repo.ListByChannel(ctx, channelID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return memberships, nil
}

func (s *service) GetMembership(ctx context.Context, channelID xid.ID, accountID account.AccountID) (*channel_membership.Membership, error) {
	membership, err := s.membership_repo.GetByChannelAndAccount(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return membership, nil
}

func (s *service) JoinPublicChannel(ctx context.Context, accountID account.AccountID, channelID xid.ID) (*channel_membership.Membership, error) {
	membership, err := s.membership_repo.Add(ctx, channelID, accountID, channel_membership.RoleMember)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return membership, nil
}

func (s *service) LeaveChannel(ctx context.Context, accountID account.AccountID, channelID xid.ID) error {
	err := s.membership_repo.Remove(ctx, channelID, accountID)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (s *service) CheckPermission(ctx context.Context, channelID xid.ID, accountID account.AccountID, permission Permission) (bool, error) {
	return s.checkPermission(ctx, channelID, accountID, permission)
}
