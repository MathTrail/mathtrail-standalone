package seal_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// The purposes, in the order they appear on the wire. Every test that has to
// cover all of them reads this list, so a purpose added without a test is a
// list that no longer matches the package.
var everyPurpose = []seal.Purpose{
	seal.PurposeCode,
	seal.PurposeAccess,
	seal.PurposeRefresh,
	seal.PurposeClient,
	seal.PurposeState,
	seal.PurposeConsent,
	seal.PurposeTaskAnswer,
}

// keyNamed turns a name into a sealing key: thirty-two bytes that are
// different for every name, the same on every run, and secret to nobody.
func keyNamed(name string) string {
	digest := sha256.Sum256([]byte(name))
	return base64.StdEncoding.EncodeToString(digest[:])
}

func newRing(t *testing.T, current, previous string) *seal.KeyRing {
	t.Helper()

	ring, err := seal.NewKeyRing(current, previous)
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	return ring
}

func sealValue(t *testing.T, ring *seal.KeyRing, purpose seal.Purpose, plaintext string, binding ...string) string {
	t.Helper()

	value, err := ring.Seal(purpose, []byte(plaintext), binding...)
	if err != nil {
		t.Fatalf("Seal(%s) error = %v, want nil", purpose, err)
	}
	return value
}

// What goes in comes out — for every purpose, bound and unbound, and for a
// plaintext that is empty as readily as one that is not.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")

	cases := []struct {
		name      string
		plaintext string
		binding   []string
	}{
		{name: "no binding", plaintext: `{"user":"OX1sT9"}`},
		{name: "one binding", plaintext: `{"user":"OX1sT9"}`, binding: []string{"OX1sT9"}},
		{name: "two bindings", plaintext: `{"answer":"C"}`, binding: []string{"OX1sT9", "task-7"}},
		{name: "an empty plaintext", plaintext: ""},
		{name: "an empty binding", plaintext: "x", binding: []string{""}},
	}

	for _, purpose := range everyPurpose {
		for _, tc := range cases {
			t.Run(string(purpose)+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				value := sealValue(t, ring, purpose, tc.plaintext, tc.binding...)

				opened, err := ring.Open(purpose, value, tc.binding...)
				if err != nil {
					t.Fatalf("Open(%s) error = %v, want nil", purpose, err)
				}
				if string(opened) != tc.plaintext {
					t.Errorf("Open(%s) = %q, want %q", purpose, opened, tc.plaintext)
				}
			})
		}
	}
}

// The wire format, pinned: every purpose has one character of its own, and a
// sealed value says which key sealed it. Changing any of this changes what
// every host already holds, so it is a test rather than a convention.
func TestEveryPurposeHasItsOwnLabel(t *testing.T) {
	t.Parallel()

	wantLabel := map[seal.Purpose]string{
		seal.PurposeCode:       "c",
		seal.PurposeAccess:     "a",
		seal.PurposeRefresh:    "r",
		seal.PurposeClient:     "d",
		seal.PurposeState:      "s",
		seal.PurposeConsent:    "k",
		seal.PurposeTaskAnswer: "t",
	}

	ring := newRing(t, keyNamed("current"), "")
	seen := make(map[string]seal.Purpose, len(everyPurpose))

	for _, purpose := range everyPurpose {
		fields := strings.Split(sealValue(t, ring, purpose, "payload"), ".")
		if len(fields) != 4 {
			t.Fatalf("Seal(%s) = %d fields, want 4", purpose, len(fields))
		}
		if fields[0] != "mt1" {
			t.Errorf("Seal(%s) version = %q, want %q", purpose, fields[0], "mt1")
		}
		if fields[1] != wantLabel[purpose] {
			t.Errorf("Seal(%s) label = %q, want %q", purpose, fields[1], wantLabel[purpose])
		}
		if fields[2] != ring.CurrentKeyID() {
			t.Errorf("Seal(%s) key id = %q, want %q", purpose, fields[2], ring.CurrentKeyID())
		}
		if other, taken := seen[fields[1]]; taken {
			t.Errorf("Seal(%s) label = %q, which %s already uses", purpose, fields[1], other)
		}
		seen[fields[1]] = purpose
	}
}

// A nonce of its own for every value: the same plaintext sealed twice must not
// be recognisable as the same plaintext.
func TestTheSameValueSealsDifferentlyEveryTime(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	first := sealValue(t, ring, seal.PurposeAccess, "the same payload")
	second := sealValue(t, ring, seal.PurposeAccess, "the same payload")

	if first == second {
		t.Error("two seals of one value are identical, want a fresh nonce each time")
	}
}

