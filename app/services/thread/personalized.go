package thread

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/alitto/pond/v2"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/thread"
	"github.com/Southclaws/storyden/app/services/authentication/session"
)

const personalizedFeedWindow = 24 * time.Hour
const personalizedSimilarLimit = 1

type PersonalizedSimilarGroup struct {
	ForThreadID post.ID
	Threads     []*thread.Thread
}

type PersonalizedFeed struct {
	SelfRecent []*thread.Thread
	Similar    []PersonalizedSimilarGroup
}

func (s *service) Personalized(ctx context.Context, channelID xid.ID) (*PersonalizedFeed, error) {
	accountIDOpt := session.GetOptAccountID(ctx)
	accountID, ok := accountIDOpt.Get()
	if !ok {
		return &PersonalizedFeed{
			SelfRecent: []*thread.Thread{},
			Similar:    []PersonalizedSimilarGroup{},
		}, nil
	}

	since := time.Now().Add(-personalizedFeedWindow)

	selfRecent, err := s.threadQuerier.ListSelfRecent(ctx, channelID, accountID, since)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if len(selfRecent) == 0 {
		return &PersonalizedFeed{
			SelfRecent: []*thread.Thread{},
			Similar:    []PersonalizedSimilarGroup{},
		}, nil
	}

	threads := make([]*thread.Thread, len(selfRecent))
	groups := make([]PersonalizedSimilarGroup, len(selfRecent))
	pool := pond.NewGroup()

	for i, entry := range selfRecent {
		i, entry := i, entry
		threads[i] = entry.Thread
		groups[i] = PersonalizedSimilarGroup{ForThreadID: entry.Thread.ID, Threads: []*thread.Thread{}}

		if entry.SentimentTag == "" || entry.PrimaryTopic == "" {
			continue
		}

		pool.SubmitErr(func() error {
			similar, err := s.threadQuerier.ListSimilarFor(
				ctx,
				channelID,
				entry.Thread.ID,
				accountID,
				entry.SentimentTag,
				entry.PrimaryTopic,
				personalizedSimilarLimit,
				accountIDOpt,
			)
			if err != nil {
				return fault.Wrap(err, fctx.With(ctx))
			}
			groups[i].Threads = similar
			return nil
		})
	}

	if err := pool.Wait(); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return &PersonalizedFeed{
		SelfRecent: threads,
		Similar:    groups,
	}, nil
}
