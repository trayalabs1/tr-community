package post_like_notify

import (
	"context"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/notification"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/services/notification/notify"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func Build() fx.Option {
	return fx.Invoke(func(
		ctx context.Context,
		lc fx.Lifecycle,
		bus *pubsub.Bus,
		notifier *notify.Notifier,
		postQuerier *post_querier.Querier,
	) {
		consumer := func(hctx context.Context) error {
			_, err := pubsub.Subscribe(ctx, bus, "post_like_notify.post_liked", func(ctx context.Context, evt *message.EventPostLiked) error {
				authorID, err := postQuerier.AuthorID(ctx, evt.PostID)
				if err != nil {
					return err
				}

				if authorID == evt.LikerID {
					return nil
				}

				itemKind := datagraph.KindPost
				if evt.PostID != evt.RootPostID {
					itemKind = datagraph.KindReply
				}

				return notifier.Send(ctx,
					authorID,
					opt.New(evt.LikerID),
					notification.EventPostLike,
					&datagraph.Ref{
						ID:   xid.ID(evt.PostID),
						Kind: itemKind,
					},
				)
			})
			return err
		}

		lc.Append(fx.StartHook(consumer))
	})
}
