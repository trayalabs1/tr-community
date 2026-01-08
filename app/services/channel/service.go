package channel

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/channel_membership"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/internal/deletable"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

var errInvalidCreate = fault.New("invalid create args", ftag.With(ftag.InvalidArgument))

type Service interface {
	Create(ctx context.Context, accountID account.AccountID, partial Partial) (*channel.Channel, error)
	Get(ctx context.Context, id channel.ChannelID) (*channel.Channel, error)
	GetBySlug(ctx context.Context, slug string) (*channel.Channel, error)
	GetUserChannels(ctx context.Context, accountID account.AccountID) ([]*channel.Channel, error)
	Update(ctx context.Context, id channel.ChannelID, partial Partial) (*channel.Channel, error)
	Delete(ctx context.Context, id channel.ChannelID) error
	SearchPublic(ctx context.Context, query string) ([]*channel.Channel, error)
	CheckAccess(ctx context.Context, channelID channel.ChannelID, accountID account.AccountID) (bool, error)
}

type Partial struct {
	Name              opt.Optional[string]
	Slug              opt.Optional[string]
	Description       opt.Optional[string]
	Visibility        opt.Optional[channel.Visibility]
	CoverImageAssetID deletable.Value[*xid.ID]
	IconAssetID       deletable.Value[*xid.ID]
	Metadata          opt.Optional[map[string]any]
}

func Build() fx.Option {
	return fx.Options(
		fx.Provide(New),
		fx.Invoke(setupSubscribers),
	)
}

type service struct {
	channel_repo    *channel.Repository
	membership_repo *channel_membership.Repository
	bus             *pubsub.Bus
}

func New(
	channel_repo *channel.Repository,
	membership_repo *channel_membership.Repository,
	bus *pubsub.Bus,
) Service {
	return &service{
		channel_repo:    channel_repo,
		membership_repo: membership_repo,
		bus:             bus,
	}
}

const generalChannelID = "general000000000"

func setupSubscribers(
	lc fx.Lifecycle,
	bus *pubsub.Bus,
	membership_repo *channel_membership.Repository,
) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		_, err := pubsub.Subscribe(ctx, bus, "channel.account_created", func(ctx context.Context, evt *message.EventAccountCreated) error {
			// Parse the general channel ID
			generalID, err := xid.FromString(generalChannelID)
			if err != nil {
				return fault.Wrap(err, fctx.With(ctx))
			}

			// Add new account to the General channel as a member
			_, err = membership_repo.Add(ctx, generalID, evt.ID, channel_membership.RoleMember)
			if err != nil {
				// Log error but don't fail - it's not critical if this fails
				return fault.Wrap(err, fctx.With(ctx))
			}
			return nil
		})
		return err
	}))
}

func (s *service) Create(ctx context.Context, accountID account.AccountID, partial Partial) (*channel.Channel, error) {
	name, ok := partial.Name.Get()
	if !ok {
		return nil, fault.Wrap(errInvalidCreate, fctx.With(ctx), fmsg.WithDesc("missing name", "Channel name is required."))
	}

	slug, ok := partial.Slug.Get()
	if !ok {
		return nil, fault.Wrap(errInvalidCreate, fctx.With(ctx), fmsg.WithDesc("missing slug", "Channel slug is required."))
	}

	description, ok := partial.Description.Get()
	if !ok {
		description = ""
	}

	visibility, ok := partial.Visibility.Get()
	if !ok {
		visibility = channel.VisibilityPublic
	}

	opts := []channel.Option{}

	coverImage, _ := partial.CoverImageAssetID.Get()
	if v, ok := coverImage.Get(); ok {
		opts = append(opts, channel.WithCoverImageAssetID(v))
	}

	icon, _ := partial.IconAssetID.Get()
	if v, ok := icon.Get(); ok {
		opts = append(opts, channel.WithIconAssetID(v))
	}

	if v, ok := partial.Metadata.Get(); ok {
		opts = append(opts, channel.WithMetadata(v))
	}

	ch, err := s.channel_repo.Create(ctx, name, slug, description, visibility, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	_, err = s.membership_repo.Add(ctx, xid.ID(ch.ID), accountID, channel_membership.RoleOwner)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return ch, nil
}

func (s *service) Get(ctx context.Context, id channel.ChannelID) (*channel.Channel, error) {
	ch, err := s.channel_repo.Get(ctx, id)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return ch, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*channel.Channel, error) {
	ch, err := s.channel_repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return ch, nil
}

func (s *service) GetUserChannels(ctx context.Context, accountID account.AccountID) ([]*channel.Channel, error) {
	channelIDs, err := s.membership_repo.GetAccountChannelIDs(ctx, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	channels := make([]*channel.Channel, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		ch, err := s.channel_repo.Get(ctx, channel.ChannelID(channelID))
		if err != nil {
			continue
		}
		channels = append(channels, ch)
	}

	return channels, nil
}

func (s *service) Update(ctx context.Context, id channel.ChannelID, partial Partial) (*channel.Channel, error) {
	opts := []channel.Option{}

	if v, ok := partial.Name.Get(); ok {
		opts = append(opts, channel.WithName(v))
	}
	if v, ok := partial.Slug.Get(); ok {
		opts = append(opts, channel.WithSlug(v))
	}
	if v, ok := partial.Description.Get(); ok {
		opts = append(opts, channel.WithDescription(v))
	}
	if v, ok := partial.Visibility.Get(); ok {
		opts = append(opts, channel.WithVisibility(v))
	}

	coverImageOpt, shouldDelete := partial.CoverImageAssetID.Get()
	if shouldDelete {
		opts = append(opts, channel.WithCoverImageAssetID(nil))
	} else if v, ok := coverImageOpt.Get(); ok {
		opts = append(opts, channel.WithCoverImageAssetID(v))
	}

	iconOpt, shouldDelete := partial.IconAssetID.Get()
	if shouldDelete {
		opts = append(opts, channel.WithIconAssetID(nil))
	} else if v, ok := iconOpt.Get(); ok {
		opts = append(opts, channel.WithIconAssetID(v))
	}

	if v, ok := partial.Metadata.Get(); ok {
		opts = append(opts, channel.WithMetadata(v))
	}

	ch, err := s.channel_repo.Update(ctx, id, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return ch, nil
}

func (s *service) Delete(ctx context.Context, id channel.ChannelID) error {
	err := s.channel_repo.Delete(ctx, id)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (s *service) SearchPublic(ctx context.Context, query string) ([]*channel.Channel, error) {
	channels, err := s.channel_repo.SearchPublic(ctx, query)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return channels, nil
}

func (s *service) CheckAccess(ctx context.Context, channelID channel.ChannelID, accountID account.AccountID) (bool, error) {
	ch, err := s.channel_repo.Get(ctx, channelID)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx))
	}

	if ch.Visibility == channel.VisibilityPublic {
		isMember, err := s.membership_repo.CheckMembership(ctx, xid.ID(channelID), accountID)
		if err != nil {
			return false, fault.Wrap(err, fctx.With(ctx))
		}
		return isMember, nil
	}

	isMember, err := s.membership_repo.CheckMembership(ctx, xid.ID(channelID), accountID)
	if err != nil {
		return false, fault.Wrap(err, fctx.With(ctx))
	}

	return isMember, nil
}
