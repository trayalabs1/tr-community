package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

type ThreadShare struct {
	ent.Schema
}

func (ThreadShare) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}, DeletedAt{}}
}

func (ThreadShare) Fields() []ent.Field {
	return []ent.Field{
		field.String("post_id").
			GoType(xid.ID{}).
			Comment("The root thread being shared into another channel"),
		field.String("channel_id").
			GoType(xid.ID{}).
			Comment("The destination channel the thread is featured in"),
		field.String("account_id").
			GoType(xid.ID{}).
			Comment("The account that shared the thread"),
		field.String("subtitle").
			Optional().
			Default("").
			Comment("Editorial subtitle shown on the featured card at the destination"),
		field.Int("pinned_rank").
			Default(0).
			Comment("Per-destination pin rank; independent of the post's own pinned_rank"),
		field.Time("pinned_at").
			Optional().
			Nillable().
			Comment("When the share was pinned in the destination channel"),
	}
}

func (ThreadShare) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("shares").
			Field("post_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),

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

func (ThreadShare) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id", "channel_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")).
			StorageKey("unique_thread_share"),
		index.Fields("channel_id", "deleted_at"),
		index.Fields("account_id"),
	}
}
