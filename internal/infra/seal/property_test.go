package seal_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// The properties here name what has to hold for every value rather than for
// the handful somebody thought of, because the inputs that break a sealer are
// exactly the ones nobody would write down: a plaintext of no bytes, a binding
// that differs from another only in where it was split, a character changed in
// the one place the format does not look at.

// genPurpose produces one of the purposes the package knows.
func genPurpose() gopter.Gen {
	constants := make([]interface{}, 0, len(everyPurpose))
	for _, purpose := range everyPurpose {
		constants = append(constants, purpose)
	}
	return gen.OneConstOf(constants...)
}

// genPlaintext produces the bytes a caller seals, from none to a few dozen.
func genPlaintext() gopter.Gen { return gen.SliceOf(gen.UInt8()) }

// genBinding produces what a caller ties a value to: a user, a task, both, or
// nothing at all.
func genBinding() gopter.Gen { return gen.SliceOf(gen.AnyString()) }

func TestSealingHoldsItsProperties(t *testing.T) {
	t.Parallel()

	ring := newRing(t, keyNamed("current"), "")
	properties := gopter.NewProperties(nil)

	properties.Property("what was sealed is what opens", prop.ForAll(
		func(plaintext []byte, purpose seal.Purpose, binding []string) bool {
			value, err := ring.Seal(purpose, plaintext, binding...)
			if err != nil {
				return false
			}
			opened, err := ring.Open(purpose, value, binding...)
			return err == nil && bytes.Equal(opened, plaintext)
		},
		genPlaintext(), genPurpose(), genBinding(),
	))

	// Every character counts, including the ones that look like decoration:
	// the version, the purpose and the key id are as protected as the
	// ciphertext, because a value that opened after any of them was rewritten
	// would be one somebody else could aim somewhere it was never meant for.
	properties.Property("a value with one character changed never opens", prop.ForAll(
		func(plaintext []byte, purpose seal.Purpose, position uint16, replacement uint8) bool {
			value, err := ring.Seal(purpose, plaintext)
			if err != nil {
				return false
			}
			changed := []byte(value)
			at := int(position) % len(changed)
			if changed[at] == replacement {
				return true
			}
			changed[at] = replacement

			_, err = ring.Open(purpose, string(changed))
			return err != nil
		},
		genPlaintext(), genPurpose(), gen.UInt16(), gen.UInt8(),
	))

	// Two generators rather than one used twice, so that each argument reads
	// as the side of the comparison it stands for.
	sealedAs, openedAs := genPurpose(), genPurpose()
	sealedWith, openedWith := genBinding(), genBinding()

	properties.Property("a value never opens as another purpose", prop.ForAll(
		func(plaintext []byte, sealedAs, openedAs seal.Purpose) bool {
			if sealedAs == openedAs {
				return true
			}
			value, err := ring.Seal(sealedAs, plaintext)
			if err != nil {
				return false
			}
			_, err = ring.Open(openedAs, value)
			return errors.Is(err, seal.ErrPurpose)
		},
		genPlaintext(), sealedAs, openedAs,
	))

	properties.Property("a value never opens against another binding", prop.ForAll(
		func(plaintext []byte, sealedWith, openedWith []string) bool {
			if slices.Equal(sealedWith, openedWith) {
				return true
			}
			value, err := ring.Seal(seal.PurposeTaskAnswer, plaintext, sealedWith...)
			if err != nil {
				return false
			}
			_, err = ring.Open(seal.PurposeTaskAnswer, value, openedWith...)
			return errors.Is(err, seal.ErrIntegrity)
		},
		genPlaintext(), sealedWith, openedWith,
	))

	properties.TestingRun(t)
}

func TestKeysHoldTheirProperties(t *testing.T) {
	t.Parallel()

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	properties := gopter.NewProperties(nil)

	properties.Property("any bytes of the right length are a key, and name themselves", prop.ForAll(
		func(secret []byte) bool {
			key, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(secret), "")
			if err != nil || len(key.CurrentKeyID()) != 6 {
				return false
			}
			if strings.ContainsFunc(key.CurrentKeyID(), func(r rune) bool { return !strings.ContainsRune(alphabet, r) }) {
				return false
			}
			again, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(secret), "")
			return err == nil && again.CurrentKeyID() == key.CurrentKeyID()
		},
		gen.SliceOfN(seal.KeySize, gen.UInt8()),
	))

	properties.Property("bytes of any other length are not a key", prop.ForAll(
		func(length int) bool {
			if length == seal.KeySize {
				return true
			}
			_, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(make([]byte, length)), "")
			return errors.Is(err, seal.ErrKey)
		},
		gen.IntRange(0, 2*seal.KeySize),
	))

	properties.TestingRun(t)
}
