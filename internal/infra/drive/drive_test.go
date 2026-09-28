package drive_test

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
)

// The parent whose Drive the cases reach, and another parent beside them.
const (
	token        = "ya29.the-token-of-a-parent"
	anotherToken = "ya29.the-token-of-another-parent"
)

// A marker the cases search by: a property of the service's own.
var marked = drive.Query{Key: "app", Value: "profile"}

// filesAt is the calls, made to the API at root and given the time given.
func filesAt(t *testing.T, root string, timeout time.Duration) drive.Files {
	t.Helper()

	files, err := drive.NewFiles(&drive.Settings{Root: root, Timeout: timeout})
	if err != nil {
		t.Fatalf("NewFiles() error = %v, want nil", err)
	}
	return files
}

// standIn is a stand-in Drive and the calls made to it.
func standIn(t *testing.T) (*drivetest.Drive, drive.Files) {
	t.Helper()

	fake := drivetest.New(t)
	return fake, filesAt(t, fake.Root(), 5*time.Second)
}

func TestSettingsNoCallCouldBeMadeWithBuildNothing(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		settings drive.Settings
	}{
		{"no root", drive.Settings{Timeout: time.Second}},
		{"a root that is not an address", drive.Settings{Root: "www.googleapis.com", Timeout: time.Second}},
		{"a root of another scheme", drive.Settings{Root: "ftp://www.googleapis.com", Timeout: time.Second}},
		{"a root with a path", drive.Settings{Root: "https://www.googleapis.com/drive/v3", Timeout: time.Second}},
		{"a root with a query", drive.Settings{Root: "https://www.googleapis.com?key=a", Timeout: time.Second}},
		{"no time for a call", drive.Settings{Root: drive.Google}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := drive.NewFiles(&tc.settings); !errors.Is(err, drive.ErrSettings) {
				t.Errorf("NewFiles(%+v) error = %v, want %v", tc.settings, err, drive.ErrSettings)
			}
		})
	}
	if _, err := drive.NewFiles(nil); !errors.Is(err, drive.ErrSettings) {
		t.Errorf("NewFiles(nil) error = %v, want %v", err, drive.ErrSettings)
	}
	if _, err := drive.NewFiles(&drive.Settings{Root: drive.Google + "/", Timeout: time.Second}); err != nil {
		t.Errorf("NewFiles() with Drive's own root error = %v, want nil", err)
	}
}

// A folder and a file in it are made, found, read and rewritten, each by one
// request, and what the service wrote is what Drive keeps.
func TestAFileIsMadeFoundReadAndRewritten(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	ctx := t.Context()

	folder, err := files.Create(ctx, token, &drive.File{
		Name: "Folder", MimeType: drive.FolderType, AppProperties: map[string]string{"app": "folder"},
	}, nil)
	if err != nil {
		t.Fatalf("Create() of a folder error = %v, want nil", err)
	}
	made, err := files.Create(ctx, token, &drive.File{
		Name: "file.json", MimeType: "application/json", Parents: []string{folder.ID},
		AppProperties: map[string]string{"app": "profile", "schema": "1"},
	}, []byte(`{"first": true}`))
	if err != nil {
		t.Fatalf("Create() of a file error = %v, want nil", err)
	}

	found, err := files.List(ctx, token, marked)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(found) != 1 {
		t.Fatalf("List() found %d files, want the one made", len(found))
	}
	got := found[0]
	switch {
	case got.ID != made.ID:
		t.Errorf("List() found %q, want %q", got.ID, made.ID)
	case got.Name != "file.json":
		t.Errorf("List() name = %q, want %q", got.Name, "file.json")
	case !slices.Equal(got.Parents, []string{folder.ID}):
		t.Errorf("List() parents = %v, want [%s]", got.Parents, folder.ID)
	case got.ModifiedTime.IsZero():
		t.Error("List() gave no time of change, want one")
	case got.WebViewLink != drivetest.WebViewLink(made.ID):
		t.Errorf("List() link = %q, want %q", got.WebViewLink, drivetest.WebViewLink(made.ID))
	}

	content, err := files.Download(ctx, token, made.ID, 1<<10)
	if err != nil {
		t.Fatalf("Download() error = %v, want nil", err)
	}
	if string(content) != `{"first": true}` {
		t.Errorf("Download() = %s, want what was made", content)
	}

	if _, err = files.Update(ctx, token, made.ID, &drive.File{
		MimeType: "application/json", AppProperties: map[string]string{"schema": "2"},
	}, []byte(`{"second": true}`)); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	kept := fileIn(t, fake, token, made.ID)
	if string(kept.Content) != `{"second": true}` {
		t.Errorf("after Update() the file holds %s, want the new content", kept.Content)
	}
	// An update names the properties it changes, and the others stay.
	want := map[string]string{"app": "profile", "schema": "2"}
	if !maps.Equal(kept.AppProperties, want) {
		t.Errorf("after Update() the properties are %v, want %v", kept.AppProperties, want)
	}

	named, err := files.Get(ctx, token, folder.ID)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if named.Name != "Folder" {
		t.Errorf("Get() name = %q, want %q", named.Name, "Folder")
	}
	if calls := fake.Calls(); !maps.Equal(calls, drivetest.Calls{"create": 2, "list": 1, "download": 1, "update": 1, "get": 1}) {
		t.Errorf("calls = %v, want one request for each", calls)
	}
}

