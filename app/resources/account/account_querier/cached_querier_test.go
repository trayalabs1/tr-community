package account_querier

import (
	"context"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/role"
	"github.com/Southclaws/storyden/app/resources/account/role/held"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/rbac"
)

type mapStore struct {
	data map[string]string
}

func newMapStore() *mapStore { return &mapStore{data: map[string]string{}} }

func (s *mapStore) Get(_ context.Context, key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", errNotFoundStub
	}
	return v, nil
}

func (s *mapStore) Set(_ context.Context, key string, object string, _ time.Duration) error {
	s.data[key] = object
	return nil
}

func (s *mapStore) Delete(_ context.Context, key string) error {
	delete(s.data, key)
	return nil
}

func (s *mapStore) HGetAll(_ context.Context, _ string) (map[string]string, error) {
	return nil, errNotFoundStub
}

func (s *mapStore) HIncrBy(_ context.Context, _ string, _ string, _ int64) (int, error) {
	return 0, errNotFoundStub
}

func (s *mapStore) HDel(_ context.Context, _ string, _ string) error { return nil }

func (s *mapStore) Expire(_ context.Context, _ string, _ time.Duration) error { return nil }

var errNotFoundStub = &stubErr{}

type stubErr struct{}

func (*stubErr) Error() string { return "not found" }

type countingLoader struct {
	acc   *account.AccountWithEdges
	calls int
}

func (l *countingLoader) GetByIDForSession(_ context.Context, _ account.AccountID) (*account.AccountWithEdges, error) {
	l.calls++
	return l.acc, nil
}

func fixtureAccount(t *testing.T) *account.AccountWithEdges {
	t.Helper()
	bio, err := datagraph.NewRichText("some **bio** text")
	require.NoError(t, err)

	return &account.AccountWithEdges{
		Account: account.Account{
			ID:        account.AccountID(xid.New()),
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
			Handle:    "fixture-handle",
			Name:      "Fixture",
			Bio:       bio,
			Kind:      account.AccountKindHuman,
			Admin:     true,
			Metadata:  map[string]any{"k": "v"},
		},
		Roles: held.Roles{
			{
				Role: role.Role{
					ID:          role.RoleID(xid.New()),
					Name:        "Member",
					Permissions: rbac.NewList(rbac.PermissionCreatePost),
				},
				Default: true,
			},
		},
	}
}

func TestCachedQuerierServesSecondCallFromCache(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	inner := &countingLoader{acc: fixtureAccount(t)}
	c := NewCachedQuerier(inner, newMapStore())

	first, err := c.GetByIDForSession(ctx, inner.acc.ID)
	r.NoError(err)
	r.Equal(inner.acc.ID, first.ID)

	second, err := c.GetByIDForSession(ctx, inner.acc.ID)
	r.NoError(err)

	r.Equal(1, inner.calls)
	r.Equal(inner.acc.ID, second.ID)
	r.Equal(inner.acc.Handle, second.Handle)
	r.Equal(inner.acc.Bio.HTML(), second.Bio.HTML())
	r.Equal(inner.acc.Roles.Roles(), second.Roles.Roles())
}

func TestCachedQuerierInvalidateForcesReload(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	inner := &countingLoader{acc: fixtureAccount(t)}
	c := NewCachedQuerier(inner, newMapStore())

	_, err := c.GetByIDForSession(ctx, inner.acc.ID)
	r.NoError(err)
	r.Equal(1, inner.calls)

	r.NoError(c.Invalidate(ctx, inner.acc.ID))

	_, err = c.GetByIDForSession(ctx, inner.acc.ID)
	r.NoError(err)
	r.Equal(2, inner.calls)
}
