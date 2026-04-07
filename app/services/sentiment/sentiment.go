package sentiment

import (
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/services/sentiment/ranker"
	"github.com/Southclaws/storyden/app/services/sentiment/rehydrator"
	"github.com/Southclaws/storyden/app/services/sentiment/sentiment_job"
)

func Build() fx.Option {
	return fx.Options(
		fx.Provide(ranker.New),
		sentiment_job.Build(),
		rehydrator.Build(),
	)
}
