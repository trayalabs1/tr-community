package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

type ReplyAdminQueue struct {
	ent.Schema
}

func (ReplyAdminQueue) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}}
}

func (ReplyAdminQueue) Fields() []ent.Field {
	return []ent.Field{
		field.String("reply_id").GoType(xid.ID{}),
		field.String("thread_id").GoType(xid.ID{}),
		field.String("channel_id").GoType(xid.ID{}),
		field.String("content_snippet").Default(""),
	}
}

func (ReplyAdminQueue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("reply", Post.Type).
			Field("reply_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("thread", Post.Type).
			Field("thread_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ReplyAdminQueue) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
