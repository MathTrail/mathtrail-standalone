package mcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// oneMessage refuses a request whose body is a batch — an array of messages —
// before the protocol's library reads it. The protocol this endpoint speaks
// carries one message a request, and the library still takes a batch from a
// request that names no version of the protocol: every message of it is then a
// call of its own, held to the paces one by one, logged and answered, and a
// megabyte of them fits in one request, which the paces never see as more than
// one. Only the first byte that is not white space is looked at; the white
// space before it is dropped, as the grammar of the message allows, and nothing
// else of the body is read here.
func oneMessage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}
		// Bounded as the library bounds the body, and counting the white space
		// dropped here with the rest, so that a body past the bound is refused
		// as too large whatever it opens with.
		body := bufio.NewReader(http.MaxBytesReader(w, r.Body, maxRequestBody))
		batch, err := batched(body)
		var tooLarge *http.MaxBytesError
		switch {
		case errors.As(err, &tooLarge):
			http.Error(w, fmt.Sprintf("request body exceeds %d bytes", tooLarge.Limit), http.StatusRequestEntityTooLarge)
			return
		case batch:
			http.Error(w, "one message a request: a batch of messages is not accepted", http.StatusBadRequest)
			return
		}
		r.Body = readCloser{Reader: body, Closer: r.Body}
		next.ServeHTTP(w, r)
	})
}

// batched says whether the body, past its white space, opens an array. The
// byte it stops at is left for whoever reads the body next. A body that ends
// in white space alone is no batch; one that could not be read says why.
func batched(body *bufio.Reader) (bool, error) {
	for {
		next, err := body.ReadByte()
		switch {
		case errors.Is(err, io.EOF):
			return false, nil
		case err != nil:
			return false, err
		}
		switch next {
		case ' ', '\t', '\n', '\r':
			continue
		}
		_ = body.UnreadByte()
		return next == '[', nil
	}
}

// readCloser is a body read through a buffer and closed where it came from.
type readCloser struct {
	io.Reader
	io.Closer
}
