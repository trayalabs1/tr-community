package channel_test

import (
	"context"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestChannelThreadFeedbackCooldownAndSentiment(t *testing.T) {
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

			ownerCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			otherCtx, other := e2e.WithAccount(root, aw, seed.Account_002_Frigg)

			ownerSession := sh.WithSession(ownerCtx)
			memberSession := sh.WithSession(memberCtx)
			otherSession := sh.WithSession(otherCtx)

			channelA, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Feedback Channel A",
				Slug:        "feedback-channel-a",
				Description: "channel A",
			}, ownerSession)
			tests.Ok(t, err, channelA)
			channelAID := channelA.JSON200.Id

			channelB, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Feedback Channel B",
				Slug:        "feedback-channel-b",
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

			feedbackProgressMeta := openapi.Metadata{"post_category": "feedback", "type": "progress"}
			feedbackOtherMeta := openapi.Metadata{"post_category": "feedback", "type": "concern"}
			bahMeta := openapi.Metadata{"post_category": "BAH"}

			var firstFeedbackID string

			t.Run("first feedback/progress stays published and gets neutral sentiment + 95-105 rank", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "First feedback",
					Body:       opt.New("<p>first</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility, "first feedback must stay published")
				firstFeedbackID = resp.JSON200.Id

				pid, perr := xid.FromString(firstFeedbackID)
				r.NoError(perr)

				ps, qerr := db.PostSentiment.Query().
					Where(ent_post_sentiment.PostID(pid)).
					Only(root)
				r.NoError(qerr, "feedback post must have a sentiment row at creation")
				r.NotNil(ps.SentimentTag, "feedback must have a sentiment tag set at creation")
				r.Equal("neutral", *ps.SentimentTag, "feedback must be tagged neutral like BAH")
				r.Equal(ent_post_sentiment.ScoringStatusScored, ps.ScoringStatus,
					"feedback sentiment must be marked Scored to bypass the AI scorer")
				r.GreaterOrEqual(ps.RankScore, 95.0, "feedback rank must be in [95, 105]")
				r.LessOrEqual(ps.RankScore, 105.0, "feedback rank must be in [95, 105]")
			})

			t.Run("second feedback/progress within window goes to review (same author)", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Second feedback same author",
					Body:       opt.New("<p>second</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility, "second feedback within 12h must be review")
			})

			t.Run("feedback/progress from different author in same channel within window goes to review", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Feedback other author",
					Body:       opt.New("<p>other</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, otherSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility, "channel-scoped: any author triggers cooldown")
			})

			t.Run("feedback with different type does NOT trigger feedback/progress cooldown", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Feedback concern",
					Body:       opt.New("<p>concern</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackOtherMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility,
					"cooldown is scoped by (category, type): feedback/concern must not be gated by feedback/progress")
			})

			t.Run("BAH does NOT trigger feedback cooldown", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelBID, openapi.ThreadInitialProps{
					Title:      "BAH in channel B",
					Body:       opt.New("<p>bah</p>").Ptr(),
					Category:   opt.New(catBID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &bahMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility, "BAH must not be affected by feedback cooldown")

				followup, err := cl.ChannelThreadCreateWithResponse(root, channelBID, openapi.ThreadInitialProps{
					Title:      "Feedback after BAH",
					Body:       opt.New("<p>feedback after bah</p>").Ptr(),
					Category:   opt.New(catBID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, memberSession)
				tests.Ok(t, err, followup)
				r.Equal(openapi.Published, followup.JSON200.Visibility,
					"feedback/progress in a channel whose only recent post is BAH must stay published (separate buckets)")
			})

			t.Run("feedback/progress in different channel stays published", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelBID, openapi.ThreadInitialProps{
					Title:      "Feedback in channel B 2",
					Body:       opt.New("<p>b2</p>").Ptr(),
					Category:   opt.New(catBID).Ptr(),
					Visibility: opt.New(openapi.Published).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, otherSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Review, resp.JSON200.Visibility,
					"a feedback/progress already exists in channel B in window: must be review")
			})

			t.Run("review-submitted feedback auto-promotes to published when no recent published feedback/progress", func(t *testing.T) {
				freshCh, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "Feedback Channel C",
					Slug:        "feedback-channel-c",
					Description: "channel C",
				}, ownerSession)
				tests.Ok(t, err, freshCh)
				freshChID := freshCh.JSON200.Id

				addMember, err := cl.ChannelMemberAddWithResponse(root, freshChID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(member.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, ownerSession)
				tests.Ok(t, err, addMember)

				freshCat, err := cl.ChannelCategoryCreateWithResponse(root, freshChID, openapi.CategoryInitialProps{
					Name:        "General C",
					Description: "general",
					Colour:      "#abcdef",
				}, ownerSession)
				tests.Ok(t, err, freshCat)

				resp, err := cl.ChannelThreadCreateWithResponse(root, freshChID, openapi.ThreadInitialProps{
					Title:      "Review-submitted feedback",
					Body:       opt.New("<p>review submitted</p>").Ptr(),
					Category:   opt.New(freshCat.JSON200.Id).Ptr(),
					Visibility: opt.New(openapi.Review).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Published, resp.JSON200.Visibility,
					"a Review-submitted feedback/progress must be auto-promoted when no published one exists in window")
			})

			t.Run("draft feedback stays draft regardless of cooldown", func(t *testing.T) {
				resp, err := cl.ChannelThreadCreateWithResponse(root, channelAID, openapi.ThreadInitialProps{
					Title:      "Feedback draft",
					Body:       opt.New("<p>draft</p>").Ptr(),
					Category:   opt.New(catAID).Ptr(),
					Visibility: opt.New(openapi.Draft).Ptr(),
					Meta:       &feedbackProgressMeta,
				}, memberSession)
				tests.Ok(t, err, resp)
				r.Equal(openapi.Draft, resp.JSON200.Visibility, "draft must not be overridden")
			})
		}))
	}))
}
