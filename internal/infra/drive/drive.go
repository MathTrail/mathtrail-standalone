// Package drive reaches the files the service keeps in a parent's Google
// Drive, through Drive's own HTTP API, version 3.
//
// It knows Drive and nothing of what the files hold. Each call is one request,
// made with the token of the parent whose Drive it is, and asks Drive for no
// more of a file than the caller reads. What Drive answers is taken as far as
// a caller can act on it — the file, or which of a handful of refusals it was
// — and no further: nothing Drive says in its own words, no identifier of a
// file and no token ever reaches an error, because an error may end up where
// a person reads it.
//
// The parent grants the drive.file scope, so Drive shows the service the
// files it made itself and none of the parent's own: whatever a search finds,
// the service put there.
package drive

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Google is where Drive's API is: both the address of the files and the one
// they are uploaded to are found under it.
const Google = "https://www.googleapis.com"

// FolderType is the type Drive gives a folder: a file with no content.
const FolderType = "application/vnd.google-apps.folder"

// Settings are what the calls are made with.
type Settings struct {
	// Root is where Drive's API is: Google, unless a test stands in.
	Root string
	// Timeout is how long one call may take, from the request to the last
	// byte of its answer.
	Timeout time.Duration
}

// ErrSettings is returned when the calls are given settings they could not
// be made with; callers branch on it with errors.Is.
var ErrSettings = errors.New("drive: settings")

// validate refuses what no call could be made with, naming it, rather than
// building a client that fails on the first parent.
func (s *Settings) validate() error {
	root, err := url.Parse(s.Root)
	switch {
	case err != nil, root.Scheme != "https" && root.Scheme != "http", root.Host == "":
		return fmt.Errorf("%w: Root must be an http or https address", ErrSettings)
	case root.RawQuery != "", root.Fragment != "", strings.TrimSuffix(root.Path, "/") != "":
		return fmt.Errorf("%w: Root must be a bare scheme and host", ErrSettings)
	case s.Timeout <= 0:
		return fmt.Errorf("%w: Timeout must be positive", ErrSettings)
	}
	return nil
}

// The refusals a caller acts on, each checked with errors.Is. A call that ends
// in none of them failed in a way nothing but a fix of the request would
// mend, and says so in the operation's own words.
var (
	// ErrNotFound means Drive has no such file, or none the service may see:
	// it was deleted for good, it is the parent's own, or it is one the
	// service may no longer reach.
	ErrNotFound = errors.New("drive: no such file")
	// ErrUnauthorized means Drive did not take the token, or took it and found
	// that it grants nothing of Drive: the parent took the access back, or
	// the token has ended.
	ErrUnauthorized = errors.New("drive: the token was not accepted")
	// ErrRateLimited means Drive asked for a pause: too many calls for this
	// parent, or for the service.
	ErrRateLimited = errors.New("drive: too many calls")
	// ErrStorageFull means the parent's Drive has no room left for what was to
	// be written.
	ErrStorageFull = errors.New("drive: the Drive is full")
	// ErrUnavailable means Drive could not be reached, or failed on its side.
	ErrUnavailable = errors.New("drive: Drive did not answer")
	// ErrTooLarge means a file holds more than the caller would read of it.
	ErrTooLarge = errors.New("drive: the file is larger than asked for")
	// ErrNotKept means Drive gives no content of an earlier revision that is
	// not kept forever.
	ErrNotKept = errors.New("drive: the revision is not kept forever")
)

