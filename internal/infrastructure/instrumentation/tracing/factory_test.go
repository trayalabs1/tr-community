package tracing

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Southclaws/storyden/internal/config"
)

func TestNewExporterRejectsSentry(t *testing.T) {
	a := assert.New(t)

	_, err := newExporter(context.Background(), config.Config{OTELProvider: "sentry"}, slog.Default())

	a.Error(err)
	a.Contains(err.Error(), "sentry OTEL_PROVIDER has been removed")
}

func TestNewExporterOTLPRequiresEndpoint(t *testing.T) {
	a := assert.New(t)

	_, err := newExporter(context.Background(), config.Config{OTELProvider: "otlp"}, slog.Default())

	a.Error(err)
	a.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT is required")
}

func TestNewExporterUnsetIsNoop(t *testing.T) {
	a := assert.New(t)

	opts, err := newExporter(context.Background(), config.Config{}, slog.Default())

	require.NoError(t, err)
	a.Empty(opts)
}
