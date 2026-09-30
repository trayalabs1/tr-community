package tracing

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/Southclaws/storyden/internal/config"
)

func TestNewExporterRejectsSentry(t *testing.T) {
	a := assert.New(t)

	_, err := newExporter(context.Background(), config.Config{OTELProvider: "sentry"}, slog.Default())

	a.Error(err)
	a.Contains(err.Error(), "sentry OTEL_PROVIDER has been removed")
}

func TestNewExporterOTLPRequiresEndpoint(t *testing.T) {
	a := assert.New(t)

	_, err := newExporter(context.Background(), config.Config{OTELProvider: "otlp"}, slog.Default())

	a.Error(err)
	a.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT is required")
}

func TestNewExporterUnsetIsNoop(t *testing.T) {
	a := assert.New(t)

	opts, err := newExporter(context.Background(), config.Config{}, slog.Default())

	require.NoError(t, err)
	a.Empty(opts)
}

func Test_newExporter_otlp_sendsConfiguredHeaders(t *testing.T) {
	var (
		mu           sync.Mutex
		gotAuthValue string
		requestSeen  bool
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotAuthValue = r.Header.Get("Authorization")
		requestSeen = true
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	endpoint, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cfg := config.Config{
		OTELProvider: "otlp",
		OTELEndpoint: *endpoint,
		OTELHeaders:  "Authorization=Basic dGVzdDp0ZXN0",
	}

	logger := slog.New(slog.DiscardHandler)

	ctx := context.Background()

	opts, err := newExporter(ctx, cfg, logger)
	require.NoError(t, err)
	require.Len(t, opts, 1)

	tp := trace.NewTracerProvider(opts...)
	defer tp.Shutdown(ctx)

	tr := tp.Tracer("test")
	_, span := tr.Start(ctx, "test-span")
	span.End()

	require.NoError(t, tp.ForceFlush(ctx))

	mu.Lock()
	defer mu.Unlock()

	require.True(t, requestSeen)
	assert.Equal(t, "Basic dGVzdDp0ZXN0", gotAuthValue)
}
