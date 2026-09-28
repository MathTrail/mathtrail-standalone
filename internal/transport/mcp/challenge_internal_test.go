package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// An answer starts as a 401 with the challenge only when the challenge was
// raised before it started, and only in place of a 200: a status of another
// kind, or an answer already under way, is left as it is.
func TestAnAnswerIsChallengedOnlyWhenItStartsAfterTheChallenge(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		answer  func(w *challenging, raise func())
		status  int
		challed bool
	}{
		{"raised, then the status", func(w *challenging, raise func()) {
			raise()
			w.WriteHeader(http.StatusOK)
		}, http.StatusUnauthorized, true},
		{"raised, then the body", func(w *challenging, raise func()) {
			raise()
			_, _ = w.Write([]byte("event"))
		}, http.StatusUnauthorized, true},
		{"raised, then a flush", func(w *challenging, raise func()) {
			raise()
			w.Flush()
		}, http.StatusUnauthorized, true},
		{"never raised", func(w *challenging, _ func()) {
			_, _ = w.Write([]byte("event"))
		}, http.StatusOK, false},
		{"raised once the answer started", func(w *challenging, raise func()) {
			_, _ = w.Write([]byte("event"))
			raise()
			_, _ = w.Write([]byte("another"))
		}, http.StatusOK, false},
		{"raised before another status", func(w *challenging, raise func()) {
			raise()
			w.WriteHeader(http.StatusBadRequest)
		}, http.StatusBadRequest, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			asked := &challenge{header: `Bearer resource_metadata="https://mcp.example/meta"`}
			tc.answer(&challenging{ResponseWriter: recorder, challenge: asked}, asked.raise)
			if recorder.Code != tc.status {
				t.Errorf("status = %d, want %d", recorder.Code, tc.status)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); (got == asked.header) != tc.challed {
				t.Errorf("WWW-Authenticate = %q, want the challenge: %t", got, tc.challed)
			}
		})
	}
}
