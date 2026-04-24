package postfilter

import (
	"entgo.io/ent/dialect/sql"

	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

func NotBAHPost() predicate.Post {
	return predicate.Post(func(s *sql.Selector) {
		s.Where(sql.Or(
			sql.IsNull(s.C(ent_post.FieldMetadata)),
			sql.P(func(b *sql.Builder) {
				b.WriteString(s.C(ent_post.FieldMetadata) + "->>'post_category' != 'BAH'")
			}),
		))
	})
}
