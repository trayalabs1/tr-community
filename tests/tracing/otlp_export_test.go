package tracing_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"google.golang.org/protobuf/proto"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func hasServerSpan(body []byte) (bool, error) {
	var req coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		return false, err
	}

	for _, rs := range req.GetResourceSpans() {
		for _, ss := range rs.GetScopeSpans() {
			for _, span := range ss.GetSpans() {
				if span.GetKind() == tracepb.Span_SPAN_KIND_SERVER {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func TestOTLPExport_withAuthHeaders(t *testing.T) {
	var (
		mu            sync.Mutex
		gotAuth       string
		requestSeen   bool
		sawServerSpan bool
	)

	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			require.NoError(t, err)
			defer gz.Close()
			body = gz
		}

		payload, err := io.ReadAll(body)
		require.NoError(t, err)

		found, err := hasServerSpan(payload)
		require.NoError(t, err)

		mu.Lock()
		gotAuth = r.Header.Get("Authorization")
		requestSeen = true
		if found {
			sawServerSpan = true
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()

	endpoint, err := url.Parse(collector.URL)
	require.NoError(t, err)

	cfg := &config.Config{
		OTELProvider: "otlp",
		OTELEndpoint: *endpoint,
		OTELHeaders:  "Authorization=Basic ZTJlOnRlc3Q=",
	}

	integration.Test(t, cfg, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		ts *httptest.Server,
	) {
		lc.Append(fx.StartHook(func() {
			resp, err := http.Get(ts.URL + "/api/info")
			require.NoError(t, err)
			resp.Body.Close()

			require.Eventually(t, func() bool {
				mu.Lock()
				defer mu.Unlock()
				return requestSeen && sawServerSpan
			}, 5*time.Second, 100*time.Millisecond, "expected the backend to export at least one HTTP server span to the fake OTLP collector")

			mu.Lock()
			defer mu.Unlock()
			require.Equal(t, "Basic ZTJlOnRlc3Q=", gotAuth)
		}))
	}))
}
