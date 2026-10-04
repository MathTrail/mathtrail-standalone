// Package geoiptest writes a small database of countries, of the format a
// deployment reads, for the tests of the packages that look a country up. Each
// network in it stands for one case a lookup has to get right, and every
// network is one the address ranges reserved for documentation and private use
// hold, so that no real household's address is ever written into a test.
//
// It is test code that lives in a package rather than a test file, because
// the tests of more than one package open a database, and a test file cannot
// be shared between packages.
package geoiptest

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

// Type is the kind of database this package writes, as the database says.
const Type = "MathTrail-Test-Country"

// Built is when the database says it was built: a fixed moment, so that two
// runs write the same file.
var Built = time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)

// The addresses a test looks up, and what each has to come out as.
const (
	// Family is where a parent's browser signs in from: an address of a
	// documentation range the database puts in Country.
	Family = "203.0.113.77"
	// Country is the country of Family's network.
	Country = "NZ"
	// FamilyV6 is an address of the documentation range of IPv6, in
	// CountryV6.
	FamilyV6 = "2001:db8:1::7"
	// CountryV6 is the country of FamilyV6's network.
	CountryV6 = "JP"
	// Private is an address of a private network, which the database puts in
	// a country anyway: a lookup that came out as that country would have
	// looked up an address no country holds.
	Private = "10.1.2.3"
	// Unlisted is an address whose network the database puts in a code no
	// country has.
	Unlisted = "198.51.100.9"
	// City is the name of a city the database keeps beside the country of
	// Family's network: whatever reads the database must never let it out.
	City = "Wellington"
	// Nowhere is an address of a documentation range the database has no
	// record of.
	Nowhere = "192.0.2.1"
)

// Write writes the database into a directory of the test's own, and returns
// the file's path.
func Write(t *testing.T) string {
	t.Helper()

	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: Type,
		BuildEpoch:   Built.Unix(),
		// The documentation and private ranges are reserved, and only a
		// database told to may hold them.
		IncludeReservedNetworks: true,
	})
	if err != nil {
		t.Fatalf("geoiptest: make the database: %v", err)
	}

	for network, record := range map[string]mmdbtype.Map{
		"203.0.113.0/24": {
			"country": mmdbtype.Map{"iso_code": mmdbtype.String(Country)},
			"city":    mmdbtype.Map{"names": mmdbtype.Map{"en": mmdbtype.String(City)}},
		},
		"2001:db8:1::/48": {"country": mmdbtype.Map{"iso_code": mmdbtype.String(CountryV6)}},
		"10.0.0.0/8":      {"country": mmdbtype.Map{"iso_code": mmdbtype.String("DE")}},
		"198.51.100.0/24": {"country": mmdbtype.Map{"iso_code": mmdbtype.String("ZZ")}},
	} {
		_, parsed, err := net.ParseCIDR(network)
		if err != nil {
			t.Fatalf("geoiptest: read the network %s: %v", network, err)
		}
		if err := tree.Insert(parsed, record); err != nil {
			t.Fatalf("geoiptest: put %s in the database: %v", network, err)
		}
	}

	path := filepath.Join(t.TempDir(), "countries.mmdb")
	file, err := os.Create(path) //nolint:gosec // a path of the test's own directory
	if err != nil {
		t.Fatalf("geoiptest: create the file: %v", err)
	}
	if _, err := tree.WriteTo(file); err != nil {
		_ = file.Close()
		t.Fatalf("geoiptest: write the database: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("geoiptest: close the file: %v", err)
	}
	return path
}
