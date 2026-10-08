package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// metadata is a description of the record a step takes.
func metadata() Metadata {
	return Metadata{
		UploadType: "software", Title: "The artifact", Description: "Code and data.",
		Creators:    []Creator{{Name: "Stub, Ann", Affiliation: "Stub Institute"}},
		AccessRight: "open", License: "mit",
	}
}

// archive writes a file to put on Zenodo, holding its words.
func archive(t *testing.T, words string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "paper-a-artifact.tar.gz")
	if err := os.WriteFile(path, []byte(words), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// draft runs the draft step against the stand-in, failing the test on an error.
func draft(t *testing.T, f *fakeZenodo, version, record, words string) string {
	t.Helper()
	var out bytes.Buffer
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, words), Version: version, Record: record, Today: "2026-10-08"}
	if err := Draft(context.Background(), f.client(), &r, &out); err != nil {
		t.Fatalf("draft of %s: %v", version, err)
	}
	return out.String()
}

// publish runs the publish step against the stand-in, failing the test on an error.
func publish(t *testing.T, f *fakeZenodo, version, record string) string {
	t.Helper()
	var out bytes.Buffer
	if err := Publish(context.Background(), f.client(), version, record, &out); err != nil {
		t.Fatalf("publish of %s: %v", version, err)
	}
	return out.String()
}

// only is the one deposition the stand-in holds that matches, failing the test
// when there is none or more than one.
func only(t *testing.T, f *fakeZenodo, match func(*fakeDeposition) bool) *fakeDeposition {
	t.Helper()
	var found []*fakeDeposition
	for _, d := range f.state() {
		if match(d) {
			found = append(found, d)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d depositions match, want 1: %+v", len(found), f.state())
	}
	return found[0]
}

// filesOf names a deposition's files with what they hold.
func filesOf(d *fakeDeposition) []string {
	var files []string
	for _, f := range d.files {
		files = append(files, f.name+"="+string(f.data))
	}
	return files
}

func TestADraftStartsARecordOfItsOwnAndSaysItsDOIs(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	out := draft(t, f, "paper-a/aaa", "", "first")
	d := only(t, f, func(*fakeDeposition) bool { return true })
	if got := filesOf(d); len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=first" {
		t.Errorf("the draft holds %v, want the archive alone", got)
	}
	if d.published || d.meta.Version != "paper-a/aaa" || d.meta.PublicationDate != "2026-10-08" || !d.meta.PrereserveDOI || d.meta.Title != "The artifact" {
		t.Errorf("the draft is %+v, want an unpublished draft of the version, dated, with its DOI kept and the record's description", d)
	}
	for _, want := range []string{"/deposit/1002", "10.5072/zenodo.1002", "record is 1001", "10.5072/zenodo.1001"} {
		if !strings.Contains(out, want) {
			t.Errorf("the draft step says %q, want %q in it", out, want)
		}
	}
}

func TestADraftOfTheSameVersionIsTakenUpAgainRatherThanMadeTwice(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	draft(t, f, "paper-a/aaa", "", "again")
	d := only(t, f, func(*fakeDeposition) bool { return true })
	if got := filesOf(d); len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=again" {
		t.Errorf("the draft holds %v, want the latest archive alone", got)
	}
}

func TestPublishPublishesTheDraftOfItsVersion(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	out := publish(t, f, "paper-a/aaa", "")
	if d := only(t, f, func(*fakeDeposition) bool { return true }); !d.published {
		t.Errorf("the draft is not published")
	}
	if !strings.Contains(out, "published as 10.5072/zenodo.1002") || !strings.Contains(out, "cited by 10.5072/zenodo.1001") {
		t.Errorf("the publish step says %q, want the version's DOI and the record's", out)
	}
}

func TestTheNextVersionHoldsItsOwnArchiveAlone(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	// A page of one deposition, so that a step that read the first page alone
	// would miss the record's versions.
	f.set(func() { f.pageSize = 1 })
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	draft(t, f, "paper-a/bbb", "1001", "second")
	next := only(t, f, func(d *fakeDeposition) bool { return !d.published })
	if got := filesOf(next); next.concept != 1001 || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=second" || next.meta.Version != "paper-a/bbb" {
		t.Errorf("the next version is %+v holding %v, want a draft of record 1001 holding its own archive alone", next, got)
	}
	first := only(t, f, func(d *fakeDeposition) bool { return d.published })
	if got := filesOf(first); first.meta.Version != "paper-a/aaa" || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=first" {
		t.Errorf("the published version became %+v holding %v", first, got)
	}
	publish(t, f, "paper-a/bbb", "1001")
}

