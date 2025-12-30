package channel_membership

import (
	"time"

	"github.com/Southclaws/dt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/internal/ent"
	ent_membership "github.com/Southclaws/storyden/internal/ent/channelmembership"
)

type MembershipID xid.ID

func (i MembershipID) String() string { return xid.ID(i).String() }

type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
)

type Membership struct {
	ID        MembershipID
	ChannelID xid.ID
	Account   account.Account
	Role      Role
	CreatedAt time.Time
}

func FromModel(m *ent.ChannelMembership) (*Membership, error) {
	acc, err := account.MapRef(m.Edges.Account)
	if err != nil {
		return nil, err
	}

	return &Membership{
		ID:        MembershipID(m.ID),
		ChannelID: m.ChannelID,
		Account:   *acc,
		Role:      Role(m.Role),
		CreatedAt: m.CreatedAt,
	}, nil
}

func FromModelSlice(memberships []*ent.ChannelMembership) ([]*Membership, error) {
	return dt.MapErr(memberships, FromModel)
}

func (r Role) String() string {
	return string(r)
}

func (r Role) ToEnt() ent_membership.Role {
	return ent_membership.Role(r)
}

func RoleFromEnt(r ent_membership.Role) Role {
	return Role(r)
}

func (r Role) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanManageSettings() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanModerateContent() bool {
	return r == RoleOwner || r == RoleAdmin || r == RoleModerator
}

func (r Role) CanDeleteChannel() bool {
	return r == RoleOwner
}
