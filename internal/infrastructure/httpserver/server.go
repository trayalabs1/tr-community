package httpserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/boot_time"
	"github.com/Southclaws/storyden/internal/config"
)

func Instrument(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "storyden",
		otelhttp.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/healthz"
		}),
		otelhttp.WithSpanNameFormatter(routeSpanName),
	)
}

func routeSpanName(operation string, r *http.Request) string {
	labeler, ok := otelhttp.LabelerFromContext(r.Context())
	if !ok {
		return operation
	}

	for _, attr := range labeler.Get() {
		if attr.Key == semconv.HTTPRouteKey {
			return r.Method + " " + attr.Value.AsString()
		}
	}

	return operation
}

func NewServer(lc fx.Lifecycle, logger *slog.Logger, cfg config.Config, router *http.ServeMux) *http.Server {
	server := &http.Server{
		Handler: Instrument(router),
		Addr:    cfg.ListenAddr,
	}

	wctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			server.BaseContext = func(ln net.Listener) context.Context { return wctx }
			go func() {
				// The HTTP server is the root node of dependency tree, because
				// it depends on everything else being initialised first. So, if
				// the app reaches this point, its considered a successful boot!

				logger.Info("storyden http server starting",
					slog.String("boot_time", time.Since(boot_time.StartedAt).String()),
					slog.String("address", cfg.ListenAddr),
					slog.String("api_address", cfg.PublicAPIAddress.String()),
					slog.String("web_address", cfg.PublicWebAddress.String()),
					slog.String("log_level", cfg.LogLevel.String()),
				)

				if err := server.ListenAndServe(); err != nil {
					logger.Error("http server stopped unexpectedly", slog.String("error", err.Error()))
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	return server
}
