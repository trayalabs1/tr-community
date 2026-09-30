package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
)

// InstrumentedTransport wraps base (or http.DefaultTransport if nil) so that
// outbound requests made through it produce client spans, named after
// serviceName.
func InstrumentedTransport(lc fx.Lifecycle, tf Factory, serviceName string, base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return otelhttp.NewTransport(base,
		otelhttp.WithTracerProvider(tf.BuildProvider(lc, serviceName)),
	)
}
