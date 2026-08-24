package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

type PostSentiment struct {
	ent.Schema
}

func (PostSentiment) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}, UpdatedAt{}}
}

func (PostSentiment) Fields() []ent.Field {
	return []ent.Field{
		field.String("post_id").
			GoType(xid.ID{}).
			Unique(),

		field.String("sentiment_tag").
			Optional().
			Nillable().
			Comment("Sentiment classification: positive, neutral, negative"),

		field.Int("positivity_score").
			Optional().
			Nillable().
			Comment("Positivity score from 0-100"),

		field.String("primary_topic").
			Optional().
			Nillable().
			Comment("Category from the v4 allowed list"),

		field.Int("feed_value_score").
			Optional().
			Nillable().
			Comment("Feed value score from 0-100 (usefulness/discussion potential)"),

		field.Enum("scoring_status").
			Values("unscored", "scored", "failed").
			Default("unscored").
			Comment("Status of sentiment scoring: unscored, scored, failed"),

		field.Float("rank_score").
			Default(0).
			Comment("v4 content_score: positivity + feed_value + quality + category_boost, computed once at classification time"),
	}
}

func (PostSentiment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("scoring_status"),
		index.Fields("rank_score"),
		index.Fields("sentiment_tag"),
	}
}

func (PostSentiment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Field("post_id").
			Ref("sentiment").
			Unique().
			Required(),
	}
}
