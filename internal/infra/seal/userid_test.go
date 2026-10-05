package seal_test

import (
	"encoding/base64"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// An account is known by the same identifier for as long as the key is
// current, whatever key is kept from before a rotation; a rotation that makes
// another key current gives it another.
func TestAUserIDFollowsTheCurrentKey(t *testing.T) {
	t.Parallel()

	const account = "google-sub:110169484474386276334"
	alone := newRing(t, keyNamed("current"), "").UserID(account)

	for _, tc := range []struct {
		name string
		ring *seal.KeyRing
		same bool
	}{
		{"the same key", newRing(t, keyNamed("current"), ""), true},
		{"the same key with a previous one", newRing(t, keyNamed("current"), keyNamed("previous")), true},
		{"after a rotation", newRing(t, keyNamed("next"), keyNamed("current")), false},
		{"another key", newRing(t, keyNamed("other"), ""), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.ring.UserID(account); (got == alone) != tc.same {
				t.Errorf("UserID() = %q beside %q under the key alone; want the same: %v", got, alone, tc.same)
			}
		})
	}
}

// During a rotation an account is known by two identifiers: the one the
// current key derives, which a sign-in is given now, and the one a sign-in
// made before the rotation was given, under the key that was current then.
// Outside a rotation there is one.
func TestAnAccountIsKnownByWhatItsTokensMayCarry(t *testing.T) {
	t.Parallel()

	const account = "google-sub:110169484474386276334"
	before := newRing(t, keyNamed("current"), "").UserID(account)
	rotated := newRing(t, keyNamed("next"), keyNamed("current"))

	if got, want := rotated.UserIDs(account), []string{rotated.UserID(account), before}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("UserIDs() during a rotation = %q, want %q", got, want)
	}
	if got := newRing(t, keyNamed("current"), "").UserIDs(account); len(got) != 1 || got[0] != before {
		t.Errorf("UserIDs() outside a rotation = %q, want %q alone", got, before)
	}
}

// An identifier says nothing of the account it was made from.
func TestAUserIDShowsNothingOfTheAccount(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	if got := ring.UserID("google-sub:"); got == ring.UserID("google-sub:1") {
		t.Errorf("UserID() = %q for two accounts, want each its own", got)
	}
	if got := ring.UserID("110169484474386276334"); got == "110169484474386276334" || len(got) != 16 {
		t.Errorf("UserID() = %q, want 16 characters that are not the account", got)
	}
}

func TestUserIDsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	properties := gopter.NewProperties(nil)

	properties.Property("an identifier is 16 characters of base64url, the same every time", prop.ForAll(
		func(account string) bool {
			id := ring.UserID(account)
			_, err := base64.RawURLEncoding.DecodeString(id)
			return len(id) == 16 && err == nil && ring.UserID(account) == id
		},
		gen.AnyString(),
	))

	properties.Property("two accounts are two identifiers", prop.ForAll(
		func(one, other string) bool {
			return one == other || ring.UserID(one) != ring.UserID(other)
		},
		gen.AnyString(), gen.Identifier(),
	))

	properties.TestingRun(t)
}
