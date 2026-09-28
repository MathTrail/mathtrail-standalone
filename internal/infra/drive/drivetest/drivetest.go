// Package drivetest stands in for Google Drive, as far as the service can see
// it: a Drive for every parent, holding the folders and the files the service
// made there beside whatever the parent keeps of their own, which the service
// is never shown — the drive.file scope in a few lines. It speaks the part of
// Drive's API the service uses, and refuses anything else out loud rather than
// answering it the way Drive might not.
//
// A test puts into a Drive what a parent or another build would have left
// there, reads back what the service wrote, and can have the next call of a
// kind refused, or left unanswered until the caller gives up. Every call is
// counted by its kind, so that what a piece of the service costs in calls can
// be held to a number.
//
// It is test code that lives in a package rather than a test file, because the
// tests of more than one package reach a parent's Drive, and a test file cannot
// be shared between packages.
package drivetest

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// The kinds of call, one for each request the service makes. They are what
// the stand-in counts, and what a test names to have the next call refused.
const (
	List     = "list"
	Get      = "get"
	Download = "download"
	Create   = "create"
	Update   = "update"
)

// File is a file in a parent's Drive, as a test puts it there and reads it
// back.
type File struct {
	ID            string
	Name          string
	MimeType      string
	Parents       []string
	AppProperties map[string]string
	// Content is what the file holds; a folder holds nothing.
	Content      []byte
	ModifiedTime time.Time
	Trashed      bool
	// Own marks a file the parent made themselves. Under drive.file the
	// service is never shown it: no search finds it, and asking for it by its
	// ID is answered as if there were no such file.
	Own bool
}

// Calls counts the calls of each kind the stand-in was asked.
type Calls map[string]int

// fault is how the next call of a kind is answered instead of served.
type fault struct {
	status int
	reason string
	// stall leaves the call without an answer until its caller gives up.
	stall bool
}

// Drive is the stand-in: one server, and a Drive behind it for every token.
type Drive struct {
	t      testing.TB
	server *httptest.Server

	mu sync.Mutex
	// drives holds each parent's files, by the token that reaches them. A
	// token not seen before opens a Drive of its own, empty.
	drives  map[string][]*File
	revoked map[string]bool
	calls   Calls
	faults  map[string][]fault
	// clock is when the latest change was made. Every change moves it on by
	// a second, so that the most recently modified file is never a tie.
	clock time.Time
}

// New starts the stand-in, and stops it when the test ends.
func New(t testing.TB) *Drive {
	t.Helper()

	d := &Drive{
		t:       t,
		drives:  map[string][]*File{},
		revoked: map[string]bool{},
		calls:   Calls{},
		faults:  map[string][]fault{},
		clock:   time.Date(2026, time.September, 28, 10, 0, 0, 0, time.UTC),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /drive/v3/files", d.serve(List, d.list))
	mux.HandleFunc("GET /drive/v3/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("alt") == "media" {
			d.serve(Download, d.download)(w, r)
			return
		}
		d.serve(Get, d.get)(w, r)
	})
	mux.HandleFunc("POST /drive/v3/files", d.serve(Create, d.createEmpty))
	mux.HandleFunc("POST /upload/drive/v3/files", d.serve(Create, d.createFile))
	mux.HandleFunc("PATCH /upload/drive/v3/files/{id}", d.serve(Update, d.update))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d.malformed(w, "%s %s is not a request the service makes", r.Method, r.URL.Path)
	})
	d.server = httptest.NewServer(mux)
	t.Cleanup(d.server.Close)
	return d
}

// Root is where the stand-in's API is, in place of Drive's own.
func (d *Drive) Root() string { return d.server.URL }

// Calls is how many calls of each kind the stand-in was asked since it
// started, or since its count was last reset.
func (d *Drive) Calls() Calls {
	d.mu.Lock()
	defer d.mu.Unlock()
	counted := Calls{}
	for kind, n := range d.calls {
		counted[kind] = n
	}
	return counted
}

// ResetCalls starts the count again from nothing.
func (d *Drive) ResetCalls() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = Calls{}
}

// Revoke takes the service's access back: from now on the token reaches
// nothing, and every call made with it is answered 401.
func (d *Drive) Revoke(token string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.revoked[token] = true
}

// Fail has the next call of a kind refused with the status given, and the
// reason Drive would give for it.
func (d *Drive) Fail(kind string, status int, reason string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faults[kind] = append(d.faults[kind], fault{status: status, reason: reason})
}

// Stall leaves the next call of a kind without an answer until its caller
// gives up on it.
func (d *Drive) Stall(kind string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faults[kind] = append(d.faults[kind], fault{stall: true})
}

