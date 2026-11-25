package sentiment_job

import (
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/services/sentiment/scorer"
)

func Build() fx.Option {
	return fx.Options(
		fx.Provide(scorer.New),
		fx.Provide(newSentimentConsumer),
		fx.Invoke(runSentimentConsumer),
	)
}