// A search finds the service's own files that carry the property, not the
// ones in the bin and not the parent's own, the most recently changed first —
// and in no Drive but the one the token reaches.
func TestASearchFindsTheMarkedFilesNewestFirst(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	marker := map[string]string{"app": "profile"}
	older := fake.Put(token, &drivetest.File{Name: "older", AppProperties: marker})
	newer := fake.Put(token, &drivetest.File{Name: "newer", AppProperties: marker})
	fake.Put(token, &drivetest.File{Name: "in the bin", AppProperties: marker, Trashed: true})
	fake.Put(token, &drivetest.File{Name: "the parent's own", AppProperties: marker, Own: true})
	fake.Put(token, &drivetest.File{Name: "unmarked", AppProperties: map[string]string{"app": "folder"}})
	fake.Put(anotherToken, &drivetest.File{Name: "another parent's", AppProperties: marker})

	found, err := files.List(t.Context(), token, marked)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	var ids []string
	for _, file := range found {
		ids = append(ids, file.ID)
	}
	if want := []string{newer, older}; !slices.Equal(ids, want) {
		t.Errorf("List() found %v, want %v", ids, want)
	}
}

// A value that holds a quote or a backslash is escaped as Drive's search
// language escapes them, and finds what holds it.
func TestAQueryValueIsEscaped(t *testing.T) {
	t.Parallel()

	query := drive.Query{Key: "a'key", Value: `a\value`}
	want := `appProperties has { key='a\'key' and value='a\\value' } and trashed = false`
	if got := query.String(); got != want {
		t.Errorf("Query.String() = %s, want %s", got, want)
	}

	fake, files := standIn(t)
	id := fake.Put(token, &drivetest.File{AppProperties: map[string]string{"a'key": `a\value`}})
	found, err := files.List(t.Context(), token, query)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(found) != 1 || found[0].ID != id {
		t.Errorf("List() found %v, want the one file holding the value", found)
	}
}

// Each refusal a caller acts on is told apart, and every other refusal is
// none of them.
func TestDrivesRefusalsAreToldApart(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		drive.ErrNotFound, drive.ErrUnauthorized, drive.ErrRateLimited,
		drive.ErrStorageFull, drive.ErrUnavailable, drive.ErrTooLarge,
	}
	for _, tc := range []struct {
		name   string
		status int
		reason string
		want   error
	}{
		{"the token not taken", http.StatusUnauthorized, "authError", drive.ErrUnauthorized},
		{"no such file", http.StatusNotFound, "notFound", drive.ErrNotFound},
		{"a pause asked for", http.StatusTooManyRequests, "rateLimitExceeded", drive.ErrRateLimited},
		{"a pause asked for this parent", http.StatusForbidden, "userRateLimitExceeded", drive.ErrRateLimited},
		{"a pause asked for the service", http.StatusForbidden, "rateLimitExceeded", drive.ErrRateLimited},
		{"the service's calls for the day spent", http.StatusForbidden, "dailyLimitExceeded", drive.ErrRateLimited},
		{"a full Drive", http.StatusForbidden, "storageQuotaExceeded", drive.ErrStorageFull},
		{"a failure of Drive's", http.StatusInternalServerError, "backendError", drive.ErrUnavailable},
		{"Drive not there", http.StatusServiceUnavailable, "backendError", drive.ErrUnavailable},
		{"a file the parent keeps from the app", http.StatusForbidden, "appNotAuthorizedToFile", nil},
		{"a request Drive cannot read", http.StatusBadRequest, "badRequest", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake, files := standIn(t)
			id := fake.Put(token, &drivetest.File{MimeType: "application/json", Content: []byte("{}")})
			fake.Fail(drivetest.Download, tc.status, tc.reason)

			_, err := files.Download(t.Context(), token, id, 1<<10)
			if err == nil {
				t.Fatalf("Download() error = nil, want a refusal")
			}
			for _, sentinel := range sentinels {
				if is, want := errors.Is(err, sentinel), errors.Is(tc.want, sentinel); is != want {
					t.Errorf("Download() error = %v; errors.Is(%v) = %t, want %t", err, sentinel, is, want)
				}
			}
		})
	}
}

