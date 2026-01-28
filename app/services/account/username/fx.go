package username

import (
	"go.uber.org/fx"
)

func Build() fx.Option {
	return fx.Module("username",
		fx.Provide(
			New,       // Username service
			NewSeeder, // Seeder service
		),
	)
}
