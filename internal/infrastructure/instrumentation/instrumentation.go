package instrumentation

import (
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/otelinit"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/spanner"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/tracing"
)

func Build() fx.Option {
	return fx.Options(
		tracing.Build(),
		otelinit.Build(),
		fx.Provide(spanner.New),
	)
}
