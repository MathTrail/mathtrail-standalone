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
	"time"

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
	// ErrCorrupted means there is a file and it cannot be read as a profile:
	// one that is no profile at all and has no earlier state that reads, or a
	// profile that breaks a rule, which is never rolled back since a newer
	// build or the parent may have written it. A file written by a newer
	// build of the service at a newer version is not corrupted: it is refused
	// with profile.ErrNewer instead, because waiting is what helps.
	ErrCorrupted = errors.New("store: the profile cannot be read")
	// ErrAccessRevoked means the account no longer lets the service reach the
	// profile: the parent took the permission back. Signing in again is what
	// helps.
	ErrAccessRevoked = errors.New("store: access to the profile was taken back")
	// ErrAccessExpired means the account's access ends before the store could
	// be sure of finishing, so nothing was asked of it. Renewed access is what
	// helps.
	ErrAccessExpired = errors.New("store: access to the profile ends too soon")
	// ErrInBin means the profile's file is in the bin of the place it is kept.
	// Nothing reads or writes it there: the parent restores it, or asks for a
	// new start.
	ErrInBin = errors.New("store: the profile is in the bin")
	// ErrBehind means the store answered with a state older than one this
	// instance wrote itself, and went on doing so when asked again. Nothing
	// was done, and the call is made again in a moment.
	ErrBehind = errors.New("store: the profile read is behind one already written")
	// ErrRestored means the file was damaged and has been put back to its last
	// state that reads. Nothing else was done: the caller says so, since what
	// came after that state is gone, and reads again.
	ErrRestored = errors.New("store: the profile was damaged and has been restored")
	// ErrStorageFull means there is no room left where the profile is kept, so
	// nothing was written.
	ErrStorageFull = errors.New("store: no room left for the profile")
	// ErrUnavailable means the place the profile is kept did not answer, or
	// asked for a pause, and went on doing so for as long as a call can wait.
	// Nothing was written.
	ErrUnavailable = errors.New("store: the profile could not be reached")
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
	// nothing can read, and profile.ErrNewer that a newer build wrote it. A
	// store that keeps a file's history may also answer ErrRestored, having
	// put a damaged file back to its last readable state, or ErrConflict,
	// having found it changed while it did; and one that keeps a bin,
	// ErrInBin.
	Load(ctx context.Context, account Account) (*profile.Profile, Revision, error)
	// Create keeps the first profile of an account. An account that has one
	// already, readable or not, answers ErrConflict and keeps what it has: a
	// profile is never replaced by a new one the parent did not ask for. One
	// whose profile is in the bin, in a store that keeps one, answers ErrInBin.
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
	// StartOver keeps a new profile in place of one nothing can read, which
	// the parent asked to start again from. The file it replaces is set aside
	// rather than deleted — it is still there to be found, and no read of the
	// profile ever takes it for one again — and so is a profile file in the
	// bin. An account with nothing to set aside is given the new profile as a
	// first one.
	//
	// A profile that can be read is never replaced: StartOver answers
	// ErrConflict and changes nothing, since what the parent was told no
	// longer holds, and reading again says what does.
	StartOver(ctx context.Context, account Account, p *profile.Profile) (Revision, error)
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

// key is what reaches the place an account's profile is kept, and until when
// it does.
type key struct {
	token string
	ends  time.Time
}

// NewAccount is the account with this identifier, reached with this token
// until the moment it ends. A token that does not end — or no token at all —
// is given the zero time.
func NewAccount(id, token string, ends time.Time) Account {
	return Account{ID: id, key: &key{token: token, ends: ends}}
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

// Ends is when the token stops opening the place the profile is kept, or the
// zero time when it does not end. A store that reaches that place with it
// starts nothing it could not finish by then.
func (a Account) Ends() time.Time {
	if a.key == nil {
		return time.Time{}
	}
	return a.key.ends
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
	// Others are files beside it that carry a profile too — an old one
	// restored from the bin, say. The store reads and writes the one above,
	// the one changed last, and never merges two: these are for the parent to
	// look at and delete.
	Others []Elsewhere
}

// Elsewhere is another file that carries a profile, as its owner would look
// for it.
type Elsewhere struct {
	// File is the name of the file.
	File string
	// Link opens the file for whoever is signed in to the account.
	Link string
}
