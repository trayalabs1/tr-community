package reply_admin_queue

import (
	"time"

	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/internal/ent"
)

type ID xid.ID

func (i ID) String() string { return xid.ID(i).String() }

type Entry struct {
	ID             ID
	ReplyID        post.ID
	ThreadID       post.ID
	ChannelID      xid.ID
	ContentSnippet string
	CreatedAt      time.Time
}

func Map(e *ent.ReplyAdminQueue) (*Entry, error) {
	return &Entry{
		ID:             ID(e.ID),
		ReplyID:        post.ID(e.ReplyID),
		ThreadID:       post.ID(e.ThreadID),
		ChannelID:      e.ChannelID,
		ContentSnippet: e.ContentSnippet,
		CreatedAt:      e.CreatedAt,
	}, nil
}
