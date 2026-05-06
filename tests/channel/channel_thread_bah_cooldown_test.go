package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/resources/settings"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
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
		settingsRepo *settings.SettingsRepository,
		bus *pubsub.Bus,
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

			channelC, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "BAH Channel C",
				Slug:        "bah-channel-c",
				Description: "channel C",
			}, ownerSession)
			tests.Ok(t, err, channelC)
			channelCID := channelC.JSON200.Id

			channelD, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "BAH Channel D",
				Slug:        "bah-channel-d",
				Description: "channel D",
			}, ownerSession)
			tests.Ok(t, err, channelD)
			channelDID := channelD.JSON200.Id

			channelE, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "BAH Channel E",
				Slug:        "bah-channel-e",
				Description: "channel E",
			}, ownerSession)
			tests.Ok(t, err, channelE)
			channelEID := channelE.JSON200.Id

			for _, ch := range []openapi.Identifier{channelAID, channelBID, channelCID, channelDID, channelEID} {
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

			catC, err := cl.ChannelCategoryCreateWithResponse(root, channelCID, openapi.CategoryInitialProps{
				Name:        "General C",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catC)
			catCID := catC.JSON200.Id

			catD, err := cl.ChannelCategoryCreateWithResponse(root, channelDID, openapi.CategoryInitialProps{
				Name:        "General D",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catD)
			catDID := catD.JSON200.Id

			catE, err := cl.ChannelCategoryCreateWithResponse(root, channelEID, openapi.CategoryInitialProps{
				Name:        "General E",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catE)
			catEID := catE.JSON200.Id

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

			t.Run("review-state BAH does not cascade subsequent BAHs into review", func(t *testing.T) {
				updated, err := settingsRepo.Set(root, settings.Settings{
					Services: opt.New(settings.ServiceSettings{
						Moderation: opt.New(settings.ModerationServiceSettings{
							WordBlockList:  opt.New([]string{}),
							WordReportList: opt.New([]string{"flaggedterm"}),
						}),
					}),
				})
				r.NoError(err)
				bus.Publish(root, &message.EventSettingsUpdated{Settings: updated})
				time.Sleep(100 * time.Millisecond)

				flagged, err := cl.ChannelThreadCreateWithResponse(root, channelCID, openapi.ThreadInitialProps{
					Title:      "Flagged BAH",
					Body:       opt.New("<p>this contains flaggedterm content</p>").Ptr(),
					Category:   opt.New(catCID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, flagged)
				r.Equal(openapi.Review, flagged.JSON200.Visibility,
					"flagged BAH should land in review (no prior BAH so cooldown is irrelevant)")

				cleared, err := settingsRepo.Set(root, settings.Settings{
					Services: opt.New(settings.ServiceSettings{
						Moderation: opt.New(settings.ModerationServiceSettings{
							WordBlockList:  opt.New([]string{}),
							WordReportList: opt.New([]string{}),
						}),
					}),
				})
				r.NoError(err)
				bus.Publish(root, &message.EventSettingsUpdated{Settings: cleared})
				time.Sleep(100 * time.Millisecond)

				next, err := cl.ChannelThreadCreateWithResponse(root, channelCID, openapi.ThreadInitialProps{
					Title:      "Next BAH after review",
					Body:       opt.New("<p>clean content</p>").Ptr(),
					Category:   opt.New(catCID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, next)
				r.Equal(openapi.Published, next.JSON200.Visibility,
					"a BAH whose only prior BAH in window is in review must stay published")
			})

			t.Run("review-submitted BAH auto-promotes to published when no recent published BAH", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelDID, openapi.ThreadInitialProps{
					Title:      "Review-submitted BAH",
					Body:       opt.New("<p>review submitted</p>").Ptr(),
					Category:   opt.New(catDID).Ptr(),
					Visibility: opt.New(openapi.Review).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility,
					"a Review-submitted BAH must be auto-promoted when no published BAH exists in window")
			})

			t.Run("review-submitted BAH stays in review when a recent published BAH exists", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelDID, openapi.ThreadInitialProps{
					Title:      "Second review-submitted BAH",
					Body:       opt.New("<p>second review</p>").Ptr(),
					Category:   opt.New(catDID).Ptr(),
					Visibility: opt.New(openapi.Review).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility,
					"a Review-submitted BAH must stay in review when a published BAH is within window")
			})

			t.Run("review-submitted BAH auto-promotes in a fresh channel even after promotes elsewhere", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelEID, openapi.ThreadInitialProps{
					Title:      "Channel E review-submitted BAH",
					Body:       opt.New("<p>fresh channel</p>").Ptr(),
					Category:   opt.New(catEID).Ptr(),
					Visibility: opt.New(openapi.Review).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility,
					"cooldown is per-channel: fresh channel must auto-promote regardless of other channels")
			})

			t.Run("review-submitted non-BAH stays in review", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelEID, openapi.ThreadInitialProps{
					Title:      "Review-submitted non-BAH",
					Body:       opt.New("<p>not a bah</p>").Ptr(),
					Category:   opt.New(catEID).Ptr(),
					Visibility: opt.New(openapi.Review).Ptr(),
					Meta:       &plainMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility,
					"non-BAH posts submitted as Review must remain in review (auto-promote is BAH-only)")
			})
		}))
	}))
}
