package poll_vote

import (
	"context"

	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/post"
	ent "github.com/Southclaws/storyden/internal/ent"
	db "github.com/Southclaws/storyden/internal/ent/pollvote"
)

type DB struct {
	db *ent.Client
}

func New(db *ent.Client) *DB {
	return &DB{db: db}
}

type OptionDef struct {
	ID   string
	Text string
}

type OptionCount struct {
	ID    string
	Text  string
	Votes int
}

type Status struct {
	TotalVotes int
	UserVote   opt.Optional[string]
	Options    []OptionCount
}

func (d *DB) Vote(ctx context.Context, postID post.ID, accountID account.AccountID, optionID string) error {
	return d.db.PollVote.Create().
		SetPostID(xid.ID(postID)).
		SetAccountID(xid.ID(accountID)).
		SetOptionID(optionID).
		OnConflictColumns(db.FieldAccountID, db.FieldPostID).
		UpdateOptionID().
		Exec(ctx)
}

func (d *DB) GetStatus(ctx context.Context, postID post.ID, accountID opt.Optional[account.AccountID], optionDefs []OptionDef) (*Status, error) {
	votes, err := d.db.PollVote.Query().
		Where(db.PostID(xid.ID(postID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	userVote := opt.NewEmpty[string]()
	for _, v := range votes {
		counts[v.OptionID]++
		if aid, ok := accountID.Get(); ok && v.AccountID == xid.ID(aid) {
			userVote = opt.New(v.OptionID)
		}
	}

	options := make([]OptionCount, len(optionDefs))
	for i, def := range optionDefs {
		options[i] = OptionCount{ID: def.ID, Text: def.Text, Votes: counts[def.ID]}
	}

	return &Status{
		TotalVotes: len(votes),
		UserVote:   userVote,
		Options:    options,
	}, nil
}
