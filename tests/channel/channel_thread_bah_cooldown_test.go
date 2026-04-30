package channel_test

import (
	"context"
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

func TestChannelThreadBAHCooldown(t *testing.T) {
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
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			otherCtx, other := e2e.WithAccount(root, aw, seed.Account_002_Frigg)

			ownerSession := sh.WithSession(ownerCtx)
			memberSession := sh.WithSession(memberCtx)
			otherSession := sh.WithSession(otherCtx)

			channelA, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "BAH Channel A",
				Slug:        "bah-channel-a",
				Description: "channel A",
			}, ownerSession)
			tests.Ok(t, err, channelA)
			channelAID := channelA.JSON200.Id

			channelB, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "BAH Channel B",
				Slug:        "bah-channel-b",
				Description: "channel B",
			}, ownerSession)
			tests.Ok(t, err, channelB)
			channelBID := channelB.JSON200.Id

			for _, ch := range []openapi.Identifier{channelAID, channelBID} {
				addMember, err := cl.ChannelMemberAddWithResponse(root, ch, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(member.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, ownerSession)
				tests.Ok(t, err, addMember)

				addOther, err := cl.ChannelMemberAddWithResponse(root, ch, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(other.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, ownerSession)
				tests.Ok(t, err, addOther)
			}

			catA, err := cl.ChannelCategoryCreateWithResponse(root, channelAID, openapi.CategoryInitialProps{
				Name:        "General A",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catA)
			catAID := catA.JSON200.Id

			catB, err := cl.ChannelCategoryCreateWithResponse(root, channelBID, openapi.CategoryInitialProps{
				Name:        "General B",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catB)
			catBID := catB.JSON200.Id

			bahMeta := openapi.Metadata{"post_category": "BAH"}
			plainMeta := openapi.Metadata{"post_category": "OTHER"}

			t.Run("first BAH stays published", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "First BAH",
					Body:       opt.New("<p>first</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility, "first BAH must stay published")
			})

			t.Run("second BAH same author within window goes to review", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Second BAH same author",
					Body:       opt.New("<p>second</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility, "second BAH within 12h must be review")
			})

			t.Run("BAH from different author in same channel within window goes to review", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "BAH other author",
					Body:       opt.New("<p>other</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, otherSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility, "channel-scoped: any author triggers cooldown")
			})

			t.Run("draft request stays draft regardless of cooldown", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "BAH draft",
					Body:       opt.New("<p>draft</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Draft).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Draft, resp.JSON200.Visibility, "draft must not be overridden")
			})

			t.Run("BAH in different channel stays published", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelBID, openapi.ThreadInitialProps{
					Title:      "BAH in channel B",
					Body:       opt.New("<p>b</p>").Ptr(),
					Category:   opt.New(catBID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility, "channel-scoped: other channel must not be affected")
			})

			t.Run("non-BAH thread is unaffected by cooldown", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Regular thread",
					Body:       opt.New("<p>regular</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &plainMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility, "non-BAH must not be gated by BAH cooldown")
			})
		}))
	}))
}
