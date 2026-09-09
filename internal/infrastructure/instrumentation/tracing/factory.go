package tracing

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
)

type factory struct {
	provider string
	opts     []trace.TracerProviderOption
}

func Build() fx.Option {
	return fx.Provide(newExporter, newTracerFactory)
}

// NewNoop returns a Factory whose tracers discard every span. Intended for
// use in tests that need a real Factory but have no interest in tracing.
func NewNoop() Factory {
	return factory{}
}

func newTracerFactory(ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	opts []trace.TracerProviderOption,
) (Factory, error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("otel error", slog.String("error", err.Error()))
	}))

	return factory{
		provider: cfg.OTELProvider,
		opts:     opts,
	}, nil
}

func newExporter(ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
) ([]trace.TracerProviderOption, error) {
	switch cfg.OTELProvider {
	case "sentry":

		if cfg.SentryDSN == "" {
			if cfg.OTELEndpoint.String() != "" {
				return nil, fault.New("OTEL_EXPORTER_OTLP_ENDPOINT is set but sentry DSN is required instead when using the sentry provider")
			}
			return nil, fault.New("sentry DSN is required when using the sentry provider")
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			return nil, err
		}

		spanProc := sentryotel.NewSentrySpanProcessor()

		// for some reason, sentry is a "span processor" not a "span exporter".
		return []trace.TracerProviderOption{
			trace.WithSpanProcessor(spanProc),
		}, nil

	case "otlp":
		endpoint := cfg.OTELEndpoint.String()
		if endpoint == "" {
			return nil, fault.New("OTEL_EXPORTER_OTLP_ENDPOINT is required when using the otlp provider")
		}

		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(endpoint),
		}

		if cfg.OTELEndpoint.Scheme != "https" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		if cfg.OTELHeaders != "" {
			headers, err := parseHeaders(cfg.OTELHeaders)
			if err != nil {
				return nil, fault.Wrap(err, fmsg.With("failed to parse OTEL_EXPORTER_OTLP_HEADERS"))
			}

			opts = append(opts, otlptracehttp.WithHeaders(headers))
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

// parseHeaders parses a comma-separated list of `key=value` pairs, matching
// the format of the standard OTEL_EXPORTER_OTLP_HEADERS environment variable.
func parseHeaders(raw string) (map[string]string, error) {
	headers := map[string]string{}

	for pair := range strings.SplitSeq(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fault.Newf("invalid header pair, expected key=value: %q", pair)
		}

		headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	return headers, nil
}

// Build constructs a new tracer for use within a system component.
func (f factory) Build(lc fx.Lifecycle, serviceName string) Tracer {
	return f.BuildProvider(lc, serviceName).Tracer("storyden")
}

// BuildProvider constructs a new TracerProvider for use within a system
// component, for handing to contrib instrumentation libraries that require
// a full provider rather than a Tracer.
func (f factory) BuildProvider(lc fx.Lifecycle, serviceName string) oteltrace.TracerProvider {
	opts := append(f.opts, trace.WithResource(resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)))

	tp := trace.NewTracerProvider(opts...)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := tp.Shutdown(ctx); err != nil {
				return fault.Wrap(err)
			}

			return nil
		},
	})

	return tp
}
