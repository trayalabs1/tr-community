package post_like_notify

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"

	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/account/notification"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/services/notification/notify"
	"github.com/Southclaws/storyden/internal/infrastructure/moengage"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func Build() fx.Option {
	return fx.Invoke(func(
		ctx context.Context,
		lc fx.Lifecycle,
		bus *pubsub.Bus,
		logger *slog.Logger,
		notifier *notify.Notifier,
		moengageSender moengage.Sender,
		accountQuerier *account_querier.Querier,
		postQuerier *post_querier.Querier,
	) {
		consumer := func(hctx context.Context) error {
			_, err := pubsub.Subscribe(ctx, bus, "post_like_notify.post_liked", func(ctx context.Context, evt *message.EventPostLiked) error {
				authorID, channelID, err := postQuerier.AuthorAndChannelID(ctx, evt.PostID)
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

				err = notifier.Send(ctx,
					authorID,
					opt.New(evt.LikerID),
					notification.EventPostLike,
					&datagraph.Ref{
						ID:   xid.ID(evt.PostID),
						Kind: itemKind,
					},
				)
				if err != nil {
					return err
				}

				if moengageSender != nil {
					var meta map[string]any
					var likerHandle string

					// parallel fetch for account info
					group, _ := errgroup.WithContext(ctx)
					group.Go(func() error {
						_, loadedMeta, err := accountQuerier.GetHandleAndMetadata(ctx, authorID)
						if err != nil {
							return err
						}
						meta = loadedMeta
						return nil
					})
					group.Go(func() error {
						handle, _, err := accountQuerier.GetHandleAndMetadata(ctx, evt.LikerID)
						if err != nil {
							return err
						}
						likerHandle = handle
						return nil
					})

					if err := group.Wait(); err != nil {
						logger.Warn("failed to load account data for moengage",
							slog.String("event", "community_card_liked_to"),
							slog.String("error", err.Error()),
						)
						return nil
					}

					caseIDValue, ok := meta["case_id"]
					caseID := ""
					if ok && caseIDValue != nil {
						caseID = fmt.Sprint(caseIDValue)
					}
					if caseID != "" {
						// async send moengage event
						go func() {
							// timeout context for moengage event after 10 seconds
							timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()

							if err := moengageSender.AddEvent(timeoutCtx, moengage.EventData{
								CaseID: caseID,
								Event:  "community_card_liked_to",
								EventAttr: map[string]any{
									"channel_id": channelID.String(),
									"thread_id":  evt.PostID.String(),
									"liked_by":   likerHandle,
								},
							}); err != nil {
								logger.Warn("failed to send moengage event",
									slog.String("event", "community_card_liked_to"),
									slog.String("error", err.Error()),
								)
							}
						}()
					}
				}

				return nil
			})
			return err
		}

		lc.Append(fx.StartHook(consumer))
	})
}
