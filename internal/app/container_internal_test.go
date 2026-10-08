package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip/geoiptest"
)

// Close runs every closer, the last registered first, so that a resource is
// closed before whatever it was built from. One that fails is logged as a
// failure, and the ones after it still run rather than being left open.
func TestCloseRunsEveryCloserLastFirstAndLogsTheOneThatFails(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.InfoLevel)
	stuck := errors.New("the exporter would not flush")
	var ran []string
	closing := func(name string, err error) func(context.Context) error {
		return func(context.Context) error {
			ran = append(ran, name)
			return err
		}
	}
	c := &Container{Logger: zap.New(core), closers: []func(context.Context) error{
		closing("telemetry", nil), closing("database", stuck), closing("store", nil),
	}}

	c.Close(t.Context())
	if want := []string{"store", "database", "telemetry"}; !slices.Equal(ran, want) {
		t.Errorf("closed %q, want %q: every closer, the last registered first", ran, want)
	}
	failed := logs.FilterMessage("close failed").All()
	if len(failed) != 1 || failed[0].Level != zapcore.ErrorLevel || failed[0].ContextMap()["error"] != stuck.Error() {
		t.Errorf("close failed lines = %v, want one error naming %q", failed, stuck)
	}
}

// The country a request came from is the one the database places the address
// it was sent from in. A request whose sender cannot be read as an address at
// all is from no country, rather than from whatever an unreadable address
// would be looked up as. The database is closed with the container.
func TestARequestWhoseSenderCannotBeReadIsFromNoCountry(t *testing.T) {
	t.Parallel()

	c := &Container{Logger: zap.NewNop()}
	countryOf, err := c.countries(&config.Config{CountryDB: geoiptest.Write(t)}, zap.NewNop())
	if err != nil {
		t.Fatalf("countries() error = %v, want the database open", err)
	}
	t.Cleanup(func() { c.Close(context.Background()) })
	if len(c.closers) != 1 {
		t.Errorf("closers = %d, want the database's alone", len(c.closers))
	}

	for _, tc := range []struct {
		name, from, want string
	}{
		{name: "an address the database places", from: geoiptest.Family + ":4711", want: geoiptest.Country},
		{name: "no address at all", from: "a-socket-of-its-own", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
			req.RemoteAddr = tc.from
			if got := countryOf(req); got != tc.want {
				t.Errorf("the country of a request sent from %q = %q, want %q", tc.from, got, tc.want)
			}
		})
	}
}
