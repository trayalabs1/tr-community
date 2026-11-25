package thread_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestThreadListChannelIDsFilter(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		aw *account_writer.Writer,
	) {
		lc.Append(fx.StartHook(func() {
			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminSession := sh.WithSession(adminCtx)

			suffix := xid.New().String()
			createChannel := func(name string) openapi.Identifier {
				resp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        name + "-" + suffix,
					Slug:        name + "-" + suffix,
					Description: "channel for channel_ids filter tests",
				}, adminSession)
				tests.Ok(t, err, resp)
				return resp.JSON200.Id
			}

			channelA := createChannel("chan-ids-a")
			channelB := createChannel("chan-ids-b")
			channelC := createChannel("chan-ids-c")

			published := openapi.Published
			createThread := func(channelID openapi.Identifier, title string) openapi.Identifier {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Title:      title,
					Body:       opt.New[openapi.PostContent](title + " body").Ptr(),
					Visibility: &published,
				}, adminSession)
				tests.Ok(t, err, resp)
				return resp.JSON200.Id
			}

			threadA := createThread(channelA, "in channel a")
			threadB := createThread(channelB, "in channel b")
			threadC := createThread(channelC, "in channel c")

			visibilities := openapi.VisibilityParam{openapi.Published}
			list := func(channels *[]openapi.Identifier) []openapi.Identifier {
				resp, err := cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Visibility: &visibilities,
					ChannelIds: channels,
				}, adminSession)
				tests.Ok(t, err, resp)
				return threadIDs(resp.JSON200.Threads)
			}

			t.Run("single_channel_returns_only_that_channel", func(t *testing.T) {
				ids := list(&[]openapi.Identifier{channelA})
				assert.Contains(t, ids, threadA)
				assert.NotContains(t, ids, threadB)
				assert.NotContains(t, ids, threadC)
			})

			t.Run("multiple_channels_returns_any_of_them", func(t *testing.T) {
				ids := list(&[]openapi.Identifier{channelA, channelC})
				assert.Contains(t, ids, threadA)
				assert.NotContains(t, ids, threadB)
				assert.Contains(t, ids, threadC)
			})

			t.Run("no_channel_filter_returns_all_channels", func(t *testing.T) {
				ids := list(nil)
				assert.Contains(t, ids, threadA)
				assert.Contains(t, ids, threadB)
				assert.Contains(t, ids, threadC)
			})
		}))
	}))
}
