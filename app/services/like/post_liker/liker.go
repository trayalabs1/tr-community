package post_liker

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/like/like_writer"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/post_querier"
	"github.com/Southclaws/storyden/app/resources/post/thread_cache"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"
)

type PostLiker struct {
	likeWriter  *like_writer.LikeWriter
	postQuerier *post_querier.Querier
	bus         *pubsub.Bus
	cache       *thread_cache.Cache
}

func New(
	likeWriter *like_writer.LikeWriter,
	postQuerier *post_querier.Querier,
	bus *pubsub.Bus,
	cache *thread_cache.Cache,
) *PostLiker {
	return &PostLiker{
		likeWriter:  likeWriter,
		postQuerier: postQuerier,
		bus:         bus,
		cache:       cache,
	}
}

func (l *PostLiker) AddPostLike(ctx context.Context, accountID account.AccountID, postID post.ID) error {
	postRef, err := l.postQuerier.Probe(ctx, postID)
	if err != nil {
		return err
	}

	if err := l.cache.Invalidate(ctx, xid.ID(postRef.Root)); err != nil {
		return err
	}

	err = l.likeWriter.AddPostLike(ctx, accountID, postID)
	if err != nil {
		return err
	}

	l.bus.Publish(ctx, &message.EventPostLiked{
		PostID:     postID,
		RootPostID: postRef.Root,
		LikerID:    accountID,
	})

	return nil
}

// AddPostLikes likes many posts on behalf of one account using a single bulk
// DB insert. Posts that don't resolve are skipped; the count of posts included
// in the bulk like is returned so callers can report partial success.
func (l *PostLiker) AddPostLikes(ctx context.Context, accountID account.AccountID, postIDs []post.ID) (int, error) {
	refs, err := l.postQuerier.ProbeMany(ctx, postIDs)
	if err != nil {
		return 0, err
	}
	if len(refs) == 0 {
		return 0, nil
	}

	ids := dt.Map(refs, func(r *post.PostRef) post.ID { return r.ID })
	if err := l.likeWriter.AddPostLikes(ctx, accountID, ids); err != nil {
		return 0, err
	}

	invalidated := map[xid.ID]struct{}{}
	for _, r := range refs {
		root := xid.ID(r.Root)
		if _, ok := invalidated[root]; !ok {
			invalidated[root] = struct{}{}
			if err := l.cache.Invalidate(ctx, root); err != nil {
				return 0, err
			}
		}

		l.bus.Publish(ctx, &message.EventPostLiked{
			PostID:     r.ID,
			RootPostID: r.Root,
			LikerID:    accountID,
		})
	}

	return len(refs), nil
}

func (l *PostLiker) RemovePostLike(ctx context.Context, accountID account.AccountID, postID post.ID) error {
	postRef, err := l.postQuerier.Probe(ctx, postID)
	if err != nil {
		return err
	}

	if err := l.cache.Invalidate(ctx, xid.ID(postRef.Root)); err != nil {
		return err
	}

	err = l.likeWriter.RemovePostLike(ctx, accountID, postID)
	if err != nil {
		return err
	}

	l.bus.Publish(ctx, &message.EventPostUnliked{
		PostID:     postID,
		RootPostID: postRef.Root,
	})

	return nil
}
