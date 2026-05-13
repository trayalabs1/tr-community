package thread_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
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

func TestThreadQuickReplyChips(t *testing.T) {
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
			a := assert.New(t)

			adminCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			memberCtx, memberAcc := e2e.WithAccount(root, aw, seed.Account_003_Baldur)

			adminSession := sh.WithSession(adminCtx)
			memberSession := sh.WithSession(memberCtx)

			chanCreate := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "Chips testing channel",
					Slug:        "chips-testing-channel",
					Description: "channel for quick-reply-chips tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := chanCreate.JSON200.Id

			tests.AssertRequest(
				cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(memberAcc.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, adminSession),
			)(t, http.StatusOK)

			createThread := func(title string, meta *openapi.Metadata) string {
				resp := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      title,
						Body:       opt.New("<p>body</p>").Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
						Meta:       meta,
					}, memberSession),
				)(t, http.StatusOK)
				return resp.JSON200.Id
			}

			getThread := func(slug string) *openapi.Thread {
				resp := tests.AssertRequest(
					cl.ThreadGetWithResponse(root, slug, &openapi.ThreadGetParams{}, memberSession),
				)(t, http.StatusOK)
				return resp.JSON200
			}

			getThreadSlug := func(threadID string) string {
				p, err := xid.FromString(threadID)
				r.NoError(err)
				post, err := db.Post.Get(root, p)
				r.NoError(err)
				return post.Slug
			}

			bahMeta := openapi.Metadata{"post_category": "BAH"}

			t.Run("positive_progress_success_story_returns_3_candidates", func(t *testing.T) {
				id := createThread("Progress thread", nil)
				mustSetSentiment(t, db, root, id, "positive", "progress_success_story")

				th := getThread(getThreadSlug(id))
				r.NotNil(th.QuickReplyChips, "quick_reply_chips should be present")
				r.Len(th.QuickReplyChips.Candidates, 3)
				a.Contains(th.QuickReplyChips.Candidates, "Amazing progress 🙌")
				a.Contains(th.QuickReplyChips.Candidates, "So happy for you")
				a.Contains(th.QuickReplyChips.Candidates, "Keep going 💪")
			})

			t.Run("bah_meta_returns_habit_candidates_without_sentiment", func(t *testing.T) {
				id := createThread("BAH thread", &bahMeta)
				// no sentiment row written — BAH is identified by meta only.

				th := getThread(getThreadSlug(id))
				r.NotNil(th.QuickReplyChips, "BAH meta should surface chips regardless of sentiment")
				r.Len(th.QuickReplyChips.Candidates, 3)
				a.Contains(th.QuickReplyChips.Candidates, "Great consistency 🔥")
				a.Contains(th.QuickReplyChips.Candidates, "Keep the streak going")
				a.Contains(th.QuickReplyChips.Candidates, "You're doing great 💪")
			})

			t.Run("bah_meta_wins_even_when_sentiment_is_negative", func(t *testing.T) {
				id := createThread("Grumpy BAH thread", &bahMeta)
				mustSetSentiment(t, db, root, id, "negative", "side_effects_safety")

				th := getThread(getThreadSlug(id))
				r.NotNil(th.QuickReplyChips, "BAH meta should override negative sentiment")
				a.Contains(th.QuickReplyChips.Candidates, "Great consistency 🔥")
			})

			t.Run("positive_other_topic_returns_nothing", func(t *testing.T) {
				id := createThread("Side-effects thread", nil)
				mustSetSentiment(t, db, root, id, "positive", "side_effects_safety")

				th := getThread(getThreadSlug(id))
				a.Nil(th.QuickReplyChips, "non-matching topic must not surface chips")
			})

			t.Run("negative_matching_topic_returns_nothing", func(t *testing.T) {
				id := createThread("Negative progress thread", nil)
				mustSetSentiment(t, db, root, id, "negative", "progress_success_story")

				th := getThread(getThreadSlug(id))
				a.Nil(th.QuickReplyChips, "negative sentiment must not surface chips on progress_success_story")
			})

			t.Run("ai_build_a_habit_without_bah_meta_returns_nothing", func(t *testing.T) {
				id := createThread("AI habit thread", nil)
				mustSetSentiment(t, db, root, id, "positive", "build_a_habit")

				th := getThread(getThreadSlug(id))
				a.Nil(th.QuickReplyChips, "build_a_habit topic from the AI scorer must NOT trigger chips; only meta.post_category=BAH does")
			})

			t.Run("unscored_thread_without_bah_meta_returns_nothing", func(t *testing.T) {
				id := createThread("Unscored thread", nil)
				// no sentiment row written

				th := getThread(getThreadSlug(id))
				a.Nil(th.QuickReplyChips, "threads without sentiment and without BAH meta must not surface chips")
			})
		}))
	}))
}

// mustSetSentiment is a local copy of the helper in tests/channel/channel_thread_personalized_test.go,
// reproduced here so this test does not depend on the channel test package.
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
