package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/rs/xid"
)

type Channel struct {
	ent.Schema
}

func (Channel) Mixin() []ent.Mixin {
	return []ent.Mixin{Identifier{}, CreatedAt{}, UpdatedAt{}}
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("The display name of the channel"),
		field.String("slug").
			Unique().
			Comment("URL-friendly identifier for the channel"),
		field.String("description").
			Optional().
			Default("").
			Comment("Description of the channel's purpose"),
		field.String("cover_image_asset_id").
			GoType(xid.ID{}).
			Optional().
			Nillable().
			Comment("Cover image for the channel"),
		field.String("icon_asset_id").
			GoType(xid.ID{}).
			Optional().
			Nillable().
			Comment("Icon for the channel"),
		field.Enum("visibility").
			Values("public", "private").
			Default("public").
			Comment("public: discoverable and joinable by anyone, private: invite-only"),
		field.JSON("metadata", map[string]any{}).
			Optional().
			Comment("Arbitrary metadata used by clients to store domain specific information"),
	}
}

func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("cover_image", Asset.Type).
			Field("cover_image_asset_id").
			Unique(),

		edge.To("icon", Asset.Type).
			Field("icon_asset_id").
			Unique(),

		edge.To("memberships", ChannelMembership.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),

		edge.To("categories", Category.Type),

		edge.To("posts", Post.Type),

		edge.To("nodes", Node.Type),
	}
}
