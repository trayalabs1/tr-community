package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

type AdminReplyTime struct {
	ent.Schema
}

func (AdminReplyTime) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}}
}

func (AdminReplyTime) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_post_id").GoType(xid.ID{}),
		field.String("admin_post_id").GoType(xid.ID{}).Unique(),
		field.Time("user_post_time"),
		field.Time("admin_post_time"),
		field.String("admin_account_id").GoType(xid.ID{}),
		field.String("admin_handle"),
		field.Int64("time_difference_seconds"),
	}
}

func (AdminReplyTime) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_post", Post.Type).
			Field("user_post_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("admin_post", Post.Type).
			Field("admin_post_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("admin_account", Account.Type).
			Field("admin_account_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (AdminReplyTime) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("admin_account_id", "admin_post_time"),
		index.Fields("admin_post_time"),
	}
}
