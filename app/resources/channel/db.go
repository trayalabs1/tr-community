package channel

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/internal/ent"
	ent_channel "github.com/Southclaws/storyden/internal/ent/channel"
	"github.com/Southclaws/storyden/internal/ent/predicate"
)

var (
	errNotFound = fault.Wrap(fault.New("channel not found"), ftag.With(ftag.NotFound))
)

type Repository struct {
	db *ent.Client
}

func New(db *ent.Client) *Repository {
	return &Repository{db}
}

type Option func(*ent.ChannelMutation)

func WithID(id ChannelID) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetID(xid.ID(id))
	}
}

func WithName(v string) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetName(v)
	}
}

func WithSlug(v string) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetSlug(v)
	}
}

func WithDescription(v string) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetDescription(v)
	}
}

func WithVisibility(v Visibility) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetVisibility(ent_channel.Visibility(v))
	}
}

func WithCoverImageAssetID(id *xid.ID) Option {
	return func(cm *ent.ChannelMutation) {
		if id == nil {
			cm.ClearCoverImage()
			return
		}
		cm.SetCoverImageAssetID(*id)
	}
}

func WithIconAssetID(id *xid.ID) Option {
	return func(cm *ent.ChannelMutation) {
		if id == nil {
			cm.ClearIcon()
			return
		}
		cm.SetIconAssetID(*id)
	}
}

func WithMetadata(v map[string]any) Option {
	return func(cm *ent.ChannelMutation) {
		cm.SetMetadata(v)
	}
}

func (r *Repository) Create(ctx context.Context, name, slug, description string, visibility Visibility, opts ...Option) (*Channel, error) {
	create := r.db.Channel.Create().
		SetName(name).
		SetSlug(slug).
		SetDescription(description).
		SetVisibility(ent_channel.Visibility(visibility))

	mut := create.Mutation()
	for _, opt := range opts {
		opt(mut)
	}

	channel, err := create.Save(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.Get(ctx, ChannelID(channel.ID))
}

func (r *Repository) Get(ctx context.Context, id ChannelID) (*Channel, error) {
	channel, err := r.db.Channel.Query().
		Where(ent_channel.ID(xid.ID(id))).
		WithCoverImage().
		WithIcon().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return FromModel(channel), nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Channel, error) {
	channel, err := r.db.Channel.Query().
		Where(ent_channel.Slug(slug)).
		WithCoverImage().
		WithIcon().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return FromModel(channel), nil
}

func (r *Repository) List(ctx context.Context, filters ...predicate.Channel) ([]*Channel, error) {
	channels, err := r.db.Channel.Query().
		Where(filters...).
		WithCoverImage().
		WithIcon().
		Order(ent.Asc(ent_channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return dt.Map(channels, FromModel), nil
}

func (r *Repository) Update(ctx context.Context, id ChannelID, opts ...Option) (*Channel, error) {
	update := r.db.Channel.UpdateOneID(xid.ID(id))

	mut := update.Mutation()
	for _, opt := range opts {
		opt(mut)
	}

	channel, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r.Get(ctx, ChannelID(channel.ID))
}

func (r *Repository) Delete(ctx context.Context, id ChannelID) error {
	err := r.db.Channel.DeleteOneID(xid.ID(id)).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fault.Wrap(errNotFound, fctx.With(ctx))
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (r *Repository) SearchPublic(ctx context.Context, query string) ([]*Channel, error) {
	filters := []predicate.Channel{
		ent_channel.VisibilityEQ(ent_channel.VisibilityPublic),
	}

	if query != "" {
		filters = append(filters, ent_channel.Or(
			ent_channel.NameContainsFold(query),
			ent_channel.DescriptionContainsFold(query),
		))
	}

	return r.List(ctx, filters...)
}
