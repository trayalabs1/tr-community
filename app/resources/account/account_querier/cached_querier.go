package account_querier

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/xid"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/role"
	"github.com/Southclaws/storyden/app/resources/account/role/held"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/internal/infrastructure/cache"
)

const sessionAccountTTL = 5 * time.Minute

type sessionAccountGetter interface {
	GetByIDForSession(ctx context.Context, id account.AccountID) (*account.AccountWithEdges, error)
}

type SessionAccountLoader interface {
	GetByIDForSession(ctx context.Context, id account.AccountID) (*account.AccountWithEdges, error)
	Invalidate(ctx context.Context, id account.AccountID) error
}

type sessionRole struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Colour      string    `json:"colour"`
	Permissions []string  `json:"permissions"`
	SortKey     float64   `json:"sort_key"`
	CreatedAt   time.Time `json:"created_at"`
	Assigned    time.Time `json:"assigned"`
	Badge       bool      `json:"badge"`
	Default     bool      `json:"default"`
}

type sessionAccount struct {
	Account account.Account `json:"account"`
	Roles   []sessionRole   `json:"roles"`
}

func encodeSessionAccount(acc *account.AccountWithEdges) sessionAccount {
	return sessionAccount{
		Account: acc.Account,
		Roles: dt.Map(acc.Roles, func(r *held.Role) sessionRole {
			return sessionRole{
				ID:          r.ID.String(),
				Name:        r.Name,
				Colour:      r.Colour,
				Permissions: dt.Map(r.Permissions.List(), func(p rbac.Permission) string { return p.String() }),
				SortKey:     r.SortKey,
				CreatedAt:   r.CreatedAt,
				Assigned:    r.Assigned,
				Badge:       r.Badge,
				Default:     r.Default,
			}
		}),
	}
}

func (s sessionAccount) decode() (*account.AccountWithEdges, error) {
	roles := make(held.Roles, len(s.Roles))
	for i, r := range s.Roles {
		id, err := xid.FromString(r.ID)
		if err != nil {
			return nil, fault.Wrap(err)
		}

		perms, err := rbac.NewPermissions(r.Permissions)
		if err != nil {
			return nil, fault.Wrap(err)
		}

		roles[i] = &held.Role{
			Role: role.Role{
				ID:          role.RoleID(id),
				Name:        r.Name,
				Colour:      r.Colour,
				Permissions: *perms,
				SortKey:     r.SortKey,
				CreatedAt:   r.CreatedAt,
			},
			Assigned: r.Assigned,
			Badge:    r.Badge,
			Default:  r.Default,
		}
	}

	return &account.AccountWithEdges{Account: s.Account, Roles: roles}, nil
}

type cachedQuerier struct {
	inner sessionAccountGetter
	store cache.Store
}

func NewCachedQuerier(inner sessionAccountGetter, store cache.Store) SessionAccountLoader {
	return &cachedQuerier{inner: inner, store: store}
}

func ProvideSessionAccountLoader(inner *Querier, store cache.Store) SessionAccountLoader {
	return NewCachedQuerier(inner, store)
}

func sessionAccountKey(id account.AccountID) string {
	return "session_account:" + id.String()
}

func (c *cachedQuerier) GetByIDForSession(ctx context.Context, id account.AccountID) (*account.AccountWithEdges, error) {
	if raw, err := c.store.Get(ctx, sessionAccountKey(id)); err == nil {
		var dto sessionAccount
		if json.Unmarshal([]byte(raw), &dto) == nil {
			if acc, err := dto.decode(); err == nil {
				return acc, nil
			}
		}
	}

	acc, err := c.inner.GetByIDForSession(ctx, id)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if payload, err := json.Marshal(encodeSessionAccount(acc)); err == nil {
		_ = c.store.Set(ctx, sessionAccountKey(id), string(payload), sessionAccountTTL)
	}

	return acc, nil
}

func (c *cachedQuerier) Invalidate(ctx context.Context, id account.AccountID) error {
	return c.store.Delete(ctx, sessionAccountKey(id))
}
