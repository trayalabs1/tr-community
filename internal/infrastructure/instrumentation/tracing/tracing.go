package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

type Factory interface {
	Build(lc fx.Lifecycle, serviceName string) Tracer

	// BuildProvider returns the full trace.TracerProvider for a component,
	// for use with contrib instrumentation libraries (otelhttp, otelecho,
	// otelsql, rueidisotel, etc) that require a provider rather than a Tracer.
	BuildProvider(lc fx.Lifecycle, serviceName string) trace.TracerProvider
}
