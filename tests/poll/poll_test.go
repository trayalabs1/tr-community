package poll_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Southclaws/opt"
	"github.com/samber/lo"
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

func TestPoll(t *testing.T) {
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

			adminCtx, adminAcc := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			user1Ctx, user1Acc := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			user2Ctx, _ := e2e.WithAccount(root, aw, seed.Account_004_Loki)
			memberCtx, memberAcc := e2e.WithAccount(root, aw, seed.Account_005_Þórr)

			adminSession := sh.WithSession(adminCtx)
			user1Session := sh.WithSession(user1Ctx)
			user2Session := sh.WithSession(user2Ctx)
			memberSession := sh.WithSession(memberCtx)

			chanCreate := tests.AssertRequest(
				cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
					Name:        "Poll testing channel",
					Slug:        "poll-testing-channel",
					Description: "channel for poll tests",
				}, adminSession),
			)(t, http.StatusOK)
			channelID := chanCreate.JSON200.Id

			tests.AssertRequest(
				cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(user1Acc.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, adminSession),
			)(t, http.StatusOK)

			tests.AssertRequest(
				cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
					AccountId: openapi.Identifier(memberAcc.ID.String()),
					Role:      openapi.ChannelMemberAddRoleMember,
				}, adminSession),
			)(t, http.StatusOK)

			catCreate := tests.AssertRequest(
				cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
					Colour:      "#ff0000",
					Description: "poll category",
					Name:        "Poll category",
				}, adminSession),
			)(t, http.StatusOK)
			categoryID := catCreate.JSON200.Id

			_ = adminAcc

			t.Run("non_admin_cannot_create_poll", func(t *testing.T) {
				meta := openapi.Metadata{
					"is_poll": true,
					"poll_options": []any{
						map[string]any{"id": "opt_a", "text": "Option A"},
						map[string]any{"id": "opt_b", "text": "Option B"},
					},
				}
				tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:    "Should fail poll",
						Category: opt.New(categoryID).Ptr(),
						Meta:     &meta,
					}, memberSession),
				)(t, http.StatusForbidden)
			})

			t.Run("poll_lifecycle", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				meta := openapi.Metadata{
					"is_poll": true,
					"poll_options": []any{
						map[string]any{"id": "opt_a", "text": "Option A"},
						map[string]any{"id": "opt_b", "text": "Option B"},
					},
				}

				pollCreate := tests.AssertRequest(
					cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
						Title:      "Poll thread",
						Category:   opt.New(categoryID).Ptr(),
						Visibility: opt.New(openapi.Published).Ptr(),
						Meta:       &meta,
					}, adminSession),
				)(t, http.StatusOK)
				r.NotNil(pollCreate.JSON200)
				a.Contains(pollCreate.JSON200.Slug, "poll-thread")

				slug := pollCreate.JSON200.Slug

				t.Run("initial_state_has_zero_votes", func(t *testing.T) {
					r := require.New(t)
					a := assert.New(t)

					pollGet := tests.AssertRequest(
						cl.ThreadGetPollWithResponse(root, slug, user1Session),
					)(t, http.StatusOK)
					r.NotNil(pollGet.JSON200)
					a.Equal(0, pollGet.JSON200.TotalVotes)
					a.True(pollGet.JSON200.UserVote.IsNull() || !pollGet.JSON200.UserVote.IsSpecified())
					r.Len(pollGet.JSON200.Options, 2)

					optA, found := lo.Find(pollGet.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_a" })
					r.True(found)
					a.Equal(0, optA.Votes)

					optB, found := lo.Find(pollGet.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_b" })
					r.True(found)
					a.Equal(0, optB.Votes)
				})

				t.Run("user1_votes_opt_a", func(t *testing.T) {
					r := require.New(t)
					a := assert.New(t)

					vote1 := tests.AssertRequest(
						cl.ThreadVotePollWithResponse(root, slug, openapi.PollVoteBody{OptionId: "opt_a"}, user1Session),
					)(t, http.StatusOK)
					r.NotNil(vote1.JSON200)
					a.Equal(1, vote1.JSON200.TotalVotes)
					r.False(vote1.JSON200.UserVote.IsNull())
					a.Equal("opt_a", vote1.JSON200.UserVote.MustGet())

					optA, found := lo.Find(vote1.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_a" })
					r.True(found)
					a.Equal(1, optA.Votes)
				})

				t.Run("user2_votes_opt_b", func(t *testing.T) {
					r := require.New(t)
					a := assert.New(t)

					vote2 := tests.AssertRequest(
						cl.ThreadVotePollWithResponse(root, slug, openapi.PollVoteBody{OptionId: "opt_b"}, user2Session),
					)(t, http.StatusOK)
					r.NotNil(vote2.JSON200)
					a.Equal(2, vote2.JSON200.TotalVotes)
					r.False(vote2.JSON200.UserVote.IsNull())
					a.Equal("opt_b", vote2.JSON200.UserVote.MustGet())
				})

				t.Run("user1_changes_vote_to_opt_b", func(t *testing.T) {
					r := require.New(t)
					a := assert.New(t)

					changed := tests.AssertRequest(
						cl.ThreadVotePollWithResponse(root, slug, openapi.PollVoteBody{OptionId: "opt_b"}, user1Session),
					)(t, http.StatusOK)
					r.NotNil(changed.JSON200)
					a.Equal(2, changed.JSON200.TotalVotes)
					r.False(changed.JSON200.UserVote.IsNull())
					a.Equal("opt_b", changed.JSON200.UserVote.MustGet())

					optA, found := lo.Find(changed.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_a" })
					r.True(found)
					a.Equal(0, optA.Votes)

					optB, found := lo.Find(changed.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_b" })
					r.True(found)
					a.Equal(2, optB.Votes)
				})

				t.Run("final_state", func(t *testing.T) {
					r := require.New(t)
					a := assert.New(t)

					finalGet := tests.AssertRequest(
						cl.ThreadGetPollWithResponse(root, slug, user1Session),
					)(t, http.StatusOK)
					r.NotNil(finalGet.JSON200)
					a.Equal(2, finalGet.JSON200.TotalVotes)
					r.False(finalGet.JSON200.UserVote.IsNull())
					a.Equal("opt_b", finalGet.JSON200.UserVote.MustGet())

					optA, found := lo.Find(finalGet.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_a" })
					r.True(found)
					a.Equal(0, optA.Votes)

					optB, found := lo.Find(finalGet.JSON200.Options, func(o openapi.PollOptionStatus) bool { return o.Id == "opt_b" })
					r.True(found)
					a.Equal(2, optB.Votes)
				})

				t.Run("admin_cannot_edit_poll", func(t *testing.T) {
					title := "Updated title"
					tests.AssertRequest(
						cl.ThreadUpdateWithResponse(root, slug, openapi.ThreadMutableProps{
							Title: &title,
						}, adminSession),
					)(t, http.StatusBadRequest)
				})

				t.Run("admin_cannot_edit_poll_via_channel", func(t *testing.T) {
					t.Parallel()
					title := "Updated title via channel"
					tests.AssertRequest(
						cl.ChannelThreadUpdateWithResponse(root, channelID, slug, openapi.ThreadMutableProps{
							Title: &title,
						}, adminSession),
					)(t, http.StatusBadRequest)
				})
			})

			_ = r
			_ = a
		}))
	}))
}
