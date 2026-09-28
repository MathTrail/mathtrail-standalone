package drivestore

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// revision is a state of the profile's file, as this store tells states
// apart: which file, the profile's own number in it, and a digest of its
// bytes.
//
// The number lets a save hold a profile to "one past the state it was read
// at" without reading the file again. The digest is what the file is compared
// with before a save: two reads of the same bytes are the same state, and a
// change of any byte is another — Drive's own revision of a file cannot be
// read in the call that reads the bytes, and this can.
type revision struct {
	file    string
	counter int
	digest  string
}

// separator divides the parts of a revision. Neither the number nor the
// digest ever holds one, so a file's ID may hold anything.
const separator = "/"

// revisionOf is the revision of these bytes, read from or written to this
// file, holding a profile of this number.
func revisionOf(file string, counter int, raw []byte) store.Revision {
	return store.Revision(file + separator + strconv.Itoa(counter) + separator + digestOf(raw))
}

// parseRevision reads a revision this store handed out, and reports whether
// it was one.
func parseRevision(r store.Revision) (revision, bool) {
	head, digest, cut := cutLast(string(r))
	if !cut {
		return revision{}, false
	}
	file, number, cut := cutLast(head)
	if !cut {
		return revision{}, false
	}
	counter, err := strconv.Atoi(number)
	if err != nil || counter < 0 || file == "" || digest == "" {
		return revision{}, false
	}
	return revision{file: file, counter: counter, digest: digest}, true
}

// cutLast splits a text at its last separator.
func cutLast(text string) (before, after string, cut bool) {
	at := strings.LastIndex(text, separator)
	if at < 0 {
		return "", "", false
	}
	return text[:at], text[at+len(separator):], true
}

// digestOf is the digest of a file's bytes.
func digestOf(raw []byte) string {
	sum := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
