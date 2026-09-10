package tracing

import (
	"context"
	"log/slog"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/otelinit"
)

type factory struct{}

func Build() fx.Option {
	return fx.Provide(newExporter, newTracerFactory)
}

func newExporter(ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
) ([]trace.TracerProviderOption, error) {
	switch cfg.OTELProvider {
	case "sentry":
		return nil, fault.New("the sentry OTEL_PROVIDER has been removed, use 'otlp' with OTEL_EXPORTER_OTLP_ENDPOINT pointed at the collector")

	case "otlp":
		endpoint := cfg.OTELEndpoint.String()
		if endpoint == "" {
			return nil, fault.New("OTEL_EXPORTER_OTLP_ENDPOINT is required when using the otlp provider")
		}

		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(endpoint),
		}

		if headers := parseOTLPHeaders(cfg.OTELHeaders); len(headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}

		if cfg.OTELEndpoint.Scheme != "https" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		otlp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to create OTLP exporter"))
		}

		return []trace.TracerProviderOption{
			trace.WithBatcher(otlp),
		}, nil

	case "logger":
		return []trace.TracerProviderOption{
			trace.WithSyncer(newLoggingTracer(logger)),
		}, nil

	default:
		return []trace.TracerProviderOption{}, nil
	}
}

func newTracerFactory(_ *otelinit.Providers) (Factory, error) {
	return factory{}, nil
}

func (f factory) Build(lc fx.Lifecycle, scopeName string) Tracer {
	return otel.Tracer(scopeName)
}
