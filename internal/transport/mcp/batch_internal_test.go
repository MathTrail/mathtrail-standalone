package mcpserver

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
)

// A body is a batch when its first byte past the white space JSON allows opens
// an array. That byte is left for whoever reads the body next, and a body that
// could not be read says why.
func TestABatchIsToldByItsFirstByte(t *testing.T) {
	t.Parallel()

	broken := errors.New("the body broke off")
	for _, test := range []struct {
		name  string
		body  io.Reader
		batch bool
		left  string
		err   error
	}{
		{"nothing", strings.NewReader(""), false, "", nil},
		{"white space alone", strings.NewReader(" \t\r\n"), false, "", nil},
		{"one message", strings.NewReader(`{"jsonrpc":"2.0"}`), false, `{"jsonrpc":"2.0"}`, nil},
		{"one message behind white space", strings.NewReader(" \n{}"), false, "{}", nil},
		{"a batch", strings.NewReader("[{}]"), true, "[{}]", nil},
		{"a batch behind white space", strings.NewReader("\r\n [ {} ]"), true, "[ {} ]", nil},
		{"a batch behind white space JSON does not allow", strings.NewReader("\v[{}]"), false, "\v[{}]", nil},
		{"a body that breaks off", iotest.ErrReader(broken), false, "", broken},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			body := bufio.NewReader(test.body)
			batch, err := batched(body)
			if batch != test.batch || !errors.Is(err, test.err) {
				t.Fatalf("batched() = %v, %v; want %v, %v", batch, err, test.batch, test.err)
			}
			if err != nil {
				return
			}
			if left, _ := io.ReadAll(body); string(left) != test.left {
				t.Errorf("the body reads on as %q, want %q", left, test.left)
			}
		})
	}
}

// Whatever a body holds, telling whether it is a batch reads nothing of it
// past its first byte that is not white space: what is left reads as the body
// did with that white space taken off, and it is a batch exactly when that
// opens an array.
func FuzzBatched(f *testing.F) {
	for _, seed := range []string{"", " ", "[", " [1]", "{}", "\t\n{\"a\":[1]}", "\v[", "  \r\n\"["} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, body string) {
		reader := bufio.NewReader(strings.NewReader(body))
		batch, err := batched(reader)
		if err != nil {
			t.Fatalf("batched(%q) error = %v, want nil", body, err)
		}
		want := strings.TrimLeft(body, " \t\r\n")
		if left, _ := io.ReadAll(reader); string(left) != want {
			t.Errorf("batched(%q) left %q to read, want %q", body, left, want)
		}
		if batch != strings.HasPrefix(want, "[") {
			t.Errorf("batched(%q) = %v, want %v", body, batch, strings.HasPrefix(want, "["))
		}
	})
}

// A request that is not a post of a body goes on to the protocol untouched: a
// batch is something only a posted body can be, and the protocol answers the
// rest in its own words.
func TestOnlyAPostedBodyIsLookedAt(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		request func(t *testing.T) *http.Request
	}{
		{"a GET", func(t *testing.T) *http.Request {
			return httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/mcp", strings.NewReader("["))
		}},
		{"a POST with no body", func(t *testing.T) *http.Request {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
			request.Body = nil
			return request
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := test.request(t)
			var reached *http.Request
			answer := httptest.NewRecorder()
			oneMessage(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = r
				w.WriteHeader(http.StatusTeapot)
			})).ServeHTTP(answer, request)
			if reached != request || answer.Code != http.StatusTeapot {
				t.Errorf("the request reached the protocol as %p, answered %d; want it passed on as it came, %p",
					reached, answer.Code, request)
			}
		})
	}
}
