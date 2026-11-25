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

// An image-only post has no text to derive a title from, so the client sends an
// empty title with a body containing just the <img>. This asserts the API
// accepts that and that the thread remains addressable afterwards.
func TestImageOnlyThreadCreate(t *testing.T) {
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
				Name:        "image-only-" + suffix,
				Slug:        "image-only-" + suffix,
				Description: "channel for image-only post tests",
			}, adminSession)
			tests.Ok(t, err, channelResp)

			imageBody := `<p></p><img src="https://example.com/a.jpg">`

			created, err := cl.ChannelThreadCreateWithResponse(root, channelResp.JSON200.Id, openapi.ThreadInitialProps{
				Title:      "",
				Body:       opt.New[openapi.PostContent](imageBody).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, created)

			a.Empty(created.JSON200.Title, "an image-only post has no title")
			r.NotEmpty(created.JSON200.Slug, "an empty title must still yield a usable slug")

			// The thread must be fetchable by the slug it was given.
			got, err := cl.ThreadGetWithResponse(root, created.JSON200.Slug, nil, adminSession)
			tests.Ok(t, err, got)
			a.Contains(got.JSON200.Body, "<img", "the image must survive round-tripping")

			// A second image-only post must not collide on slug with the first.
			second, err := cl.ChannelThreadCreateWithResponse(root, channelResp.JSON200.Id, openapi.ThreadInitialProps{
				Title:      "",
				Body:       opt.New[openapi.PostContent](imageBody).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, second)
			a.NotEqual(created.JSON200.Slug, second.JSON200.Slug,
				"two titleless posts must not share a slug")
		}))
	}))
}