// Nothing a sealed value is made of may be read out of it.
func TestASealedValueShowsNothing(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	const answer = "the correct option is C"
	value := sealValue(t, ring, seal.PurposeTaskAnswer, answer, "OX1sT9", "task-7")

	if strings.Contains(value, answer) {
		t.Errorf("the sealed value carries its plaintext: %q", value)
	}
	for _, bound := range []string{"OX1sT9", "task-7"} {
		if strings.Contains(value, bound) {
			t.Errorf("the sealed value carries what it is bound to: %q in %q", bound, value)
		}
	}
}

// Every way a value can fail to open, and the refusal each one gets. The
// refusals are what T47 onwards branch on, so they are checked by identity
// rather than by message.
func TestOpenRefuses(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	const purpose = seal.PurposeAccess
	binding := []string{"OX1sT9", "task-7"}
	value := sealValue(t, ring, purpose, `{"user":"OX1sT9"}`, binding...)

	// withField rebuilds a sealed value with one of its four fields replaced.
	withField := func(index int, replacement string) string {
		fields := strings.Split(value, ".")
		fields[index] = replacement
		return strings.Join(fields, ".")
	}
	// withChangedByte flips one character of the ciphertext.
	withChangedByte := func() string {
		fields := strings.Split(value, ".")
		payload := []byte(fields[3])
		if payload[0] == 'A' {
			payload[0] = 'B'
		} else {
			payload[0] = 'A'
		}
		fields[3] = string(payload)
		return strings.Join(fields, ".")
	}

	cases := []struct {
		name    string
		value   string
		binding []string
		want    error
	}{
		{name: "nothing at all", value: "", binding: binding, want: seal.ErrMalformed},
		{name: "not a sealed value", value: "hello", binding: binding, want: seal.ErrMalformed},
		{name: "too few fields", value: "mt1.a." + ring.CurrentKeyID(), binding: binding, want: seal.ErrMalformed},
		{name: "too many fields", value: value + ".extra", binding: binding, want: seal.ErrMalformed},
		{name: "another version", value: withField(0, "mt2"), binding: binding, want: seal.ErrMalformed},
		{name: "a label nobody uses", value: withField(1, "z"), binding: binding, want: seal.ErrMalformed},
		{name: "an empty label", value: withField(1, ""), binding: binding, want: seal.ErrMalformed},
		{name: "the payload is not base64", value: withField(3, "not base64!"), binding: binding, want: seal.ErrMalformed},
		{name: "the payload is padded", value: withField(3, "AAAA="), binding: binding, want: seal.ErrMalformed},
		{name: "the payload is too short", value: withField(3, "AAAAAAAAAAAA"), binding: binding, want: seal.ErrMalformed},
		{name: "an empty payload", value: withField(3, ""), binding: binding, want: seal.ErrMalformed},

		{name: "sealed for another purpose", value: sealValue(t, ring, seal.PurposeRefresh, "x", binding...), binding: binding, want: seal.ErrPurpose},

		{name: "a key nobody has", value: withField(2, "aaaaaa"), binding: binding, want: seal.ErrUnknownKey},
		{name: "an empty key id", value: withField(2, ""), binding: binding, want: seal.ErrUnknownKey},

		{name: "a changed byte", value: withChangedByte(), binding: binding, want: seal.ErrIntegrity},
		{name: "another binding", value: value, binding: []string{"OX1sT9", "task-8"}, want: seal.ErrIntegrity},
		{name: "another user's binding", value: value, binding: []string{"Zk4pQ2", "task-7"}, want: seal.ErrIntegrity},
		{name: "no binding at all", value: value, binding: nil, want: seal.ErrIntegrity},
		{name: "one binding too many", value: value, binding: []string{"OX1sT9", "task-7", "extra"}, want: seal.ErrIntegrity},
		{name: "the binding in the other order", value: value, binding: []string{"task-7", "OX1sT9"}, want: seal.ErrIntegrity},
		{name: "the binding split elsewhere", value: sealValue(t, ring, purpose, "x", "ab", "c"), binding: []string{"a", "bc"}, want: seal.ErrIntegrity},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opened, err := ring.Open(purpose, tc.value, tc.binding...)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Open() error = %v, want %v", err, tc.want)
			}
			if opened != nil {
				t.Errorf("Open() = %q with an error, want nothing", opened)
			}
		})
	}
}

// The base64 of a sealed value has one spelling. Without that, the unused bits
// of its last character could be rewritten and it would still open: a value
// that accepts more than it was sealed as.
func TestTheLastCharacterCannotBeRewritten(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	value := sealValue(t, ring, seal.PurposeAccess, "a payload of some length")

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	original := value[len(value)-1]

	for _, replacement := range []byte(alphabet) {
		if replacement == original {
			continue
		}
		rewritten := value[:len(value)-1] + string(replacement)
		if _, err := ring.Open(seal.PurposeAccess, rewritten); err == nil {
			t.Errorf("Open() opened a value ending %q instead of %q", replacement, original)
		}
	}
}

