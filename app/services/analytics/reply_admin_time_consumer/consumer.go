package reply_admin_time_consumer

import (
	"context"
	"log/slog"

	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/ent"
	ent_account "github.com/Southclaws/storyden/internal/ent/account"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

func Build() fx.Option {
	return fx.Invoke(func(
		ctx context.Context,
		lc fx.Lifecycle,
		bus *pubsub.Bus,
		logger *slog.Logger,
		db *ent.Client,
	) {
		consumer := func(_ context.Context) error {
			_, err := pubsub.Subscribe(ctx, bus, "analytics.admin_replied", func(ctx context.Context, evt *message.EventAdminReplied) error {
				parentPost, err := db.Post.Query().
					Where(ent_post.IDEQ(xid.ID(evt.ParentPostID))).
					Select(ent_post.FieldAccountPosts, ent_post.FieldCreatedAt).
					Only(ctx)
				if err != nil {
					logger.WarnContext(ctx, "analytics consumer: failed to probe parent post",
						slog.String("parent_post_id", evt.ParentPostID.String()),
						slog.String("error", err.Error()))
					return nil
				}

				parentAuthor, err := db.Account.Query().
					Where(ent_account.IDEQ(parentPost.AccountPosts)).
					Select(ent_account.FieldAdmin).
					Only(ctx)
				if err != nil {
					logger.WarnContext(ctx, "analytics consumer: failed to fetch parent author",
						slog.String("error", err.Error()))
					return nil
				}
				if parentAuthor.Admin {
					return nil
				}

				adminAcc, err := db.Account.Query().
					Where(ent_account.IDEQ(xid.ID(evt.AdminAccountID))).
					Select(ent_account.FieldHandle).
					Only(ctx)
				if err != nil {
					logger.WarnContext(ctx, "analytics consumer: failed to fetch admin account",
						slog.String("error", err.Error()))
					return nil
				}

				timeDiff := int64(evt.AdminPostTime.Sub(parentPost.CreatedAt).Seconds())
				if timeDiff < 0 {
					timeDiff = 0
				}

				_, err = db.AdminReplyTime.Create().
					SetUserPostID(xid.ID(evt.ParentPostID)).
					SetAdminPostID(xid.ID(evt.AdminPostID)).
					SetUserPostTime(parentPost.CreatedAt).
					SetAdminPostTime(evt.AdminPostTime).
					SetAdminAccountID(xid.ID(evt.AdminAccountID)).
					SetAdminHandle(adminAcc.Handle).
					SetTimeDifferenceSeconds(timeDiff).
					Save(ctx)
				if err != nil {
					if ent.IsConstraintError(err) {
						return nil
					}
					return err
				}
				return nil
			})
			if err != nil {
				logger.WarnContext(ctx, "analytics consumer: failed to subscribe", slog.String("error", err.Error()))
			}
			return nil
		}
		lc.Append(fx.StartHook(consumer))
	})
}
