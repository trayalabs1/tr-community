package bindings

import (
	"context"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/services/authentication/provider/traya"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/middleware/session_cookie"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type TrayaAuth struct {
	tp *traya.Provider
	cj *session_cookie.Jar
	si *session.Issuer
}

func NewTrayaAuth(tp *traya.Provider, cj *session_cookie.Jar, si *session.Issuer) TrayaAuth {
	return TrayaAuth{tp, cj, si}
}

func (i *TrayaAuth) AuthTrayaToken(ctx context.Context, request openapi.AuthTrayaTokenRequestObject) (openapi.AuthTrayaTokenResponseObject, error) {
	acc, err := i.tp.AuthenticateWithToken(ctx, request.Params.Token)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	t, err := i.si.Issue(ctx, acc.ID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Check if user needs to set a username (temporary handle)
	needsUsername := strings.HasPrefix(acc.Handle, "temp_")

	return openapi.AuthTrayaToken200JSONResponse{
		AuthSuccessOKJSONResponse: openapi.AuthSuccessOKJSONResponse{
			Body: openapi.AuthSuccess{
				Id:            acc.ID.String(),
				NeedsUsername: &needsUsername,
			},
			Headers: openapi.AuthSuccessOKResponseHeaders{
				SetCookie: i.cj.Create(*t).String(),
			},
		},
	}, nil
}
