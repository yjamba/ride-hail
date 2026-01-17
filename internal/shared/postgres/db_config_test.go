package postgres

import "testing"

func TestNormalizeSSLMode(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "disabled", in: "disabled", want: "disable"},
		{name: "enabled", in: "enabled", want: "require"},
		{name: "true", in: "true", want: "require"},
		{name: "false", in: "false", want: "disable"},
		{name: "prefer", in: "prefer", want: "prefer"},
		{name: "verify-ca", in: "verify-ca", want: "verify-ca"},
		{name: "trimmed", in: "  DISABLE  ", want: "disable"},
		{name: "unknown", in: "nope", want: "disable"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeSSLMode(tc.in); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestNewDBConfig_NormalizesSSLMode(t *testing.T) {
	cfg := NewDBConfig("localhost", "5432", "user", "pass", "db", "disabled")
	if cfg.SSLMode != "disable" {
		t.Fatalf("expected sslmode disable, got %q", cfg.SSLMode)
	}
}