// File is what the service knows of a file in Drive: as much of it as a call
// asks for, and nothing of what it holds. Written to Drive, it carries only
// the fields that are set.
type File struct {
	// ID is how Drive names the file for good, whatever it is called.
	ID string `json:"id,omitempty"`
	// Name is what the parent sees the file called.
	Name string `json:"name,omitempty"`
	// MimeType is what kind of file it is, FolderType for a folder.
	MimeType string `json:"mimeType,omitempty"`
	// Parents holds the folder the file is in.
	Parents []string `json:"parents,omitempty"`
	// AppProperties are the service's own properties of the file, which no
	// other app and no search of the parent's sees. An update sets the ones
	// it names and keeps the others.
	AppProperties map[string]string `json:"appProperties,omitempty"`
	// ModifiedTime is when the file last changed, by anybody's hand.
	ModifiedTime time.Time `json:"modifiedTime,omitzero"`
	// WebViewLink opens the file in a browser, for whoever may see it.
	WebViewLink string `json:"webViewLink,omitempty"`
	// Trashed says the file is in the bin. The parent puts it there, and
	// nothing the service writes takes it out.
	Trashed bool `json:"trashed,omitempty"`
}

// Query is a search for the files that carry one property of the service's
// own with one value: those outside the bin, or those in it.
type Query struct {
	Key   string
	Value string
	// InBin looks in the bin, and only there.
	InBin bool
}

// String is the query in Drive's own search language.
func (q Query) String() string {
	return fmt.Sprintf("appProperties has { key='%s' and value='%s' } and trashed = %t", quoted(q.Key), quoted(q.Value), q.InBin)
}

// quoted escapes a value for a string of Drive's search language, which
// escapes a quote and the backslash with a backslash.
func quoted(value string) string {
	return strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value)
}

// Files are the calls the service makes to the files in a parent's Drive. Each
// method is one request to Drive, made with the parent's token.
type Files interface {
	// List finds the files a query matches, the most recently modified first,
	// with their names, their folders, when they changed and where they open.
	List(ctx context.Context, token string, query Query) ([]File, error)
	// Get reads what a file is called.
	Get(ctx context.Context, token, id string) (File, error)
	// Download reads what a file holds, and refuses a file that holds more
	// than limit bytes with ErrTooLarge.
	Download(ctx context.Context, token, id string, limit int64) ([]byte, error)
	// Create makes a file of the metadata given, holding the content when
	// there is any; a folder is made with none. It answers the new file's ID.
	Create(ctx context.Context, token string, meta *File, content []byte) (File, error)
	// Update makes the change to a file, and answers the file as it is
	// afterwards: its ID, whether it is in the bin, and its properties.
	Update(ctx context.Context, token, id string, change *Change) (File, error)
}

// Change is what an update does to a file.
type Change struct {
	// Meta holds what the update sets: the name, the type, and the properties
	// it names, the others kept as they are.
	Meta *File
	// Content replaces what the file holds, when there is any. A change with
	// none sets the metadata alone and makes no revision.
	Content []byte
	// Keep marks the revision the content makes to be kept forever: Drive
	// never purges it, and gives what it held once it is no longer the file's
	// content. A revision kept forever cannot be let go again, only deleted.
	Keep bool
}

// Revision is one state of a file's content that Drive keeps: the file's
// content as one upload left it.
type Revision struct {
	// ID is how Drive names the revision, among the file's.
	ID string `json:"id"`
	// ModifiedTime is when the content was uploaded.
	ModifiedTime time.Time `json:"modifiedTime"`
	// KeepForever says Drive never purges the revision. Drive purges one that
	// is not after about thirty days, or sooner when a file has a hundred
	// such, and gives the content of only those kept forever.
	KeepForever bool `json:"keepForever"`
}

// Revisions are the calls the service makes to the history of a file in a
// parent's Drive. Each method is one request to Drive, made with the parent's
// token.
type Revisions interface {
	// List reads every revision of a file Drive keeps, in no order Drive
	// promises.
	List(ctx context.Context, token, file string) ([]Revision, error)
	// Download reads what a revision held, and refuses one that holds more
	// than limit bytes with ErrTooLarge and one not kept forever with
	// ErrNotKept.
	Download(ctx context.Context, token, file, revision string, limit int64) ([]byte, error)
	// Keep marks a revision to be kept forever.
	Keep(ctx context.Context, token, file, revision string) error
	// Delete removes a revision for good: the one way to let go of a
	// revision kept forever.
	Delete(ctx context.Context, token, file, revision string) error
}
