package main

import (
	"context"
	"crypto/md5" //nolint:gosec // Zenodo reports what it stored by MD5; the sum is compared, and guards nothing
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DraftRequest is what the draft step is asked to put on Zenodo.
type DraftRequest struct {
	Metadata Metadata
	Archive  string // the file the version holds
	Version  string // the name of the paper's release it is
	Record   string // the record's concept id, or nothing for a record of its own
	Today    string // the version's date, YYYY-MM-DD
}

// Draft puts an archive into a draft of the record and says where the draft
// is and which DOI it will have, leaving its publication to a person. A draft
// of the same version left by an earlier run is taken up again, and a version
// already published leaves nothing to do.
func Draft(ctx context.Context, c Client, r *DraftRequest, out io.Writer) error {
	d, published, err := draftFor(ctx, c, r)
	if err != nil {
		return err
	}
	if published {
		_, err = fmt.Fprintf(out, "Version %s of record %s is published already, so there is nothing to draft.\n", r.Version, r.Record)
		return err
	}
	// The draft says what it is before it holds anything, so that a run cut
	// short leaves a draft the next run knows by its title.
	m := r.Metadata
	m.Version, m.PublicationDate, m.PrereserveDOI = r.Version, r.Today, true
	if d, err = c.Update(ctx, d.ID, &m); err != nil {
		return err
	}
	err = replaceFiles(ctx, c, &d, r.Archive)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "The draft of version %s is ready to be looked at: %s\n%s",
		r.Version, d.Links.HTML, doisOf(d.ReservedDOI(), d.ConceptRecID))
	return err
}

// Publish publishes the draft of a version. A version already published
// leaves nothing to do; a version with no draft, or with two, is refused.
func Publish(ctx context.Context, c Client, version, record string, out io.Writer) error {
	drafts, err := c.List(ctx, "draft")
	if err != nil {
		return err
	}
	d, err := oneOf(drafts, func(d *Deposition) bool {
		return d.Metadata.Version == version && (record == "" || d.ConceptRecID == record)
	})
	if err != nil {
		return err
	}
	if d.ID == 0 {
		return nothingToPublish(ctx, c, version, record, out)
	}
	if d, err = c.Publish(ctx, d.ID); err != nil {
		return err
	}
	shared := d.ConceptDOI
	if shared == "" {
		shared = conceptDOI(d.DOI, d.ConceptRecID)
	}
	_, err = fmt.Fprintf(out, "Version %s is published as %s.\nIts record is %s; every version of it is cited by %s.\n",
		version, d.DOI, d.ConceptRecID, shared)
	return err
}

// nothingToPublish says why no draft of a version is there to publish: the
// version is published already, or the draft step never ran for it. With no
// record given, the version published in any record of the owner's counts,
// and the record is named for the steps after it.
func nothingToPublish(ctx context.Context, c Client, version, record string, out io.Writer) error {
	published, err := c.List(ctx, "published")
	if err != nil {
		return err
	}
	for i := range published {
		d := &published[i]
		if d.Metadata.Version == version && (record == "" || d.ConceptRecID == record) {
			_, err = fmt.Fprintf(out, "Version %s is published already, as %s, in record %s: the variable ZENODO_RECORD names that record.\n", version, d.DOI, d.ConceptRecID)
			return err
		}
	}
	return fmt.Errorf("no draft holds version %s: run the draft step first", version)
}

// draftFor finds the draft a version goes into, and says when the version is
// published already. For a record of its own it is the draft an earlier run
// left of a record with the same title, or a new record; for a record, the
// draft its owner has open, or a new draft of the version after its latest.
func draftFor(ctx context.Context, c Client, r *DraftRequest) (Deposition, bool, error) {
	version, record := r.Version, r.Record
	drafts, err := c.List(ctx, "draft")
	if err != nil {
		return Deposition{}, false, err
	}
	if record == "" {
		var own Deposition
		own, err = draftOfItsOwn(ctx, c, drafts, r.Metadata.Title)
		return own, false, err
	}
	latest, ok, err := latestPublished(ctx, c, record)
	switch {
	case err != nil:
		return Deposition{}, false, err
	case !ok:
		return Deposition{}, false, fmt.Errorf("record %s has no published version of its owner's to follow", record)
	case latest.Metadata.Version == version:
		return latest, true, nil
	}
	open, err := oneOf(drafts, func(d *Deposition) bool { return d.ConceptRecID == record })
	if err != nil || open.ID != 0 {
		return open, false, err
	}
	next, err := c.NewVersion(ctx, latest.ID)
	if err != nil {
		return Deposition{}, false, err
	}
	if next.Links.LatestDraft == "" {
		return Deposition{}, false, errors.New("the answer to a new version named no draft of it")
	}
	d, err := c.Get(ctx, next.Links.LatestDraft)
	return d, false, err
}

