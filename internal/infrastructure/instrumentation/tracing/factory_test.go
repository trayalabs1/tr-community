package tracing

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/Southclaws/storyden/internal/config"
)

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
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}

	cfg := config.Config{
		OTELProvider: "otlp",
		OTELEndpoint: *endpoint,
		OTELHeaders:  "Authorization=Basic dGVzdDp0ZXN0",
	}

	logger := slog.New(slog.DiscardHandler)

	ctx := context.Background()

	opts, err := newExporter(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("newExporter returned error: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 TracerProviderOption, got %d", len(opts))
	}

	tp := trace.NewTracerProvider(opts...)
	defer tp.Shutdown(ctx)

	tr := tp.Tracer("test")
	_, span := tr.Start(ctx, "test-span")
	span.End()

	if err := tp.ForceFlush(ctx); err != nil {
		t.Fatalf("ForceFlush returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !requestSeen {
		t.Fatal("expected the fake OTLP collector to receive an export request, got none")
	}
	if gotAuthValue != "Basic dGVzdDp0ZXN0" {
		t.Errorf("Authorization header = %q, want %q", gotAuthValue, "Basic dGVzdDp0ZXN0")
	}
}
