package notify_querier

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/samber/lo"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/notification"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/post_search"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/reply_querier"
	"github.com/Southclaws/storyden/internal/ent"
	entnotification "github.com/Southclaws/storyden/internal/ent/notification"
)

type Querier struct {
	db           *ent.Client
	postSearcher post_search.Repository
	replyQuerier *reply_querier.Querier
}

func New(db *ent.Client, postSearcher post_search.Repository, replyQuerier *reply_querier.Querier) *Querier {
	return &Querier{db: db, postSearcher: postSearcher, replyQuerier: replyQuerier}
}

func (n *Querier) ListNotifications(
	ctx context.Context,
	accountID account.AccountID,
	pp pagination.Parameters,
	read opt.Optional[bool],
) (*pagination.Result[*notification.Notification], error) {
	query := n.db.Notification.Query().
		Where(entnotification.OwnerAccountID(xid.ID(accountID)))

	if v, ok := read.Get(); ok {
		query.Where(entnotification.Read(v))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	r, err := query.
		Order(ent.Desc(entnotification.FieldCreatedAt), ent.Desc(entnotification.FieldID)).
		Offset(pp.Offset()).
		Limit(pp.Limit()).
		WithSource().
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	refs, err := dt.MapErr(r, notification.Map)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	paged := pagination.NewPageResult(pp, total, refs)

	ns, err := n.hydrateRefs(ctx, paged.Items)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	result := pagination.ConvertPageResult(paged, []*notification.Notification(ns))

	return &result, nil
}

func (n *Querier) hydrateRefs(ctx context.Context, refs notification.NotificationRefs) (notification.Notifications, error) {
	grouped := lo.GroupBy(refs, func(n *notification.NotificationRef) datagraph.Kind {
		return n.ItemRef.OrZero().Kind
	})

	pids := dt.Map(grouped[datagraph.KindPost], func(n *notification.NotificationRef) post.ID {
		return post.ID(n.ItemRef.OrZero().ID)
	})
	posts, err := n.postSearcher.GetMany(ctx, pids...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	pg := lo.KeyBy(posts, func(p *post.Post) post.ID { return p.ID })

	rids := dt.Map(grouped[datagraph.KindReply], func(n *notification.NotificationRef) post.ID {
		return post.ID(n.ItemRef.OrZero().ID)
	})
	replies, err := n.replyQuerier.GetMany(ctx, rids...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	rg := lo.KeyBy(replies, func(r *reply.Reply) post.ID { return r.ID })

	ns := dt.Map(refs, func(r *notification.NotificationRef) *notification.Notification {
		switch r.ItemRef.OrZero().Kind {
		case datagraph.KindPost:
			p := pg[post.ID(r.ItemRef.OrZero().ID)]

			if p == nil {
				// Post was deleted, skip.
				return nil
			}

			return &notification.Notification{
				ID:     r.ID,
				Event:  r.Event,
				Item:   p,
				Source: r.Source,
				Time:   r.Time,
				Read:   r.Read,
			}

		case datagraph.KindReply:
			rep := rg[post.ID(r.ItemRef.OrZero().ID)]

			if rep == nil {
				// Reply was deleted, skip.
				return nil
			}

			return &notification.Notification{
				ID:     r.ID,
				Event:  r.Event,
				Item:   rep,
				Source: r.Source,
				Time:   r.Time,
				Read:   r.Read,
			}
		}

		return &notification.Notification{
			ID:     r.ID,
			Event:  r.Event,
			Source: r.Source,
			Time:   r.Time,
			Read:   r.Read,
		}
	})

	ns = dt.Filter(ns, func(n *notification.Notification) bool {
		return n != nil
	})

	return ns, nil
}
