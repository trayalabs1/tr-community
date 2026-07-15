package thread_share

import (
	"context"
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/internal/ent"
	ent_threadshare "github.com/Southclaws/storyden/internal/ent/threadshare"
)

var (
	errNotFound      = fault.Wrap(fault.New("thread share not found"), ftag.With(ftag.NotFound))
	errAlreadyExists = fault.Wrap(fault.New("thread share already exists"), ftag.With(ftag.AlreadyExists))
)

type Repository struct {
	db *ent.Client
}

func New(db *ent.Client) *Repository {
	return &Repository{db}
}

func (r *Repository) Create(ctx context.Context, postID xid.ID, channelID xid.ID, accountID xid.ID, subtitle string) (*Share, error) {
	_, err := r.db.ThreadShare.Create().
		SetPostID(postID).
		SetChannelID(channelID).
		SetAccountID(accountID).
		SetSubtitle(subtitle).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return r.Get(ctx, postID, channelID)
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.Get(ctx, postID, channelID)
}

func (r *Repository) CreateMany(ctx context.Context, postID xid.ID, channelIDs []xid.ID, accountID xid.ID, subtitle string) ([]*Share, error) {
	existing, err := r.db.ThreadShare.Query().
		Where(
			ent_threadshare.PostID(postID),
			ent_threadshare.ChannelIDIn(channelIDs...),
			ent_threadshare.DeletedAtIsNil(),
		).
		Select(ent_threadshare.FieldChannelID).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	skip := make(map[xid.ID]struct{}, len(existing))
	for _, e := range existing {
		skip[e.ChannelID] = struct{}{}
	}

	builders := make([]*ent.ThreadShareCreate, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if _, ok := skip[channelID]; ok {
			continue
		}
		builders = append(builders, r.db.ThreadShare.Create().
			SetPostID(postID).
			SetChannelID(channelID).
			SetAccountID(accountID).
			SetSubtitle(subtitle))
	}

	if len(builders) > 0 {
		if err := r.db.ThreadShare.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
	}

	return r.ListByPost(ctx, postID)
}

func (r *Repository) Delete(ctx context.Context, postID, channelID xid.ID) error {
	_, err := r.db.ThreadShare.Update().
		Where(
			ent_threadshare.PostID(postID),
			ent_threadshare.ChannelID(channelID),
			ent_threadshare.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (r *Repository) ListByPost(ctx context.Context, postID xid.ID) ([]*Share, error) {
	shares, err := r.db.ThreadShare.Query().
		Where(
			ent_threadshare.PostID(postID),
			ent_threadshare.DeletedAtIsNil(),
		).
		WithChannel().
		Order(ent.Asc(ent_threadshare.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.MapErr(shares, FromModel)
}

func (r *Repository) ListByChannel(ctx context.Context, channelID xid.ID) ([]*Share, error) {
	shares, err := r.db.ThreadShare.Query().
		Where(
			ent_threadshare.ChannelID(channelID),
			ent_threadshare.DeletedAtIsNil(),
		).
		WithChannel().
		Order(ent.Asc(ent_threadshare.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.MapErr(shares, FromModel)
}

func (r *Repository) SetPin(ctx context.Context, postID, channelID xid.ID, pinned bool) (*Share, error) {
	update := r.db.ThreadShare.Update().
		Where(
			ent_threadshare.PostID(postID),
			ent_threadshare.ChannelID(channelID),
			ent_threadshare.DeletedAtIsNil(),
		)

	if pinned {
		update.SetPinnedRank(1).SetPinnedAt(time.Now())
	} else {
		update.SetPinnedRank(0).ClearPinnedAt()
	}

	_, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.Get(ctx, postID, channelID)
}

func (r *Repository) Get(ctx context.Context, postID, channelID xid.ID) (*Share, error) {
	share, err := r.db.ThreadShare.Query().
		Where(
			ent_threadshare.PostID(postID),
			ent_threadshare.ChannelID(channelID),
			ent_threadshare.DeletedAtIsNil(),
		).
		WithChannel().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return FromModel(share)
}
