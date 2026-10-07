package account

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// flushing is the connection to a browser sent back, which notes how much of
// the page had gone out to the browser while the sign-in had yet to hear what
// the browser brought.
type flushing struct {
	*httptest.ResponseRecorder
	back    chan cameBack
	unheard int
}

func (f *flushing) Flush() {
	if len(f.back) == 0 {
		f.unheard = f.Body.Len()
	}
	f.ResponseRecorder.Flush()
}

// The browser sent back has its whole page before the sign-in hears what it
// brought, a code or a refusal alike: once the sign-in has heard, it may be
// over, and close the way back, before a page written after that reaches the
// browser.
func TestThePageGoesOutWholeBeforeTheSignInHears(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, query string }{
		{"a code", "code=the-code&state=sent"},
		{"a refusal", "error=access_denied&state=sent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := &way{back: make(chan cameBack, 1), sent: true, state: "sent"}
			browser := &flushing{ResponseRecorder: httptest.NewRecorder(), back: w.back}
			w.serve(browser, httptest.NewRequestWithContext(t.Context(), http.MethodGet, callbackPath+"?"+tc.query, http.NoBody))

			if len(w.back) != 1 {
				t.Fatalf("the sign-in heard %d times, want once", len(w.back))
			}
			page := browser.Body.Len()
			if page == 0 || browser.unheard != page {
				t.Errorf("%d bytes of a %d-byte page went out before the sign-in heard, want all of them", browser.unheard, page)
			}
			if length := browser.Header().Get("Content-Length"); length != strconv.Itoa(page) {
				t.Errorf("the page's Content-Length = %q, want %d, so that the page is whole without the end of the reply", length, page)
			}
		})
	}
}
