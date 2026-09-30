package otelroute_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Southclaws/storyden/app/transports/http/middleware/otelroute"
	"github.com/Southclaws/storyden/internal/infrastructure/httpserver"
)

func newInstrumentedStack(t *testing.T) (http.Handler, *tracetest.InMemoryExporter) {
	exporter := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter)))
	otel.SetTextMapPropagator(propagation.TraceContext{})

	router := echo.New()
	router.Use(otelroute.Middleware())
	router.GET("/api/v1/threads/:thread_mark", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("/", router)

	return httpserver.Instrument(mux), exporter
}

func TestExportedSpanNameIsRouteTemplate(t *testing.T) {
	a := assert.New(t)

	handler, exporter := newInstrumentedStack(t)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/threads/01H8XYZABCDEF", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	a.Equal("GET /api/v1/threads/:thread_mark", spans[0].Name)
}

func TestExportedSpanNameFallsBackForUnroutedRequests(t *testing.T) {
	a := assert.New(t)

	handler, exporter := newInstrumentedStack(t)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nothing/here/123", nil))

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	a.Equal("storyden", spans[0].Name)
}
