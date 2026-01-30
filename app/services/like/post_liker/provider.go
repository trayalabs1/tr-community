package post_liker

import (
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/services/like/post_like_notify"
)

func Build() fx.Option {
	return fx.Options(
		fx.Provide(New),
		post_like_notify.Build(),
	)
}
