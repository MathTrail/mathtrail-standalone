package main

import (
	"bytes"
	"context"
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

func TestDOIFromID(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ id, want string }{
		{"10.1016/j.compedu.2011.02.003", "10.1016/j.compedu.2011.02.003"},
		{"doi:10.1016/j.compedu.2011.02.003", "10.1016/j.compedu.2011.02.003"},
		{"https://doi.org/10.1016/J.COMPEDU.2011.02.003", "10.1016/J.COMPEDU.2011.02.003"},
		{"arXiv:2503.16460v2", "10.48550/arXiv.2503.16460"},
		{"2402.15861", "10.48550/arXiv.2402.15861"},
		{" 10.48550/arXiv.2402.15861 ", "10.48550/arXiv.2402.15861"},
		{"http://doi.org/10.5281/zenodo.3554625", "10.5281/zenodo.3554625"},
		{"https://dx.doi.org/10.1016/j.compedu.2016.03.017", "10.1016/j.compedu.2016.03.017"},
		{"https://arxiv.org/abs/2404.18796v2", "10.48550/arXiv.2404.18796"},
		{"https://arxiv.org/pdf/2404.18796.pdf", "10.48550/arXiv.2404.18796"},
	} {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			if got := DOIFromID(tt.id); got != tt.want {
				t.Errorf("DOIFromID(%q) = %q, want %q", tt.id, got, tt.want)
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
