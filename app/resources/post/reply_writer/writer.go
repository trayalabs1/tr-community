package reply_writer

import (
	"context"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/post"
	"github.com/Southclaws/storyden/app/resources/post/reply"
	"github.com/Southclaws/storyden/app/resources/post/reply_querier"
	"github.com/Southclaws/storyden/app/resources/visibility"
	"github.com/Southclaws/storyden/internal/ent"
	ent_post "github.com/Southclaws/storyden/internal/ent/post"
)

type Writer struct {
	db      *ent.Client
	querier *reply_querier.Querier
}

func New(db *ent.Client, querier *reply_querier.Querier) *Writer {
	return &Writer{db: db, querier: querier}
}

type Option func(*ent.PostMutation)

func WithID(id post.ID) Option {
	return func(pm *ent.PostMutation) {
		pm.SetID(xid.ID(id))
	}
}

func WithContent(v datagraph.Content) Option {
	return func(pm *ent.PostMutation) {
		pm.SetBody(v.HTML())
		pm.SetShort(v.Short())
	}
}

func WithReplyTo(v post.ID) Option {
	return func(pm *ent.PostMutation) {
		pm.SetReplyToID(xid.ID(v))
	}
}

func WithMeta(meta map[string]any) Option {
	return func(m *ent.PostMutation) {
		m.SetMetadata(meta)
	}
}

func WithAssets(ids ...asset.AssetID) Option {
	return func(m *ent.PostMutation) {
		m.AddAssetIDs(ids...)
	}
}

func WithLink(id xid.ID) Option {
	return func(pm *ent.PostMutation) {
		pm.SetLinkID(id)
	}
}

func WithContentLinks(ids ...xid.ID) Option {
	return func(pm *ent.PostMutation) {
		pm.AddContentLinkIDs(ids...)
	}
}

func WithVisibility(v visibility.Visibility) Option {
	return func(pm *ent.PostMutation) {
		pm.SetVisibility(ent_post.Visibility(v.String()))
	}
}

func WithChannel(id xid.ID) Option {
	return func(pm *ent.PostMutation) {
		pm.SetChannelID(id)
	}
}

func (d *Writer) Create(
	ctx context.Context,
	authorID account.AccountID,
	parentID post.ID,
	opts ...Option,
) (*reply.Reply, error) {
	tx, err := d.db.Tx(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to start transaction"), fctx.With(ctx))
	}
	defer tx.Rollback()

	thread, err := tx.Post.Get(ctx, xid.ID(parentID))
	if err != nil {
		if ent.IsNotFound(err) {
			err = fault.Wrap(err, ftag.With(ftag.NotFound), fmsg.WithDesc("not found", "Thread not found."))
		}

		return nil, fault.Wrap(err, fmsg.With("failed to get parent thread"), fctx.With(ctx))
	}

	if thread.RootPostID != nil {
		return nil, fault.New("attempt to create post under non-thread post", fmsg.WithDesc("invalid parent", "Cannot reply to a non-thread post."))
	}

	q := tx.Post.
		Create().
		SetUpdatedAt(time.Now()).
		SetLastReplyAt(time.Now()).
		SetRootID(xid.ID(parentID)).
		SetAuthorID(xid.ID(authorID)).
		SetChannelID(thread.ChannelID)

	for _, fn := range opts {
		fn(q.Mutation())
	}

	p, err := q.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			err = fault.Wrap(err, ftag.With(ftag.InvalidArgument), fmsg.WithDesc("invalid reply", "Failed to create reply."))
		}

		return nil, fault.Wrap(err, fmsg.With("failed to create reply"), fctx.With(ctx))
	}

	p, err = tx.Post.Query().
		Where(ent_post.IDEQ(p.ID)).
		WithAuthor().
		WithRoot(func(pq *ent.PostQuery) {
			pq.WithAuthor()
		}).
		WithAssets().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = fault.Wrap(err, ftag.With(ftag.NotFound), fmsg.WithDesc("not found", "Reply not found after creation."))
		}

		return nil, fault.Wrap(err, fmsg.With("failed to query created reply"), fctx.With(ctx))
	}

	err = tx.Post.
		UpdateOneID(xid.ID(parentID)).
		SetLastReplyAt(time.Now()).
		Exec(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to update parent thread timestamp"), fctx.With(ctx))
	}

	if err = tx.Commit(); err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to commit transaction"), fctx.With(ctx))
	}

	return d.querier.Get(ctx, post.ID(p.ID))
}

type BulkItem struct {
	ParentID post.ID
	Opts     []Option
}

