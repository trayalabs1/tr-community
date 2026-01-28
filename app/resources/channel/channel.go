package channel

import (
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/internal/ent"
)

type ChannelID xid.ID

func (i ChannelID) String() string { return xid.ID(i).String() }

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type Channel struct {
	ID          ChannelID
	Name        string
	Slug        string
	Description string
	CoverImage  opt.Optional[asset.Asset]
	Icon        opt.Optional[asset.Asset]
	Visibility  Visibility
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromModel(c *ent.Channel) *Channel {
	coverImage := opt.Map(opt.NewPtr(c.Edges.CoverImage), func(a ent.Asset) asset.Asset {
		return *asset.Map(&a)
	})

	icon := opt.Map(opt.NewPtr(c.Edges.Icon), func(a ent.Asset) asset.Asset {
		return *asset.Map(&a)
	})

	return &Channel{
		ID:          ChannelID(c.ID),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		CoverImage:  coverImage,
		Icon:        icon,
		Visibility:  Visibility(c.Visibility),
		Metadata:    c.Metadata,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func FromModelSlice(channels []*ent.Channel) []*Channel {
	return dt.Map(channels, FromModel)
}
