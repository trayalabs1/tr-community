package thread_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestThreadEdgeIndexApplied(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		db *sql.DB,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			var name string
			err := db.QueryRowContext(root,
				`SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'tag_posts_post_id'`,
			).Scan(&name)
			r.NoError(err)
			r.Equal("tag_posts_post_id", name)
		}))
	}))
}
