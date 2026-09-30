package reqlog

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	otel_trace "go.opentelemetry.io/otel/trace"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}

func TestWithLoggerSetsSpanAttributesOnNormalRequest(t *testing.T) {
	a := assert.New(t)
	require := require.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	m := New(testLogger())

	handler := m.WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/things?foo=bar&baz=qux", nil)
	ctx, span := tracer.Start(req.Context(), "GET /v1/things")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	span.End()

	require.Equal(http.StatusTeapot, rec.Code)

	spans := exporter.GetSpans()
	require.Len(spans, 1)

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}

	a.Equal(http.MethodGet, attrs["http.request.method"])
	a.Equal("baz=qux&foo=bar", attrs["url.query"])
	a.Equal("418", attrs["http.response.status_code"])
	a.NotEmpty(attrs["duration"])
	a.Equal(codes.Unset, spans[0].Status.Code)
}

func TestWithLoggerRecordsQueryStringVerbatim(t *testing.T) {
	a := assert.New(t)
	require := require.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	m := New(testLogger())

	handler := m.WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/session?token=super-secret-token-value&user=alice", nil)
	ctx, span := tracer.Start(req.Context(), "GET /v1/session")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	span.End()

	require.Equal(http.StatusOK, rec.Code)

	spans := exporter.GetSpans()
	require.Len(spans, 1)

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}

	require.Contains(attrs, "url.query")
	a.Equal(req.URL.Query().Encode(), attrs["url.query"])
	a.Contains(attrs["url.query"], "token=super-secret-token-value")
	a.NotContains(attrs["url.query"], "REDACTED")
	a.NotContains(attrs["url.query"], "***")
}

func TestWithLoggerRecoversFromPanicAndSetsSpanError(t *testing.T) {
	a := assert.New(t)
	require := require.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	m := New(testLogger())

	handler := m.WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/explode", nil)
	ctx, span := tracer.Start(req.Context(), "GET /v1/explode")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	require.NotPanics(func() {
		handler.ServeHTTP(rec, req)
	})
	span.End()

	a.Equal(http.StatusInternalServerError, rec.Code)

	spans := exporter.GetSpans()
	require.Len(spans, 1)

	a.Equal(codes.Error, spans[0].Status.Code)
	a.Contains(spans[0].Status.Description, "boom")

	require.Len(spans[0].Events, 1)
	a.Equal("exception", spans[0].Events[0].Name)

	eventAttrs := map[string]string{}
	for _, kv := range spans[0].Events[0].Attributes {
		eventAttrs[string(kv.Key)] = kv.Value.Emit()
	}
	a.Equal("boom", eventAttrs["exception.message"])
}

func TestWithLoggerRecoversFromErrorPanic(t *testing.T) {
	a := assert.New(t)
	require := require.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	m := New(testLogger())

	handler := m.WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nilSlice []int
		_ = nilSlice[5]
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/index-out-of-range", nil)
	ctx, span := tracer.Start(req.Context(), "GET /v1/index-out-of-range")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	require.NotPanics(func() {
		handler.ServeHTTP(rec, req)
	})
	span.End()

	a.Equal(http.StatusInternalServerError, rec.Code)

	spans := exporter.GetSpans()
	require.Len(spans, 1)
	a.Equal(codes.Error, spans[0].Status.Code)
}

func TestWithLoggerUsesAmbientSpanFromContext(t *testing.T) {
	require := require.New(t)

	m := New(testLogger())

	handler := m.WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := otel_trace.SpanFromContext(r.Context())
		require.True(span.SpanContext().IsValid())
		w.WriteHeader(http.StatusNoContent)
	}))

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	req := httptest.NewRequest(http.MethodGet, "/v1/no-content", nil)
	ctx, span := tracer.Start(req.Context(), "GET /v1/no-content")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	span.End()

	require.Equal(http.StatusNoContent, rec.Code)
}
