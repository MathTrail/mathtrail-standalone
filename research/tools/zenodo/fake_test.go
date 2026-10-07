package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

// fakeToken is the only token the stand-in for Zenodo lets in.
const fakeToken = "the-token"

// fakeZenodo stands in for Zenodo's deposit API: its depositions held in
// memory, the token checked on every request, and the rules Zenodo keeps — one
// open draft a record, a new version only from the latest published one, and
// a draft published only once it holds a file and a version.
type fakeZenodo struct {
	mu        sync.Mutex
	server    *httptest.Server
	nextID    int64
	deps      map[int64]*fakeDeposition
	wrongSum  bool     // reports a sum that is not of what it stored
	pageSize  int      // the most depositions a page of the list holds, or every one
	elsewhere bool     // answers a new version with a draft at another address
	broken    bool     // fails every upload
	requests  []string // each request's method and path, in order
}

type fakeDeposition struct {
	id, concept int64
	meta        Metadata
	files       []fakeFile
	published   bool
}

type fakeFile struct {
	id, name string
	data     []byte
}

func newFakeZenodo(t *testing.T) *fakeZenodo {
	t.Helper()
	f := &fakeZenodo{nextID: 1000, deps: map[int64]*fakeDeposition{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/deposit/depositions", f.list)
	mux.HandleFunc("POST /api/deposit/depositions", f.create)
	mux.HandleFunc("GET /api/deposit/depositions/{id}", f.get)
	mux.HandleFunc("PUT /api/deposit/depositions/{id}", f.update)
	mux.HandleFunc("POST /api/deposit/depositions/{id}/actions/newversion", f.newVersion)
	mux.HandleFunc("POST /api/deposit/depositions/{id}/actions/publish", f.publish)
	mux.HandleFunc("GET /api/deposit/depositions/{id}/files", f.files)
	mux.HandleFunc("DELETE /api/deposit/depositions/{id}/files/{file}", f.deleteFile)
	mux.HandleFunc("PUT /files/{id}/{name}", f.upload)
	f.server = httptest.NewServer(f.authorized(mux))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeZenodo) client() Client {
	return Client{HTTP: f.server.Client(), URL: f.server.URL, Token: fakeToken}
}

func (f *fakeZenodo) authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.requests = append(f.requests, r.Method+" "+r.URL.Path)
		f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+fakeToken {
			http.Error(w, `{"message": "the token is not valid"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// doiOf is the DOI the stand-in gives a deposition, under the sandbox's prefix.
func doiOf(id int64) string { return "10.5072/zenodo." + strconv.FormatInt(id, 10) }

func (f *fakeZenodo) view(d *fakeDeposition) map[string]any {
	id := strconv.FormatInt(d.id, 10)
	v := map[string]any{
		"id": d.id, "conceptrecid": strconv.FormatInt(d.concept, 10), "submitted": d.published,
		"metadata": map[string]any{"title": d.meta.Title, "version": d.meta.Version, "prereserve_doi": map[string]any{"doi": doiOf(d.id), "recid": d.id}},
		"links":    map[string]any{"bucket": f.server.URL + "/files/" + id, "html": f.server.URL + "/deposit/" + id},
	}
	if d.published {
		v["doi"], v["conceptdoi"], v["record_id"] = doiOf(d.id), doiOf(d.concept), d.id
	}
	return v
}

func answer(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// deposition is the deposition the request's path names, or nothing, having
// answered 404.
func (f *fakeZenodo) deposition(w http.ResponseWriter, r *http.Request) *fakeDeposition {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if d, ok := f.deps[id]; err == nil && ok {
		return d
	}
	http.Error(w, `{"message": "no such deposition"}`, http.StatusNotFound)
	return nil
}

func (f *fakeZenodo) list(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	published := r.URL.Query().Get("status") == "published"
	views := []map[string]any{}
	for id := int64(0); id <= f.nextID; id++ {
		if d, ok := f.deps[id]; ok && d.published == published {
			views = append(views, f.view(d))
		}
	}
	size := f.pageSize
	if size == 0 {
		size = len(views) + 1
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	start := (page - 1) * size
	if err != nil || page < 1 || start >= len(views) {
		answer(w, http.StatusOK, []map[string]any{})
		return
	}
	answer(w, http.StatusOK, views[start:min(start+size, len(views))])
}

func (f *fakeZenodo) create(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID += 2
	d := &fakeDeposition{id: f.nextID, concept: f.nextID - 1}
	f.deps[d.id] = d
	answer(w, http.StatusCreated, f.view(d))
}

func (f *fakeZenodo) get(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d := f.deposition(w, r); d != nil {
		answer(w, http.StatusOK, f.view(d))
	}
}

func (f *fakeZenodo) update(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	var body struct {
		Metadata Metadata `json:"metadata"`
	}
	if d.published || json.NewDecoder(r.Body).Decode(&body) != nil {
		http.Error(w, `{"message": "not a draft, or not metadata"}`, http.StatusBadRequest)
		return
	}
	d.meta = body.Metadata
	answer(w, http.StatusOK, f.view(d))
}

func (f *fakeZenodo) newVersion(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	for _, other := range f.deps {
		if other.concept == d.concept && (!other.published || other.id > d.id) {
			http.Error(w, `{"message": "a draft is open, or this is not the latest version"}`, http.StatusBadRequest)
			return
		}
	}
	f.nextID++
	next := &fakeDeposition{id: f.nextID, concept: d.concept, meta: d.meta, files: append([]fakeFile(nil), d.files...)}
	f.deps[next.id] = next
	v := f.view(d)
	links, ok := v["links"].(map[string]any)
	if !ok {
		http.Error(w, `{"message": "no links"}`, http.StatusInternalServerError)
		return
	}
	links["latest_draft"] = f.server.URL + "/api/deposit/depositions/" + strconv.FormatInt(next.id, 10)
	if f.elsewhere {
		links["latest_draft"] = "http://elsewhere.invalid/api/deposit/depositions/" + strconv.FormatInt(next.id, 10)
	}
	answer(w, http.StatusCreated, v)
}

func (f *fakeZenodo) publish(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	if d.published || len(d.files) == 0 || d.meta.Version == "" {
		http.Error(w, `{"message": "published already, or with no file or no version"}`, http.StatusBadRequest)
		return
	}
	d.published = true
	answer(w, http.StatusAccepted, f.view(d))
}

func (f *fakeZenodo) files(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	list := []map[string]any{}
	for _, file := range d.files {
		list = append(list, map[string]any{"id": file.id, "filename": file.name})
	}
	answer(w, http.StatusOK, list)
}

func (f *fakeZenodo) deleteFile(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	for i, file := range d.files {
		if file.id == r.PathValue("file") && !d.published {
			d.files = append(d.files[:i], d.files[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, `{"message": "no such file in a draft"}`, http.StatusNotFound)
}

func (f *fakeZenodo) upload(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	d := f.deposition(w, r)
	if d == nil {
		return
	}
	if err != nil || d.published {
		http.Error(w, `{"message": "not a draft"}`, http.StatusBadRequest)
		return
	}
	if f.broken {
		http.Error(w, `{"message": "the bucket is down"}`, http.StatusInternalServerError)
		return
	}
	if r.ContentLength < 0 {
		http.Error(w, `{"message": "an upload names its length"}`, http.StatusLengthRequired)
		return
	}
	sum := md5.Sum(data)
	if f.wrongSum {
		sum[0] ^= 0xff
	}
	name := r.PathValue("name")
	d.files = append(d.files, fakeFile{id: "file-" + name + "-" + strconv.Itoa(len(d.files)), name: name, data: data})
	answer(w, http.StatusOK, map[string]any{"key": name, "checksum": "md5:" + hex.EncodeToString(sum[:])})
}

// set changes how the stand-in behaves, under its lock.
func (f *fakeZenodo) set(change func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	change()
}

// state reads copies of the stand-in's depositions under its lock.
func (f *fakeZenodo) state() map[int64]*fakeDeposition {
	f.mu.Lock()
	defer f.mu.Unlock()
	copied := map[int64]*fakeDeposition{}
	for id, d := range f.deps {
		c := *d
		c.files = append([]fakeFile(nil), d.files...)
		copied[id] = &c
	}
	return copied
}
