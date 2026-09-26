// Package seal turns a value the service must protect into a string it can
// hand out, and back again.
//
// Everything the service's own key protects goes through here: the codes and
// tokens of the sign-in, the client registrations, the context of an
// authorization request, the consent cookie, and the part of a task that gives
// its answer away. One envelope, one key ring, and a closed list of purposes —
// so that mistaking one sealed value for another is not something a caller can
// do by accident.
//
// A sealed value reads mt1.<purpose>.<key id>.<base64url>. It is opaque to
// whoever holds it and carries no claims anyone else can read. What protects
// it is XChaCha20-Poly1305 under a subkey belonging to its purpose alone, with
// the purpose, the key id and whatever the caller bound the value to — a user,
// a task — authenticated beside the ciphertext. So a value opens only for the
// purpose it was sealed for and only against the binding it was sealed with:
// an access token cannot be presented as a refresh token, and one child's
// sealed answer cannot be opened as another child's.
//
// The four refusals below are what a caller branches on. None of them is ever
// relayed to whoever sent the value: an unknown key and a forged ciphertext
// are answered the same way, and the log line about either carries no
// ciphertext.
package seal

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
)

// The refusals of Open. Every one of them is checked with errors.Is.
var (
	// ErrMalformed means the string is not shaped like a sealed value at all.
	ErrMalformed = errors.New("seal: malformed value")
	// ErrPurpose means the purposes do not match: the value was sealed for
	// another one, or the caller named a purpose this package does not have.
	ErrPurpose = errors.New("seal: wrong purpose")
	// ErrUnknownKey means the key the value names is not on the ring: it was
	// sealed by a key that has since been retired, or by nobody at all.
	ErrUnknownKey = errors.New("seal: unknown key")
	// ErrIntegrity means the value did not survive its own check: a changed
	// byte, or a binding that is not the one it was sealed with.
	ErrIntegrity = errors.New("seal: integrity check failed")
)

const (
	// envelopeVersion is the first field of every sealed value. It changes
	// when the shape or the cryptography does, and never otherwise.
	envelopeVersion = "mt1"
	// envelopeSeparator is what the four fields are joined with. It is outside
	// the base64url alphabet, so no field can contain one.
	envelopeSeparator = "."
	// envelopeFields is how many fields a sealed value has: the version, the
	// purpose, the key id and the ciphertext.
	envelopeFields = 4
)

// Seal protects plaintext under the current key and returns the string to hand
// out. The binding is whatever the value belongs to — a user, a task — and
// opening it again requires exactly the same binding, in the same order.
//
// Two calls with the same arguments return different strings: every value gets
// its own random nonce.
func (r *KeyRing) Seal(purpose Purpose, plaintext []byte, binding ...string) (string, error) {
	label, known := labels[purpose]
	if !known {
		return "", fmt.Errorf("%w: %q is not a purpose", ErrPurpose, purpose)
	}

	// The nonce goes first and the ciphertext is appended to it, so the reader
	// gets them in the order Open needs. How much room that takes is left to
	// the append rather than worked out here: a capacity is a length added to
	// two constants, arithmetic on a length can overflow, and an allocation
	// sized by it is worth more care than the one allocation it saves.
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("seal: read a nonce: %w", err)
	}

	sealed := r.current.aead[purpose].Seal(nonce, nonce, plaintext,
		additionalData(purpose, r.current.id, binding))

	return strings.Join([]string{
		envelopeVersion,
		label,
		r.current.id,
		base64.RawURLEncoding.EncodeToString(sealed),
	}, envelopeSeparator), nil
}

// Open reverses Seal. It returns the plaintext only when the value was sealed
// by a key this ring still carries, for this purpose, against this binding,
// and has not been touched since.
func (r *KeyRing) Open(purpose Purpose, value string, binding ...string) ([]byte, error) {
	if _, known := labels[purpose]; !known {
		return nil, fmt.Errorf("%w: %q is not a purpose", ErrPurpose, purpose)
	}

	// One more piece than a sealed value has is enough to know it has too many,
	// and a value of a million dots is not split into a million strings first.
	fields := strings.SplitN(value, envelopeSeparator, envelopeFields+1)
	if len(fields) != envelopeFields || fields[0] != envelopeVersion || fields[2] == "" {
		return nil, ErrMalformed
	}
	label, keyID, payload := fields[1], fields[2], fields[3]

	sealedFor, ours := purposesByLabel[label]
	if !ours {
		return nil, ErrMalformed
	}
	if sealedFor != purpose {
		return nil, fmt.Errorf("%w: sealed as %s, opened as %s", ErrPurpose, sealedFor, purpose)
	}

	// Nothing read out of the value reaches the error: a key id is whatever
	// the sender wrote there, and a refusal that repeated it would carry a
	// stranger's text into our own log.
	key, onTheRing := r.key(keyID)
	if !onTheRing {
		return nil, ErrUnknownKey
	}

	// Strict, so that one sealed value has exactly one spelling: without it
	// the unused bits of the last character could be changed and the value
	// would still open, which is a signature that accepts more than it signed.
	raw, err := base64.RawURLEncoding.Strict().DecodeString(payload)
	if err != nil {
		return nil, ErrMalformed
	}
	if len(raw) < chacha20poly1305.NonceSizeX+chacha20poly1305.Overhead {
		return nil, ErrMalformed
	}

	plaintext, err := key.aead[purpose].Open(nil,
		raw[:chacha20poly1305.NonceSizeX], raw[chacha20poly1305.NonceSizeX:],
		additionalData(purpose, key.id, binding))
	if err != nil {
		return nil, ErrIntegrity
	}
	return plaintext, nil
}

// additionalData is everything a ciphertext is tied to without being hidden:
// the purpose it was sealed for, the key that sealed it, and whatever the
// caller bound it to.
//
// Each field is written with its length in front, so that no two different
// sets of fields can produce the same bytes. Without that, a value bound to
// ["ab", "c"] and one bound to ["a", "bc"] would be tied to the same thing,
// and a user id ending in a digit could be read as a task id starting with one.
func additionalData(purpose Purpose, keyID string, binding []string) []byte {
	// One byte of length in front of every field, which is what a uvarint costs
	// for anything shorter than 128 bytes — a purpose, a key id and a binding
	// are all identifiers, and none of them comes near that.
	size := len(purpose) + len(keyID) + 2
	for _, bound := range binding {
		size += len(bound) + 1
	}

	data := make([]byte, 0, size)
	data = appendField(data, string(purpose))
	data = appendField(data, keyID)
	for _, bound := range binding {
		data = appendField(data, bound)
	}
	return data
}

func appendField(data []byte, field string) []byte {
	data = binary.AppendUvarint(data, uint64(len(field)))
	return append(data, field...)
}
