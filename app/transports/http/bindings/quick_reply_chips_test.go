package bindings

import (
	"testing"

	"github.com/Southclaws/opt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
)

func TestQuickReplyChipsForDoubtChannel(t *testing.T) {
	chips := quickReplyChipsFor(&thread.Thread{Channel: opt.New(channel.Channel{Slug: doubtChannelSlug})})
	require.NotNil(t, chips)
	assert.Equal(t, []string{"Same doubt 🤔", "Helpful 👍"}, chips.Candidates)

	bah := quickReplyChipsFor(&thread.Thread{
		Channel: opt.New(channel.Channel{Slug: doubtChannelSlug}),
		Post:    post.Post{Meta: map[string]any{"post_category": "BAH"}},
	})
	require.NotNil(t, bah)
	assert.Equal(t, buildAHabitChips, bah.Candidates)

	assert.Nil(t, quickReplyChipsFor(&thread.Thread{Channel: opt.New(channel.Channel{Slug: "other"})}))
	assert.Nil(t, quickReplyChipsFor(&thread.Thread{}))
}
