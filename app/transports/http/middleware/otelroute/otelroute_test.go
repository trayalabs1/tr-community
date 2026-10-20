package otelroute

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestMiddlewareSetsRouteTemplate(t *testing.T) {
	a := assert.New(t)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	router := echo.New()
	router.Use(Middleware())
	router.GET("/v1/threads/:thread_mark", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/threads/01H8XYZABCDEF", nil)
	ctx, span := tracer.Start(req.Context(), "GET /")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	span.End()

	require.Equal(t, http.StatusOK, rec.Code)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	a.Equal("GET /v1/threads/:thread_mark", spans[0].Name)

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	a.Equal("/v1/threads/:thread_mark", attrs["http.route"])
	a.NotContains(spans[0].Name, "01H8XYZABCDEF")
}
