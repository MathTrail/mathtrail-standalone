package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// records is a bibliography with one good entry and one of each fault the
// check exists to catch.
const records = `@article{good,
  author  = {Klinkenberg, S. and Straatemeier, M. and van der Maas, H.L.J.},
  title   = {Computer adaptive practice of Maths ability using a new item response model for on the fly ability and difficulty estimation},
  journal = {Computers \& Education},
  year    = {2011},
  doi     = {10.1016/j.compedu.2011.02.003},
}

@article{changedtitle,
  author  = {Klinkenberg, S. and Straatemeier, M. and van der Maas, H.L.J.},
  title   = {Adaptive practice of maths},
  journal = {Computers \& Education},
  year    = {2011},
  doi     = {10.1016/j.compedu.2011.02.003},
}

@misc{extraauthor,
  author        = {Gupta, Adit and Reddig, Jennifer and Calo, Tommaso and Weitekamp, Daniel and MacLellan, Christopher J. and Koedinger, Kenneth},
  title         = {Beyond Final Answers: Evaluating Large Language Models for Math Tutoring},
  year          = {2025},
  eprint        = {2503.16460},
  archiveprefix = {arXiv},
}

@article{nosuchdoi,
  author = {Nobody, N.},
  title  = {Nothing},
  year   = {2020},
  doi    = {10.9999/no-such-doi},
}

@article{retracted,
  author = {Wang, Q. and Zhang, J. and Cui, L.},
  title  = {Downregulation of long noncoding RNA LINC01419},
  year   = {2019},
  doi    = {10.1177/1758835919874651},
}
`

func writeBib(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "refs.bib")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckCommandReportsEveryFault(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", writeBib(t, records), "check"}, fixtureRegistries(t), &stdout, &stderr)

	if status != 1 {
		t.Errorf("status = %d, want 1; stderr: %s", status, &stderr)
	}
	for _, want := range []string{
		"good: ok",
		"changedtitle: title",
		"extraauthor: authors",
		"nosuchdoi: no registry holds 10.9999/no-such-doi",
		"retracted: Crossref records a retraction",
		"5 entries, 4 with problems",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, &stdout)
		}
	}
}

func TestCheckCommandPassesACleanBibliography(t *testing.T) {
	t.Parallel()
	good, _, _ := strings.Cut(records, "\n\n")
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", writeBib(t, good), "check"}, fixtureRegistries(t), &stdout, &stderr)

	if status != 0 {
		t.Errorf("status = %d, want 0; output: %s%s", status, &stdout, &stderr)
	}
}

func TestAddCommandWritesEntriesFromTheRecordsOnly(t *testing.T) {
	t.Parallel()
	path := writeBib(t, "% The bibliography.\n")
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", path, "add",
		"https://doi.org/10.1016/j.compedu.2011.02.003", "arXiv:2503.16460v1", "10.1177/1758835919874651", "10.9999/no-such-doi",
	}, fixtureRegistries(t), &stdout, &stderr)

	if status != 1 {
		t.Errorf("status = %d, want 1: two of the four could not be added; output: %s%s", status, &stdout, &stderr)
	}
	for _, want := range []string{
		"https://doi.org/10.1016/j.compedu.2011.02.003: added as klinkenberg2011computer",
		"arXiv:2503.16460v1: added as gupta2025beyond",
		"10.1177/1758835919874651: not added: Crossref records a retraction",
		"10.9999/no-such-doi: not added",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, &stdout)
		}
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ParseBib(string(written))
	if err != nil || len(entries) != 2 || !strings.HasPrefix(string(written), "% The bibliography.\n\n@article{klinkenberg2011computer,") {
		t.Fatalf("the file holds %d entries, %v:\n%s", len(entries), err, written)
	}

	var again bytes.Buffer
	status = run(context.Background(), []string{"-bib", path, "add", "10.48550/arXiv.2503.16460"}, fixtureRegistries(t), &again, &stderr)
	if status != 1 || !strings.Contains(again.String(), "already in the file as gupta2025beyond") {
		t.Errorf("adding a work twice: status %d, output %q; want it refused", status, &again)
	}
	var checked bytes.Buffer
	if status := run(context.Background(), []string{"-bib", path, "check"}, fixtureRegistries(t), &checked, &stderr); status != 0 {
		t.Errorf("the entries add wrote do not check out: status %d\n%s", status, &checked)
	}
}

func TestRunRefusesWhatItCannotDo(t *testing.T) {
	t.Parallel()
	for name, args := range map[string][]string{
		"no command":        {},
		"an unknown one":    {"verify"},
		"add with no ids":   {"add"},
		"check with an id":  {"check", "10.1/x"},
		"a missing file":    {"-bib", filepath.Join(t.TempDir(), "none.bib"), "check"},
		"a broken file":     {"-bib", writeBib(t, "@misc{a, title = {x}"), "check"},
		"an unknown flag":   {"-verbose", "check"},
		"a broken file add": {"-bib", writeBib(t, "@misc{a"), "add", "10.1/x"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), args, fixtureRegistries(t), &stdout, &stderr); status != 2 {
				t.Errorf("status = %d, want 2", status)
			}
		})
	}
}

