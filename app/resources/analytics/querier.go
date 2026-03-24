package analytics

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/internal/ent"
	ent_account "github.com/Southclaws/storyden/internal/ent/account"
	ent_art "github.com/Southclaws/storyden/internal/ent/adminreplytime"
	ent_channel "github.com/Southclaws/storyden/internal/ent/channel"
	ent_membership "github.com/Southclaws/storyden/internal/ent/channelmembership"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

type Querier struct {
	db *ent.Client
}

func New(db *ent.Client) *Querier {
	return &Querier{db: db}
}

func (q *Querier) GetReport(ctx context.Context, start, end time.Time) (*Report, error) {
	onboardings, err := q.channelOnboardings(ctx, start, end)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	posts, err := q.channelPosts(ctx, start, end)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	replyTimes, err := q.adminReplyTimes(ctx, start, end)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return &Report{
		ChannelOnboardings: onboardings,
		ChannelPosts:       posts,
		AdminReplyTimes:    replyTimes,
	}, nil
}

func (q *Querier) channelOnboardings(ctx context.Context, start, end time.Time) ([]ChannelStat, error) {
	var rows []struct {
		ChannelID string `json:"channel_id"`
		Count     int    `json:"count"`
	}
	err := q.db.ChannelMembership.Query().
		Where(
			ent_membership.CreatedAtGTE(start),
			ent_membership.CreatedAtLTE(end),
		).
		GroupBy(ent_membership.FieldChannelID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	ids := make([]xid.ID, 0, len(rows))
	countByID := make(map[xid.ID]int, len(rows))
	for _, row := range rows {
		id, err := xid.FromString(row.ChannelID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
		countByID[id] = row.Count
	}

	channels, err := q.db.Channel.Query().
		Where(ent_channel.IDIn(ids...)).
		Select(ent_channel.FieldID, ent_channel.FieldName).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	stats := make([]ChannelStat, 0, len(channels))
	for _, ch := range channels {
		stats = append(stats, ChannelStat{ChannelName: ch.Name, Count: countByID[ch.ID]})
	}
	return stats, nil
}

func (q *Querier) channelPosts(ctx context.Context, start, end time.Time) ([]ChannelStat, error) {
	var rows []struct {
		ChannelID string `json:"channel_id"`
		Count     int    `json:"count"`
	}
	err := q.db.Post.Query().
		Where(
			ent_post.CreatedAtGTE(start),
			ent_post.CreatedAtLTE(end),
			ent_post.HasAuthorWith(ent_account.AdminEQ(false)),
			ent_post.RootPostIDIsNil(),
		).
		GroupBy(ent_post.FieldChannelID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	ids := make([]xid.ID, 0, len(rows))
	countByID := make(map[xid.ID]int, len(rows))
	for _, row := range rows {
		id, err := xid.FromString(row.ChannelID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
		countByID[id] = row.Count
	}

	channels, err := q.db.Channel.Query().
		Where(ent_channel.IDIn(ids...)).
		Select(ent_channel.FieldID, ent_channel.FieldName).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	stats := make([]ChannelStat, 0, len(channels))
	for _, ch := range channels {
		stats = append(stats, ChannelStat{ChannelName: ch.Name, Count: countByID[ch.ID]})
	}
	return stats, nil
}

func (q *Querier) adminReplyTimes(ctx context.Context, start, end time.Time) ([]AdminReplyTimeStat, error) {
	var rows []struct {
		AdminHandle string  `json:"admin_handle"`
		Mean        float64 `json:"avg"`
	}
	err := q.db.AdminReplyTime.Query().
		Where(
			ent_art.AdminPostTimeGTE(start),
			ent_art.AdminPostTimeLTE(end),
		).
		GroupBy(ent_art.FieldAdminHandle).
		Aggregate(ent.Mean(ent_art.FieldTimeDifferenceSeconds)).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	stats := make([]AdminReplyTimeStat, 0, len(rows))
	for _, row := range rows {
		stats = append(stats, AdminReplyTimeStat{
			AdminHandle:    row.AdminHandle,
			AvgTimeMinutes: row.Mean / 60.0,
		})
	}
	return stats, nil
}
