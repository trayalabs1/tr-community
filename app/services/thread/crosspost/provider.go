package crosspost

import "go.uber.org/fx"

func Build() fx.Option {
	return fx.Options(
		fx.Provide(newCrossposter),
		fx.Invoke(runCrossposter),
	)
}
