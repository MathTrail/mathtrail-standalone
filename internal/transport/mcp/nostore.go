package mcpserver

import "net/http"

// neverStored makes every answer of the endpoint forbid being stored anywhere
// on its way: an answer carries a child's profile, tasks and progress.
//
// The protocol library sets a caching rule of its own on every answer —
// revalidate before reuse, and transform nothing — and the first half still
// lets a cache keep a copy. Its rule is replaced as the headers leave, which is
// the one moment after the library has set it. Transforming nothing is kept:
// it is what stops a proxy from holding back a streamed answer to compress it.
func neverStored(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&noStore{ResponseWriter: w}, r)
	})
}

// noStore is a response writer that forbids storing the answer as it starts
// to send it.
type noStore struct {
	http.ResponseWriter
	started bool
}

// WriteHeader sends the status, with the rule in place.
func (w *noStore) WriteHeader(status int) {
	w.forbidStoring()
	w.ResponseWriter.WriteHeader(status)
}

// Write sends part of the body, with the rule in place if nothing has been sent
// yet.
func (w *noStore) Write(body []byte) (int, error) {
	w.forbidStoring()
	return w.ResponseWriter.Write(body)
}

// Flush sends what has been written so far, with the rule in place: a flush
// before anything was written sends the headers too.
func (w *noStore) Flush() {
	w.forbidStoring()
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

// Unwrap is the writer underneath, for anything that looks past this one.
func (w *noStore) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// forbidStoring sets the rule once, before the headers leave.
func (w *noStore) forbidStoring() {
	if w.started {
		return
	}
	w.started = true
	w.Header().Set("Cache-Control", "no-store, no-transform")
}
