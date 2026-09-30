package tracing

import "testing"

func Test_parseOTLPHeaders(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]string
	}{
		{
			name: "empty string returns empty map",
			raw:  "",
			want: map[string]string{},
		},
		{
			name: "single pair",
			raw:  "Authorization=Basic dXNlcjpwYXNz",
			want: map[string]string{"Authorization": "Basic dXNlcjpwYXNz"},
		},
		{
			name: "multiple pairs",
			raw:  "Authorization=Basic dXNlcjpwYXNz,x-last9-tenant=community",
			want: map[string]string{
				"Authorization":  "Basic dXNlcjpwYXNz",
				"x-last9-tenant": "community",
			},
		},
		{
			name: "trims surrounding whitespace around pairs",
			raw:  " Authorization = Basic dXNlcjpwYXNz , x-last9-tenant = community ",
			want: map[string]string{
				"Authorization":  "Basic dXNlcjpwYXNz",
				"x-last9-tenant": "community",
			},
		},
		{
			name: "skips malformed pair without '='",
			raw:  "Authorization=Basic dXNlcjpwYXNz,malformed,x-last9-tenant=community",
			want: map[string]string{
				"Authorization":  "Basic dXNlcjpwYXNz",
				"x-last9-tenant": "community",
			},
		},
		{
			name: "value containing '=' is preserved (base64 padding)",
			raw:  "Authorization=Basic dXNlcjpwYXNzMTIz==",
			want: map[string]string{"Authorization": "Basic dXNlcjpwYXNzMTIz=="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOTLPHeaders(tt.raw)

			if len(got) != len(tt.want) {
				t.Fatalf("got %d headers, want %d: got=%v want=%v", len(got), len(tt.want), got, tt.want)
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("header %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