// Put keeps a file in the Drive a token reaches, as a parent or something
// other than this service would have left it, and answers its ID. A file with
// no ID is given one, and one with no time of change is changed now.
func (d *Drive) Put(token string, file *File) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	kept := file.copied()
	if kept.ID == "" {
		kept.ID = rand.Text()
	}
	if kept.ModifiedTime.IsZero() {
		kept.ModifiedTime = d.tick()
	}
	d.drives[token] = append(d.drives[token], kept)
	return kept.ID
}

// Files is everything in the Drive a token reaches, the parent's own files
// among it, in the order the files were made: copies, which a test may read
// and change without changing the Drive.
func (d *Drive) Files(token string) []*File {
	d.mu.Lock()
	defer d.mu.Unlock()

	files := make([]*File, 0, len(d.drives[token]))
	for _, file := range d.drives[token] {
		files = append(files, file.copied())
	}
	return files
}

// Edit changes a file the way a parent would in Drive itself: renamed, moved,
// put in the bin, rewritten. The file is changed now, unless the change says
// when.
func (d *Drive) Edit(token, id string, change func(*File)) {
	d.t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()

	file := d.find(token, id)
	if file == nil {
		d.t.Fatalf("drivetest: Edit() of a file the Drive does not hold")
		return
	}
	edited := file.copied()
	edited.ModifiedTime = time.Time{}
	change(edited)
	if edited.ModifiedTime.IsZero() {
		edited.ModifiedTime = d.tick()
	}
	*file = *edited
}

// Delete removes a file for good, bin and all.
func (d *Drive) Delete(token, id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.drives[token] = slices.DeleteFunc(d.drives[token], func(file *File) bool { return file.ID == id })
}

// handler is how the stand-in serves one kind of call to a parent's Drive.
type handler func(w http.ResponseWriter, r *http.Request, token string)

// serve counts a call, holds it to its token, and answers it the way a test
// asked the next call of its kind to be answered, or as Drive would.
func (d *Drive) serve(kind string, serve handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		d.calls[kind]++
		token, signed := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		revoked := d.revoked[token]
		faults := d.faults[kind]
		var planned *fault
		if len(faults) > 0 {
			planned = &faults[0]
			d.faults[kind] = faults[1:]
		}
		d.mu.Unlock()

		switch {
		case !signed || token == "" || revoked:
			refuse(w, http.StatusUnauthorized, "authError", "Invalid Credentials")
		case planned != nil && planned.stall:
			<-r.Context().Done()
		case planned != nil:
			refuse(w, planned.status, planned.reason, "Refused, as the test asked.")
		default:
			serve(w, r, token)
		}
	}
}

// tick moves the clock on by a second and answers the new time. The caller
// holds the lock.
func (d *Drive) tick() time.Time {
	d.clock = d.clock.Add(time.Second)
	return d.clock
}

// find is a file of the Drive a token reaches, the parent's own included, or
// nil. The caller holds the lock.
func (d *Drive) find(token, id string) *File {
	for _, file := range d.drives[token] {
		if file.ID == id {
			return file
		}
	}
	return nil
}

// visible is a file the service may see, or nil: under drive.file the
// parent's own files are not there for it at all. The caller holds the lock.
func (d *Drive) visible(token, id string) *File {
	if file := d.find(token, id); file != nil && !file.Own {
		return file
	}
	return nil
}

// malformed refuses a request no correct caller makes, and fails the test
// that made it with what was wrong: every such request is a fault of the
// service's own, and Drive's answer to it would say less.
func (d *Drive) malformed(w http.ResponseWriter, format string, args ...any) {
	d.t.Errorf("drivetest: "+format, args...)
	refuse(w, http.StatusBadRequest, "badRequest", "Bad Request")
}

// copied is a file that shares nothing with the one it was copied from.
func (f *File) copied() *File {
	file := *f
	file.Parents = slices.Clone(f.Parents)
	file.Content = slices.Clone(f.Content)
	if f.AppProperties != nil {
		file.AppProperties = make(map[string]string, len(f.AppProperties))
		for key, value := range f.AppProperties {
			file.AppProperties[key] = value
		}
	}
	return &file
}

// refuse answers a call the way Drive refuses one: a status, and the reason
// in Drive's error format.
func refuse(w http.ResponseWriter, status int, reason, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code":    status,
		"message": message,
		"errors":  []map[string]string{{"domain": "global", "reason": reason, "message": message}},
	}})
}

// writeJSON answers with a JSON body.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
