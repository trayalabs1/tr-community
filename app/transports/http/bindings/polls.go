package bindings

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"
	"github.com/Southclaws/opt"
	"github.com/oapi-codegen/nullable"

	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/resources/poll_vote"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	thread_service "github.com/Southclaws/storyden/app/services/thread"
	"github.com/Southclaws/storyden/app/services/thread_mark"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type Polls struct {
	thread_mark_svc thread_mark.Service
	thread_svc      thread_service.Service
	db              *poll_vote.DB
}

func NewPolls(
	thread_mark_svc thread_mark.Service,
	thread_svc thread_service.Service,
	db *poll_vote.DB,
) Polls {
	return Polls{thread_mark_svc, thread_svc, db}
}

func (i *Polls) ThreadGetPoll(ctx context.Context, request openapi.ThreadGetPollRequestObject) (openapi.ThreadGetPollResponseObject, error) {
	postID, err := i.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	accountID := session.GetOptAccountID(ctx)

	thread, err := i.thread_svc.Get(ctx, postID, pagination.Parameters{})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	optionDefs := deserialisePollOptionDefs(thread.Meta)

	status, err := i.db.GetStatus(ctx, postID, accountID, optionDefs)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadGetPoll200JSONResponse{
		ThreadGetPollOKJSONResponse: openapi.ThreadGetPollOKJSONResponse(serialisePollStatus(status)),
	}, nil
}

func (i *Polls) ThreadVotePoll(ctx context.Context, request openapi.ThreadVotePollRequestObject) (openapi.ThreadVotePollResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	postID, err := i.thread_mark_svc.Lookup(ctx, string(request.ThreadMark))
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	thread, err := i.thread_svc.Get(ctx, postID, pagination.Parameters{})
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	optionDefs := deserialisePollOptionDefs(thread.Meta)

	validOption := false
	for _, def := range optionDefs {
		if def.ID == request.Body.OptionId {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, fault.Wrap(
			fault.New("invalid poll option"),
			fctx.With(ctx),
			ftag.With(ftag.InvalidArgument),
		)
	}

	if err := i.db.Vote(ctx, postID, accountID, request.Body.OptionId); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	status, err := i.db.GetStatus(ctx, postID, opt.New(accountID), optionDefs)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.ThreadVotePoll200JSONResponse{
		ThreadVotePollOKJSONResponse: openapi.ThreadVotePollOKJSONResponse(serialisePollStatus(status)),
	}, nil
}

func deserialisePollOptionDefs(meta map[string]any) []poll_vote.OptionDef {
	raw, ok := meta["poll_options"].([]any)
	if !ok {
		return nil
	}

	return dt.Map(raw, func(item any) poll_vote.OptionDef {
		m, _ := item.(map[string]any)
		id, _ := m["id"].(string)
		text, _ := m["text"].(string)
		return poll_vote.OptionDef{ID: id, Text: text}
	})
}

func serialisePollStatus(s *poll_vote.Status) openapi.PollStatus {
	var userVote nullable.Nullable[string]
	if v, ok := s.UserVote.Get(); ok {
		userVote = nullable.NewNullableWithValue(v)
	}

	return openapi.PollStatus{
		TotalVotes: s.TotalVotes,
		UserVote:   userVote,
		Options: dt.Map(s.Options, func(o poll_vote.OptionCount) openapi.PollOptionStatus {
			return openapi.PollOptionStatus{
				Id:    o.ID,
				Text:  o.Text,
				Votes: o.Votes,
			}
		}),
	}
}