// A rotation is invisible to whoever holds a sealed value: the previous key
// still opens what it sealed, and everything new is sealed with the current
// one. Once the previous key falls off the ring, what it sealed is gone.
func TestRotation(t *testing.T) {
	t.Parallel()

	const plaintext = `{"user":"OX1sT9"}`
	before := newRing(t, keyNamed("first"), "")
	old := sealValue(t, before, seal.PurposeRefresh, plaintext, "OX1sT9")

	during := newRing(t, keyNamed("second"), keyNamed("first"))
	opened, err := during.Open(seal.PurposeRefresh, old, "OX1sT9")
	if err != nil {
		t.Fatalf("Open() error = %v, want the previous key to open what it sealed", err)
	}
	if string(opened) != plaintext {
		t.Errorf("Open() = %q, want %q", opened, plaintext)
	}

	fresh := sealValue(t, during, seal.PurposeRefresh, plaintext, "OX1sT9")
	if keyID := strings.Split(fresh, ".")[2]; keyID != during.CurrentKeyID() {
		t.Errorf("a fresh value names key %q, want the current key %q", keyID, during.CurrentKeyID())
	}
	if during.PreviousKeyID() != before.CurrentKeyID() {
		t.Errorf("PreviousKeyID() = %q, want %q", during.PreviousKeyID(), before.CurrentKeyID())
	}

	// The key that fell off the end: what it sealed is refused like anything
	// else nobody has a key for.
	after := newRing(t, keyNamed("second"), "")
	if after.PreviousKeyID() != "" {
		t.Errorf("PreviousKeyID() = %q, want empty on a ring with one key", after.PreviousKeyID())
	}
	if _, err := after.Open(seal.PurposeRefresh, old, "OX1sT9"); !errors.Is(err, seal.ErrUnknownKey) {
		t.Errorf("Open() error = %v, want %v", err, seal.ErrUnknownKey)
	}

	// And the old ring cannot read what the new key sealed.
	if _, err := before.Open(seal.PurposeRefresh, fresh, "OX1sT9"); !errors.Is(err, seal.ErrUnknownKey) {
		t.Errorf("Open() error = %v, want %v", err, seal.ErrUnknownKey)
	}
}

// The identifier comes from the key, so the two cannot be configured apart.
func TestKeyIDComesFromTheKey(t *testing.T) {
	t.Parallel()

	first, err := seal.ParseKey(keyNamed("first"))
	if err != nil {
		t.Fatalf("ParseKey() error = %v, want nil", err)
	}
	again, err := seal.ParseKey(keyNamed("first"))
	if err != nil {
		t.Fatalf("ParseKey() error = %v, want nil", err)
	}
	second, err := seal.ParseKey(keyNamed("second"))
	if err != nil {
		t.Fatalf("ParseKey() error = %v, want nil", err)
	}

	if first.ID() != again.ID() {
		t.Errorf("one key has two identifiers: %q and %q", first.ID(), again.ID())
	}
	if first.ID() == second.ID() {
		t.Errorf("two keys share the identifier %q", first.ID())
	}

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	if len(first.ID()) != 6 {
		t.Errorf("ID() = %q, want six characters", first.ID())
	}
	if strings.ContainsFunc(first.ID(), func(r rune) bool { return !strings.ContainsRune(alphabet, r) }) {
		t.Errorf("ID() = %q, want base64url characters only", first.ID())
	}

	// Whitespace around a secret read out of a file is not part of the key.
	padded, err := seal.ParseKey("  " + keyNamed("first") + "\n")
	if err != nil {
		t.Fatalf("ParseKey() error = %v, want the surrounding whitespace to be ignored", err)
	}
	if padded.ID() != first.ID() {
		t.Errorf("ID() = %q, want %q", padded.ID(), first.ID())
	}
}

// A key the service could not seal with stops it at the door, whichever of
// the two keys it was handed as.
func TestKeysThatAreNotKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		encoded string
	}{
		{name: "not base64", encoded: "not base64 at all!"},
		{name: "base64 of too few bytes", encoded: base64.StdEncoding.EncodeToString(make([]byte, seal.KeySize-1))},
		{name: "base64 of too many bytes", encoded: base64.StdEncoding.EncodeToString(make([]byte, seal.KeySize+1))},
		{
			name: "the url alphabet instead of the standard one",
			encoded: strings.ReplaceAll(
				base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xff}, seal.KeySize)), "/", "_"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := seal.ParseKey(tc.encoded); !errors.Is(err, seal.ErrKey) {
				t.Errorf("ParseKey() error = %v, want %v", err, seal.ErrKey)
			}
			if _, err := seal.NewKeyRing(tc.encoded, ""); !errors.Is(err, seal.ErrKey) {
				t.Errorf("NewKeyRing(current) error = %v, want %v", err, seal.ErrKey)
			}
			if _, err := seal.NewKeyRing(keyNamed("current"), tc.encoded); !errors.Is(err, seal.ErrKey) {
				t.Errorf("NewKeyRing(previous) error = %v, want %v", err, seal.ErrKey)
			}
		})
	}
}

