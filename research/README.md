# Research

MathTrail's research program: one paper, paper A. A second paper, about an evolving textbook, was planned and has been dropped (Q60).

- **Paper A** covers MathTrail as it is built. The chat's own model writes each task. A small deterministic service chooses what to practise, admits only tasks that a sandboxed brute-force solver confirms, and keeps the answer sealed until the child answers.

The plan is [RUN.md](RUN.md), in Russian like the product's plan. It holds:
- tasks S00–S66 (paper B's tasks were removed, so the numbers have gaps);
- the rules G1–G10;
- the author's decisions Q01–Q20;
- the questions only the author can answer;
- the log of the autonomous run.

Everything else in this folder is in English, except the Russian copy of the paper draft (Q59).

## Layout

What later tasks will create is listed too, with the task that writes it.

| Path | What it holds | Written by |
|---|---|---|
| `RUN.md` | The plan, its rules, decisions, critical questions and run log | the author and every task |
| `questions.md` | Decisions made while running the tasks, from Q21 on | every task |
| `ai-use-log.md` | What AI tools did in each task and what a person checked (G7) | every task |
| `evidence/` | The claims about MathTrail (`ledger.md`) and its prototype (`prototype.md`), with their proof (G1), the claims files and the scripts that compute facts at a commit; the claims of the service's real use, which began after the product's pin, are proven at a later commit of their own that `use-stats.txt` names (Q88); `premises.md` audits the Gemini prompt | S02–S04 |
| `literature/` | Search protocol, notes per area, `refs.bib`, the unverified `seed.bib` from the prototype's notes, the numbers taken from the works cited (`numbers.txt`), venues, the writing standard, exemplar papers | S01, S03, S06, S08–S10, S16, S31–S35 |
| `experiments/` | Protocols, each frozen with an OpenTimestamps proof beside it (`.ots`), and the code, data and results of the experiments | S07, S36–S46 |
| `paper-a/` | Paper A: the LaTeX source (`main.tex`, `preamble.tex`, `sections/`, `figures/`), the numbers it cites and the version the named build prints (`generated/numbers.tex`, `generated/version.tex`, written by the build), the names that carry digits the check of numbers typed by hand lets through (`names-with-digits.txt`), the 14-page text the sections are typeset from (`draft.md`) and the long text it was cut from (`draft-extended.md`, Q80), each with a Russian copy for the author (`draft.ru.md`, `draft-extended.ru.md`, Q59); the PDFs go to `build/` (Q67) | the first draft (S56–S63), S55 on |
| `release/` | The artifact of paper A: the script that assembles it from the product at the pin and the experiments (`assemble.sh`), the one that shows it reproduces a result on its own (`check.sh`), its README and the note its anonymous copy carries; the tarballs go to `release/build/` (Q84) | S65 |
| `containers/` | Pinned images for what the research runs outside the devcontainer, such as the clients of other model families (Q55) | S07 on |
| `tools/` | `ledger`, which resolves and renders the claims and writes the numbers they state (S63); `handtyped`, which refuses a number typed by hand into the paper's LaTeX (S63); `isolation`, which shows what a model's client would offer the model, without calling it; `macros`, which turns computed numbers into the macros the paper cites them by (S55); `citecheck`, which writes and checks the bibliography against Crossref and DataCite, and whose `cited` command refuses a cited key no checked bibliography holds (S55); `live`, which writes what the paper prints of the service's real use from its monthly snapshot of public totals, `site/research/live.json`, and refuses one that shows what the public views hide (Q88) | S02, S07, S08, S55 |
| `archive/` | Local copies of files outside git, ignored by git: memory for this machine, not evidence (Q07) | S02, S03 |
| `.cache/` | A clone of the prototype's public repository, which the ledger reads its proofs from; ignored by git, `just research prototype-fetch` recreates it (Q31) | S03 |
| `go.mod`, `justfile`, `wiring_test.go` | The research module | S00 |

## The research module

- **What it is.** Code for experiments and tools lives in a separate Go module, `github.com/MathTrail/mathtrail-standalone/research`. It imports the product's packages, including those under `internal/`, at `v0.1.53`, the tag on the commit the paper's experiments were run on (Q87), so experiments exercise the product's own code as the paper describes it, rather than a copy of it or whatever the working tree holds now. The product never imports the research module.
- **What the product's tooling sees.**
  - The product's `ci-*` recipes and the pre-commit hook's build and tests do not see the module.
  - Anything that walks files rather than Go packages does: `just fmt` and `just fmt-check` (gofmt over the whole tree), the pre-commit hook's gofmt, and gitleaks.
  - CI runs the module's own checks on every pull request and every commit of `main`, and a red one blocks the merge as the product's do (R231):
    - `just research ci-lint` in the toolchain image, beside the product's lint;
    - `just research ci-test` in the toolchain image too, as a check of its own, "Research tests", since one experiment's test takes minutes under the race detector;
    - `just research ci-paper` on the runner, as the check "Paper A and its evidence", since it needs Docker, the whole history and the network.
  - CodeQL builds the module after the product and reports on its code. Its coverage goes to Codecov under the flag `research`, which decides nothing, and to SonarCloud, which reads all but `draft-ui/`.
  - `.github/workflows/research.yml` holds the experiments of paper A to the results kept here (R234). On every pull request and every release, the checks "Reproduce faultinject", "Reproduce learnersim" and "Reproduce perf" run an experiment again whenever its program or its results changed, as `just research reproduce` does, and a run that gives other bytes is red. E-A4's times stay those of the run the paper describes, and only its package sizes are compared. A release and every pull request then build paper A's named and anonymous PDF, with the paper's own checks, and hand them on as the artifact `paper-a` (R239). The named one prints its version, the date and the short hash of the commit, in a footnote on its first page. No Russian PDF is built. From the same commit comes the data of the site's page "Research", which takes the named PDF in, whatever placeholder it still prints, and both go to the site's build; a pull request's site is built from them and published by no one (R248).
- **Checks**, run from the repository root:

  ```sh
  just research test          # tests, with the race detector
  just research lint          # gofmt -s and golangci-lint over this module, ShellCheck over its scripts
  just research ci-lint       # what a change must pass with Go and golangci-lint: gofmt -s and golangci-lint, licenses, deps-check
  just research ci-test       # the tests, with a coverage profile in coverage.out
  just research ci-paper      # what a change must pass with Docker: ShellCheck, ledger-check, paper-a
  just research reproduce faultinject  # run an experiment again and hold it to its results: every file of E-A1 and E-A3, E-A4's package sizes
  just research experiment-key faultinject  # the key a reproduction is kept under: the program, the processor, the comparison and the results
  just research vuln          # govulncheck, pinned by the product's justfile
  just research licenses      # go-licenses against the product's allow-list (R21)
  just research tidy          # go.sum after the product's go.mod changed
  just research deps-check    # every module whose packages both builds compile, at the product's version
  just research ledger-check  # every claim, product and prototype, still proven; both tables up to date
  just research seed-check    # every link of the prototype's notes is in literature/seed.bib
  just research citecheck     # every entry of literature/refs.bib as Crossref or DataCite records it
  just research paper-a       # paper A: its numbers, its citations and names checked, the named and the anonymous PDF, no line past the margins and the anonymous one in 14 pages
  just research paper-a-final # the anonymous PDF for submission; fails while a TBD is left
  just research paper-a-figure-check  # a changed CSV changes a figure; the same sources make the same PDF
  just research timestamp FILE        # freeze a protocol: its SHA-256, and an OpenTimestamps proof as FILE.ots beside it; FILE is a path inside research/, such as experiments/PROTOCOL-A-offline.md
  just research timestamp-upgrade FILE  # complete that proof with its Bitcoin attestation, hours later
  just research faultinject   # experiment E-A1: defects injected into the reference tasks, results in experiments/faultinject/results
  just research learnersim    # experiment E-A3: simulated learners under the service and the baselines, results in experiments/learnersim/results
  just research perf          # experiment E-A4: what a review costs on this machine, and the package's size, results in experiments/perf/results
  just research release [named|anonymous]        # the artifact of paper A as a tarball in release/build; anonymous for double-blind review
  just research release-check [named|anonymous]  # unpack the artifact, rerun the experiments from it and compare what they write with the results it ships
  ```

- **A new Go dependency** of the research module needs `just research licenses`, `just research vuln` and `just research deps-check` to pass (Q26, Q27). Experiments live in non-test packages: go-licenses does not read what test files import.

- **Figures** are drawn by pgfplots from CSV files, inside the paper build. Python is used only when a task needs a library Go lacks, and runs in a pinned container. Tools that only build a paper or run an analysis in a container are dependencies of no code in the repository, so the product's licence list does not govern them; S55 records their licences with the build (Q24).

## Evidence and numbers

The full rules are G1–G10 in [RUN.md](RUN.md). In short:

- **Claims.** Every claim about MathTrail or its prototype points to the ledger. The proof is `file:line` at a pinned commit, a fact a script computed at that commit, or a quote and hash of a file in the public prototype repository at a pinned commit.
- **Numbers.** Every number in a paper comes from a saved script over saved data or, when it is another work's, from `literature/numbers.txt`, which names the passage it comes from. It reaches the LaTeX only through generated macros: `\stat{source}{key}`, which stops the build when nothing defines it (Q67), and its relatives that round it or spell it out (Q81). The build refuses a number typed by hand; a number the text chooses, such as a confidence level, is marked `\given`.
- **Citations.** Every citation is checked against its Crossref or DataCite record by `just research citecheck` (Q61). No child data is ever used.
