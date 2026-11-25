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

func TestChannelThreads(t *testing.T) {
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

			// Create test users: owner, admin, member, non-member
			ownerCtx, owner := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			adminCtx, admin := e2e.WithAccount(root, aw, seed.Account_002_Frigg)
			memberCtx, member := e2e.WithAccount(root, aw, seed.Account_003_Baldur)
			nonMemberCtx, _ := e2e.WithAccount(root, aw, seed.Account_004_Loki)

			ownerSession := sh.WithSession(ownerCtx)
			adminSession := sh.WithSession(adminCtx)
			memberSession := sh.WithSession(memberCtx)
			nonMemberSession := sh.WithSession(nonMemberCtx)

			ownerID := owner.ID
			adminID := admin.ID
			memberID := member.ID

			// Test: Create a channel
			createResp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Test Channel",
				Slug:        "test-channel",
				Description: "A test channel for thread operations",
			}, ownerSession)
			tests.Ok(t, err, createResp)

			channelID := createResp.JSON200.Id

			// Add admin and member to the channel
			addAdminResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(adminID.String()),
				Role:      openapi.ChannelMemberAddRoleAdmin,
			}, ownerSession)
			tests.Ok(t, err, addAdminResp)

			addMemberResp, err := cl.ChannelMemberAddWithResponse(root, channelID, openapi.ChannelMemberAdd{
				AccountId: openapi.Identifier(memberID.String()),
				Role:      openapi.ChannelMemberAddRoleMember,
			}, ownerSession)
			tests.Ok(t, err, addMemberResp)

			// Create a category in the channel for threads
			catResp, err := cl.ChannelCategoryCreateWithResponse(root, channelID, openapi.CategoryInitialProps{
				Name:        "General Discussion",
				Description: "General topics",
				Colour:      "#ff0000",
			}, ownerSession)
			tests.Ok(t, err, catResp)
			categoryID := catResp.JSON200.Id

			// Test: List threads in empty channel
			listResp1, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, listResp1)
			r.Equal(0, listResp1.JSON200.Results, "empty channel should have zero threads")

			// Test: Owner creates a thread
			thread1Body := "<p>This is a test thread in the channel</p>"
			thread1Create, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Test Thread 1",
				Body:       opt.New(thread1Body).Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, ownerSession)
			tests.Ok(t, err, thread1Create)
			r.Equal(ownerID.String(), thread1Create.JSON200.Author.Id)
			r.Equal("Test Thread 1", thread1Create.JSON200.Title)
			r.Contains(thread1Create.JSON200.Slug, "test-thread-1")
			thread1ID := thread1Create.JSON200.Id
			thread1Slug := thread1Create.JSON200.Slug

			// Test: Admin can create threads
			thread2Create, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Test Thread 2",
				Body:       opt.New("<p>Admin thread</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, adminSession)
			tests.Ok(t, err, thread2Create)
			r.Equal(adminID.String(), thread2Create.JSON200.Author.Id)
			thread2Slug := thread2Create.JSON200.Slug

			// Test: Member can create threads
			thread3Create, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Test Thread 3",
				Body:       opt.New("<p>Member thread</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, memberSession)
			tests.Ok(t, err, thread3Create)
			r.Equal(memberID.String(), thread3Create.JSON200.Author.Id)

			// Test: Non-member cannot create threads
			thread4Create, err := cl.ChannelThreadCreateWithResponse(root, channelID, openapi.ThreadInitialProps{
				Title:      "Test Thread 4",
				Body:       opt.New("<p>Non-member thread</p>").Ptr(),
				Category:   opt.New(categoryID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, nonMemberSession)
			r.NoError(err)
			r.Nil(thread4Create.JSON200, "non-member should not be able to create threads")
			r.NotNil(thread4Create.JSONDefault, "should return error response")

			// Test: List threads
			listResp2, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, listResp2)
			r.Equal(3, listResp2.JSON200.Results, "should have 3 threads")

			// Test: Member can list threads
			listResp3, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, memberSession)
			tests.Ok(t, err, listResp3)
			r.Equal(3, listResp3.JSON200.Results, "member should see all threads")

			// Test: Non-member cannot list threads
			listResp4, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, nonMemberSession)
			r.NoError(err)
			r.Nil(listResp4.JSON200, "non-member should not be able to list threads")

			// Test: Get a specific thread
			getResp1, err := cl.ChannelThreadGetWithResponse(root, channelID, thread1Slug, nil, ownerSession)
			tests.Ok(t, err, getResp1)
			r.Equal(thread1ID, getResp1.JSON200.Id)
			r.Equal("Test Thread 1", getResp1.JSON200.Title)

			// Test: Member can get thread
			getResp2, err := cl.ChannelThreadGetWithResponse(root, channelID, thread1Slug, nil, memberSession)
			tests.Ok(t, err, getResp2)
			r.Equal(thread1ID, getResp2.JSON200.Id)

			// Test: Non-member cannot get thread
			getResp3, err := cl.ChannelThreadGetWithResponse(root, channelID, thread1Slug, nil, nonMemberSession)
			r.NoError(err)
			r.Nil(getResp3.JSON200, "non-member should not be able to get thread")

			// Test: Owner updates their thread
			newTitle := "Updated Test Thread 1"
			updateResp1, err := cl.ChannelThreadUpdateWithResponse(root, channelID, thread1Slug, openapi.ThreadMutableProps{
				Title: &newTitle,
			}, ownerSession)
			tests.Ok(t, err, updateResp1)
			r.Equal(newTitle, updateResp1.JSON200.Title)

			// Test: Admin can update their own thread
			adminNewTitle := "Updated Admin Thread"
			updateResp2, err := cl.ChannelThreadUpdateWithResponse(root, channelID, thread2Slug, openapi.ThreadMutableProps{
				Title: &adminNewTitle,
			}, adminSession)
			tests.Ok(t, err, updateResp2)
			r.Equal(adminNewTitle, updateResp2.JSON200.Title)

			// Test: Member cannot update owner's thread (not their own)
			memberUpdateTitle := "Member trying to update"
			updateResp3, err := cl.ChannelThreadUpdateWithResponse(root, channelID, thread1Slug, openapi.ThreadMutableProps{
				Title: &memberUpdateTitle,
			}, memberSession)
			r.NoError(err)
			r.Nil(updateResp3.JSON200, "member should not be able to update other's threads")

			// Test: Non-member cannot update any thread
			updateResp4, err := cl.ChannelThreadUpdateWithResponse(root, channelID, thread1Slug, openapi.ThreadMutableProps{
				Title: &memberUpdateTitle,
			}, nonMemberSession)
			r.NoError(err)
			r.Nil(updateResp4.JSON200, "non-member should not be able to update threads")

			// Test: Create replies
			reply1Create, err := cl.ChannelReplyCreateWithResponse(root, channelID, thread1Slug, openapi.ReplyInitialProps{
				Body: "First reply to thread 1",
			}, memberSession)
			tests.Ok(t, err, reply1Create)
			r.Equal(memberID.String(), reply1Create.JSON200.Author.Id)
			reply1ID := reply1Create.JSON200.Id

			reply2Create, err := cl.ChannelReplyCreateWithResponse(root, channelID, thread1Slug, openapi.ReplyInitialProps{
				Body: "Second reply to thread 1",
			}, adminSession)
			tests.Ok(t, err, reply2Create)
			r.Equal(adminID.String(), reply2Create.JSON200.Author.Id)

			// Test: Reply to a reply
			reply3Create, err := cl.ChannelReplyCreateWithResponse(root, channelID, thread1Slug, openapi.ReplyInitialProps{
				Body:    "Reply to first reply",
				ReplyTo: &reply1ID,
			}, ownerSession)
			tests.Ok(t, err, reply3Create)
			r.Equal(ownerID.String(), reply3Create.JSON200.Author.Id)
			r.NotNil(reply3Create.JSON200.ReplyTo)
			r.Equal(reply1ID, reply3Create.JSON200.ReplyTo.Id)

			// Test: Non-member cannot create replies
			reply4Create, err := cl.ChannelReplyCreateWithResponse(root, channelID, thread1Slug, openapi.ReplyInitialProps{
				Body: "Non-member reply",
			}, nonMemberSession)
			r.NoError(err)
			r.Nil(reply4Create.JSON200, "non-member should not be able to create replies")

			// Test: Get thread with replies
			getWithReplies, err := cl.ChannelThreadGetWithResponse(root, channelID, thread1Slug, nil, ownerSession)
			tests.Ok(t, err, getWithReplies)
			r.Len(getWithReplies.JSON200.Replies.Replies, 3, "thread should have 3 replies")

			// Test: Delete a thread (owner can delete their own)
			deleteResp1, err := cl.ChannelThreadDeleteWithResponse(root, channelID, thread1Slug, ownerSession)
			tests.Ok(t, err, deleteResp1)

			// Test: Verify thread is deleted (should not be accessible)
			getDeletedResp, err := cl.ChannelThreadGetWithResponse(root, channelID, thread1Slug, nil, ownerSession)
			r.NoError(err)
			r.Nil(getDeletedResp.JSON200, "deleted thread should not be accessible")

			// Test: Final thread list should have 2 threads (one was deleted)
			listFinal, err := cl.ChannelThreadListWithResponse(root, channelID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, listFinal)
			r.Equal(2, listFinal.JSON200.Results, "should have 2 threads after deletion")
		}))
	}))
}