// A token Drive no longer takes reaches nothing, whatever it asks for.
func TestARevokedTokenReachesNothing(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	fake.Put(token, &drivetest.File{AppProperties: map[string]string{"app": "profile"}})
	fake.Revoke(token)

	if _, err := files.List(t.Context(), token, marked); !errors.Is(err, drive.ErrUnauthorized) {
		t.Errorf("List() with a revoked token error = %v, want %v", err, drive.ErrUnauthorized)
	}
}

// No error names the file it was about, repeats what Drive said about it, or
// carries the token: whatever went wrong, and however.
func TestNoErrorNamesTheFileOrCarriesTheToken(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	gone := drivetest.New(t)
	stopped := httptest.NewServer(http.NotFoundHandler())
	stopped.Close()

	for _, tc := range []struct {
		name string
		root string
		// spoil makes the download of the file with this ID go wrong.
		spoil func(id string)
	}{
		{"a file Drive does not have", gone.Root(), func(string) {}},
		{"a refusal Drive explains", fake.Root(), func(string) {
			fake.Fail(drivetest.Download, http.StatusForbidden, "appNotAuthorizedToFile")
		}},
		{"a call that takes too long", fake.Root(), func(string) { fake.Stall(drivetest.Download) }},
		// A reason that is no word of Drive's list is not repeated, whatever
		// it holds — here the file's own ID.
		{"a reason that is no word of Drive's", fake.Root(), func(id string) {
			fake.Fail(drivetest.Download, http.StatusForbidden, "no access to "+id)
		}},
		{"a Drive that is not there", stopped.URL, func(string) {}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := fake.Put(token, &drivetest.File{MimeType: "application/json", Content: []byte("{}")})
			tc.spoil(id)
			files := filesAt(t, tc.root, 200*time.Millisecond)

			_, err := files.Download(t.Context(), token, id, 1<<10)
			if err == nil {
				t.Fatalf("Download() error = nil, want one")
			}
			for _, secret := range []string{id, token, "File not found"} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("Download() error = %q, which carries %q", err, secret)
				}
			}
		})
	}
}

// A file that holds more than the caller would read is refused, and one that
// holds exactly as much is read whole.
func TestADownloadStopsAtItsLimit(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	content := []byte("0123456789")
	id := fake.Put(token, &drivetest.File{MimeType: "application/json", Content: content})

	if _, err := files.Download(t.Context(), token, id, int64(len(content))-1); !errors.Is(err, drive.ErrTooLarge) {
		t.Errorf("Download() past its limit error = %v, want %v", err, drive.ErrTooLarge)
	}
	read, err := files.Download(t.Context(), token, id, int64(len(content)))
	if err != nil {
		t.Fatalf("Download() at its limit error = %v, want nil", err)
	}
	if !bytes.Equal(read, content) {
		t.Errorf("Download() = %q, want %q", read, content)
	}
}

// Content goes to Drive as the kind of file it is, and content whose kind is
// not named is refused before anything is sent.
func TestContentOfNoNamedKindIsNotSent(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	id := fake.Put(token, &drivetest.File{MimeType: "application/json", Content: []byte("{}")})

	if _, err := files.Create(t.Context(), token, &drive.File{Name: "file.json"}, []byte("{}")); err == nil {
		t.Error("Create() of content of no named kind: error = nil, want a refusal")
	}
	if _, err := files.Update(t.Context(), token, id, &drive.File{}, []byte("{}")); err == nil {
		t.Error("Update() with content of no named kind: error = nil, want a refusal")
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("refused uploads cost %v, want no call", calls)
	}
}

// A call Drive leaves unanswered ends when its time is up, as a call that
// took too long, and as Drive not answering.
func TestACallThatTakesTooLongEnds(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	files := filesAt(t, fake.Root(), 50*time.Millisecond)
	fake.Stall(drivetest.List)

	_, err := files.List(context.Background(), token, marked)
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, drive.ErrUnavailable) {
		t.Errorf("List() error = %v, want %v and %v", err, context.DeadlineExceeded, drive.ErrUnavailable)
	}
}

