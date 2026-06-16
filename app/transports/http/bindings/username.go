package bindings

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/account/username"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type Username struct {
	usernameService *username.Service
	accountQuery    *account_querier.Querier
}

func NewUsername(usernameService *username.Service, accountQuery *account_querier.Querier) Username {
	return Username{
		usernameService: usernameService,
		accountQuery:    accountQuery,
	}
}

// UsernameCheck checks if a username is available
func (u *Username) UsernameCheck(ctx context.Context, request openapi.UsernameCheckRequestObject) (openapi.UsernameCheckResponseObject, error) {
	available, err := u.usernameService.CheckAvailability(ctx, request.Params.Username)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.UsernameCheck200JSONResponse{
		Available: available,
	}, nil
}

// UsernameSet sets the username for the authenticated account
func (u *Username) UsernameSet(ctx context.Context, request openapi.UsernameSetRequestObject) (openapi.UsernameSetResponseObject, error) {
	accountID, err := session.GetAccountID(ctx)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	_, err = u.usernameService.SetUsername(ctx, accountID, request.Body.Username)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Fetch the updated account with edges for serialization
	acc, err := u.accountQuery.GetByID(ctx, accountID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.UsernameSet200JSONResponse{
		AccountUpdateOKJSONResponse: openapi.AccountUpdateOKJSONResponse(serialiseAccount(acc)),
	}, nil
}

// AdminRegenerateTempHandles regenerates handles for all accounts that still
// carry a temporary handle assigned during signup.
func (u *Username) AdminRegenerateTempHandles(ctx context.Context, request openapi.AdminRegenerateTempHandlesRequestObject) (openapi.AdminRegenerateTempHandlesResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	var cursor string
	if request.Params.After != nil {
		cursor = *request.Params.After
	}

	var batchSize int
	if request.Params.Limit != nil {
		batchSize = *request.Params.Limit
	}

	result, err := u.usernameService.RegenerateTempHandles(ctx, cursor, batchSize)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	var nextCursor *string
	if result.NextCursor != "" {
		nextCursor = &result.NextCursor
	}

	return openapi.AdminRegenerateTempHandles200JSONResponse{
		AdminRegenerateTempHandlesOKJSONResponse: openapi.AdminRegenerateTempHandlesOKJSONResponse{
			Processed:  result.Processed,
			Updated:    result.Updated,
			Skipped:    result.Skipped,
			Failed:     result.Failed,
			NextCursor: nextCursor,
		},
	}, nil
}
