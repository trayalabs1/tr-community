package bindings

import (
	"context"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/ftag"

	"github.com/Southclaws/storyden/app/resources/analytics"
	"github.com/Southclaws/storyden/app/resources/rbac"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/app/transports/http/openapi"
)

type Analytics struct {
	querier *analytics.Querier
}

func NewAnalytics(querier *analytics.Querier) *Analytics {
	return &Analytics{querier: querier}
}

func (a *Analytics) AdminAnalyticsGet(ctx context.Context, request openapi.AdminAnalyticsGetRequestObject) (openapi.AdminAnalyticsGetResponseObject, error) {
	if err := session.Authorise(ctx, nil, rbac.PermissionAdministrator); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	start, end := request.Params.Start, request.Params.End
	if !end.After(start) {
		return nil, fault.Wrap(fault.New("end must be after start"), fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	report, err := a.querier.GetReport(ctx, start, end)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return openapi.AdminAnalyticsGet200JSONResponse{
		AdminAnalyticsGetOKJSONResponse: openapi.AdminAnalyticsGetOKJSONResponse{
			ChannelOnboardings: dt.Map(report.ChannelOnboardings, func(s analytics.ChannelStat) openapi.ChannelAnalyticsStat {
				return openapi.ChannelAnalyticsStat{ChannelName: s.ChannelName, Count: s.Count}
			}),
			ChannelPosts: dt.Map(report.ChannelPosts, func(s analytics.ChannelStat) openapi.ChannelAnalyticsStat {
				return openapi.ChannelAnalyticsStat{ChannelName: s.ChannelName, Count: s.Count}
			}),
			AdminReplyTimes: dt.Map(report.AdminReplyTimes, func(s analytics.AdminReplyTimeStat) openapi.AdminReplyTimeStat {
				return openapi.AdminReplyTimeStat{AdminHandle: s.AdminHandle, AvgTimeMinutes: s.AvgTimeMinutes}
			}),
		},
	}, nil
}
