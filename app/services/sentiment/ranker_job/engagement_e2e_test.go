package ranker_job_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/services/sentiment/ranker_job"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_likepost "github.com/Southclaws/storyden/internal/ent/likepost"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestGetDailyIncrements(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		db *ent.Client,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, adminAccount := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)
			adminAccountID := xid.ID(adminAccount.ID)

			// like_posts has a unique (account_id, post_id) constraint, so
			// each like on the same post needs a distinct account.
			_, liker1 := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			_, liker2 := e2e.WithAccount(root, aw, seed.Account_004_Loki)
			_, liker3 := e2e.WithAccount(root, aw, seed.Account_005_Þórr)
			likerIDs := []xid.ID{xid.ID(liker1.ID), xid.ID(liker2.ID), xid.ID(liker3.ID)}

			suffix := xid.New().String()
			channelResp := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "ranker-job-" + suffix,
					Slug:        "ranker-job-" + suffix,
					Description: "channel for ranker job engagement tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := channelResp.JSON200.Id

			mkThread := func(title string) (xid.ID, string) {
				resp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title,
						Body:       opt.New[openapi.PostContent]("body for " + title).Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
					}, adminSession),
				)(t, http.StatusOK)
				id, err := xid.FromString(resp.JSON200.Id)
				r.NoError(err)
				return id, resp.JSON200.Slug
			}

			withinWindow, withinWindowSlug := mkThread("within window " + suffix)
			outsideWindow, outsideWindowSlug := mkThread("outside window " + suffix)
			untouched, _ := mkThread("untouched " + suffix)

			since := time.Now().Add(-24 * time.Hour)

			// 3 likes inside the window on withinWindow -> +6.0 (3 * 2.0)
			for _, likerID := range likerIDs {
				_, err := db.LikePost.Create().
					SetPostID(withinWindow).
					SetAccountID(likerID).
					Save(root)
				r.NoError(err)
			}

			// 1 like backdated before the window on outsideWindow -> excluded
			backdated, err := db.LikePost.Create().
				SetPostID(outsideWindow).
				SetAccountID(adminAccountID).
				Save(root)
			r.NoError(err)
			err = db.LikePost.UpdateOneID(backdated.ID).
				Modify(func(u *sql.UpdateBuilder) {
					u.Set(ent_likepost.FieldCreatedAt, since.Add(-1*time.Hour))
				}).
				Exec(root)
			r.NoError(err)

			mkReply := func(rootSlug string) xid.ID {
				resp := tests.AssertRequest(
					cl.ReplyCreateWithResponse(root, rootSlug, openapi.ReplyInitialProps{
						Body: "a reply",
					}, adminSession),
				)(t, http.StatusOK)
				id, err := xid.FromString(resp.JSON200.Id)
				r.NoError(err)
				return id
			}

			// 2 replies inside the window on withinWindow -> +1.0 (2 * 0.5)
			mkReply(withinWindowSlug)
			mkReply(withinWindowSlug)

			// 1 reply backdated before the window on outsideWindow -> excluded
			outsideReplyID := mkReply(outsideWindowSlug)
			err = db.Post.UpdateOneID(outsideReplyID).
				Modify(func(u *sql.UpdateBuilder) {
					u.Set(ent_post.FieldCreatedAt, since.Add(-2*time.Hour))
				}).
				Exec(root)
			r.NoError(err)

			deltas, err := ranker_job.GetDailyIncrements(root, db, since)
			r.NoError(err)

			r.InDelta(6.0+1.0, deltas[withinWindow], 0.001, "withinWindow should get 3 likes*2 + 2 replies*0.5 = 7.0")
			_, ok := deltas[outsideWindow]
			r.False(ok, "outsideWindow's like/reply were both backdated before the window, must not appear")
			_, ok = deltas[untouched]
			r.False(ok, "untouched post has no engagement at all, must not appear")
		}))
	}))
}
