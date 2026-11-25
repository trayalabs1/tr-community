package like_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestBulkLikePosts(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			userCtx, userAcc := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			adminSession := sh.WithSession(adminCtx)
			userSession := sh.WithSession(userCtx)

			suffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "bulk-like-" + suffix,
				Slug:        "bulk-like-" + suffix,
				Description: "channel for bulk like tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			published := openapi.Published
			t1, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Bulk like thread 1",
				Body:       opt.New[openapi.PostContent]("one").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, t1)
			t2, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Bulk like thread 2",
				Body:       opt.New[openapi.PostContent]("two").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, t2)

			t.Run("single_request_likes_all_posts", func(t *testing.T) {
				resp, err := cl.LikePostAddManyWithResponse(root, openapi.LikePostAddManyJSONRequestBody{
					PostIds: []openapi.Identifier{t1.JSON200.Id, t2.JSON200.Id},
				}, userSession)
				tests.Ok(t, err, resp)
				r.NotNil(resp.JSON200)
				a.Equal(2, resp.JSON200.Requested)
				a.Equal(2, resp.JSON200.Liked)

				get1, err := cl.LikePostGetWithResponse(root, t1.JSON200.Id, adminSession)
				tests.Ok(t, err, get1)
				r.Len(get1.JSON200.Likes, 1)
				a.Equal(userAcc.ID.String(), get1.JSON200.Likes[0].Owner.Id)

				get2, err := cl.LikePostGetWithResponse(root, t2.JSON200.Id, adminSession)
				tests.Ok(t, err, get2)
				r.Len(get2.JSON200.Likes, 1)
				a.Equal(userAcc.ID.String(), get2.JSON200.Likes[0].Owner.Id)
			})

			t.Run("idempotent_on_repeat", func(t *testing.T) {
				resp, err := cl.LikePostAddManyWithResponse(root, openapi.LikePostAddManyJSONRequestBody{
					PostIds: []openapi.Identifier{t1.JSON200.Id},
				}, userSession)
				tests.Ok(t, err, resp)
				a.Equal(1, resp.JSON200.Liked)

				get1, err := cl.LikePostGetWithResponse(root, t1.JSON200.Id, adminSession)
				tests.Ok(t, err, get1)
				r.Len(get1.JSON200.Likes, 1)
			})
		}))
	}))
}
