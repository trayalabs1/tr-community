package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"
)

func Build() fx.Option {
	return fx.Provide(func() (c Config, err error) {
		if err = envconfig.Process("", &c); err != nil {
			return c, fault.Wrap(err, fmsg.With("failed to parse configuration from environment variables"))
		}

		t := reflect.TypeOf(c)
		v := reflect.ValueOf(c)
		for i := range t.NumField() {
			field := t.Field(i)
			if key := field.Tag.Get("envconfig"); key != "" {
				fmt.Fprintf(os.Stderr, "[config] %s=%v\n", key, v.Field(i).Interface())
			}
		}

		return
	})
}
