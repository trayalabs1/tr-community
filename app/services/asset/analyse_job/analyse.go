package analyse_job

import (
	"context"
	"net/http"
	"time"

	"github.com/Southclaws/opt"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/library/node_writer"
	"github.com/Southclaws/storyden/app/services/asset/analyse"
	"github.com/Southclaws/storyden/app/services/asset/asset_upload"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/tracing"
)

type analyseConsumer struct {
	analyser   *analyse.Analyser
	uploader   *asset_upload.Uploader
	nodeWriter *node_writer.Writer
	httpClient *http.Client
}

func newAnalyseConsumer(
	lc fx.Lifecycle,
	tf tracing.Factory,
	analyser *analyse.Analyser,
	uploader *asset_upload.Uploader,
	nodeWriter *node_writer.Writer,
) *analyseConsumer {
	return &analyseConsumer{
		analyser:   analyser,
		uploader:   uploader,
		nodeWriter: nodeWriter,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tracing.InstrumentedTransport(lc, tf, "asset-download", nil),
		},
	}
}

func (i *analyseConsumer) analyseAsset(ctx context.Context, id asset.AssetID, fillrule opt.Optional[asset.ContentFillCommand]) error {
	return i.analyser.Analyse(ctx, id, fillrule)
}
