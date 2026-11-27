package crosspost

import (
	"context"
	"fmt"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/mark"
	"github.com/Southclaws/storyden/app/resources/message"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/internal/ent"
	ent_channel "github.com/Southclaws/storyden/internal/ent/channel"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
	"github.com/Southclaws/storyden/internal/infrastructure/pubsub"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
)

type Crossposter struct {
	db  *ent.Client
	bus *pubsub.Bus
}

func newCrossposter(db *ent.Client, bus *pubsub.Bus) *Crossposter {
	return &Crossposter{db: db, bus: bus}
}

func (c *Crossposter) findCopy(ctx context.Context, sourceID xid.ID) (*ent.Post, error) {
	p, err := c.db.Post.Query().
		Where(
			ent_post.DeletedAtIsNil(),
			func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(ent_post.FieldMetadata, sourceID.String(), sqljson.Path(post.MetaCrosspostSourceID)))
			},
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	return p, nil
}

func (c *Crossposter) loadSource(ctx context.Context, id xid.ID) (*ent.Post, error) {
	p, err := c.db.Post.Query().
		Where(ent_post.IDEQ(id), ent_post.DeletedAtIsNil()).
		WithChannel().
		WithAuthor().
		WithAssets().
		WithTags().
		WithSentiment().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fault.Wrap(err, fctx.With(ctx))
	}
	return p, nil
}

func toCandidate(p *ent.Post) candidate {
	cand := candidate{
		IsRoot:     p.RootPostID == nil,
		IsShare:    p.ReferencePostID != nil,
		Visibility: p.Visibility.String(),
	}
	if p.Metadata != nil {
		_, cand.IsCopy = p.Metadata[post.MetaCrosspostSourceID]
	}
	if p.Edges.Channel != nil {
		cand.ChannelSlug = p.Edges.Channel.Slug
	}
	if a := p.Edges.Author; a != nil {
		cand.IsAdmin = a.Admin
		if v, ok := a.Metadata["case_id"]; ok && v != nil {
			cand.CaseID = fmt.Sprint(v)
		}
	}
	if s := p.Edges.Sentiment; s != nil {
		cand.SentimentStatus = s.ScoringStatus.String()
		if s.SentimentTag != nil {
			cand.SentimentTag = *s.SentimentTag
		}
		if s.PrimaryTopic != nil {
			cand.Topic = *s.PrimaryTopic
		}
	}
	return cand
}

func (c *Crossposter) OnScored(ctx context.Context, id xid.ID) error {
	src, err := c.loadSource(ctx, id)
	if err != nil || src == nil {
		return err
	}

	existing, err := c.findCopy(ctx, id)
	if err != nil || existing != nil {
		return err
	}

	return c.create(ctx, src)
}

func (c *Crossposter) OnPublished(ctx context.Context, id xid.ID) error {
	src, err := c.loadSource(ctx, id)
	if err != nil || src == nil {
		return err
	}

	existing, err := c.findCopy(ctx, id)
	if err != nil {
		return err
	}
	if existing != nil {
		return c.setVisibility(ctx, existing, ent_post.VisibilityPublished)
	}

	return c.create(ctx, src)
}

func (c *Crossposter) OnUnpublished(ctx context.Context, id xid.ID) error {
	existing, err := c.findCopy(ctx, id)
	if err != nil || existing == nil {
		return err
	}

	src, err := c.loadSource(ctx, id)
	if err != nil {
		return err
	}
	if src == nil {
		return c.setVisibility(ctx, existing, ent_post.VisibilityArchived)
	}

	return c.setVisibility(ctx, existing, src.Visibility)
}

func (c *Crossposter) OnDeleted(ctx context.Context, id xid.ID) error {
	existing, err := c.findCopy(ctx, id)
	if err != nil || existing == nil {
		return err
	}
	return c.setVisibility(ctx, existing, ent_post.VisibilityArchived)
}

func (c *Crossposter) setVisibility(ctx context.Context, copy *ent.Post, v ent_post.Visibility) error {
	if copy.Visibility == v {
		return nil
	}

	err := c.db.Post.UpdateOneID(copy.ID).SetVisibility(v).SetUpdatedAt(time.Now()).Exec(ctx)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	if v == ent_post.VisibilityPublished {
		c.bus.Publish(ctx, &message.EventThreadPublished{ID: post.ID(copy.ID)})
	} else {
		c.bus.Publish(ctx, &message.EventThreadUnpublished{ID: post.ID(copy.ID)})
	}

	return nil
}

func (c *Crossposter) create(ctx context.Context, src *ent.Post) error {
	if !toCandidate(src).eligible() {
		return nil
	}

	target, err := c.db.Channel.Query().Where(ent_channel.Slug(TargetChannelSlug)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	meta := make(map[string]any, len(src.Metadata)+2)
	for k, v := range src.Metadata {
		meta[k] = v
	}
	meta[post.MetaCrosspostSourceID] = src.ID.String()
	meta[post.MetaCrosspostSourceChannelID] = src.ChannelID.String()
	meta[post.MetaCrosspostSourceChannelName] = src.Edges.Channel.Name
	meta[post.MetaCrosspostSourceChannelSlug] = src.Edges.Channel.Slug

	now := time.Now()
	create := c.db.Post.Create().
		SetTitle(src.Title).
		SetBody(src.Body).
		SetShort(src.Short).
		SetMetadata(meta).
		SetVisibility(ent_post.VisibilityPublished).
		SetChannelID(target.ID).
		SetAccountPosts(src.AccountPosts).
		SetLastReplyAt(now).
		SetUpdatedAt(now)

	if !src.LinkID.IsNil() {
		create.SetLinkID(src.LinkID)
	}
	for _, a := range src.Edges.Assets {
		create.AddAssetIDs(a.ID)
	}
	for _, t := range src.Edges.Tags {
		create.AddTagIDs(t.ID)
	}

	copy, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return fault.Wrap(err, fctx.With(ctx))
	}

	err = c.db.Post.UpdateOneID(copy.ID).
		SetSlug(fmt.Sprintf("%s-%s", copy.ID, mark.Slugify(src.Title))).
		Exec(ctx)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	if s := src.Edges.Sentiment; s != nil {
		sc := c.db.PostSentiment.Create().
			SetPostID(copy.ID).
			SetScoringStatus(s.ScoringStatus).
			SetRankScore(s.RankScore).
			SetNillableSentimentTag(s.SentimentTag).
			SetNillablePositivityScore(s.PositivityScore).
			SetNillablePrimaryTopic(s.PrimaryTopic).
			SetNillableFeedValueScore(s.FeedValueScore)
		if err := sc.Exec(ctx); err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}
	}

	c.bus.Publish(ctx, &message.EventThreadPublished{ID: post.ID(copy.ID)})

	return nil
}
