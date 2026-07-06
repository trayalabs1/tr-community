package asset

import (
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/services/asset/analyse"
	"github.com/Southclaws/storyden/app/services/asset/analyse_job"
	"github.com/Southclaws/storyden/app/services/asset/asset_download"
	"github.com/Southclaws/storyden/app/services/asset/asset_upload"
	"github.com/Southclaws/storyden/app/services/asset/convert_job"
)

func Build() fx.Option {
	return fx.Options(
		analyse_job.Build(),
		convert_job.Build(),
		fx.Provide(
			analyse.New,
			asset_upload.New,
			asset_download.New,
		),
	)
}
