package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Metadata is what a version of the record says of itself, in the shape
// Zenodo's deposit API takes. The file it is read from describes the record;
// the step adds the version, its date and the DOI it asks Zenodo to keep.
type Metadata struct {
	UploadType         string              `json:"upload_type"`
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	Creators           []Creator           `json:"creators"`
	AccessRight        string              `json:"access_right"`
	License            string              `json:"license"`
	Keywords           []string            `json:"keywords,omitempty"`
	RelatedIdentifiers []RelatedIdentifier `json:"related_identifiers,omitempty"`
	Version            string              `json:"version,omitempty"`
	PublicationDate    string              `json:"publication_date,omitempty"`
	PrereserveDOI      bool                `json:"prereserve_doi,omitempty"`
}

// Creator is an author of the record, named "Family, Given".
type Creator struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
	ORCID       string `json:"orcid,omitempty"`
}

// RelatedIdentifier is a work the record is related to, and how.
type RelatedIdentifier struct {
	Identifier   string `json:"identifier"`
	Relation     string `json:"relation"`
	ResourceType string `json:"resource_type,omitempty"`
}

// ReadMetadata reads the record's description from a file, refusing a field
// it does not know, and one a step sets itself.
func ReadMetadata(path string) (Metadata, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Metadata{}, fmt.Errorf("the record's description: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var m Metadata
	if err := decoder.Decode(&m); err != nil {
		return Metadata{}, fmt.Errorf("the record's description, %s: %w", path, err)
	}
	if err := m.check(); err != nil {
		return Metadata{}, fmt.Errorf("the record's description, %s, %w", path, err)
	}
	return m, nil
}

// check holds a description to what a published record must say. A creator
// still named by a placeholder is refused before any draft is made: a record's
// authors are public once it is published, and a draft is what gets published.
func (m *Metadata) check() error {
	switch {
	case m.Version != "" || m.PublicationDate != "" || m.PrereserveDOI:
		return errors.New("sets a version, a date or a DOI, which each step sets itself")
	case !slices.Contains([]string{"software", "dataset"}, m.UploadType):
		return fmt.Errorf("gives the type %q, and an artifact of code and data is software or a dataset", m.UploadType)
	case strings.TrimSpace(m.Title) == "" || strings.TrimSpace(m.Description) == "":
		return errors.New("gives no title or no description")
	case m.AccessRight != "open" || m.License == "":
		return errors.New("is not open under a licence")
	case len(m.Creators) == 0:
		return errors.New("names no creator")
	}
	for _, c := range m.Creators {
		if strings.TrimSpace(c.Name) == "" || strings.ContainsAny(c.Name+c.Affiliation, "[]") {
			return fmt.Errorf("names a creator by a placeholder or by nothing, %q, and a record's authors are public once it is published", c.Name)
		}
	}
	return nil
}