func TestAVersionPublishedAlreadyLeavesNothingToDo(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	before := len(f.state())
	if out := draft(t, f, "paper-a/aaa", "1001", "first"); !strings.Contains(out, "published already") {
		t.Errorf("the draft step says %q, want that the version is published already", out)
	}
	if out := publish(t, f, "paper-a/aaa", "1001"); !strings.Contains(out, "published already") {
		t.Errorf("the publish step says %q, want that the version is published already", out)
	}
	if after := len(f.state()); after != before {
		t.Errorf("%d depositions after, %d before", after, before)
	}
}

func TestAnOpenDraftOfTheRecordIsTakenUpForTheNextVersion(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	draft(t, f, "paper-a/bbb", "1001", "second")
	draft(t, f, "paper-a/ccc", "1001", "third")
	open := only(t, f, func(d *fakeDeposition) bool { return !d.published })
	if got := filesOf(open); open.meta.Version != "paper-a/ccc" || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=third" {
		t.Errorf("the open draft is %+v holding %v, want version paper-a/ccc holding the third archive alone", open, got)
	}
}

func TestASecondRecordOfTheSameTitleIsRefused(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	before := len(f.state())
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "second"), Version: "paper-a/bbb", Today: "2026-10-08"}
	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "record 1001") || !strings.Contains(err.Error(), "ZENODO_RECORD") {
		t.Errorf("a draft with no record after the record was published gives %v, want a refusal naming record 1001", err)
	}
	if after := len(f.state()); after != before {
		t.Errorf("%d depositions after, %d before: a second record was started", after, before)
	}
}

func TestTheFirstRecordsDraftIsTakenUpForAnotherVersion(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	draft(t, f, "paper-a/bbb", "", "second")
	d := only(t, f, func(*fakeDeposition) bool { return true })
	if got := filesOf(d); d.meta.Version != "paper-a/bbb" || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=second" {
		t.Errorf("the draft is %+v holding %v, want the first record's draft taken up for the newer version", d, got)
	}
}

func TestADraftCutShortIsTakenUpByTheNextRun(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	f.set(func() { f.broken = true })
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Today: "2026-10-08"}
	if err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{}); err == nil {
		t.Fatal("a draft whose upload failed gives no error")
	}
	f.set(func() { f.broken = false })
	draft(t, f, "paper-a/bbb", "", "second")
	d := only(t, f, func(*fakeDeposition) bool { return true })
	if got := filesOf(d); d.meta.Version != "paper-a/bbb" || len(got) != 1 || got[0] != "paper-a-artifact.tar.gz=second" {
		t.Errorf("the draft is %+v holding %v, want the draft the failed run left taken up", d, got)
	}
}

func TestPublishingTheFirstRecordAgainSaysItIsOut(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	if out := publish(t, f, "paper-a/aaa", ""); !strings.Contains(out, "published already") || !strings.Contains(out, "record 1001") {
		t.Errorf("publishing again says %q, want that the version is out, in record 1001", out)
	}
}

func TestPublishRefusesAVersionNoDraftHolds(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	err := Publish(context.Background(), f.client(), "paper-a/zzz", "", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "run the draft step first") {
		t.Errorf("publish with no draft gives %v, want a refusal naming the draft step", err)
	}
}

func TestAFileStoredUnderAnotherSumIsRefused(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	f.set(func() { f.wrongSum = true })
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "first"), Version: "paper-a/aaa", Today: "2026-10-08"}
	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "sums to") {
		t.Errorf("a draft whose file was stored under another sum gives %v, want a refusal", err)
	}
}

func TestAWrongTokenIsRefusedByZenodo(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	c := f.client()
	c.Token = "another"
	err := Publish(context.Background(), c, "paper-a/aaa", "", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "another") {
		t.Errorf("a wrong token gives %v, want Zenodo's refusal, which does not quote the token", err)
	}
}

