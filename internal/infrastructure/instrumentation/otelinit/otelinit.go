package otelinit

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
)

type Providers struct {
	Tracer *sdktrace.TracerProvider
	Meter  *sdkmetric.MeterProvider
}

func Build() fx.Option {
	return fx.Options(
		fx.Provide(newResource),
		fx.Provide(New),
	)
}

func newResource(cfg config.Config) *resource.Resource {
	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		version = info.Main.Version
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(version),
		semconv.HostName(hostname),
		attribute.String("deployment.environment.name", cfg.DeploymentEnvironment),
	)
}

func New(
	ctx context.Context,
	lc fx.Lifecycle,
	cfg config.Config,
	logger *slog.Logger,
	res *resource.Resource,
	spanOpts []sdktrace.TracerProviderOption,
) (*Providers, error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("otel error", slog.String("error", err.Error()))
	}))

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	opts := make([]sdktrace.TracerProviderOption, 0, len(spanOpts)+2)
	opts = append(opts, spanOpts...)
	opts = append(opts,
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.OTELTracesSamplerArg),
		)),
	)

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	providers := &Providers{Tracer: tp}

	if cfg.OTELProvider == "otlp" {
		mp, err := newMeterProvider(ctx, cfg, res)
		if err != nil {
			return nil, fault.Wrap(err)
		}

		otel.SetMeterProvider(mp)
		providers.Meter = mp

		if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to start runtime metrics"))
		}
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if providers.Meter != nil {
				if err := providers.Meter.Shutdown(ctx); err != nil {
					logger.Error("failed to shut down meter provider", slog.String("error", err.Error()))
				}
			}

			if err := providers.Tracer.Shutdown(ctx); err != nil {
				return fault.Wrap(err)
			}

			return nil
		},
	})

	return providers, nil
}

func newMeterProvider(ctx context.Context, cfg config.Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	endpoint := cfg.OTELEndpoint.String()
	if endpoint == "" {
		return nil, fault.New("OTEL_EXPORTER_OTLP_ENDPOINT is required when using the otlp provider")
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(endpoint),
	}

	if cfg.OTELEndpoint.Scheme != "https" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to create OTLP metric exporter"))
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(cfg.OTELMetricExportInterval),
		)),
	), nil
}
