package tracing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestOTLPExport_withAuthHeaders(t *testing.T) {
	var (
		mu          sync.Mutex
		gotAuth     string
		requestSeen bool
	)

	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotAuth = r.Header.Get("Authorization")
		requestSeen = true
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
				return requestSeen
			}, 5*time.Second, 100*time.Millisecond, "expected the backend to export at least one span to the fake OTLP collector")

			mu.Lock()
			defer mu.Unlock()
			require.Equal(t, "Basic ZTJlOnRlc3Q=", gotAuth)
		}))
	}))
}
