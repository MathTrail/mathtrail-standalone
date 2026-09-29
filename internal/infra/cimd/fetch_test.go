package cimd_test

import (
	"errors"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
)

// The fetcher the service builds reaches no address of the machine it runs
// on, nor the platform's metadata server: an address inside is refused however
// it is written, before anything is dialled — a dial there would fail as
// unreachable instead, since nothing on this machine listens on https's port.
func TestTheServicesFetcherReachesNothingOfItsOwn(t *testing.T) {
	t.Parallel()

	fetcher := cimd.NewFetcher()
	for _, address := range []string{
		"https://127.0.0.1/oauth/metadata",
		"https://localhost/oauth/metadata",
		"https://[::ffff:127.0.0.1]/oauth/metadata",
		"https://169.254.169.254/computeMetadata/v1/instance",
		"https://[fd20:ce::254]/computeMetadata/v1/instance",
	} {
		if _, err := fetcher.Fetch(t.Context(), address); !errors.Is(err, cimd.ErrAddress) {
			t.Errorf("Fetch(%q) error = %v, want ErrAddress", address, err)
		}
	}

	for _, address := range []string{
		"http://127.0.0.1/oauth/metadata",
		"https://127.0.0.1:8443/oauth/metadata",
	} {
		if _, err := fetcher.Fetch(t.Context(), address); !errors.Is(err, cimd.ErrClientURL) {
			t.Errorf("Fetch(%q) error = %v, want ErrClientURL", address, err)
		}
	}
}
