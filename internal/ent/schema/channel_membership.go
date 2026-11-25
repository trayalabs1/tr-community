package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

var (
	ChannelRoleOwner     = "owner"
	ChannelRoleAdmin     = "admin"
	ChannelRoleModerator = "moderator"
	ChannelRoleMember    = "member"
)

var ChannelRoles = []string{
	ChannelRoleOwner,
	ChannelRoleAdmin,
	ChannelRoleModerator,
	ChannelRoleMember,
}

type ChannelMembership struct {
	ent.Schema
}

func (ChannelMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}}
}

func (ChannelMembership) Fields() []ent.Field {
	return []ent.Field{
		field.String("channel_id").
			GoType(xid.ID{}).
			Comment("The channel this membership belongs to"),
		field.String("account_id").
			GoType(xid.ID{}).
			Comment("The account that is a member of the channel"),
		field.Enum("role").
			Values(ChannelRoles...).
			Default(ChannelRoleMember).
			Comment("The role of the member in the channel: owner > admin > moderator > member"),
	}
}

func (ChannelMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),

		edge.To("account", Account.Type).
			Field("account_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ChannelMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "account_id").
			Unique().
			StorageKey("unique_channel_member"),
		index.Fields("account_id"),
		index.Fields("channel_id", "role"),
	}
}