func TestRunSaysWhenARegistryDidNotAnswer(t *testing.T) {
	t.Parallel()
	failing := "@article{down,\n  author = {A, B},\n  title = {T},\n  year = {2020},\n  doi = {10.5555/fail},\n}\n"
	var checked, added, stderr bytes.Buffer

	checkStatus := run(context.Background(), []string{"-bib", writeBib(t, failing), "check"}, fixtureRegistries(t), &checked, &stderr)
	addStatus := run(context.Background(), []string{"-bib", writeBib(t, ""), "add", "10.5555/fail"}, fixtureRegistries(t), &added, &stderr)

	if checkStatus != 2 || !strings.Contains(checked.String(), "down: not checked") {
		t.Errorf("check: status %d, output %q; want 2 and the entry not checked", checkStatus, &checked)
	}
	if addStatus != 2 || !strings.Contains(added.String(), "10.5555/fail: not added") {
		t.Errorf("add: status %d, output %q; want 2 and the work not added", addStatus, &added)
	}
}

func TestAddRefusesAWorkAnEntryCitesByItsAddress(t *testing.T) {
	t.Parallel()
	path := writeBib(t, "@article{old,\n  doi = {https://doi.org/10.1016/j.compedu.2011.02.003},\n}\n")
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", path, "add", "10.1016/J.COMPEDU.2011.02.003"}, fixtureRegistries(t), &stdout, &stderr)

	if status != 1 || !strings.Contains(stdout.String(), "already in the file as old") {
		t.Errorf("status %d, output %q; want the work refused as already there", status, &stdout)
	}
}

// An entry with nothing to look it up by cannot be checked, and the check
// says so without asking any registry: the registries here have no HTTP
// client, and asking one would panic.
func TestCheckCommandReportsAnEntryThatCannotBeLookedUp(t *testing.T) {
	t.Parallel()
	bib := writeBib(t, "@misc{web2024page,\n  title = {A page},\n  url   = {https://example.org},\n}\n")
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", bib, "check"}, Registries{}, &stdout, &stderr)

	want := "web2024page: cannot be checked: it has neither a DOI nor an arXiv id to check it by\n1 entries, 1 with problems\n"
	if status != 1 || stdout.String() != want {
		t.Errorf("check = %d, %q; want 1 and %q", status, &stdout, want)
	}
}

// hangingUp is a registry that closes every connection without an answer.
func hangingUp(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Error("the stand-in cannot take the connection over")
			return
		}
		if conn, _, err := hijacker.Hijack(); err == nil {
			_ = conn.Close()
		}
	}))
	t.Cleanup(server.Close)
	return server
}

// A registry that does not answer has said nothing of the work: the entry is
// not checked, the run fails as one that could not do its job, and DataCite,
// which would say it holds no such DOI, is not asked instead.
func TestCheckCommandTellsARegistryThatDidNotAnswerFromAMissingWork(t *testing.T) {
	t.Parallel()
	good, _, _ := strings.Cut(records, "\n\n")
	datacite := answering(t, http.StatusNotFound, "")
	registries := Registries{Client: datacite.Client(), Crossref: hangingUp(t).URL, DataCite: datacite.URL, UserAgent: "citecheck-test"}
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", writeBib(t, good), "check"}, registries, &stdout, &stderr)

	if status != 2 || !strings.Contains(stdout.String(), "good: not checked: crossref:") || strings.Contains(stdout.String(), "no registry holds") {
		t.Errorf("check = %d, %q; want 2 and the entry not checked", status, &stdout)
	}
	if n := datacite.asked.Load(); n != 0 {
		t.Errorf("DataCite was asked %d times, want none", n)
	}
}

// add starts a bibliography that is not there yet: the file it writes holds
// the new entry alone.
func TestAddCommandStartsABibliographyThatIsNotThere(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "refs.bib")
	var stdout, stderr bytes.Buffer

	status := run(context.Background(), []string{"-bib", path, "add", "10.1016/j.compedu.2011.02.003"}, fixtureRegistries(t), &stdout, &stderr)

	if status != 0 {
		t.Fatalf("add = %d, want 0; output: %s%s", status, &stdout, &stderr)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ParseBib(string(written))
	if err != nil || len(entries) != 1 || !strings.HasPrefix(string(written), "@article{klinkenberg2011computer,\n") {
		t.Errorf("the file holds %d entries, %v:\n%s\nwant the one entry added", len(entries), err, written)
	}
}

// A bibliography add cannot read, or cannot write where it is, fails the
// command as one that could not do its job, and names the file.
func TestAddCommandFailsOnABibliographyItCannotReadOrWrite(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		path func(dir string) string
		want string
	}{
		{"a directory", func(dir string) string { return dir }, "read "},
		{"a file in a directory that is not there", func(dir string) string { return filepath.Join(dir, "missing", "refs.bib") }, "write "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := tc.path(t.TempDir())
			var stdout, stderr bytes.Buffer

			status := run(context.Background(), []string{"-bib", path, "add", "10.1016/j.compedu.2011.02.003"}, fixtureRegistries(t), &stdout, &stderr)

			if want := "citecheck: " + tc.want + path; status != 2 || !strings.Contains(stderr.String(), want) {
				t.Errorf("add = %d, stderr %q; want 2 and %q", status, &stderr, want)
			}
		})
	}
}
