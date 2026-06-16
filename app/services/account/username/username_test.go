package username

import (
	"fmt"
	"strings"
	"testing"
)

func TestHandlePrefixFromName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple lowercase", "rahul", "rahu"},
		{"uppercase truncated", "PRIYANKA", "priy"},
		{"short name", "raj", "raj"},
		{"spaces stripped", "John Doe", "john"},
		{"non-alphanumeric stripped", "a.b_c-d!", "abcd"},
		{"unicode stripped", "Renée", "rene"},
		{"empty", "", ""},
		{"only symbols yields empty", "!!!", ""},
		{"digits kept", "u2 band", "u2ba"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handlePrefixFromName(tt.in); got != tt.want {
				t.Fatalf("handlePrefixFromName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGeneratedHandleIsValid(t *testing.T) {
	names := []string{"rahul", "raj", "PRIYANKA", "John Doe", "a.b_c-d", "u2"}

	for _, name := range names {
		prefix := handlePrefixFromName(name)
		if prefix == "" {
			t.Fatalf("expected usable prefix for %q", name)
		}

		for digits := 0; digits < 10000; digits += 999 {
			candidate := fmt.Sprintf("%s%04d", prefix, digits)
			if err := validateUsername(candidate); err != nil {
				t.Fatalf("generated handle %q (name=%q) failed validation: %v", candidate, name, err)
			}
		}
	}
}

func TestGenerateHandle(t *testing.T) {
	for _, name := range []string{"rahul", "raj", "PRIYANKA", "John Doe", "u2"} {
		for i := 0; i < 100; i++ {
			h := generateHandle(name)
			if err := validateUsername(h); err != nil {
				t.Fatalf("generateHandle(%q) = %q failed validation: %v", name, h, err)
			}
			if !strings.HasPrefix(h, handlePrefixFromName(name)) {
				t.Fatalf("generateHandle(%q) = %q missing name prefix", name, h)
			}
		}
	}
}
