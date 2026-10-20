package otelroute

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			route := c.Path()
			if route == "" {
				return next(c)
			}

			ctx := c.Request().Context()

			if labeler, ok := otelhttp.LabelerFromContext(ctx); ok {
				labeler.Add(semconv.HTTPRoute(route))
			}

			span := trace.SpanFromContext(ctx)
			if span.IsRecording() {
				span.SetName(c.Request().Method + " " + route)
				span.SetAttributes(semconv.HTTPRoute(route))
			}

			return next(c)
		}
	}
}
