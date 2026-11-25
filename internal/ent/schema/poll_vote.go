package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

type PollVote struct {
	ent.Schema
}

func (PollVote) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}}
}

func (PollVote) Fields() []ent.Field {
	return []ent.Field{
		field.String("account_id").GoType(xid.ID{}),
		field.String("post_id").GoType(xid.ID{}),
		field.String("option_id"),
	}
}

func (PollVote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("poll_votes").
			Field("account_id").
			Unique().
			Required(),
		edge.From("post", Post.Type).
			Ref("poll_votes").
			Field("post_id").
			Unique().
			Required(),
	}
}

func (PollVote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "post_id").
			Unique().
			StorageKey("unique_poll_vote_per_user"),
		index.Fields("post_id", "option_id"),
	}
}
