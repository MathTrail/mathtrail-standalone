package main

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// logged is a copy of the stand-in's request log.
func logged(f *fakeZenodo) []string {
	var log []string
	f.set(func() { log = slices.Clone(f.requests) })
	return log
}

// requestsLike are the requests of the stand-in's log that start with the
// method and path given, such as "PUT /files/".
func requestsLike(f *fakeZenodo, prefix string) []string {
	var found []string
	for _, r := range logged(f) {
		if strings.HasPrefix(r, prefix) {
			found = append(found, r)
		}
	}
	return found
}

// failOn makes the stand-in fail every request that matches, from now on, and
// starts its request log afresh.
func failOn(f *fakeZenodo, match func(*http.Request) bool) {
	f.set(func() { f.fail, f.requests = match, nil })
}

// nextVersion is the request for a draft of version paper-a/bbb in record 1001,
// which holds version paper-a/aaa, published.
func nextVersion(t *testing.T, f *fakeZenodo) DraftRequest {
	t.Helper()
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	return DraftRequest{Metadata: metadata(), Archive: archive(t, "second"), Version: "paper-a/bbb", Record: "1001", Today: "2026-10-08"}
}

// A first draft starts a record only once the owner's published records are
// known: one of them may hold the title already, and a second record would
// give the artifact a second DOI for good.
func TestAFirstDraftStartsNoRecordWhenThePublishedOnesCannotBeListed(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	failOn(f, func(r *http.Request) bool { return r.URL.Query().Get("status") == "published" })
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Today: "2026-10-08"}

	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("a draft whose list of published records failed gives %v, want that failure", err)
	}
	if slices.Contains(logged(f), "POST /api/deposit/depositions") {
		t.Errorf("it started a record anyway: %v", logged(f))
	}
}

// A draft whose upload failed holds no file, and Zenodo refuses to publish it:
// the step passes that refusal on, says nothing of a version that is not out,
// and the draft stays a draft.
func TestPublishPassesOnTheRefusalOfADraftWithNoFile(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	f.set(func() { f.broken = true })
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Today: "2026-10-08"}
	if err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{}); err == nil {
		t.Fatal("a draft whose upload failed gives no error")
	}
	var out bytes.Buffer

	err := Publish(context.Background(), f.client(), "paper-a/aaa", "", &out)

	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("publishing a draft with no file gives %v, want Zenodo's refusal", err)
	}
	if d := only(t, f, func(*fakeDeposition) bool { return true }); d.published {
		t.Errorf("the draft is %+v, want it unpublished", d)
	}
	if out.Len() != 0 {
		t.Errorf("the publish step says %q, want nothing", out.String())
	}
}

// Two drafts of one version, each in a record of its own, are not chosen
// between: a published version cannot be taken back. Named by its record, the
// one meant is published and the other is left a draft.
func TestPublishChoosesNoDraftOfTwoUnlessTheRecordIsNamed(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	other := metadata()
	other.Title = "Another artifact"
	r := DraftRequest{Metadata: other, Archive: archive(t, "other"), Version: "paper-a/aaa", Today: "2026-10-08"}
	if err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{}); err != nil {
		t.Fatalf("draft of another record: %v", err)
	}

	err := Publish(context.Background(), f.client(), "paper-a/aaa", "", &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "2 drafts match") {
		t.Errorf("publish with two drafts of the version gives %v, want a refusal to choose", err)
	}
	for _, d := range f.state() {
		if d.published {
			t.Errorf("deposition %d of record %d was published anyway", d.id, d.concept)
		}
	}

	publish(t, f, "paper-a/aaa", "1001")

	if d := only(t, f, func(d *fakeDeposition) bool { return d.published }); d.concept != 1001 {
		t.Errorf("published deposition %d of record %d, want the draft of record 1001", d.id, d.concept)
	}
}

// A record named for the next version must hold a published version of its
// owner's: a mistyped id would otherwise leave a draft somewhere, or a record
// nobody meant.
func TestADraftForARecordWithNothingPublishedIsRefused(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Record: "4242", Today: "2026-10-08"}

	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "record 4242 has no published version") {
		t.Errorf("a draft for an unknown record gives %v, want a refusal naming record 4242", err)
	}
	if posts := requestsLike(f, "POST "); len(posts) != 0 {
		t.Errorf("it asked Zenodo %v, want nothing made", posts)
	}
}

// An archive that is not there stops the step with an error that names it.
// The step takes the predecessor's files out of the new draft before it opens
// the archive, so the draft is left holding no file; the next run, given the
// archive, takes that draft up.
func TestAMissingArchiveLeavesAnEmptyDraftTheNextRunTakesUp(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	r := nextVersion(t, f)
	r.Archive = filepath.Join(t.TempDir(), "paper-a-artifact.tar.gz")

	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "the archive") {
		t.Errorf("a draft of an archive that is not there gives %v, want an error naming the archive", err)
	}
	if left := only(t, f, func(d *fakeDeposition) bool { return !d.published }); len(left.files) != 0 {
		t.Errorf("the draft left holds %v, want no file", filesOf(left))
	}

	draft(t, f, "paper-a/bbb", "1001", "second")

	next := only(t, f, func(d *fakeDeposition) bool { return !d.published })
	if got := filesOf(next); len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=second" {
		t.Errorf("the draft holds %v, want the archive alone", got)
	}
}

