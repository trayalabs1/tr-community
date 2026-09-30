package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
)

func Test_parseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "single header",
			raw:  "Authorization=Basic dG9rZW4=",
			want: map[string]string{"Authorization": "Basic dG9rZW4="},
		},
		{
			name: "multiple headers",
			raw:  "Authorization=Basic abc,X-Custom-Header=value",
			want: map[string]string{
				"Authorization":   "Basic abc",
				"X-Custom-Header": "value",
			},
		},
		{
			name: "trims whitespace around pairs",
			raw:  " Authorization=Basic abc , X-Custom-Header=value ",
			want: map[string]string{
				"Authorization":   "Basic abc",
				"X-Custom-Header": "value",
			},
		},
		{
			name: "empty string",
			raw:  "",
			want: map[string]string{},
		},
		{
			name: "trailing comma is ignored",
			raw:  "Authorization=Basic abc,",
			want: map[string]string{"Authorization": "Basic abc"},
		},
		{
			name:    "missing equals sign",
			raw:     "Authorization",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaders(tt.raw)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_factory_BuildProvider(t *testing.T) {
	f := factory{}
	lc := fxtest.NewLifecycle(t)

	provider := f.BuildProvider(lc, "test-component")
	require.NotNil(t, provider)

	tracer := provider.Tracer("test-tracer")
	require.NotNil(t, tracer)

	lc.RequireStart().RequireStop()
}

func Test_factory_Build(t *testing.T) {
	f := factory{}
	lc := fxtest.NewLifecycle(t)

	tracer := f.Build(lc, "test-component")
	require.NotNil(t, tracer)

	lc.RequireStart().RequireStop()
}