// draftOfItsOwn is the draft of a record of its own: the draft an earlier run
// left of a record with the title, or a new record. A record of the title
// published already is refused, since its versions follow it: a second record
// would give the artifact a second DOI for good.
func draftOfItsOwn(ctx context.Context, c Client, drafts []Deposition, title string) (Deposition, error) {
	published, err := c.List(ctx, "published")
	if err != nil {
		return Deposition{}, err
	}
	for i := range published {
		if published[i].Metadata.Title == title {
			return Deposition{}, fmt.Errorf("record %s holds this title already: give its id, the variable ZENODO_RECORD, rather than start a second record", published[i].ConceptRecID)
		}
	}
	left, err := oneOf(drafts, func(d *Deposition) bool { return d.Metadata.Title == title })
	if err != nil || left.ID != 0 {
		return left, err
	}
	return c.Create(ctx)
}

// oneOf is the one deposition that matches, or none; two are refused rather
// than one of them chosen.
func oneOf(ds []Deposition, match func(*Deposition) bool) (Deposition, error) {
	var found []int
	for i := range ds {
		if match(&ds[i]) {
			found = append(found, i)
		}
	}
	switch len(found) {
	case 0:
		return Deposition{}, nil
	case 1:
		return ds[found[0]], nil
	}
	return Deposition{}, fmt.Errorf("%d drafts match, where one is taken up again; discard all but one on Zenodo", len(found))
}

// latestPublished is the newest published version of a record, the one with
// the highest id.
func latestPublished(ctx context.Context, c Client, record string) (Deposition, bool, error) {
	published, err := c.List(ctx, "published")
	if err != nil {
		return Deposition{}, false, err
	}
	latest := -1
	for i := range published {
		if published[i].ConceptRecID == record && (latest < 0 || published[i].ID > published[latest].ID) {
			latest = i
		}
	}
	if latest < 0 {
		return Deposition{}, false, nil
	}
	return published[latest], true, nil
}

// replaceFiles leaves the archive the only file of the draft: a new version
// holds its predecessor's files until they are taken out. What Zenodo stored
// is held to the file's own sum.
func replaceFiles(ctx context.Context, c Client, d *Deposition, archive string) error {
	files, err := c.Files(ctx, d.ID)
	if err != nil {
		return err
	}
	for _, f := range files {
		err = c.DeleteFile(ctx, d.ID, f.ID)
		if err != nil {
			return err
		}
	}
	file, err := os.Open(filepath.Clean(archive))
	if err != nil {
		return fmt.Errorf("the archive: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("the archive: %w", err)
	}
	sum := md5.New() //nolint:gosec // compared with the sum Zenodo reports, and guards nothing
	stored, err := c.Upload(ctx, d.Links.Bucket, filepath.Base(archive), io.TeeReader(file, sum), info.Size())
	if err != nil {
		return err
	}
	if want := "md5:" + hex.EncodeToString(sum.Sum(nil)); stored != want {
		return fmt.Errorf("what was stored of %s sums to %s, and the file to %s", filepath.Base(archive), stored, want)
	}
	return nil
}

// doisOf says which DOI a draft will have and which DOI every version of its
// record shares.
func doisOf(reserved, record string) string {
	var b strings.Builder
	if reserved != "" {
		fmt.Fprintf(&b, "Once published it is %s.\n", reserved)
	}
	if record != "" {
		fmt.Fprintf(&b, "Its record is %s; every version of it is cited by %s.\n", record, conceptDOI(reserved, record))
	}
	return b.String()
}

// conceptDOI is the DOI a record's every version shares: Zenodo's prefix,
// read from a DOI of one of its versions, and the record's concept id.
func conceptDOI(doi, record string) string {
	i := strings.LastIndex(doi, ".")
	if i < 0 || record == "" {
		return "the DOI Zenodo shows on the record's page"
	}
	return doi[:i+1] + record
}
