package drive

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// answered is a body that remembers whether it was read to its end and
// closed: what a connection needs of an answer before it can carry the next
// call.
type answered struct {
	io.Reader
	ended, closed bool
}

func (a *answered) Read(p []byte) (int, error) {
	n, err := a.Reader.Read(p)
	if errors.Is(err, io.EOF) {
		a.ended = true
	}
	return n, err
}

func (a *answered) Close() error {
	a.closed = true
	return nil
}

// answering is a transport that answers every request with one answer.
type answering struct {
	status int
	body   *answered
}

func (a answering) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: a.status, Header: http.Header{}, Body: a.body, Request: r}, nil
}

// Every answer is read to its end and closed — a file's, a refusal, and one
// whose JSON ends before the answer does — so that the connection it came on
// can carry the next call instead of a new one paying for its handshake.
func TestEveryAnswerIsReadToItsEnd(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		status int
		body   string
	}{
		"a file":                              {http.StatusOK, `{"id": "a-file", "name": "a name"}` + "\n\n"},
		"a refusal":                           {http.StatusForbidden, `{"error": {"errors": [{"reason": "x"}]}}` + "\n\n"},
		"a refusal longer than is read of it": {http.StatusForbidden, strings.Repeat(" ", maxRefusal+10)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			body := &answered{Reader: strings.NewReader(tc.body)}
			calls := &files{client: &http.Client{Transport: answering{status: tc.status, body: body}}, root: Google, timeout: time.Second}
			_, _ = calls.Get(t.Context(), "a-token", "a-file")
			if !body.ended || !body.closed {
				t.Errorf("the answer was read to its end %t and closed %t, want both", body.ended, body.closed)
			}
		})
	}
}
