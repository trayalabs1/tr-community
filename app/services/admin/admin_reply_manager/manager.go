package admin_reply_manager

import (
	"context"
	"slices"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/reply_querier"
	"github.com/Southclaws/storyden/app/resources/rbac"
)

type Manager struct {
	replies  *reply_querier.Querier
	accounts *account_querier.Querier
}

func New(
	replies *reply_querier.Querier,
	accounts *account_querier.Querier,
) *Manager {
	return &Manager{replies: replies, accounts: accounts}
}

type ListOpts struct {
	// RepliedBy restricts results to a single admin. When absent, replies from
	// every admin are returned.
	RepliedBy     opt.Optional[account.AccountID]
	CreatedAfter  opt.Optional[time.Time]
	CreatedBefore opt.Optional[time.Time]
}

// Admins lists every account the platform treats as an administrator.
func (m *Manager) Admins(ctx context.Context) ([]*account.Account, error) {
	admins, err := m.accounts.ListByHeldPermission(ctx, rbac.PermissionAdministrator)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return admins, nil
}

// List returns admin-authored replies, newest first. Without a RepliedBy filter
// the results span all admins, so the admin set is resolved up front and used
// as the author filter.
func (m *Manager) List(
	ctx context.Context,
	page pagination.Parameters,
	opts ListOpts,
) (pagination.Result[*reply.Reply], error) {
	admins, err := m.Admins(ctx)
	if err != nil {
		return pagination.Result[*reply.Reply]{}, fault.Wrap(err, fctx.With(ctx))
	}

	adminIDs := dt.Map(admins, func(a *account.Account) xid.ID { return xid.ID(a.ID) })

	if requested, ok := opts.RepliedBy.Get(); ok {
		if !slices.Contains(adminIDs, xid.ID(requested)) {
			return pagination.Result[*reply.Reply]{}, fault.New("requested account is not an administrator",
				fctx.With(ctx),
				ftag.With(ftag.InvalidArgument),
				fmsg.WithDesc("not an admin", "The selected account is not an administrator."))
		}

		adminIDs = []xid.ID{xid.ID(requested)}
	}

	if len(adminIDs) == 0 {
		return pagination.NewPageResult(page, 0, []*reply.Reply{}), nil
	}

	filters := reply_querier.ListFilters{AuthorIDs: adminIDs}
	opts.CreatedAfter.Call(func(t time.Time) { filters.CreatedAfter = &t })
	opts.CreatedBefore.Call(func(t time.Time) { filters.CreatedBefore = &t })

	result, err := m.replies.List(ctx, page, filters)
	if err != nil {
		return pagination.Result[*reply.Reply]{}, fault.Wrap(err, fctx.With(ctx))
	}

	return result, nil
}
