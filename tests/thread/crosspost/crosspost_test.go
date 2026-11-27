package crosspost_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/services/thread/crosspost"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	ent_post_sentiment "github.com/Southclaws/storyden/internal/ent/postsentiment"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
	"github.com/Southclaws/storyden/tests"
)

func TestCrosspostLifecycle(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		cl *openapi.ClientWithResponses,
		sh *e2e.SessionHelper,
		db *ent.Client,
		aw *account_writer.Writer,
		cp *crosspost.Crossposter,
		pq *post_querier.Querier,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin,
				account_writer.WithMetadata(map[string]any{"case_id": "0123"}))
			adminSession := sh.WithSession(adminCtx)

			authorCtx, author := e2e.WithAccount(root, aw, seed.Account_004_Loki,
				account_writer.WithMetadata(map[string]any{"case_id": "0456"}))
			authorSession := sh.WithSession(authorCtx)

			leadCtx, lead := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			leadSession := sh.WithSession(leadCtx)

			source := tests.AssertRequest(cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Month 5 Masters", Slug: "month-5-masters", Description: "source",
			}, adminSession))(t, http.StatusOK)
			target := tests.AssertRequest(cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name: "Traya Explorers", Slug: crosspost.TargetChannelSlug, Description: "leads",
			}, adminSession))(t, http.StatusOK)

			tests.AssertRequest(cl.ChannelMemberAddWithResponse(root, target.JSON200.Id, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(lead.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession))(t, http.StatusOK)
			tests.AssertRequest(cl.ChannelMemberAddWithResponse(root, source.JSON200.Id, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(author.ID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, adminSession))(t, http.StatusOK)

			targetID, err := xid.FromString(target.JSON200.Id)
			r.NoError(err)

			created := tests.AssertRequest(cl.ChannelThreadCreateWithResponse(root, source.JSON200.Id, openapi.ThreadInitialProps{
				Title:      "Visible regrowth after five months",
				Body:       opt.New[openapi.PostContent]("<p>Baby hairs everywhere.</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, authorSession))(t, http.StatusOK)
			originalID, err := xid.FromString(created.JSON200.Id)
			r.NoError(err)

			adminPost := tests.AssertRequest(cl.ChannelThreadCreateWithResponse(root, source.JSON200.Id, openapi.ThreadInitialProps{
				Title:      "Coach progress showcase",
				Body:       opt.New[openapi.PostContent]("<p>Look at these results.</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession))(t, http.StatusOK)
			adminPostID, err := xid.FromString(adminPost.JSON200.Id)
			r.NoError(err)

			control := tests.AssertRequest(cl.ChannelThreadCreateWithResponse(root, source.JSON200.Id, openapi.ThreadInitialProps{
				Title:      "A tip nobody classified",
				Body:       opt.New[openapi.PostContent]("<p>Drink water.</p>").Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession))(t, http.StatusOK)
			controlID, err := xid.FromString(control.JSON200.Id)
			r.NoError(err)

			score := func(id xid.ID, tag, topic string) {
				err := db.PostSentiment.Create().
					SetPostID(id).
					SetSentimentTag(tag).
					SetPrimaryTopic(topic).
					SetPositivityScore(90).
					SetFeedValueScore(80).
					SetRankScore(1.5).
					SetScoringStatus(ent_post_sentiment.ScoringStatusScored).
					OnConflictColumns(ent_post_sentiment.FieldPostID).
					UpdateNewValues().
					Exec(root)
				r.NoError(err)
			}

			copiesOf := func(id xid.ID) []*ent.Post {
				ps, err := db.Post.Query().
					Where(ent_post.ChannelID(targetID), ent_post.DeletedAtIsNil()).
					All(root)
				r.NoError(err)
				var out []*ent.Post
				for _, p := range ps {
					if p.Metadata[post.MetaCrosspostSourceID] == id.String() {
						out = append(out, p)
					}
				}
				return out
			}

			t.Run("ineligible posts are not copied", func(t *testing.T) {
				score(controlID, "positive", "TIPS & EXPERIENCES")
				r.NoError(cp.OnScored(root, controlID))
				r.Empty(copiesOf(controlID))

				score(adminPostID, "positive", "RESULTS & PROGRESS")
				r.NoError(cp.OnScored(root, adminPostID))
				r.NoError(cp.OnPublished(root, adminPostID))
				r.Empty(copiesOf(adminPostID))

				score(originalID, "neutral", "RESULTS & PROGRESS")
				r.NoError(cp.OnScored(root, originalID))
				r.Empty(copiesOf(originalID))
			})

			var copyID xid.ID

			t.Run("eligible post is copied once", func(t *testing.T) {
				score(originalID, "positive", "RESULTS & PROGRESS")
				r.NoError(cp.OnScored(root, originalID))
				r.NoError(cp.OnScored(root, originalID))
				r.NoError(cp.OnPublished(root, originalID))

				copies := copiesOf(originalID)
				r.Len(copies, 1)
				c := copies[0]
				copyID = c.ID

				r.Equal(xid.ID(author.ID), c.AccountPosts)
				r.Equal("Visible regrowth after five months", c.Title)
				r.Equal(ent_post.VisibilityPublished, c.Visibility)
				r.Nil(c.ReferencePostID)
				r.Equal(source.JSON200.Id, c.Metadata[post.MetaCrosspostSourceChannelID])
				r.Equal("Month 5 Masters", c.Metadata[post.MetaCrosspostSourceChannelName])

				s, err := db.PostSentiment.Query().Where(ent_post_sentiment.PostID(c.ID)).Only(root)
				r.NoError(err)
				r.Equal("RESULTS & PROGRESS", *s.PrimaryTopic)
			})

			t.Run("copy exposes the cohort channel", func(t *testing.T) {
				got := tests.AssertRequest(cl.ThreadGetWithResponse(root, copyID.String(), nil, leadSession))(t, http.StatusOK)
				r.NotNil(got.JSON200.CohortChannel)
				r.Equal("Month 5 Masters", got.JSON200.CohortChannel.Name)
				r.Equal(target.JSON200.Id, got.JSON200.ChannelId)
			})

			t.Run("notification guards can tell the copy from a normal thread", func(t *testing.T) {
				isCopy, err := pq.IsCrosspost(root, post.ID(copyID))
				r.NoError(err)
				r.True(isCopy)

				isOriginal, err := pq.IsCrosspost(root, post.ID(originalID))
				r.NoError(err)
				r.False(isOriginal)
			})

			t.Run("lead can reply to the copy and the original stays untouched", func(t *testing.T) {
				tests.AssertRequest(cl.ReplyCreateWithResponse(root, copyID.String(), openapi.ReplyInitialProps{
					Body: "<p>Amazing, what did you change?</p>",
				}, leadSession))(t, http.StatusOK)

				orig, err := db.Post.Query().Where(ent_post.RootPostID(originalID)).Count(root)
				r.NoError(err)
				r.Zero(orig)

				onCopy, err := db.Post.Query().Where(ent_post.RootPostID(copyID)).Count(root)
				r.NoError(err)
				r.Equal(1, onCopy)
			})

			t.Run("customer profile listing hides the copy", func(t *testing.T) {
				list := tests.AssertRequest(cl.ThreadListWithResponse(root, &openapi.ThreadListParams{
					Author: &author.Handle,
				}, authorSession))(t, http.StatusOK)

				ids := make([]string, len(list.JSON200.Threads))
				for i, th := range list.JSON200.Threads {
					ids[i] = th.Id
				}
				r.Contains(ids, originalID.String())
				r.NotContains(ids, copyID.String())
			})

			visibilityOfCopy := func() ent_post.Visibility {
				c, err := db.Post.Get(root, copyID)
				r.NoError(err)
				return c.Visibility
			}

			update := func(v openapi.Visibility) {
				tests.AssertRequest(cl.ThreadUpdateWithResponse(root, created.JSON200.Slug, openapi.ThreadMutableProps{
					Visibility: &v,
				}, adminSession))(t, http.StatusOK)
			}

			t.Run("copy mirrors visibility of the original and is never deleted", func(t *testing.T) {
				update(openapi.Review)
				r.NoError(cp.OnUnpublished(root, originalID))
				r.Equal(ent_post.VisibilityReview, visibilityOfCopy())

				update(openapi.Published)
				r.NoError(cp.OnPublished(root, originalID))
				r.Equal(ent_post.VisibilityPublished, visibilityOfCopy())
				r.Len(copiesOf(originalID), 1)

				tests.AssertRequest(cl.ThreadDeleteWithResponse(root, created.JSON200.Slug, adminSession))(t, http.StatusOK)
				r.NoError(cp.OnDeleted(root, originalID))
				r.Equal(ent_post.VisibilityArchived, visibilityOfCopy())

				c, err := db.Post.Get(root, copyID)
				r.NoError(err)
				r.Nil(c.DeletedAt)
			})
		}))
	}))
}
