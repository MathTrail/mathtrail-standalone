package collector_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry/collector"
)

// A root carries a scheme, a host and at most a path: what else an address can
// hold either has no business there or would land after the signals' paths.
func TestARootIsHeldToTheRule(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		raw     string
		refused bool
	}{
		{name: "a bare host", raw: "https://telemetry.example"},
		{name: "a trailing slash", raw: "https://telemetry.example/"},
		{name: "a port", raw: "http://127.0.0.1:4318"},
		{name: "a prefix", raw: "https://telemetry.example/otlp"},
		{name: "an at sign in its path", raw: "https://telemetry.example/tenants/@team"},
		{name: "no scheme", raw: "telemetry.example", refused: true},
		{name: "a host and a port, no scheme", raw: "localhost:4318", refused: true},
		{name: "a scheme we do not speak", raw: "grpc://telemetry.example", refused: true},
		{name: "no host", raw: "https:///v1", refused: true},
		{name: "empty", raw: "", refused: true},
		{name: "not an address at all", raw: "http://[::1", refused: true},
		{name: "a user", raw: "https://someone@telemetry.example", refused: true},
		{name: "a query", raw: "https://telemetry.example?key=value", refused: true},
		{name: "an empty query", raw: "https://telemetry.example?", refused: true},
		{name: "a fragment", raw: "https://telemetry.example#part", refused: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			root, err := collector.Root(c.raw)
			if (err != nil) != c.refused {
				t.Fatalf("Root(%q) error = %v, want refused: %v", c.raw, err, c.refused)
			}
			if !c.refused && root.String() != c.raw {
				t.Errorf("Root(%q) = %q, want the address as it was written", c.raw, root)
			}
		})
	}
}

// A root that carries a secret is refused without repeating it, whatever shape
// the rest of the address takes and whether or not it can be read at all: the
// refusal goes to the log of a start that failed, where a secret must not.
func TestARefusalNeverRepeatsASecret(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"https://someone:hunter2@telemetry.example",
		"https://someone:hunter 2@telemetry.example",
		"https://someone:hunter%zz@telemetry.example",
		"someone:hunter2@telemetry.example",
		"https:someone:hunter2@telemetry.example",
		"https://telemetry.example?key=hunter2",
		"telemetry.example?key=hunter2",
		"https://telemetry.example#hunter2",
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, err := collector.Root(raw)
			if err == nil {
				t.Fatal("Root() error = nil, want the address refused")
			}
			if strings.Contains(err.Error(), "hunter") {
				t.Errorf("error = %q, want it without the secret", err)
			}
		})
	}
}
