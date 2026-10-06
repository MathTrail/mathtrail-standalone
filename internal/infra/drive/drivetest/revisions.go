package drivetest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const (
	// maxKept is how many revisions of a file Drive keeps forever at most.
	maxKept = 200
	// maxPurgeable is how many revisions of a file not kept forever Drive
	// holds on to once another is uploaded: past it, the earliest is purged.
	maxPurgeable = 100
	// defaultRevisionPage is how many revisions a list answers with when it
	// does not say.
	defaultRevisionPage = 200
)

// listRevisions answers the revisions of a file, with the fields asked for,
// in no order worth relying on: Drive promises none, so the stand-in keeps
// none either.
func (d *Drive) listRevisions(w http.ResponseWriter, r *http.Request, token string) {
	query := r.URL.Query()
	fields, err := revisionListSelection(query.Get("fields"))
	if err != nil {
		d.malformed(w, "fields: %v", err)
		return
	}
	page := defaultRevisionPage
	if given := query.Get("pageSize"); given != "" {
		if page, err = pageSize(given); err != nil {
			d.malformed(w, "pageSize: %v", err)
			return
		}
	}
	id := r.PathValue("id")
	d.mu.Lock()
	var revisions []Revision
	file := d.visible(token, id)
	if file != nil {
		revisions = file.copied().Revisions
	}
	d.mu.Unlock()
	if file == nil {
		notFound(w, id)
		return
	}

	slices.SortFunc(revisions, func(a, b Revision) int { return strings.Compare(a.ID, b.ID) })
	listed := make([]map[string]any, 0, min(page, len(revisions)))
	for i := range revisions[:min(page, len(revisions))] {
		listed = append(listed, revisions[i].pick(fields))
	}
	answer := map[string]any{"revisions": listed}
	if len(revisions) > page {
		answer["nextPageToken"] = "the-next-page"
	}
	writeJSON(w, http.StatusOK, answer)
}

// downloadRevision answers what a revision of a file held — only when it is
// kept forever, as Drive gives no other.
func (d *Drive) downloadRevision(w http.ResponseWriter, r *http.Request, token string) {
	if r.URL.Query().Get("alt") != "media" {
		d.malformed(w, "the service reads what a revision held, and nothing else of it")
		return
	}
	id, wanted := r.PathValue("id"), r.PathValue("revision")
	d.mu.Lock()
	var revision *Revision
	file := d.visible(token, id)
	if file != nil {
		if at := file.revisionAt(wanted); at >= 0 {
			copied := file.copied().Revisions[at]
			revision = &copied
		}
	}
	d.mu.Unlock()

	switch {
	case file == nil:
		notFound(w, id)
	case revision == nil:
		revisionNotFound(w, wanted)
	case !revision.KeepForever:
		// The reason as the example in Drive's guide to its errors spells it.
		Refuse(w, http.StatusForbidden, "download_restricted_for_revision",
			"This revision cannot be downloaded by the authenticated user.")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		// What a revision held, served back to the client of the service that
		// wrote it, as Drive serves it: no browser reads it.
		_, _ = w.Write(revision.Content)
	}
}

