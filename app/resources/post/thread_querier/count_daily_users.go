package thread_querier

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

var istLocation = time.FixedZone("IST", 5*3600+30*60)

func istDayWindow(now time.Time) (start, end time.Time) {
	nowIST := now.In(istLocation)
	start = time.Date(nowIST.Year(), nowIST.Month(), nowIST.Day(), 0, 0, 0, 0, istLocation)
	end = start.Add(24 * time.Hour)
	return start, end
}

func (d *Querier) CountUniqueAuthorsToday(ctx context.Context, channelID xid.ID) (int, error) {
	start, end := istDayWindow(time.Now())

	authorIDs, err := d.db.Post.Query().
		Where(
			ent_post.DeletedAtIsNil(),
			ent_post.RootPostIDIsNil(),
			ent_post.ChannelID(channelID),
			ent_post.CreatedAtGTE(start),
			ent_post.CreatedAtLT(end),
		).
		Unique(true).
		Select(ent_post.FieldAccountPosts).
		Strings(ctx)
	if err != nil {
		return 0, fault.Wrap(err, fctx.With(ctx))
	}

	return len(authorIDs), nil
}
