package thread_test

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

func TestReplyCreateMany(t *testing.T) {
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
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			channelResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "bulk-reply-" + suffix,
				Slug:        "bulk-reply-" + suffix,
				Description: "channel for bulk reply tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)
			channelID := channelResp.JSON200.Id

			published := openapi.Published
			t1, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Bulk reply thread 1",
				Body:       opt.New[openapi.PostContent]("one").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, t1)
			t2, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Bulk reply thread 2",
				Body:       opt.New[openapi.PostContent]("two").Ptr(),
				Visibility: &published,
			}, adminSession)
			tests.Ok(t, err, t2)

			type item = struct {
				Body       openapi.PostContent `json:"body"`
				ThreadMark openapi.ThreadMark  `json:"thread_mark"`
			}

			t.Run("single_request_creates_replies_with_distinct_bodies", func(t *testing.T) {
				resp, err := cl.ReplyCreateManyWithResponse(root, openapi.ReplyCreateManyJSONRequestBody{
					Items: []item{
						{ThreadMark: openapi.ThreadMark(t1.JSON200.Slug), Body: "reply alpha"},
						{ThreadMark: openapi.ThreadMark(t2.JSON200.Slug), Body: "reply beta"},
					},
				}, adminSession)
				tests.Ok(t, err, resp)
				r.NotNil(resp.JSON200)
				a.Equal(2, resp.JSON200.Requested)
				a.Equal(2, resp.JSON200.Created)

				get1, err := cl.ThreadGetWithResponse(root, t1.JSON200.Id, nil, adminSession)
				tests.Ok(t, err, get1)
				r.Len(get1.JSON200.Replies.Replies, 1)
				a.Contains(get1.JSON200.Replies.Replies[0].Body, "reply alpha")

				get2, err := cl.ThreadGetWithResponse(root, t2.JSON200.Id, nil, adminSession)
				tests.Ok(t, err, get2)
				r.Len(get2.JSON200.Replies.Replies, 1)
				a.Contains(get2.JSON200.Replies.Replies[0].Body, "reply beta")
			})

			t.Run("multiple_replies_same_thread_in_one_call", func(t *testing.T) {
				t3, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Title:      "Bulk reply thread 3",
					Body:       opt.New[openapi.PostContent]("three").Ptr(),
					Visibility: &published,
				}, adminSession)
				tests.Ok(t, err, t3)

				resp, err := cl.ReplyCreateManyWithResponse(root, openapi.ReplyCreateManyJSONRequestBody{
					Items: []item{
						{ThreadMark: openapi.ThreadMark(t3.JSON200.Slug), Body: "first"},
						{ThreadMark: openapi.ThreadMark(t3.JSON200.Slug), Body: "second"},
					},
				}, adminSession)
				tests.Ok(t, err, resp)
				a.Equal(2, resp.JSON200.Created)

				get3, err := cl.ThreadGetWithResponse(root, t3.JSON200.Id, nil, adminSession)
				tests.Ok(t, err, get3)
				r.Len(get3.JSON200.Replies.Replies, 2)
			})

			t.Run("invalid_thread_mark_is_skipped", func(t *testing.T) {
				resp, err := cl.ReplyCreateManyWithResponse(root, openapi.ReplyCreateManyJSONRequestBody{
					Items: []item{
						{ThreadMark: openapi.ThreadMark(t1.JSON200.Slug), Body: "another reply"},
						{ThreadMark: openapi.ThreadMark("nonexistent-thread-mark"), Body: "orphan reply"},
					},
				}, adminSession)
				tests.Ok(t, err, resp)
				a.Equal(2, resp.JSON200.Requested)
				a.Equal(1, resp.JSON200.Created)
			})
		}))
	}))
}
