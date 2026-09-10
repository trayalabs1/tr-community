package otelinit

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx/fxtest"

	"github.com/Southclaws/storyden/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		ServiceName:              "storyden-test",
		DeploymentEnvironment:    "test",
		OTELTracesSamplerArg:     1.0,
		OTELMetricExportInterval: 0,
	}
}

func TestNewRegistersGlobalTracerProvider(t *testing.T) {
	a := assert.New(t)
	lc := fxtest.NewLifecycle(t)

	exporter := tracetest.NewInMemoryExporter()

	p, err := New(
		context.Background(),
		lc,
		testConfig(),
		slog.Default(),
		newResource(testConfig()),
		[]sdktrace.TracerProviderOption{sdktrace.WithSyncer(exporter)},
	)
	require.NoError(t, err)
	require.NotNil(t, p.Tracer)

	_, span := otel.Tracer("scope-under-test").Start(context.Background(), "unit")
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	a.Equal("unit", spans[0].Name)
	a.Equal("scope-under-test", spans[0].InstrumentationScope.Name)

	attrs := map[string]string{}
	for _, kv := range spans[0].Resource.Attributes() {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	a.Equal("storyden-test", attrs["service.name"])
	a.Equal("test", attrs["deployment.environment.name"])
}

func TestNewRegistersPropagator(t *testing.T) {
	a := assert.New(t)
	lc := fxtest.NewLifecycle(t)

	_, err := New(
		context.Background(),
		lc,
		testConfig(),
		slog.Default(),
		newResource(testConfig()),
		[]sdktrace.TracerProviderOption{},
	)
	require.NoError(t, err)

	ctx, span := otel.Tracer("propagation").Start(context.Background(), "outbound")
	defer span.End()

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/", nil)
	require.NoError(t, err)

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	a.NotEmpty(req.Header.Get("traceparent"))

	extracted := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(req.Header))

	a.Equal(
		trace.SpanContextFromContext(ctx).TraceID().String(),
		trace.SpanContextFromContext(extracted).TraceID().String(),
	)
}

func TestNewSkipsMeterProviderWhenNotOTLP(t *testing.T) {
	a := assert.New(t)
	lc := fxtest.NewLifecycle(t)

	p, err := New(
		context.Background(),
		lc,
		testConfig(),
		slog.Default(),
		newResource(testConfig()),
		[]sdktrace.TracerProviderOption{},
	)
	require.NoError(t, err)

	a.Nil(p.Meter)
}
