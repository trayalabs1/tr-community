package channel_test

import (
	"context"
	"testing"
	"time"

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

func TestChannelThreadListPersonalized(t *testing.T) {
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

			ownerCtx, owner := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			otherCtx, other := e2e.WithAccount(root, aw, seed.Account_002_Frigg)

			ownerSession := sh.WithSession(ownerCtx)
			memberSession := sh.WithSession(memberCtx)
			otherSession := sh.WithSession(otherCtx)

			memberID := member.ID
			otherID := other.ID
			_ = memberID
			_ = otherID
			_ = owner

			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Personalized Channel",
				Slug:        "personalized-channel",
				Description: "channel for personalized feed test",
			}, ownerSession)
			tests.Ok(t, err, createResp)
			channelID := createResp.JSON200.Id

			addMemberResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(member.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			tests.Ok(t, err, addMemberResp)

			addOtherResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(other.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			tests.Ok(t, err, addOtherResp)

			catResp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        "General",
				Description: "general",
				Colour:      "#abcdef",
			}, ownerSession)
			tests.Ok(t, err, catResp)
			categoryID := catResp.JSON200.Id

			t.Run("no recent self-post returns empty arrays", func(t *testing.T) {
				resp, err := cl.ChannelThreadListPersonalizedWithResponse(root, channelID, memberSession)
				tests.Ok(t, err, resp)
				r.NotNil(resp.JSON200)
				r.Empty(resp.JSON200.SelfRecent)
				r.Empty(resp.JSON200.Similar)
			})

			selfPublished, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "My recent thread",
				Body:       opt.New("<p>my body</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, memberSession)
			tests.Ok(t, err, selfPublished)
			selfPublishedID := selfPublished.JSON200.Id

			selfReview, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "My in-review thread",
				Body:       opt.New("<p>review body</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Review).Ptr(),
			}, memberSession)
			tests.Ok(t, err, selfReview)
			selfReviewID := selfReview.JSON200.Id

			t.Run("self_recent includes published and review threads", func(t *testing.T) {
				resp, err := cl.ChannelThreadListPersonalizedWithResponse(root, channelID, memberSession)
				tests.Ok(t, err, resp)

				ids := map[string]struct{}{}
				for _, th := range resp.JSON200.SelfRecent {
					ids[th.Id] = struct{}{}
				}
				r.Contains(ids, selfPublishedID)
				r.Contains(ids, selfReviewID)
			})

			t.Run("similar group is empty when self post is unscored", func(t *testing.T) {
				resp, err := cl.ChannelThreadListPersonalizedWithResponse(root, channelID, memberSession)
				tests.Ok(t, err, resp)

				for _, group := range resp.JSON200.Similar {
					r.Empty(group.Threads, "unscored self post should have empty similar group")
				}
			})

			matchTag := "positive"
			matchTopic := "scalp_health"
			mustSetSentiment(t, db, root, selfPublishedID, matchTag, matchTopic)

			otherSimilar, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Other matching thread",
				Body:       opt.New("<p>other body</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, otherSession)
			tests.Ok(t, err, otherSimilar)
			otherSimilarID := otherSimilar.JSON200.Id
			mustSetSentiment(t, db, root, otherSimilarID, matchTag, matchTopic)

			otherDifferent, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Different topic",
				Body:       opt.New("<p>different</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, otherSession)
			tests.Ok(t, err, otherDifferent)
			otherDifferentID := otherDifferent.JSON200.Id
			mustSetSentiment(t, db, root, otherDifferentID, matchTag, "different_topic")

			otherDraft, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Draft thread",
				Body:       opt.New("<p>draft</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Draft).Ptr(),
			}, otherSession)
			tests.Ok(t, err, otherDraft)
			otherDraftID := otherDraft.JSON200.Id
			mustSetSentiment(t, db, root, otherDraftID, matchTag, matchTopic)

			t.Run("similar returns matching published thread and excludes mismatches", func(t *testing.T) {
				resp, err := cl.ChannelThreadListPersonalizedWithResponse(root, channelID, memberSession)
				tests.Ok(t, err, resp)

				var group *openapi.ChannelPersonalizedSimilarGroup
				for i := range resp.JSON200.Similar {
					if resp.JSON200.Similar[i].ForThreadId == openapi.Identifier(selfPublishedID) {
						group = &resp.JSON200.Similar[i]
						break
					}
				}
				r.NotNil(group, "expected similar group anchored on self published thread")

				ids := map[string]struct{}{}
				for _, th := range group.Threads {
					ids[th.Id] = struct{}{}
				}
				r.Contains(ids, otherSimilarID, "matching published thread must appear")
				r.NotContains(ids, otherDifferentID, "different topic must be excluded")
				r.NotContains(ids, otherDraftID, "non-published must be excluded")
				r.NotContains(ids, selfPublishedID, "anchor thread itself must be excluded")
			})

			t.Run("similar returns at most two threads", func(t *testing.T) {
				for i := 0; i < 3; i++ {
					extra, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      "extra-" + xid.New().String(),
						Body:       opt.New("<p>extra</p>").Ptr(),
						Category:   opt.New(categoryID).Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
					}, otherSession)
					tests.Ok(t, err, extra)
					mustSetSentiment(t, db, root, extra.JSON200.Id, matchTag, matchTopic)
				}

				resp, err := cl.ChannelThreadListPersonalizedWithResponse(root, channelID, memberSession)
				tests.Ok(t, err, resp)

				for _, group := range resp.JSON200.Similar {
					if group.ForThreadId != openapi.Identifier(selfPublishedID) {
						continue
					}
					r.LessOrEqual(len(group.Threads), 2, "similar limit must be 2")
				}
			})
		}))
	}))
}

func mustSetSentiment(t *testing.T, db *ent.Client, ctx context.Context, postID string, tag string, topic string) {
	t.Helper()

	pid, err := xid.FromString(postID)
	require.NoError(t, err)

	existing, err := db.PostSentiment.Query().
		Where(ent_post_sentiment.PostID(pid)).
		First(ctx)
	if err == nil && existing != nil {
		_, err = db.PostSentiment.UpdateOneID(existing.ID).
			SetSentimentTag(tag).
			SetPrimaryTopic(topic).
			SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
			SetUpdatedAt(time.Now()).
			Save(ctx)
		require.NoError(t, err)
		return
	}

	_, err = db.PostSentiment.Create().
		SetPostID(pid).
		SetSentimentTag(tag).
		SetPrimaryTopic(topic).
		SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
		Save(ctx)
	require.NoError(t, err)
}
