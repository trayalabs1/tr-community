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
)

func TestPanicRecoveryKeepsStackTraceOnSpanAndLog(t *testing.T) {
	a := assert.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	handler := New(logger).WithLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/explode?x=1", nil)
	ctx, span := tp.Tracer("test").Start(req.Context(), "root")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	span.End()

	a.Equal(http.StatusInternalServerError, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	a.Equal(codes.Error, spans[0].Status.Code)

	var stack string
	for _, ev := range spans[0].Events {
		if ev.Name != "exception" {
			continue
		}
		for _, kv := range ev.Attributes {
			if string(kv.Key) == "trace" {
				stack = kv.Value.AsString()
			}
		}
	}
	a.Contains(stack, "requestlog_test.go")

	a.Contains(logs.String(), "GET /explode: boom")
	a.Contains(logs.String(), "trace=")
	a.Contains(logs.String(), "trace_id="+span.SpanContext().TraceID().String())
}

func TestRequestAnnotatesAmbientSpanWithoutCreatingOne(t *testing.T) {
	a := assert.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	handler := New(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))).WithLogger()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/tea?flavour=earl", nil)
	ctx, span := tp.Tracer("test").Start(req.Context(), "root")

	handler.ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	a.Equal("GET", attrs["http.request.method"])
	a.Equal("flavour=earl", attrs["url.query"])
	a.Equal("418", attrs["http.response.status_code"])
}
