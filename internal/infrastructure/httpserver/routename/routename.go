package routename

import "context"

type contextKey struct{}

type carrier struct {
	route string
}

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, &carrier{})
}

func Set(ctx context.Context, route string) {
	if c, ok := ctx.Value(contextKey{}).(*carrier); ok {
		c.route = route
	}
}

func Get(ctx context.Context) (string, bool) {
	c, ok := ctx.Value(contextKey{}).(*carrier)
	if !ok || c.route == "" {
		return "", false
	}

	return c.route, true
}