// Outside a rotation there is no previous key, and a ring says so rather than
// refusing to be built. A current key, though, is not optional.
func TestAPreviousKeyIsOptional(t *testing.T) {
	t.Parallel()

	for _, previous := range []string{"", "   ", "\n"} {
		ring, err := seal.NewKeyRing(keyNamed("current"), previous)
		if err != nil {
			t.Fatalf("NewKeyRing(%q) error = %v, want nil", previous, err)
		}
		if ring.PreviousKeyID() != "" {
			t.Errorf("PreviousKeyID() = %q, want empty", ring.PreviousKeyID())
		}
	}

	for _, current := range []string{"", "   "} {
		if _, err := seal.NewKeyRing(current, ""); !errors.Is(err, seal.ErrKey) {
			t.Errorf("NewKeyRing(%q) error = %v, want %v", current, err, seal.ErrKey)
		}
	}
}

// No refusal, anywhere, may repeat the key it was handed.
func TestARefusalNeverCarriesTheKey(t *testing.T) {
	t.Parallel()

	tooLong := base64.StdEncoding.EncodeToString([]byte("a secret nobody may read back out of an error"))

	_, err := seal.ParseKey(tooLong)
	if err == nil {
		t.Fatal("ParseKey() error = nil, want a refusal")
	}
	if strings.Contains(err.Error(), tooLong) || strings.Contains(err.Error(), "secret") {
		t.Errorf("ParseKey() error = %q, want it to carry nothing of the key", err)
	}
}

// A purpose the package does not know is a mistake in the caller, not a value
// to guess at.
func TestAnUnknownPurposeIsRefused(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	const invented = seal.Purpose("session")

	if _, err := ring.Seal(invented, []byte("x")); !errors.Is(err, seal.ErrPurpose) {
		t.Errorf("Seal() error = %v, want %v", err, seal.ErrPurpose)
	}
	if _, err := ring.Open(invented, sealValue(t, ring, seal.PurposeAccess, "x")); !errors.Is(err, seal.ErrPurpose) {
		t.Errorf("Open() error = %v, want %v", err, seal.ErrPurpose)
	}
}

// A sealed value arrives from whoever sent it, and it is the one input to this
// package that is entirely theirs to write. Whatever it says, opening it ends
// in one of the four refusals — never in a panic, and never in a plaintext
// that was not sealed.
func FuzzOpen(f *testing.F) {
	ring, err := seal.NewKeyRing(keyNamed("current"), keyNamed("previous"))
	if err != nil {
		f.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	const plaintext = `{"user":"OX1sT9"}`
	value, err := ring.Seal(seal.PurposeAccess, []byte(plaintext), "OX1sT9")
	if err != nil {
		f.Fatalf("Seal() error = %v, want nil", err)
	}

	f.Add(value)
	f.Add(value[:len(value)-1] + "A")
	f.Add("")
	f.Add("mt1")
	f.Add("mt1.a.aaaaaa.")
	f.Add("mt1.a." + ring.CurrentKeyID() + ".AAAA")
	f.Add("mt1..." + strings.Repeat("A", 100))
	f.Add("....")
	f.Add("mt1.a.\xff\xfe.AAAA")

	f.Fuzz(func(t *testing.T, candidate string) {
		opened, err := ring.Open(seal.PurposeAccess, candidate, "OX1sT9")
		if err == nil {
			// The only string that opens is the one this ring sealed, and it
			// opens to exactly what went in.
			if string(opened) != plaintext {
				t.Errorf("Open(%q) = %q, want %q", candidate, opened, plaintext)
			}
			return
		}
		switch {
		case errors.Is(err, seal.ErrMalformed), errors.Is(err, seal.ErrPurpose),
			errors.Is(err, seal.ErrUnknownKey), errors.Is(err, seal.ErrIntegrity):
		default:
			t.Errorf("Open(%q) error = %v, want one a caller can branch on", candidate, err)
		}
		if opened != nil {
			t.Errorf("Open(%q) = %q alongside an error, want nothing", candidate, opened)
		}
	})
}
