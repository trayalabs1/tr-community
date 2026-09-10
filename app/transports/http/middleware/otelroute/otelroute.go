package otelroute

import (
	"github.com/labstack/echo/v4"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/Southclaws/storyden/internal/infrastructure/httpserver/routename"
)

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			route := c.Path()
			if route == "" {
				return next(c)
			}

			routename.Set(c.Request().Context(), c.Request().Method+" "+route)

			span := trace.SpanFromContext(c.Request().Context())
			if span.IsRecording() {
				span.SetName(c.Request().Method + " " + route)
				span.SetAttributes(semconv.HTTPRoute(route))
			}

			return next(c)
		}
	}
}
