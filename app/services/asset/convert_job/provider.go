package convert_job

import (
	"go.uber.org/fx"
)

func Build() fx.Option {
	return fx.Options(
		fx.Provide(New),
		fx.Invoke(runConvertConsumer),
	)
}
