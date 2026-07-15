package thread_share

import (
	"time"

	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/internal/ent"
)

type Share struct {
	PostID     xid.ID
	Channel    channel.Channel
	AccountID  xid.ID
	Subtitle   string
	PinnedRank int
	Pinned     bool
	CreatedAt  time.Time
}

func FromModel(m *ent.ThreadShare) (*Share, error) {
	return &Share{
		PostID:     m.PostID,
		Channel:    *channel.FromModel(m.Edges.Channel),
		AccountID:  m.AccountID,
		Subtitle:   m.Subtitle,
		PinnedRank: m.PinnedRank,
		Pinned:     m.PinnedRank > 0 || m.PinnedAt != nil,
		CreatedAt:  m.CreatedAt,
	}, nil
}