// keepRevision marks a revision of a file to be kept forever. A revision so
// marked cannot be let go again, only deleted, and a file keeps at most
// maxKept so.
func (d *Drive) keepRevision(w http.ResponseWriter, r *http.Request, token string) {
	fields, err := fileSelectionOf(r.URL.Query().Get("fields"), revisionFields)
	if err != nil {
		d.malformed(w, "fields: %v", err)
		return
	}
	var asked struct {
		KeepForever *bool `json:"keepForever"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&asked); err != nil || asked.KeepForever == nil {
		d.malformed(w, "the change of a revision is its keepForever, and nothing else")
		return
	}
	id, wanted := r.PathValue("id"), r.PathValue("revision")
	d.mu.Lock()
	defer d.mu.Unlock()
	file := d.visible(token, id)
	if file == nil {
		notFound(w, id)
		return
	}
	at := file.revisionAt(wanted)
	switch {
	case at < 0:
		revisionNotFound(w, wanted)
		return
	case !*asked.KeepForever && file.Revisions[at].KeepForever:
		Refuse(w, http.StatusBadRequest, "illegalKeepForeverModification",
			"Bad Request. Cannot update a revision to false that is marked as keepForever.")
		return
	case *asked.KeepForever && !file.Revisions[at].KeepForever && file.kept() >= maxKept:
		refuseKeeping(w)
		return
	}
	file.Revisions[at].KeepForever = *asked.KeepForever
	writeJSON(w, http.StatusOK, file.Revisions[at].pick(fields))
}

// deleteRevision removes a revision of a file for good. The service lets go of
// old revisions kept forever this way, and never of what a file holds now:
// that would be a fault of its own.
func (d *Drive) deleteRevision(w http.ResponseWriter, r *http.Request, token string) {
	id, wanted := r.PathValue("id"), r.PathValue("revision")
	d.mu.Lock()
	file := d.visible(token, id)
	at := -1
	if file != nil {
		at = file.revisionAt(wanted)
	}
	switch {
	case file == nil:
		d.mu.Unlock()
		notFound(w, id)
		return
	case at < 0:
		d.mu.Unlock()
		revisionNotFound(w, wanted)
		return
	case at == len(file.Revisions)-1:
		d.mu.Unlock()
		d.malformed(w, "the service deleted what a file holds now")
		return
	}
	file.Revisions = slices.Delete(file.Revisions, at, at+1)
	d.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// revisionAt is where a revision is in the file's history, or -1.
func (f *File) revisionAt(id string) int {
	return slices.IndexFunc(f.Revisions, func(revision Revision) bool { return revision.ID == id })
}

// kept is how many revisions of the file are kept forever.
func (f *File) kept() int {
	kept := 0
	for i := range f.Revisions {
		if f.Revisions[i].KeepForever {
			kept++
		}
	}
	return kept
}

// purge lets go of the earliest revisions Drive may purge, past the number it
// holds on to once another has been uploaded. What a file holds now is never
// one of them.
func (f *File) purge() {
	for {
		var purgeable []int
		for i := range f.Revisions[:len(f.Revisions)-1] {
			if !f.Revisions[i].KeepForever {
				purgeable = append(purgeable, i)
			}
		}
		if len(purgeable) <= maxPurgeable {
			return
		}
		f.Revisions = slices.Delete(f.Revisions, purgeable[0], purgeable[0]+1)
	}
}

// Purge lets go of every revision of a file that is neither kept forever nor
// what it holds now, as Drive does with them after about thirty days.
func (d *Drive) Purge(token, id string) {
	d.t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()

	file := d.find(token, id)
	if file == nil {
		d.t.Fatalf("drivetest: Purge() of a file the Drive does not hold")
		return
	}
	if len(file.Revisions) == 0 {
		return
	}
	head := file.Revisions[len(file.Revisions)-1].ID
	file.Revisions = slices.DeleteFunc(file.Revisions, func(revision Revision) bool {
		return !revision.KeepForever && revision.ID != head
	})
}

// pick is a revision as Drive answers with it: the fields asked for.
func (rev *Revision) pick(fields []string) map[string]any {
	answer := map[string]any{}
	for _, field := range fields {
		switch field {
		case "kind":
			answer["kind"] = "drive#revision"
		case "id":
			answer["id"] = rev.ID
		case "modifiedTime":
			answer["modifiedTime"] = rev.ModifiedTime.UTC().Format("2006-01-02T15:04:05.000Z")
		case "keepForever":
			answer["keepForever"] = rev.KeepForever
		}
	}
	return answer
}

// refuseKeeping answers a revision to be kept forever beyond the most Drive
// keeps. Drive's documentation does not say how it refuses one, so the
// stand-in refuses it as a request Drive would not take.
func refuseKeeping(w http.ResponseWriter) {
	Refuse(w, http.StatusBadRequest, "badRequest", "At most "+strconv.Itoa(maxKept)+" revisions can be kept forever.")
}

// revisionNotFound answers a call about a revision the file does not have,
// naming it in the sentence, as Drive does.
func revisionNotFound(w http.ResponseWriter, id string) {
	Refuse(w, http.StatusNotFound, "notFound", fmt.Sprintf("Revision not found: %s.", id))
}
