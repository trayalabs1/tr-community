package channel_membership

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/internal/ent"
	ent_membership "github.com/Southclaws/storyden/internal/ent/channelmembership"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

var (
	errNotFound      = fault.Wrap(fault.New("membership not found"), ftag.With(ftag.NotFound))
	errAlreadyExists = fault.Wrap(fault.New("membership already exists"), ftag.With(ftag.AlreadyExists))
)

type Repository struct {
	db *ent.Client
}

func New(db *ent.Client) *Repository {
	return &Repository{db}
}

func (r *Repository) Add(ctx context.Context, channelID xid.ID, accountID account.AccountID, role Role) (*Membership, error) {
	membership, err := r.db.ChannelMembership.Create().
		SetChannelID(channelID).
		SetAccountID(xid.ID(accountID)).
		SetRole(role.ToEnt()).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fault.Wrap(errAlreadyExists, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.Get(ctx, MembershipID(membership.ID))
}

func (r *Repository) BulkAdd(ctx context.Context, channelID xid.ID, accountIDs []account.AccountID, role Role) ([]*Membership, error) {
	builders := make([]*ent.ChannelMembershipCreate, len(accountIDs))
	for i, accountID := range accountIDs {
		builders[i] = r.db.ChannelMembership.Create().
			SetChannelID(channelID).
			SetAccountID(xid.ID(accountID)).
			SetRole(role.ToEnt())
	}

	_, err := r.db.ChannelMembership.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.ListByChannel(ctx, channelID)
}

func (r *Repository) Get(ctx context.Context, id MembershipID) (*Membership, error) {
	membership, err := r.db.ChannelMembership.Query().
		Where(ent_membership.ID(xid.ID(id))).
		WithAccount(func(aq *ent.AccountQuery) {
			aq.WithAccountRoles(func(arq *ent.AccountRolesQuery) {
				arq.WithRole()
			})
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result, err := FromModel(membership)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return result, nil
}

func (r *Repository) GetByChannelAndAccount(ctx context.Context, channelID xid.ID, accountID account.AccountID) (*Membership, error) {
	membership, err := r.db.ChannelMembership.Query().
		Where(
			ent_membership.ChannelID(channelID),
			ent_membership.AccountID(xid.ID(accountID)),
		).
		WithAccount(func(aq *ent.AccountQuery) {
			aq.WithAccountRoles(func(arq *ent.AccountRolesQuery) {
				arq.WithRole()
			})
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result, err := FromModel(membership)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return result, nil
}

func (r *Repository) ListByChannel(ctx context.Context, channelID xid.ID, filters ...predicate.ChannelMembership) ([]*Membership, error) {
	baseFilters := []predicate.ChannelMembership{
		ent_membership.ChannelID(channelID),
	}
	baseFilters = append(baseFilters, filters...)

	memberships, err := r.db.ChannelMembership.Query().
		Where(baseFilters...).
		WithAccount(func(aq *ent.AccountQuery) {
			aq.WithAccountRoles(func(arq *ent.AccountRolesQuery) {
				arq.WithRole()
			})
		}).
		Order(ent.Asc(ent_membership.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.MapErr(memberships, FromModel)
}

func (r *Repository) ListByAccount(ctx context.Context, accountID account.AccountID) ([]*Membership, error) {
	memberships, err := r.db.ChannelMembership.Query().
		Where(ent_membership.AccountID(xid.ID(accountID))).
		WithAccount(func(aq *ent.AccountQuery) {
			aq.WithAccountRoles(func(arq *ent.AccountRolesQuery) {
				arq.WithRole()
			})
		}).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.MapErr(memberships, FromModel)
}

func (r *Repository) UpdateRole(ctx context.Context, channelID xid.ID, accountID account.AccountID, role Role) (*Membership, error) {
	_, err := r.db.ChannelMembership.Update().
		Where(
			ent_membership.ChannelID(channelID),
			ent_membership.AccountID(xid.ID(accountID)),
		).
		SetRole(role.ToEnt()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Need to query to get the membership ID since Update doesn't return the entity
	membership, err := r.GetByChannelAndAccount(ctx, channelID, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return membership, nil
}

func (r *Repository) Remove(ctx context.Context, channelID xid.ID, accountID account.AccountID) error {
	_, err := r.db.ChannelMembership.Delete().
		Where(
			ent_membership.ChannelID(channelID),
			ent_membership.AccountID(xid.ID(accountID)),
		).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (r *Repository) CheckMembership(ctx context.Context, channelID xid.ID, accountID account.AccountID) (bool, error) {
	exists, err := r.db.ChannelMembership.Query().
		Where(
			ent_membership.ChannelID(channelID),
			ent_membership.AccountID(xid.ID(accountID)),
		).
		Exist(ctx)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx))
	}

	return exists, nil
}

func (r *Repository) GetAccountChannelIDs(ctx context.Context, accountID account.AccountID) ([]xid.ID, error) {
	memberships, err := r.db.ChannelMembership.Query().
		Where(ent_membership.AccountID(xid.ID(accountID))).
		Select(ent_membership.FieldChannelID).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.Map(memberships, func(m *ent.ChannelMembership) xid.ID {
		return m.ChannelID
	}), nil
}

func (r *Repository) CountMembers(ctx context.Context, channelID xid.ID) (int, error) {
	count, err := r.db.ChannelMembership.Query().
		Where(ent_membership.ChannelID(channelID)).
		Count(ctx)
	if err != nil {
		return 0, fault.Wrap(err, fctx.With(ctx))
	}

	return count, nil
}
