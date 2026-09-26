// Package store is where the profile is kept, as everything that reads or
// writes it sees it: an account, the profile it holds, and the revision that
// profile was read at.
//
// A profile is read, computed over and written back whole. Between the read
// and the write the file may have moved on — another tab, another device, the
// parent editing it by hand — and a write made from a revision that is no
// longer the current one is refused rather than laid over what is there. The
// change is then made again from the state that won, by whoever made it the
// first time: nothing here knows how to make it.
//
// Where the file actually lives — in the parent's own Drive, or in the memory
// of the process — is the business of an implementation, each in a package of
// its own. This package is what they share: the operations, the refusals, and
// how the bytes a store keeps are read as a profile.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The refusals of a store. A caller branches on them with errors.Is, because
// what the person at the other end is told is different for each.
var (
	// ErrNotFound means the account has no profile.
	ErrNotFound = errors.New("store: no profile")
	// ErrConflict means the stored profile is not the one the write was made
	// from: it moved on after it was read, or it exists where a new one was to
	// be created. Nothing was written, and the caller reads again.
	ErrConflict = errors.New("store: the stored profile is not the one expected")
	// ErrCorrupted means there is a file and it cannot be read as a profile.
	// A file written by a newer build of the service is not corrupted: it is
	// refused with profile.ErrNewer instead, because waiting is what helps.
	ErrCorrupted = errors.New("store: the profile cannot be read")
	// ErrAccessRevoked means the account no longer lets the service reach the
	// profile: the parent took the permission back.
	ErrAccessRevoked = errors.New("store: access to the profile was taken back")
)

// Storage keeps one profile per account.
//
// Two things are refused before anything is looked at. An account that names
// no one reaches no one's profile, whatever its token would open. A profile
// that breaks its own rules is never written: Create and Save refuse it with
// profile.ErrInvalid, whatever the store holds, and leave it as it was.
type Storage interface {
	// Load reads the account's profile and the revision it was read at.
	// ErrNotFound means the account has none, ErrCorrupted that it has a file
	// nothing can read, and profile.ErrNewer that a newer build wrote it.
	Load(ctx context.Context, account Account) (*profile.Profile, Revision, error)
	// Create keeps the first profile of an account. An account that has one
	// already, readable or not, answers ErrConflict and keeps what it has: a
	// profile is never replaced by a new one the parent did not ask for.
	Create(ctx context.Context, account Account, p *profile.Profile) (Revision, error)
	// Save writes the profile over the revision it was computed from, and
	// answers the revision it wrote. ErrConflict means the stored profile has
	// moved on since, and ErrNotFound that there is none: a save never makes
	// one, because starting over is something a parent asks for.
	//
	// The profile's own revision has to be one past the stored one — the
	// touch every write carries. The history of the file is pinned by that
	// number and an early read is noticed by it, so a profile touched never or
	// twice is refused, and not as a conflict: reading again would not help.
	Save(ctx context.Context, account Account, p *profile.Profile, expected Revision) (Revision, error)
	// Export says where the person the account belongs to can find the file
	// themselves. The file is the export: there is no second copy of it, and
	// no format of its own.
	Export(ctx context.Context, account Account) (Location, error)
}

// Account is whose profile a call is about. Two accounts are the same account
// when their identifiers are.
type Account struct {
	// ID is the identifier the sign-in derived for the account. It keys what a
	// store holds while the process runs and nothing durable, because it
	// changes at the first sign-in after the keys it was derived with rotate.
	ID string
	// key opens somebody's files, so it sits behind a pointer nothing outside
	// this package can name. Whatever prints or encodes a value by walking its
	// fields then never reaches it: fmt writes a pointer it finds inside a
	// value as an address, and encoding/json skips a field it cannot name.
	key *key
	// Two accounts made alike hold different pointers, so == would answer
	// that they differ. This makes the compiler refuse the question instead.
	_ [0]func()
}

// key is what reaches the place an account's profile is kept.
type key struct{ token string }

// NewAccount is the account with this identifier, reached with this token.
func NewAccount(id, token string) Account {
	return Account{ID: id, key: &key{token: token}}
}

// Token opens the place the profile is kept when that place is the account's
// own rather than the service's. A store that keeps profiles itself does not
// read it, and an account made without one has none.
func (a Account) Token() string {
	if a.key == nil {
		return ""
	}
	return a.key.token
}

// Format writes the account as its identifier, in whatever verb and with
// whatever flags it was asked for, so that an account in an error or a log
// line reads as its identifier would and never takes the token along.
func (a Account) Format(state fmt.State, verb rune) {
	_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), a.ID)
}

// Revision is a state of the stored file, as the store that holds it tells
// states apart. It means nothing anywhere else: a caller keeps the one it was
// handed and gives it back, and never compares or makes one.
type Revision string

// Location is where the person a profile belongs to can find it, in the words
// they would look for it by. The zero Location says there is nowhere a person
// could open it, which is the truth about a profile kept in memory.
type Location struct {
	// Folder is the name of the folder the file is in.
	Folder string
	// File is the name of the file.
	File string
	// Link opens the file for whoever is signed in to the account.
	Link string
}
