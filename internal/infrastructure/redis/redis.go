package redis

import (
	"fmt"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidisotel"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/tracing"
)

func Build() fx.Option {
	return fx.Options(
		fx.Provide(newRedis),
	)
}

func newRedis(lc fx.Lifecycle, tf tracing.Factory, cfg config.Config) (rueidis.Client, error) {
	if cfg.RedisURL.String() == "" {
		return nil, nil
	}

	password, _ := cfg.RedisURL.User.Password()

	client, err := rueidisotel.NewClient(rueidis.ClientOption{
		InitAddress:      []string{cfg.RedisURL.Host},
		Username:         cfg.RedisURL.User.Username(),
		Password:         password,
		DisableCache:     true,
		ConnWriteTimeout: 5 * time.Second,
	}, rueidisotel.WithTracerProvider(tf.BuildProvider(lc, "redis")))
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With(fmt.Sprintf("failed to connect to redis at %s", cfg.RedisURL.Host)))
	}

	return client, nil
}
