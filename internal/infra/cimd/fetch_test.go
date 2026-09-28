package cimd_test

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
)

// The fetcher the service builds reaches no address of the machine it runs
// on, nor the platform's metadata server: a document served on the loopback is
// refused however its address is written, and nothing reaches the server.
func TestTheServicesFetcherReachesNothingOfItsOwn(t *testing.T) {
	t.Parallel()

	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.StartTLS()
	t.Cleanup(server.Close)
	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("the server's address: %v", err)
	}

	fetcher := cimd.NewFetcher()
	for _, address := range []string{
		server.URL + "/oauth/metadata",
		"https://localhost:" + port + "/oauth/metadata",
		"https://[::ffff:127.0.0.1]:" + port + "/oauth/metadata",
		"https://169.254.169.254/computeMetadata/v1/instance",
		"https://[fd20:ce::254]/computeMetadata/v1/instance",
	} {
		if _, err := fetcher.Fetch(t.Context(), address); !errors.Is(err, cimd.ErrAddress) {
			t.Errorf("Fetch(%q) error = %v, want ErrAddress", address, err)
		}
	}
	if got := connections.Load(); got != 0 {
		t.Errorf("%d connections reached the server, want none", got)
	}

	plain := strings.Replace(server.URL, "https://", "http://", 1) + "/oauth/metadata"
	if _, err := fetcher.Fetch(t.Context(), plain); !errors.Is(err, cimd.ErrClientURL) {
		t.Errorf("Fetch(%q) error = %v, want ErrClientURL", plain, err)
	}
}
