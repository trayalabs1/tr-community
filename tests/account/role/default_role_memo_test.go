package role_test

import (
	"context"
	"sync/atomic"
	"testing"

	entgo "entgo.io/ent"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/role/role_querier"
	"github.com/Southclaws/storyden/internal/ent"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestDefaultRolesMemoized(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		db *ent.Client,
		rq *role_querier.Querier,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			var roleQueries int64
			db.Intercept(ent.InterceptFunc(func(next ent.Querier) ent.Querier {
				return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
					if entgo.QueryFromContext(ctx).Type == "Role" {
						atomic.AddInt64(&roleQueries, 1)
					}
					return next.Query(ctx, query)
				})
			}))

			_, err := rq.GetMemberRole(root)
			r.NoError(err)
			after1 := atomic.LoadInt64(&roleQueries)

			_, err = rq.GetGuestRole(root)
			r.NoError(err)
			_, err = rq.GetMemberRole(root)
			r.NoError(err)
			after3 := atomic.LoadInt64(&roleQueries)

			r.Equal(after1, after3, "subsequent default-role lookups should be served from memo, issuing no new Role queries")
		}))
	}))
}
