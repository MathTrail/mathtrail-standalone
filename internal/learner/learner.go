// Package learner derives the name a child's activity is counted under in the
// lines the service writes: one name for a child through a calendar month,
// which leads back to no child and which the next month's name cannot be told
// from. With it the children active on a day, a week or a month can be counted
// for as long as those lines are kept, and nothing that outlives them has to
// know who any child is.
//
// It counts children; the bench of simulated children that shares its stem is
// another thing.
package learner

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// KeySize is how many random bytes the secret has.
const KeySize = 32

const (
	// monthInfo prefixes what a month's key is derived from. The version is
	// part of it, so a later scheme derives other keys from the same secret
	// rather than quietly reusing these.
	monthInfo = "mathtrail/v1/learner/"
	// monthLayout writes a month the way its key is derived from it.
	monthLayout = "2006-01"
	// studentInfo prefixes the child's identifier inside the digest, so that
	// nothing else ever derived under a month's key can come out as a name.
	studentInfo = "student:"
	// nameLength is how many base64url characters a name keeps: 96 bits, far
	// past any number of children a month could count.
	nameLength = 16
)

// ErrKey is returned when a value handed in as the secret is not one. Neither
// it nor anything wrapping it ever carries the secret.
var ErrKey = errors.New("learner: invalid key")

// Key is the secret the names are derived from. It is a secret of its own,
// apart from the keys the service seals with: those are rotated, and a name
// that changed with them would count a child twice in the month of a rotation.
//
// A month's key is a keyed digest of the month under the secret, and a name a
// keyed digest of the child's identifier under the month's key. Whoever holds
// the names of two months holds two unrelated strings; whoever holds the
// secret and a child's identifier can count that child again, and nobody else
// can.
type Key struct {
	secret []byte
}

// NewKey reads the secret from the standard base64 of exactly KeySize random
// bytes a deployment carries. Whitespace around it is ignored, since a secret
// read out of a file arrives with a newline on the end. What went wrong is
// said without the value.
func NewKey(encoded string) (*Key, error) {
	secret, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("%w: not base64", ErrKey)
	}
	if len(secret) != KeySize {
		return nil, fmt.Errorf("%w: %d bytes, want %d", ErrKey, len(secret), KeySize)
	}
	return &Key{secret: secret}, nil
}

// RandomKey is a key of its own for a process that was given none. The names
// it derives mean something only while the process runs, which is all a
// developer's machine needs of them.
func RandomKey() *Key {
	secret := make([]byte, KeySize)
	// The source of randomness never fails short of the program itself
	// failing, so there is no error to handle.
	_, _ = rand.Read(secret)
	return &Key{secret: secret}
}

// Of is the name the child with this identifier is counted under in the
// calendar month, in UTC, that the moment falls in. The same child gets the
// same name all month, whoever signs in for it and however often, and another
// name from the first second of the next month.
func (k *Key) Of(studentID string, at time.Time) string {
	month := digest(k.secret, monthInfo+at.UTC().Format(monthLayout))
	name := digest(month, studentInfo+studentID)
	return base64.RawURLEncoding.EncodeToString(name)[:nameLength]
}

// digest is the keyed digest of a text under a key.
func digest(key []byte, text string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(text))
	return mac.Sum(nil)
}
