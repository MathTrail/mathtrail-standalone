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

// The purposes, in the order they appear on the wire, and the label each one
// travels under. Both lists are written out by hand: they are what the tests
// know, held against what the package does rather than read out of it.
//
// TestTheTestsKnowEveryPurpose is what keeps them honest — nothing else here
// would notice a purpose that was added to the package and not to these lines.
var everyPurpose = []seal.Purpose{
	seal.PurposeCode,
	seal.PurposeAccess,
	seal.PurposeRefresh,
	seal.PurposeClient,
	seal.PurposeState,
	seal.PurposeConsent,
	seal.PurposeTaskAnswer,
}

var labelOf = map[seal.Purpose]string{
	seal.PurposeCode:       "c",
	seal.PurposeAccess:     "a",
	seal.PurposeRefresh:    "r",
	seal.PurposeClient:     "d",
	seal.PurposeState:      "s",
	seal.PurposeConsent:    "k",
	seal.PurposeTaskAnswer: "t",
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
		if fields[1] != labelOf[purpose] {
			t.Errorf("Seal(%s) label = %q, want %q", purpose, fields[1], labelOf[purpose])
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

// The closed list, kept closed. Every test that covers all the purposes reads
// everyPurpose or labelOf, so a purpose added to the package and not to those
// lines would quietly lose its round trip, its label and its properties —
// and nothing would say so.
//
// What the package actually accepts is therefore read back out of it, one
// label at a time. The order of the checks in Open is what makes that
// readable: a label nobody uses is refused as malformed, and a label that is
// in use gets as far as the purpose or the key. This covers labels of one
// character, which is every label the format has room for.
func TestTheTestsKnowEveryPurpose(t *testing.T) {
	t.Parallel()

	// A key id the ring does not carry, so that a label in use is refused for
	// the key or the purpose rather than for the shape of the value.
	const absentKey = "aaaaaa"

	ring := newRing(t, keyNamed("current"), "")
	if ring.CurrentKeyID() == absentKey {
		t.Fatalf("the test key is named %q, which the probe assumes nothing is", absentKey)
	}

	inUse := make(map[string]struct{})
	for symbol := byte('!'); symbol <= '~'; symbol++ {
		label := string(symbol)
		if label == "." {
			continue // the separator: no field of a sealed value can hold one
		}
		_, err := ring.Open(seal.PurposeAccess, "mt1."+label+"."+absentKey+".AAAA")
		if !errors.Is(err, seal.ErrMalformed) {
			inUse[label] = struct{}{}
		}
	}

	for purpose, label := range labelOf {
		if _, known := inUse[label]; !known {
			t.Errorf("the tests give %s the label %q, which the package does not know", purpose, label)
		}
		delete(inUse, label)
	}
	for label := range inUse {
		t.Errorf("the package knows the label %q, and no test names a purpose for it", label)
	}

	// And the two lists the tests keep must name the same purposes as each other.
	if len(everyPurpose) != len(labelOf) {
		t.Errorf("everyPurpose names %d purposes and labelOf names %d, want the same set",
			len(everyPurpose), len(labelOf))
	}
	for _, purpose := range everyPurpose {
		if _, named := labelOf[purpose]; !named {
			t.Errorf("everyPurpose names %s, and labelOf does not", purpose)
		}
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

// Every way a value can fail to open, and the refusal each one gets. Callers
// branch on the refusals, so they are checked by identity rather than by
// message.
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
		{name: "an empty key id", value: withField(2, ""), binding: binding, want: seal.ErrMalformed},

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

	first, err := seal.NewKeyRing(keyNamed("first"), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	again, err := seal.NewKeyRing(keyNamed("first"), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	second, err := seal.NewKeyRing(keyNamed("second"), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}

	if first.CurrentKeyID() != again.CurrentKeyID() {
		t.Errorf("one key has two identifiers: %q and %q", first.CurrentKeyID(), again.CurrentKeyID())
	}
	if first.CurrentKeyID() == second.CurrentKeyID() {
		t.Errorf("two keys share the identifier %q", first.CurrentKeyID())
	}

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	if len(first.CurrentKeyID()) != 6 {
		t.Errorf("CurrentKeyID() = %q, want six characters", first.CurrentKeyID())
	}
	if strings.ContainsFunc(first.CurrentKeyID(), func(r rune) bool { return !strings.ContainsRune(alphabet, r) }) {
		t.Errorf("CurrentKeyID() = %q, want base64url characters only", first.CurrentKeyID())
	}

	// Whitespace around a secret read out of a file is not part of the key.
	padded, err := seal.NewKeyRing("  "+keyNamed("first")+"\n", "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want the surrounding whitespace to be ignored", err)
	}
	if padded.CurrentKeyID() != first.CurrentKeyID() {
		t.Errorf("CurrentKeyID() = %q, want %q", padded.CurrentKeyID(), first.CurrentKeyID())
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

// A ring of one key twice is a rotation that only looks done, and it is
// refused; a key that is not one is refused with the name of the one it was,
// so that whoever reads the refusal knows which of the two secrets to fix.
func TestARingIsTwoKeysOrOne(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, current, previous string
		which                   error
	}{
		{"a current key that is not one", "half a key", "", seal.ErrCurrentKey},
		{"a previous key that is not one", keyNamed("current"), "half a key", seal.ErrPreviousKey},
		{"one key twice", keyNamed("current"), keyNamed("current"), seal.ErrPreviousKey},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := seal.NewKeyRing(test.current, test.previous); !errors.Is(err, seal.ErrKey) || !errors.Is(err, test.which) {
				t.Errorf("NewKeyRing() error = %v, want %v about %v", err, seal.ErrKey, test.which)
			}
		})
	}
}

// The zero ring holds no key, and a value that names no key — the one thing
// a zero ring could mistake for its own — is malformed rather than opened.
func TestAZeroRingOpensNothing(t *testing.T) {
	t.Parallel()

	value := sealValue(t, newRing(t, keyNamed("current"), ""), seal.PurposeAccess, "x")
	fields := strings.Split(value, ".")
	fields[2] = ""

	var zero seal.KeyRing
	if _, err := zero.Open(seal.PurposeAccess, strings.Join(fields, ".")); !errors.Is(err, seal.ErrMalformed) {
		t.Errorf("Open(a value naming no key) error = %v, want %v", err, seal.ErrMalformed)
	}
	if _, err := zero.Open(seal.PurposeAccess, value); !errors.Is(err, seal.ErrUnknownKey) {
		t.Errorf("Open(a value of another ring) error = %v, want %v", err, seal.ErrUnknownKey)
	}
}

// No refusal, anywhere, may repeat the key it was handed.
func TestARefusalNeverCarriesTheKey(t *testing.T) {
	t.Parallel()

	tooLong := base64.StdEncoding.EncodeToString([]byte("a secret nobody may read back out of an error"))

	_, err := seal.NewKeyRing(tooLong, "")
	if err == nil {
		t.Fatal("NewKeyRing() error = nil, want a refusal")
	}
	if strings.Contains(err.Error(), tooLong) || strings.Contains(err.Error(), "secret") {
		t.Errorf("NewKeyRing() error = %q, want it to carry nothing of the key", err)
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

// A ring with its purpose already chosen is the same ring: what it seals, the
// ring opens under that purpose, and what another purpose sealed it refuses.
func TestARingWithOnePurpose(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	answers := ring.For(seal.PurposeTaskAnswer)

	value, err := answers.Seal([]byte(`{"answer":"C"}`), "OX1sT9", "task-7")
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	opened, err := answers.Open(value, "OX1sT9", "task-7")
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	if string(opened) != `{"answer":"C"}` {
		t.Errorf("Open() = %q, want what was sealed", opened)
	}

	// The purpose is the one it was bound to, and nothing else opens it.
	if _, err := ring.Open(seal.PurposeAccess, value, "OX1sT9", "task-7"); !errors.Is(err, seal.ErrPurpose) {
		t.Errorf("Open() as another purpose error = %v, want %v", err, seal.ErrPurpose)
	}
	if _, err := answers.Open(value, "Zk4pQ2", "task-7"); !errors.Is(err, seal.ErrIntegrity) {
		t.Errorf("Open() with another binding error = %v, want %v", err, seal.ErrIntegrity)
	}
}
