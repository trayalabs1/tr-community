package otelhttp

import (
	"net/http"

	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type Middleware struct{}

func New() *Middleware {
	return &Middleware{}
}

// WithTracing wraps every request in an HTTP server span. Uses the process's
// global TracerProvider unless one is passed explicitly (tests pass their own
// to assert against, production relies on tracing.factory.Build having called
// otel.SetTracerProvider).
func (m *Middleware) WithTracing(tp trace.TracerProvider) func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("http.server",
		otelhttp.WithTracerProvider(tp),
	)
}
