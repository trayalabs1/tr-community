package thread_share

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread_querier"
	"github.com/Southclaws/storyden/app/resources/post/thread_share"
	"github.com/Southclaws/storyden/app/services/authentication/session"
)

var errNotRootThread = fault.New("thread is not a root thread", ftag.With(ftag.InvalidArgument))

type Service interface {
	Share(ctx context.Context, threadID post.ID, channelIDs []xid.ID, subtitle string) ([]*thread_share.Share, error)
	Unshare(ctx context.Context, threadID post.ID, channelID xid.ID) error
	SetPin(ctx context.Context, threadID post.ID, channelID xid.ID, pinned bool) (*thread_share.Share, error)
	ListForThread(ctx context.Context, threadID post.ID) ([]*thread_share.Share, error)
}

func Build() fx.Option {
	return fx.Options(fx.Provide(New))
}

type service struct {
	share_repo     *thread_share.Repository
	thread_querier *thread_querier.Querier
}

func New(share_repo *thread_share.Repository, thread_querier *thread_querier.Querier) Service {
	return &service{
		share_repo:     share_repo,
		thread_querier: thread_querier,
	}
}

func (s *service) Share(ctx context.Context, threadID post.ID, channelIDs []xid.ID, subtitle string) ([]*thread_share.Share, error) {
	ref, err := s.thread_querier.Probe(ctx, threadID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if ref.Root != threadID {
		return nil, fault.Wrap(errNotRootThread, fctx.With(ctx))
	}

	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	shares, err := s.share_repo.CreateMany(ctx, xid.ID(threadID), channelIDs, xid.ID(accountID), subtitle)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return shares, nil
}

func (s *service) Unshare(ctx context.Context, threadID post.ID, channelID xid.ID) error {
	err := s.share_repo.Delete(ctx, xid.ID(threadID), channelID)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}

func (s *service) SetPin(ctx context.Context, threadID post.ID, channelID xid.ID, pinned bool) (*thread_share.Share, error) {
	share, err := s.share_repo.SetPin(ctx, xid.ID(threadID), channelID, pinned)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return share, nil
}

func (s *service) ListForThread(ctx context.Context, threadID post.ID) ([]*thread_share.Share, error) {
	shares, err := s.share_repo.ListByPost(ctx, xid.ID(threadID))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return shares, nil
}
