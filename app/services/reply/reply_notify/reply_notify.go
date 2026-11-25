package reply_notify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Southclaws/opt"
	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/account/notification"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/services/notification/notify"
	"github.com/Southclaws/storyden/internal/infrastructure/moengage"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
	"github.com/rs/xid"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
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
			_, err := pubsub.Subscribe(ctx, bus, "reply_notify.reply_created", func(ctx context.Context, evt *message.EventThreadReplyCreated) error {
				errs := []error{}

				_, channelID, err := postQuerier.AuthorAndChannelID(ctx, evt.ThreadID)
				if err != nil {
					return err
				}

				if evt.ReplyAuthorID != evt.ThreadAuthorID {
					err := notifier.Send(ctx,
						evt.ThreadAuthorID,
						opt.New(evt.ReplyAuthorID),
						notification.EventThreadReply,
						&datagraph.Ref{
							ID:   xid.ID(evt.ThreadID),
							Kind: datagraph.KindPost,
						},
					)
					errs = append(errs, err)

					errs = append(errs, sendMoengage(ctx, logger, moengageSender, accountQuerier, evt.ThreadAuthorID, evt.ReplyAuthorID, channelID, evt.ThreadID))
				}

				if rtid, ok := evt.ReplyToAuthorID.Get(); ok && rtid != evt.ReplyAuthorID {
					err := notifier.Send(ctx,
						rtid,
						opt.New(evt.ReplyAuthorID),
						notification.EventReplyToReply,
						&datagraph.Ref{
							ID:   xid.ID(evt.ReplyID),
							Kind: datagraph.KindReply,
						},
					)
					errs = append(errs, err)

					errs = append(errs, sendMoengage(ctx, logger, moengageSender, accountQuerier, rtid, evt.ReplyAuthorID, channelID, evt.ThreadID))
				}

				return errors.Join(errs...)
			})
			return err
		}

		lc.Append(fx.StartHook(consumer))
	})
}

func sendMoengage(
	ctx context.Context,
	logger *slog.Logger,
	moengageSender moengage.Sender,
	accountQuerier *account_querier.Querier,
	targetID account.AccountID,
	replyAuthorID account.AccountID,
	channelID xid.ID,
	threadID post.ID,
) error {
	if moengageSender == nil {
		return nil
	}

	var meta map[string]any
	var replyAuthorHandle string

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error {
		_, loadedMeta, err := accountQuerier.GetHandleAndMetadata(ctx, targetID)
		if err != nil {
			return err
		}
		meta = loadedMeta
		return nil
	})
	group.Go(func() error {
		handle, _, err := accountQuerier.GetHandleAndMetadata(ctx, replyAuthorID)
		if err != nil {
			return err
		}
		replyAuthorHandle = handle
		return nil
	})

	if err := group.Wait(); err != nil {
		logger.Warn("failed to load account data for moengage",
			slog.String("event", "community_card_replied_to"),
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
		go func() {
			// timeout context for moengage event after 10 seconds
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := moengageSender.AddEvent(timeoutCtx, moengage.EventData{
				CaseID: caseID,
				Event:  "community_card_replied_to",
				EventAttr: map[string]any{
					"channel_id": channelID.String(),
					"thread_id":  threadID.String(),
					"reply_by":   replyAuthorHandle,
				},
			}); err != nil {
				logger.Warn("failed to send moengage event",
					slog.String("event", "community_card_replied_to"),
					slog.String("error", err.Error()),
				)
			}
		}()
	}

	return nil
}