// CreateBulk inserts many replies in a single multi-row INSERT inside one
// transaction, then updates each affected thread's LastReplyAt. Parents that
// don't exist or aren't threads are skipped. Returns the created reply IDs
// paired with their parent thread IDs.
func (d *Writer) CreateBulk(
	ctx context.Context,
	authorID account.AccountID,
	items []BulkItem,
) ([]CreatedReply, error) {
	if len(items) == 0 {
		return nil, nil
	}

	tx, err := d.db.Tx(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to start transaction"), fctx.With(ctx))
	}
	defer tx.Rollback()

	parentIDs := make([]xid.ID, 0, len(items))
	seen := map[xid.ID]struct{}{}
	for _, item := range items {
		pid := xid.ID(item.ParentID)
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		parentIDs = append(parentIDs, pid)
	}

	parents, err := tx.Post.Query().
		Where(ent_post.IDIn(parentIDs...), ent_post.RootPostIDIsNil()).
		Select(ent_post.FieldID, ent_post.FieldChannelID, ent_post.FieldAccountPosts).
		All(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to load parent threads"), fctx.With(ctx))
	}

	channelByParent := make(map[xid.ID]xid.ID, len(parents))
	authorByParent := make(map[xid.ID]xid.ID, len(parents))
	for _, p := range parents {
		channelByParent[p.ID] = p.ChannelID
		authorByParent[p.ID] = p.AccountPosts
	}

	now := time.Now()
	builders := make([]*ent.PostCreate, 0, len(items))
	parentForBuilder := make([]xid.ID, 0, len(items))
	for _, item := range items {
		pid := xid.ID(item.ParentID)
		channelID, ok := channelByParent[pid]
		if !ok {
			// Parent missing or not a thread — skip it.
			continue
		}

		b := tx.Post.Create().
			SetUpdatedAt(now).
			SetLastReplyAt(now).
			SetRootID(pid).
			SetAuthorID(xid.ID(authorID)).
			SetChannelID(channelID)

		for _, fn := range item.Opts {
			fn(b.Mutation())
		}

		builders = append(builders, b)
		parentForBuilder = append(parentForBuilder, pid)
	}

	if len(builders) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to commit transaction"), fctx.With(ctx))
		}
		return nil, nil
	}

	saved, err := tx.Post.CreateBulk(builders...).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			err = fault.Wrap(err, ftag.With(ftag.InvalidArgument), fmsg.WithDesc("invalid reply", "Failed to create replies."))
		}
		return nil, fault.Wrap(err, fmsg.With("failed to create replies"), fctx.With(ctx))
	}

	// Bump LastReplyAt for every parent that received a reply, in one UPDATE.
	bumpedParents := make([]xid.ID, 0, len(parentForBuilder))
	bumped := map[xid.ID]struct{}{}
	for _, pid := range parentForBuilder {
		if _, ok := bumped[pid]; ok {
			continue
		}
		bumped[pid] = struct{}{}
		bumpedParents = append(bumpedParents, pid)
	}
	if err := tx.Post.Update().
		Where(ent_post.IDIn(bumpedParents...)).
		SetLastReplyAt(now).
		Exec(ctx); err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to update parent thread timestamps"), fctx.With(ctx))
	}

	if err := tx.Commit(); err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to commit transaction"), fctx.With(ctx))
	}

	created := make([]CreatedReply, len(saved))
	for i, p := range saved {
		parentID := parentForBuilder[i]
		created[i] = CreatedReply{
			ReplyID:        post.ID(p.ID),
			ParentID:       post.ID(parentID),
			ThreadAuthorID: account.AccountID(authorByParent[parentID]),
		}
	}
	return created, nil
}

type CreatedReply struct {
	ReplyID        post.ID
	ParentID       post.ID
	ThreadAuthorID account.AccountID
}

func (d *Writer) Update(ctx context.Context, id post.ID, opts ...Option) (*reply.Reply, error) {
	update := d.db.Post.UpdateOneID(xid.ID(id))
	mutate := update.Mutation()

	for _, fn := range opts {
		fn(mutate)
	}

	if _, set := mutate.IndexedAt(); !set {
		mutate.SetUpdatedAt(time.Now())
	}

	err := update.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = fault.Wrap(err, ftag.With(ftag.NotFound), fmsg.WithDesc("not found", "Reply not found."))
		}

		return nil, fault.Wrap(err, fmsg.With("failed to update reply"), fctx.With(ctx))
	}

	p, err := d.db.Post.
		Query().
		Where(ent_post.IDEQ(xid.ID(id))).
		WithAuthor().
		WithRoot(func(pq *ent.PostQuery) {
			pq.WithAuthor()
		}).
		WithAssets().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = fault.Wrap(err, ftag.With(ftag.NotFound), fmsg.WithDesc("not found", "Reply not found after update."))
		}

		return nil, fault.Wrap(err, fmsg.With("failed to query updated reply"), fctx.With(ctx))
	}

	return reply.Map(p)
}

func (d *Writer) Delete(ctx context.Context, id post.ID) error {
	err := d.db.Post.
		UpdateOneID(xid.ID(id)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = fault.Wrap(err, ftag.With(ftag.NotFound), fmsg.WithDesc("not found", "Reply not found."))
		}

		return fault.Wrap(err, fmsg.With("failed to delete reply"), fctx.With(ctx))
	}

	return nil
}
