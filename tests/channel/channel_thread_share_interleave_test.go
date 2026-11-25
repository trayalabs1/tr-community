package channel_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func metaCategory(t openapi.ThreadReference) string {
	if t.Meta == nil {
		return ""
	}
	v, _ := (*t.Meta)["post_category"].(string)
	return v
}

func metaType(t openapi.ThreadReference) string {
	if t.Meta == nil {
		return ""
	}
	v, _ := (*t.Meta)["type"].(string)
	return v
}

func TestChannelThreadShareInterleave(t *testing.T) {
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

			ownerCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			ownerSession := sh.WithSession(ownerCtx)

			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Share Interleave Channel",
				Slug:        "share-interleave-channel",
				Description: "Feed share interleave test",
			}, ownerSession)
			tests.Ok(t, err, createResp)
			channelID := createResp.JSON200.Id

			// 20 organic posts (non-share categories). Created oldest-first, so the
			// created_at DESC organic ordering yields o19, o18, ... o0.
			for i := 0; i < 20; i++ {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
					Title:      fmt.Sprintf("Organic %d", i),
					Body:       opt.New(fmt.Sprintf("<p>organic %d</p>", i)).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
				}, ownerSession)
				tests.Ok(t, err, resp)
			}

			// One share per rotation step, in order:
			// feedback/story, feedback/progress, BAH/21, tip. Distinct
			// (category,type) pairs so none trips another's cooldown into review.
			storyShare, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Share Story",
				Body:       opt.New("<p>story</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
				Meta:       &openapi.Metadata{"post_category": "feedback", "type": "story"},
			}, ownerSession)
			tests.Ok(t, err, storyShare)

			progShare, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Share Progress",
				Body:       opt.New("<p>progress</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
				Meta:       &openapi.Metadata{"post_category": "feedback", "type": "progress"},
			}, ownerSession)
			tests.Ok(t, err, progShare)

			bah21Share, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Share BAH 21",
				Body:       opt.New("<p>bah 21</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
				Meta:       &openapi.Metadata{"post_category": "BAH", "type": "21"},
			}, ownerSession)
			tests.Ok(t, err, bah21Share)

			tipShare, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Share Tip",
				Body:       opt.New("<p>tip</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
				Meta:       &openapi.Metadata{"post_category": "tip"},
			}, ownerSession)
			tests.Ok(t, err, tipShare)

			listResp, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, listResp)

			threads := listResp.JSON200.Threads
			r.GreaterOrEqual(len(threads), 20, "expected at least 20 items in the merged feed")

			// Positions 5, 10, 15, 20 (1-indexed) are share slots in rotation order.
			r.Equal("feedback", metaCategory(threads[4]), "slot 5 -> feedback share")
			r.Equal("story", metaType(threads[4]), "slot 5 -> feedback/story")

			r.Equal("feedback", metaCategory(threads[9]), "slot 10 -> feedback share")
			r.Equal("progress", metaType(threads[9]), "slot 10 -> feedback/progress")

			r.Equal("BAH", metaCategory(threads[14]), "slot 15 -> BAH share")
			r.Equal("21", metaType(threads[14]), "slot 15 -> BAH/21")

			r.Equal("tip", metaCategory(threads[19]), "slot 20 -> tip share")

			// Every non-share slot must be an organic (non-share-category) post.
			shareSlots := map[int]bool{4: true, 9: true, 14: true, 19: true}
			shareCats := map[string]bool{"BAH": true, "feedback": true, "tip": true}
			for i := 0; i < 20; i++ {
				if shareSlots[i] {
					continue
				}
				r.False(shareCats[metaCategory(threads[i])], "organic slot %d must not be a share", i)
			}

			// The ordered sequence is cached; a second identical request must
			// return the exact same order (served from cache).
			listResp2, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, listResp2)
			threads2 := listResp2.JSON200.Threads
			r.Equal(len(threads), len(threads2), "cached request returns same count")
			for i := range threads {
				r.Equal(threads[i].Id, threads2[i].Id, "cached order matches at index %d", i)
			}

			// A filtered request bypasses the cache and the interleave: excluding
			// BAH means the BAH share must not occupy slot 5.
			excludeBah := true
			filtered, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{
				ExcludeBah: &excludeBah,
			}, ownerSession)
			tests.Ok(t, err, filtered)
			for _, th := range filtered.JSON200.Threads {
				r.NotEqual("BAH", metaCategory(th), "exclude_bah request must not contain BAH posts")
			}
		}))
	}))
}
