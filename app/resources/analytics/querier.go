package analytics

import (
	"context"
	"slices"
	"strings"
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

	perAdmin, overall, err := q.threadsReplied(ctx, start, end)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return &Report{
		ChannelOnboardings:  onboardings,
		ChannelPosts:        posts,
		AdminReplyTimes:     replyTimes,
		AdminThreadsReplied: perAdmin,
		ThreadsReplied:      overall,
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
			ent_art.UserPostTimeGTE(start),
			ent_art.UserPostTimeLTE(end),
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

// threadsReplied counts distinct threads that received a published admin reply
// in the period. A reply made to another reply has reply_to_post_id set, a reply
// made directly to the thread does not.
func (q *Querier) threadsReplied(ctx context.Context, start, end time.Time) ([]AdminThreadsRepliedStat, ThreadsRepliedCounts, error) {
	var rows []struct {
		AccountID string  `json:"account_posts"`
		RootID    string  `json:"root_post_id"`
		ReplyToID *string `json:"reply_to_post_id"`
	}
	err := q.db.Post.Query().
		Where(
			ent_post.CreatedAtGTE(start),
			ent_post.CreatedAtLTE(end),
			ent_post.DeletedAtIsNil(),
			ent_post.VisibilityEQ(ent_post.VisibilityPublished),
			ent_post.RootPostIDNotNil(),
			ent_post.HasAuthorWith(ent_account.AdminEQ(true)),
		).
		Select(ent_post.FieldAccountPosts, ent_post.FieldRootPostID, ent_post.FieldReplyToPostID).
		Scan(ctx, &rows)
	if err != nil {
		return nil, ThreadsRepliedCounts{}, fault.Wrap(err, fctx.With(ctx))
	}

	type sets struct{ all, direct, toReply map[string]struct{} }
	newSets := func() *sets {
		return &sets{map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}}
	}
	add := func(s *sets, root string, isToReply bool) {
		s.all[root] = struct{}{}
		if isToReply {
			s.toReply[root] = struct{}{}
		} else {
			s.direct[root] = struct{}{}
		}
	}
	counts := func(s *sets) ThreadsRepliedCounts {
		return ThreadsRepliedCounts{Total: len(s.all), Direct: len(s.direct), ToReply: len(s.toReply)}
	}

	overall := newSets()
	byAdmin := map[string]*sets{}
	for _, r := range rows {
		isToReply := r.ReplyToID != nil && *r.ReplyToID != ""
		add(overall, r.RootID, isToReply)
		s, ok := byAdmin[r.AccountID]
		if !ok {
			s = newSets()
			byAdmin[r.AccountID] = s
		}
		add(s, r.RootID, isToReply)
	}

	ids := make([]xid.ID, 0, len(byAdmin))
	for k := range byAdmin {
		if id, err := xid.FromString(k); err == nil {
			ids = append(ids, id)
		}
	}
	accounts, err := q.db.Account.Query().
		Where(ent_account.IDIn(ids...)).
		Select(ent_account.FieldID, ent_account.FieldHandle).
		All(ctx)
	if err != nil {
		return nil, ThreadsRepliedCounts{}, fault.Wrap(err, fctx.With(ctx))
	}

	stats := make([]AdminThreadsRepliedStat, 0, len(accounts))
	for _, a := range accounts {
		stats = append(stats, AdminThreadsRepliedStat{
			AdminHandle:          a.Handle,
			ThreadsRepliedCounts: counts(byAdmin[a.ID.String()]),
		})
	}
	slices.SortFunc(stats, func(a, b AdminThreadsRepliedStat) int {
		if a.Total != b.Total {
			return b.Total - a.Total
		}
		return strings.Compare(a.AdminHandle, b.AdminHandle)
	})

	return stats, counts(overall), nil
}