// An answer that breaks off, or does not read as one of Drive's, is Drive not
// answering — the same as no answer at all, and not a fault of the service's.
func TestAnAnswerThatBreaksOffIsDriveNotAnswering(t *testing.T) {
	t.Parallel()

	for name, answer := range map[string]http.HandlerFunc{
		"cut short": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000")
			_, _ = w.Write([]byte(`{"files": [{"id": "a-fi`))
		},
		"not Drive's": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<html>a page of somebody else's</html>`))
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(answer)
			t.Cleanup(server.Close)
			files := filesAt(t, server.URL, 5*time.Second)
			if _, err := files.List(t.Context(), token, marked); !errors.Is(err, drive.ErrUnavailable) {
				t.Errorf("List() error = %v, want %v", err, drive.ErrUnavailable)
			}
			if _, err := files.Download(t.Context(), token, "a-file", 1<<10); name == "cut short" && !errors.Is(err, drive.ErrUnavailable) {
				t.Errorf("Download() error = %v, want %v", err, drive.ErrUnavailable)
			}
		})
	}
}

// An answer about a file says which file it is, or it is not taken for one;
// and a call about a file names one, or it is not sent — as it stands it would
// reach the address of the search.
func TestAFileIsAlwaysNamed(t *testing.T) {
	t.Parallel()

	nameless := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/drive/v3/files" {
			_, _ = w.Write([]byte(`{"files": [{"name": "a file of no ID"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"name": "a file of no ID"}`))
	}))
	t.Cleanup(nameless.Close)
	files := filesAt(t, nameless.URL, 5*time.Second)
	meta := &drive.File{Name: "file.json", MimeType: "application/json"}
	ctx := t.Context()

	for call, err := range map[string]error{
		"List":   second(files.List(ctx, token, marked)),
		"Get":    second(files.Get(ctx, token, "a-file")),
		"Create": second(files.Create(ctx, token, meta, []byte("{}"))),
		"Update": second(files.Update(ctx, token, "a-file", meta, []byte("{}"))),
	} {
		if !errors.Is(err, drive.ErrUnavailable) {
			t.Errorf("%s() answered with no ID: error = %v, want %v", call, err, drive.ErrUnavailable)
		}
	}

	fake, named := standIn(t)
	for call, err := range map[string]error{
		"Get":      second(named.Get(ctx, token, "")),
		"Download": second(named.Download(ctx, token, "", 1<<10)),
		"Update":   second(named.Update(ctx, token, "", meta, []byte("{}"))),
	} {
		if err == nil {
			t.Errorf("%s() of no file: error = nil, want a refusal", call)
		}
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("calls of no file cost %v, want none sent", calls)
	}
}

// second is the error of a call's two results.
func second[T any](_ T, err error) error { return err }

// A call its caller gave up on ends as the caller's own doing, not as Drive
// not answering.
func TestACallGivenUpOnIsNoFailureOfDrives(t *testing.T) {
	t.Parallel()

	fake, files := standIn(t)
	fake.Stall(drivetest.List)
	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(20*time.Millisecond, cancel)

	_, err := files.List(ctx, token, marked)
	if !errors.Is(err, context.Canceled) || errors.Is(err, drive.ErrUnavailable) {
		t.Errorf("List() given up on: error = %v, want %v and not %v", err, context.Canceled, drive.ErrUnavailable)
	}
}

// A redirect is not followed, so the token goes nowhere Drive's API is not.
func TestARedirectIsNotFollowed(t *testing.T) {
	t.Parallel()

	var reached atomic.Bool
	elsewhere := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached.Store(true) }))
	t.Cleanup(elsewhere.Close)
	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(redirecting.Close)

	_, err := filesAt(t, redirecting.URL, time.Second).Get(t.Context(), token, "a-file")
	if err == nil {
		t.Error("Get() answered with a redirect: error = nil, want one")
	}
	if reached.Load() {
		t.Error("the redirect was followed, want it refused")
	}
}

// fileIn is the file the stand-in keeps under this ID, or the end of the case.
func fileIn(t *testing.T, fake *drivetest.Drive, token, id string) *drivetest.File {
	t.Helper()

	for _, file := range fake.Files(token) {
		if file.ID == id {
			return file
		}
	}
	t.Fatalf("the Drive holds no file %q", id)
	return nil
}