// A step that fails inside a draft stops where it failed: no archive is sent
// beside a predecessor's file that could not be listed or taken out, and no
// file is touched once the draft's description could not be written.
func TestAFailureInsideADraftStopsTheStepThere(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		fail     func(*http.Request) bool
		unwanted func(request string) bool
	}{
		{
			name: "listing the predecessor's files",
			fail: func(r *http.Request) bool {
				return r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/files")
			},
			unwanted: func(request string) bool { return strings.HasPrefix(request, "PUT /files/") },
		},
		{
			name:     "taking out a predecessor's file",
			fail:     func(r *http.Request) bool { return r.Method == http.MethodDelete },
			unwanted: func(request string) bool { return strings.HasPrefix(request, "PUT /files/") },
		},
		{
			name: "writing the description",
			fail: func(r *http.Request) bool {
				return r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/")
			},
			unwanted: func(request string) bool { return strings.Contains(request, "/files") },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeZenodo(t)
			r := nextVersion(t, f)
			failOn(f, tc.fail)

			err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

			if err == nil || !strings.Contains(err.Error(), "500") {
				t.Errorf("the draft step gives %v, want the failure", err)
			}
			if after := slices.DeleteFunc(logged(f), func(request string) bool { return !tc.unwanted(request) }); len(after) != 0 {
				t.Errorf("the step went on to %v", after)
			}
		})
	}
}

// Zenodo made the draft of the next version, but its answer did not say
// where. The step says so rather than guessing, and the next run finds that
// draft open in the record and takes it up instead of asking for another.
func TestANewVersionThatNamesNoDraftIsTakenUpByTheNextRun(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	r := nextVersion(t, f)
	f.set(func() { f.noLatestDraft = true })

	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "named no draft") {
		t.Errorf("a new version that names no draft gives %v, want that said", err)
	}

	f.set(func() { f.noLatestDraft = false })
	draft(t, f, "paper-a/bbb", "1001", "second")

	next := only(t, f, func(d *fakeDeposition) bool { return !d.published })
	if got := filesOf(next); next.concept != 1001 || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=second" {
		t.Errorf("the draft is %+v holding %v, want the one Zenodo made, holding the archive alone", next, got)
	}
}

// The DOI every version shares is what a reader cites, so the publish step
// names it even when Zenodo's answer leaves it out, from the version's DOI and
// the record's id.
func TestPublishNamesTheSharedDOIWhenZenodoLeavesItOut(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	f.set(func() { f.noConceptDOI = true })
	draft(t, f, "paper-a/aaa", "", "first")

	out := publish(t, f, "paper-a/aaa", "")

	if want := "every version of it is cited by 10.5072/zenodo.1001"; !strings.Contains(out, want) {
		t.Errorf("the publish step says %q, want %q in it", out, want)
	}
}

// A server that gave the same page whatever page was asked for would be read
// forever; the list stops after as many pages as an owner's depositions could
// fill, and nothing is published.
func TestAListThatNeverEndsIsCutOff(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	f.set(func() { f.samePage, f.requests = true, nil })

	err := Publish(context.Background(), f.client(), "paper-a/aaa", "", &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "the draft depositions run past 100 pages") {
		t.Errorf("an endless list gives %v, want it cut off", err)
	}
	if pages := requestsLike(f, "GET /api/deposit/depositions"); len(pages) != mostPages {
		t.Errorf("read %d pages, want %d", len(pages), mostPages)
	}
	if posts := requestsLike(f, "POST "); len(posts) != 0 {
		t.Errorf("it asked Zenodo %v, want nothing published", posts)
	}
}

// The draft step decides what to make from what Zenodo holds; when it cannot
// read that, it makes nothing: no record, and no version.
func TestADraftMakesNothingFromAnAnswerItCouldNotGet(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		next     bool // a draft of the next version of a published record, rather than a first draft
		fail     func(*http.Request) bool
		unwanted func(request string) bool
	}{
		{
			name:     "the drafts, before a first draft",
			fail:     func(r *http.Request) bool { return r.URL.Query().Get("status") == "draft" },
			unwanted: func(request string) bool { return strings.HasPrefix(request, "POST ") },
		},
		{
			name:     "the published versions, before the next version",
			next:     true,
			fail:     func(r *http.Request) bool { return r.URL.Query().Get("status") == "published" },
			unwanted: func(request string) bool { return strings.HasPrefix(request, "POST ") },
		},
		{
			name: "the next version itself",
			next: true,
			fail: func(r *http.Request) bool { return strings.HasSuffix(r.URL.Path, "/actions/newversion") },
			unwanted: func(request string) bool {
				return strings.HasPrefix(request, "GET /api/deposit/depositions/") || strings.HasPrefix(request, "PUT ")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeZenodo(t)
			r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Today: "2026-10-08"}
			if tc.next {
				r = nextVersion(t, f)
			}
			before := len(f.state())
			failOn(f, tc.fail)

			err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})

			if err == nil || !strings.Contains(err.Error(), "500") {
				t.Errorf("the draft step gives %v, want the failure", err)
			}
			if after := slices.DeleteFunc(logged(f), func(request string) bool { return !tc.unwanted(request) }); len(after) != 0 {
				t.Errorf("the step went on to %v", after)
			}
			if after := len(f.state()); after != before {
				t.Errorf("%d depositions after, %d before", after, before)
			}
		})
	}
}

// With no draft of the version to publish, the step tells a version published
// already from one never drafted only once it has read the published
// versions; when it cannot, it says that failure rather than sending the
// person to draft a version that may be out already.
func TestPublishDoesNotGuessWhyNoDraftHoldsTheVersion(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	failOn(f, func(r *http.Request) bool { return r.URL.Query().Get("status") == "published" })

	err := Publish(context.Background(), f.client(), "paper-a/aaa", "", &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "run the draft step first") {
		t.Errorf("publish with the published versions unreadable gives %v, want that failure and no advice to draft", err)
	}
}
