package seal

import (
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// KeySize is how many bytes a sealing key has.
	KeySize = chacha20poly1305.KeySize

	// keyIDLength is how much of a key's digest names it. Six base64url
	// characters are 36 bits: far too few to say anything about the key, and
	// far more than enough to tell two live keys apart.
	keyIDLength = 6

	// subkeyInfo prefixes the context string a purpose's subkey is derived
	// from. The version is part of it, so a later scheme derives different
	// subkeys from the same key rather than quietly reusing these.
	subkeyInfo = "mathtrail/v1/"
)

// ErrKey is returned when a value handed in as a sealing key is not one.
// Neither it nor anything wrapping it ever carries the key itself.
var ErrKey = errors.New("seal: invalid key")

// ErrCurrentKey and ErrPreviousKey say which of the two keys of a ring a
// refusal of NewKeyRing is about. They come wrapped in ErrKey, beside it.
var (
	ErrCurrentKey  = errors.New("the current key")
	ErrPreviousKey = errors.New("the previous key")
)

// sealingKey is one version of the sealing key. The bytes it was read from are
// never kept: what it holds is the identifier derived from them and one AEAD
// per purpose, so the key material itself cannot be read back out, logged or
// passed on. The identifier is the first characters of the key's own digest,
// which is why a key and its id can never be configured apart; it is public —
// every sealed value carries it — and says nothing about the key.
type sealingKey struct {
	id   string
	aead map[Purpose]cipher.AEAD
}

// readKey reads a key from the base64 a deployment carries, and derives
// everything that is used in its place. The encoding is standard base64 of
// exactly KeySize random bytes; surrounding whitespace is ignored, because a
// secret read out of a file arrives with a newline on the end. What went wrong
// is said without the name of the refusal, which the caller puts first.
func readKey(encoded string) (sealingKey, error) {
	secret, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return sealingKey{}, errors.New("not base64")
	}
	if len(secret) != KeySize {
		return sealingKey{}, fmt.Errorf("%d bytes, want %d", len(secret), KeySize)
	}

	digest := sha256.Sum256(secret)
	key := sealingKey{
		id:   base64.RawURLEncoding.EncodeToString(digest[:])[:keyIDLength],
		aead: make(map[Purpose]cipher.AEAD, len(labels)),
	}

	// Every subkey is derived once, when the process starts. A purpose that
	// could not be derived is a key that would fail at the worst possible
	// moment instead, so all of them are made here or none is.
	for purpose := range labels {
		subkey, err := hkdf.Key(sha256.New, secret, nil, subkeyInfo+string(purpose), KeySize)
		if err != nil {
			return sealingKey{}, fmt.Errorf("derive the %s subkey: %w", purpose, err)
		}
		aead, err := chacha20poly1305.NewX(subkey)
		if err != nil {
			return sealingKey{}, fmt.Errorf("%s: %w", purpose, err)
		}
		key.aead[purpose] = aead
	}
	return key, nil
}

// KeyRing is the set of keys a running service seals and opens with: the
// current one, which does both, and — while a rotation is under way — the
// previous one, which only opens. A value sealed before the rotation therefore
// keeps working until it expires on its own, and everything issued from now on
// is sealed with the current key.
//
// A ring is made by NewKeyRing. The zero KeyRing holds no key: it opens
// nothing, and sealing with it is a mistake of the caller's.
type KeyRing struct {
	current  sealingKey
	previous *sealingKey
}

// NewKeyRing builds the ring from the two encoded keys a deployment carries.
// A previous key of nothing but whitespace means there is none, which is the
// shape outside a rotation. The same key twice is refused: it is a rotation
// that only looks done, and what the retired key sealed would stop opening the
// moment it was deployed.
func NewKeyRing(current, previous string) (*KeyRing, error) {
	currentKey, err := readKey(current)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrKey, ErrCurrentKey, err)
	}

	ring := &KeyRing{current: currentKey}
	if strings.TrimSpace(previous) == "" {
		return ring, nil
	}

	previousKey, err := readKey(previous)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %w", ErrKey, ErrPreviousKey, err)
	}
	if previousKey.id == currentKey.id {
		return nil, fmt.Errorf("%w: %w is the current one, and a rotation needs two", ErrKey, ErrPreviousKey)
	}
	ring.previous = &previousKey
	return ring, nil
}

// CurrentKeyID is the identifier of the key everything is sealed with now.
func (r *KeyRing) CurrentKeyID() string { return r.current.id }

// PreviousKeyID is the identifier of the key kept from before a rotation, and
// empty when the ring carries only one key.
func (r *KeyRing) PreviousKeyID() string {
	if r.previous == nil {
		return ""
	}
	return r.previous.id
}

// key finds the key a sealed value names, and reports whether the ring still
// carries it.
func (r *KeyRing) key(id string) (sealingKey, bool) {
	switch {
	case id == r.current.id:
		return r.current, true
	case r.previous != nil && id == r.previous.id:
		return *r.previous, true
	default:
		return sealingKey{}, false
	}
}

// PurposeRing is the ring with one purpose already chosen. A caller that only
// ever protects one kind of value takes one of these and stops carrying the
// purpose to every call — and stops being able to pass the wrong one.
type PurposeRing struct {
	ring    *KeyRing
	purpose Purpose
}

// For binds a purpose to the ring.
func (r *KeyRing) For(purpose Purpose) PurposeRing {
	return PurposeRing{ring: r, purpose: purpose}
}

// Seal protects a value under the chosen purpose.
func (p PurposeRing) Seal(plaintext []byte, binding ...string) (string, error) {
	return p.ring.Seal(p.purpose, plaintext, binding...)
}

// Open reads one back.
func (p PurposeRing) Open(value string, binding ...string) ([]byte, error) {
	return p.ring.Open(p.purpose, value, binding...)
}
