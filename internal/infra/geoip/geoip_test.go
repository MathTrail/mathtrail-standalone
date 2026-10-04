package geoip_test

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/country"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip/geoiptest"
)

func open(t *testing.T) geoip.Database {
	t.Helper()

	database, err := geoip.Open(geoiptest.Write(t))
	if err != nil {
		t.Fatalf("Open() error = %v, want none", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v, want none", err)
		}
	})
	return database
}

// A file that is not there, or is no database, is refused when it is opened,
// rather than at the first parent who signs in.
func TestAFileThatIsNoDatabaseIsRefused(t *testing.T) {
	t.Parallel()

	notADatabase := filepath.Join(t.TempDir(), "countries.mmdb")
	if err := os.WriteFile(notADatabase, []byte("country,code\nNZ,203.0.113.0/24\n"), 0o600); err != nil {
		t.Fatalf("write the file: %v", err)
	}
	empty := filepath.Join(t.TempDir(), "empty.mmdb")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatalf("write the file: %v", err)
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{"no file", filepath.Join(t.TempDir(), "missing.mmdb")},
		{"a file of another format", notADatabase},
		{"an empty file", empty},
		{"a directory", t.TempDir()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			database, err := geoip.Open(tc.path)
			if !errors.Is(err, geoip.ErrDatabase) {
				t.Fatalf("Open() error = %v, want ErrDatabase", err)
			}
			if database != nil {
				t.Errorf("Open() = %v beside its error, want nothing", database)
			}
		})
	}
}

// An address comes out as the country its network is in, and as nothing when
// no country holds it, when the database has no record of it, or when the
// record names a code no country has.
func TestAnAddressIsLookedUp(t *testing.T) {
	t.Parallel()

	database := open(t)
	for _, tc := range []struct {
		name string
		addr string
		want string
	}{
		{"an address in a country", geoiptest.Family, geoiptest.Country},
		{"an IPv6 address in a country", geoiptest.FamilyV6, geoiptest.CountryV6},
		{"an IPv4 address written as IPv6", "::ffff:" + geoiptest.Family, geoiptest.Country},
		{"an address of a private network the database puts in a country", geoiptest.Private, ""},
		{"an address the database puts in a code no country has", geoiptest.Unlisted, ""},
		{"an address the database has no record of", geoiptest.Nowhere, ""},
		{"this machine", "127.0.0.1", ""},
		{"this machine over IPv6", "::1", ""},
		{"a local link", "fe80::1", ""},
		{"an IPv6 address of a private network", "fd00::1", ""},
		{"no address", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var addr netip.Addr
			if tc.addr != "" {
				addr = netip.MustParseAddr(tc.addr)
			}
			if got := database.Country(addr); got != tc.want {
				t.Errorf("Country(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}

func TestADatabaseSaysWhatItIs(t *testing.T) {
	t.Parallel()

	about := open(t).About()
	if about.Type != geoiptest.Type || !about.Built.Equal(geoiptest.Built) {
		t.Errorf("About() = %+v, want type %q built %v", about, geoiptest.Type, geoiptest.Built)
	}
}

func TestLookupsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	database := open(t)
	properties := gopter.NewProperties(nil)

	properties.Property("any address comes out as a country of the list or as nothing", prop.ForAll(
		func(bytes []byte) bool {
			addr, _ := netip.AddrFromSlice(bytes)
			code := database.Country(addr)
			return code == "" || country.Known(code)
		},
		gen.OneGenOf(gen.SliceOfN(4, gen.UInt8()), gen.SliceOfN(16, gen.UInt8())),
	))

	properties.TestingRun(t)
}

// A lookup may still be running when the service stops and lets go of the
// database: the server gives up waiting for a request it cannot drain, and
// leaves it running. Close waits for the lookups under way, and a lookup after
// it finds no country, rather than reading a file no longer in memory.
func TestALookupAsTheDatabaseClosesFindsACountryOrNothing(t *testing.T) {
	t.Parallel()

	database, err := geoip.Open(geoiptest.Write(t))
	if err != nil {
		t.Fatalf("Open() error = %v, want none", err)
	}
	family := netip.MustParseAddr(geoiptest.Family)

	// The lookups go on until Close has returned, so that Close always runs
	// while lookups are under way.
	const lookers = 8
	var started, done sync.WaitGroup
	started.Add(lookers)
	done.Add(lookers)
	closed := make(chan struct{})
	for range lookers {
		go func() {
			defer done.Done()
			lookUpUntil(t, database, family, started.Done, closed)
		}()
	}
	started.Wait()
	if err := database.Close(); err != nil {
		t.Errorf("Close() error = %v, want none", err)
	}
	close(closed)
	done.Wait()

	if got := database.Country(family); got != "" {
		t.Errorf("Country() after Close = %q, want nothing", got)
	}
}

// lookUpUntil looks the family's address up again and again until stop is
// closed, says it has started once its first lookup is done, and holds every
// lookup to the country of the address or to nothing.
func lookUpUntil(t *testing.T, database geoip.Database, family netip.Addr, started func(), stop <-chan struct{}) {
	t.Helper()

	for first := true; ; first = false {
		if got := database.Country(family); got != geoiptest.Country && got != "" {
			t.Errorf("Country() = %q while the database closes, want %q or nothing", got, geoiptest.Country)
		}
		if first {
			started()
		}
		select {
		case <-stop:
			return
		default:
		}
	}
}
