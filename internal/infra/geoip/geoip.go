// Package geoip tells the country an address was handed out in, from a
// database of the MaxMind DB format — DB-IP's IP-to-Country Lite in a
// deployment — read from a file when the service starts and kept open until it
// stops.
//
// Nothing leaves the process: the database is a file of the service's own, and
// an address is looked up in memory and forgotten. Only a country's code comes
// out, and only a code of the closed list a line may carry.
package geoip

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/country"
)

// ErrDatabase is returned when the file is no database that can be read: it
// is missing, unreadable, or not of the format.
var ErrDatabase = errors.New("geoip: the database does not open")

// Database is an open database of countries.
type Database interface {
	// Country is the code of the country the address was handed out in, as
	// the list of countries writes it, or empty when that is not known: an
	// address of a network nobody's country holds — a private one, this
	// machine's own, one of a local link — an address the database has no
	// record of, and a code the list does not have.
	Country(addr netip.Addr) string
	// About says what the database is and when it was built.
	About() About
	// Close lets go of the file once the lookups under way have finished. A
	// lookup after it finds no country.
	Close() error
}

// About is what a database says of itself: the kind of database it is, as its
// maker names it, and when it was built.
type About struct {
	Type  string
	Built time.Time
}

// database is a file of the format, mapped into memory. A lookup may still be
// running when the file is let go of — one for a request nobody waits for any
// longer — so the lock keeps the file mapped until the lookups under way have
// finished, and reading memory no longer mapped would end the process. A
// lookup after that finds the reader closed, which answers it with an error:
// no country.
type database struct {
	mu     sync.RWMutex
	reader *maxminddb.Reader
}

// Open opens the database in the file at path. The file is mapped into
// memory, so it must not change while it is open.
func Open(path string) (Database, error) {
	reader, err := maxminddb.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDatabase, err)
	}
	return &database{reader: reader}, nil
}

func (d *database) Country(addr netip.Addr) string {
	addr = addr.Unmap()
	// An address no country holds is not looked up at all, so that a database
	// that names one anyway cannot put a country on it.
	if !addr.IsValid() || !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return ""
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Only the code is read out of the record: whatever else a database keeps
	// beside it never enters the process.
	var code string
	if err := d.reader.Lookup(addr).DecodePath(&code, "country", "iso_code"); err != nil || !country.Known(code) {
		return ""
	}
	return code
}

func (d *database) About() About {
	return About{Type: d.reader.Metadata.DatabaseType, Built: d.reader.Metadata.BuildTime().UTC()}
}

func (d *database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.reader.Close(); err != nil {
		return fmt.Errorf("geoip: close the database: %w", err)
	}
	return nil
}
