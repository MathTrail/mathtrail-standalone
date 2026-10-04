// Package country holds the countries a family can say it lives in, and the
// regions of the countries that have them, by their ISO 3166 codes: the closed
// lists a code is held to before a profile keeps it or a line carries it.
//
// The lists live in places.json beside this file, which the widget's form
// reads too, so that the codes a parent can choose and the codes the service
// accepts are one list.
package country

import (
	// The blank import is what makes the embed directive below work.
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
)

//go:embed places.json
var placesJSON []byte

// places is the file as it is written: every country by its ISO 3166-1
// alpha-2 code, and, for a country a profile may name a region of, each region
// by its ISO 3166-2 code with its name in English.
type places struct {
	Countries []string                     `json:"countries"`
	Regions   map[string]map[string]string `json:"regions"`
}

// known is the file, read once. It is part of the binary, so a file that does
// not read is a build that must not ship, and the tests of this package are
// what stop it.
var known = mustRead(placesJSON)

func mustRead(raw []byte) places {
	var read places
	if err := json.Unmarshal(raw, &read); err != nil {
		panic(fmt.Sprintf("country: the embedded list does not read: %v", err))
	}
	return read
}

// Known reports whether a code is a country of the list, written as the list
// writes it: two capital letters. The list is the officially assigned codes of
// ISO 3166-1, and XK, the code Kosovo is known by where ISO assigns none.
func Known(code string) bool {
	return slices.Contains(known.Countries, code)
}

// KnownRegion reports whether a code is one of the regions of the country, as
// ISO 3166-2 writes it: the United States by its states and the District of
// Columbia, such as US-TX. A country the list names no regions of has none a
// profile may name.
func KnownRegion(country, region string) bool {
	_, listed := known.Regions[country][region]
	return listed
}

// RegionName is the name of a region in English, or empty when the code is
// none of the list's.
func RegionName(region string) string {
	for _, regions := range known.Regions {
		if name, listed := regions[region]; listed {
			return name
		}
	}
	return ""
}
