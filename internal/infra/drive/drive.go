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
	// it was deleted for good, or it is the parent's own.
	ErrNotFound = errors.New("drive: no such file")
	// ErrUnauthorized means Drive did not take the token: the parent took the
	// access back, or the token has ended.
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
}

// Query is a search for the files that carry one property of the service's
// own with one value, and are not in the bin.
type Query struct {
	Key   string
	Value string
}

// String is the query in Drive's own search language.
func (q Query) String() string {
	return fmt.Sprintf("appProperties has { key='%s' and value='%s' } and trashed = false", quoted(q.Key), quoted(q.Value))
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
	// Update replaces what a file holds with content, and sets the
	// properties meta names while keeping the others. It answers the file's
	// ID.
	Update(ctx context.Context, token, id string, meta *File, content []byte) (File, error)
}