func TestChannelThreadIsolation(t *testing.T) {
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

			// Create owner account
			ownerCtx, _ := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			ownerSession := sh.WithSession(ownerCtx)

			// Create two channels
			channel1Resp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Channel 1",
				Slug:        "channel-1-isolation",
				Description: "First channel",
			}, ownerSession)
			tests.Ok(t, err, channel1Resp)
			channel1ID := channel1Resp.JSON200.Id

			channel2Resp, err := cl.ChannelCreateWithResponse(root, openapi.ChannelInitialProps{
				Name:        "Channel 2",
				Slug:        "channel-2-isolation",
				Description: "Second channel",
			}, ownerSession)
			tests.Ok(t, err, channel2Resp)
			channel2ID := channel2Resp.JSON200.Id

			// Create categories in each channel
			cat1Resp, err := cl.ChannelCategoryCreateWithResponse(root, channel1ID, openapi.CategoryInitialProps{
				Name:        "Category 1",
				Description: "Category in channel 1",
				Colour:      "#ff0000",
			}, ownerSession)
			tests.Ok(t, err, cat1Resp)
			cat1ID := cat1Resp.JSON200.Id

			cat2Resp, err := cl.ChannelCategoryCreateWithResponse(root, channel2ID, openapi.CategoryInitialProps{
				Name:        "Category 2",
				Description: "Category in channel 2",
				Colour:      "#00ff00",
			}, ownerSession)
			tests.Ok(t, err, cat2Resp)
			cat2ID := cat2Resp.JSON200.Id

			// Create a thread in channel 1
			thread1Create, err := cl.ChannelThreadCreateWithResponse(root, channel1ID, openapi.ThreadInitialProps{
				Title:      "Thread in Channel 1",
				Body:       opt.New("<p>This thread belongs to channel 1</p>").Ptr(),
				Category:   opt.New(cat1ID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, ownerSession)
			tests.Ok(t, err, thread1Create)
			thread1Slug := thread1Create.JSON200.Slug

			// Create a thread in channel 2
			thread2Create, err := cl.ChannelThreadCreateWithResponse(root, channel2ID, openapi.ThreadInitialProps{
				Title:      "Thread in Channel 2",
				Body:       opt.New("<p>This thread belongs to channel 2</p>").Ptr(),
				Category:   opt.New(cat2ID).Ptr(),
				Visibility: opt.New(openapi.Published).Ptr(),
			}, ownerSession)
			tests.Ok(t, err, thread2Create)
			thread2Slug := thread2Create.JSON200.Slug

			// Test: Cannot access channel 1's thread from channel 2
			wrongChannelGet, err := cl.ChannelThreadGetWithResponse(root, channel2ID, thread1Slug, nil, ownerSession)
			r.NoError(err)
			r.Nil(wrongChannelGet.JSON200, "should not be able to access thread from wrong channel")

			// Test: Cannot access channel 2's thread from channel 1
			wrongChannelGet2, err := cl.ChannelThreadGetWithResponse(root, channel1ID, thread2Slug, nil, ownerSession)
			r.NoError(err)
			r.Nil(wrongChannelGet2.JSON200, "should not be able to access thread from wrong channel")

			// Test: List threads in channel 1 should only show channel 1's thread
			list1, err := cl.ChannelThreadListWithResponse(root, channel1ID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, list1)
			r.Equal(1, list1.JSON200.Results)
			r.Equal("Thread in Channel 1", list1.JSON200.Threads[0].Title)

			// Test: List threads in channel 2 should only show channel 2's thread
			list2, err := cl.ChannelThreadListWithResponse(root, channel2ID, &openapi.ChannelThreadListParams{}, ownerSession)
			tests.Ok(t, err, list2)
			r.Equal(1, list2.JSON200.Results)
			r.Equal("Thread in Channel 2", list2.JSON200.Threads[0].Title)

			// Test: Cannot update thread from wrong channel
			newTitle := "Trying to update from wrong channel"
			wrongUpdate, err := cl.ChannelThreadUpdateWithResponse(root, channel2ID, thread1Slug, openapi.ThreadMutableProps{
				Title: &newTitle,
			}, ownerSession)
			r.NoError(err)
			r.Nil(wrongUpdate.JSON200, "should not be able to update thread from wrong channel")

			// Test: Cannot delete thread from wrong channel
			wrongDelete, err := cl.ChannelThreadDeleteWithResponse(root, channel2ID, thread1Slug, ownerSession)
			r.NoError(err)
			r.NotEqual(200, wrongDelete.StatusCode(), "should not be able to delete thread from wrong channel")

			// Test: Cannot create reply to thread from wrong channel
			wrongReply, err := cl.ChannelReplyCreateWithResponse(root, channel2ID, thread1Slug, openapi.ReplyInitialProps{
				Body: "Reply from wrong channel",
			}, ownerSession)
			r.NoError(err)
			r.Nil(wrongReply.JSON200, "should not be able to reply to thread from wrong channel")
		}))
	}))
}
