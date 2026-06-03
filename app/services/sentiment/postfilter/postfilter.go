package postfilter

import (
	"entgo.io/ent/dialect/sql"

	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

// NotPrescoredPost excludes posts whose post_category is prescored at creation
// (neutral tag + fixed rank score) and therefore must not be re-scored by the
// AI sentiment pipeline. Keep this list in sync with isPrescoredCategory in
// app/services/thread/create.go.
func NotPrescoredPost() predicate.Post {
	return predicate.Post(func(s *sql.Selector) {
		s.Where(sql.P(func(b *sql.Builder) {
			cat := s.C(ent_post.FieldMetadata) + "->>'post_category'"
			b.WriteString("COALESCE(" + cat + ", '') NOT IN ('BAH', 'feedback')")
		}))
	})
}

// NotBAHPost is retained for backward compatibility; new callers should use
// NotPrescoredPost which covers all prescored categories.
//
// Deprecated: use NotPrescoredPost.
func NotBAHPost() predicate.Post {
	return NotPrescoredPost()
}