// runZenodo runs the command against the stand-in with a token, and gives its
// exit status, what it said and what it complained of.
func runZenodo(t *testing.T, f *fakeZenodo, token string, args ...string) (code int, said, complained string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	client := Client{HTTP: f.server.Client(), Token: token}
	args = append(args, "-url", f.server.URL)
	now := func() time.Time { return time.Date(2026, 10, 8, 23, 59, 0, 0, time.UTC) }
	code = run(context.Background(), args, client, now, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// metadataFile writes a description of the record with the creators given.
func metadataFile(t *testing.T, creators string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "zenodo.json")
	text := `{"upload_type": "software", "title": "The artifact", "description": "Code and data.",
	"creators": ` + creators + `, "access_right": "open", "license": "mit"}`
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTheCommandDraftsAndPublishes(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	meta := metadataFile(t, `[{"name": "Stub, Ann"}]`)
	code, out, errs := runZenodo(t, f, fakeToken, "draft", "-archive", archive(t, "first"), "-version", "paper-a/aaa", "-metadata", meta)
	if code != 0 || !strings.Contains(out, "ready to be looked at") {
		t.Fatalf("draft exits %d saying %q, %q", code, out, errs)
	}
	if d := only(t, f, func(*fakeDeposition) bool { return true }); d.meta.PublicationDate != "2026-10-08" {
		t.Errorf("the version is dated %q, want the day the step ran, in UTC", d.meta.PublicationDate)
	}
	if code, out, errs = runZenodo(t, f, fakeToken, "publish", "-version", "paper-a/aaa"); code != 0 || !strings.Contains(out, "published as") {
		t.Errorf("publish exits %d saying %q, %q", code, out, errs)
	}
}

func TestTheCommandRefusesBeforeAnyRequest(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "zenodo.json")
	for _, tc := range []struct {
		name, token, why string
		args             []string
		creators         string // the creators of a description given with -metadata, or none
	}{
		{"no token", "", "ZENODO_TOKEN", []string{"publish", "-version", "paper-a/aaa"}, ""},
		{"no version", fakeToken, "-version", []string{"publish"}, ""},
		{"no archive", fakeToken, "-archive", []string{"draft", "-version", "paper-a/aaa"}, ""},
		{"no step", fakeToken, "usage", []string{"upload"}, ""},
		{"a placeholder for an author", fakeToken, "placeholder", []string{"draft", "-version", "paper-a/aaa", "-archive", "x"},
			`[{"name": "[Author]", "affiliation": "[Affiliation]"}]`},
		{"an unknown flag", fakeToken, "flag provided but not defined: -verbose", []string{"publish", "-version", "paper-a/aaa", "-verbose"}, ""},
		{"a stray argument", fakeToken, "usage", []string{"publish", "-version", "paper-a/aaa", "now"}, ""},
		{"a description that is not there", fakeToken, "the record's description: open " + missing, []string{"draft", "-version", "paper-a/aaa", "-archive", "x", "-metadata", missing}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeZenodo(t)
			args := tc.args
			if tc.creators != "" {
				args = append(args, "-metadata", metadataFile(t, tc.creators))
			}
			code, _, errs := runZenodo(t, f, tc.token, args...)
			if code != 2 || !strings.Contains(errs, tc.why) {
				t.Errorf("exits %d complaining %q, want 2 and %q", code, errs, tc.why)
			}
			f.set(func() {
				if len(f.requests) != 0 {
					t.Errorf("it asked Zenodo %v before refusing", f.requests)
				}
			})
		})
	}
}

// recordFile writes a description of the record with one field set to a value,
// or as it is with no field given.
func recordFile(t *testing.T, field string, value any) string {
	t.Helper()
	record := map[string]any{
		"upload_type": "software", "title": "The artifact", "description": "Code and data.",
		"creators": []any{map[string]any{"name": "Stub, Ann"}}, "access_right": "open", "license": "mit",
	}
	if field != "" {
		record[field] = value
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "zenodo.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTheRecordsDescriptionIsHeldToWhatARecordSays(t *testing.T) {
	t.Parallel()
	if _, err := ReadMetadata(recordFile(t, "", nil)); err != nil {
		t.Fatalf("a description with nothing wrong gives %v", err)
	}
	for _, tc := range []struct {
		name, field string
		value       any
	}{
		{"a field Zenodo does not take", "edition", "1"},
		{"a version, which the step sets", "version", "1"},
		{"another kind of upload", "upload_type", "poster"},
		{"access that is not open", "access_right", "closed"},
		{"no licence", "license", ""},
		{"no title", "title", " "},
		{"no creators", "creators", []any{}},
		{"a creator with no name", "creators", []any{map[string]any{"name": " "}}},
		{"an affiliation still to come", "creators", []any{map[string]any{"name": "Stub, Ann", "affiliation": "[Affiliation]"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ReadMetadata(recordFile(t, tc.field, tc.value)); err == nil {
				t.Errorf("%s set to %v reads with no error", tc.field, tc.value)
			}
		})
	}
}

func TestTheRepositorysDescriptionIsOneTheDraftStepSends(t *testing.T) {
	t.Parallel()
	if _, err := ReadMetadata(filepath.Join("..", "..", "release", "zenodo.json")); err != nil {
		t.Errorf("release/zenodo.json gives %v, want a description the draft step sends", err)
	}
}

func TestALinkThatLeavesZenodoIsNotFollowed(t *testing.T) {
	t.Parallel()
	f := newFakeZenodo(t)
	draft(t, f, "paper-a/aaa", "", "first")
	publish(t, f, "paper-a/aaa", "")
	f.set(func() { f.elsewhere = true })
	r := DraftRequest{Metadata: metadata(), Archive: archive(t, "second"), Version: "paper-a/bbb", Record: "1001", Today: "2026-10-08"}
	err := Draft(context.Background(), f.client(), &r, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "is not under") {
		t.Errorf("a draft at another address gives %v, want a refusal to follow it with the token", err)
	}
}
