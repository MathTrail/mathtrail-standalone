# Research decisions and questions

This file holds the decisions made while running the research tasks, numbered from Q21.

Other questions are kept elsewhere:
- Q01–Q20 — the author's answers and the defaults the author delegated to the executor — are in [RUN.md](RUN.md), in the section «Решения по открытым вопросам».
- Critical questions only the author can answer, the ones that block work, are in RUN.md, «Критические вопросы автору».
- Questions about the author's reasons and intentions that block nothing are collected at the end of this file, marked "answer when convenient".

## Format

Each entry is a heading **Qnn. Title** followed by the fields S00 of the plan asks for:

- **Question** — what had to be decided;
- **Why it matters** — what goes wrong without an answer;
- **Options** — the ones considered, when there was a real choice;
- **Recommendation** — kept while a question is still open, dropped once decided;
- **Decision** — what was chosen and why; an answer from the author is quoted;
- **Who and when** — the author or the executor, the task, the date;
- **Blocks** — the tasks that cannot start or finish without it, or the tasks it affects once decided.

One field is added to the plan's list: **Sources** — files, pages or papers the decision rests on, which rule G1 asks every claim to have.

A decision can be revised by a later task on new facts; the revision is a new entry that names the one it replaces.

## Decisions

### Q21. The research module is a nested module that imports the product

- **Question:** where the research Go code lives. Q06 set the default; S00 had to prove it works.
- **Why it matters:** without it no experiment can call the product's own code, and would have to re-implement it.
- **Decision:** a separate module, `github.com/MathTrail/mathtrail-standalone/research`, with `replace github.com/MathTrail/mathtrail-standalone => ../`.
  - Go's rule for `internal/` is about import paths. The research module's path lies inside the product's, so it may import `internal/` packages.
  - `wiring_test.go` imports `internal/domain/rating` and passes.
  - The product's package-based tooling does not see the module:
    - `go list ./...` at the root lists no research package;
    - CodeQL analyses what `go build ./...` at the root compiles, which excludes the nested module.
  - What walks files rather than packages does see it: the root's `fmt` and `fmt-check`, run by `ci-lint`, format every Go file of the repository, and so do the pre-commit hook and gitleaks. `just research lint` checks the module's formatting, so a research task does not leave the product's `ci-lint` red.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** S08, S37 and every task with code.
- **Sources:** `research/go.mod`, `research/wiring_test.go`, `.github/workflows/codeql.yml`.

### Q22. The research recipes live in `research/justfile`

- **Question:** how the research module is built and checked without growing the product's justfile.
- **Why it matters:** without it, each research task would grow the product's justfile, which another session edits in parallel.
- **Decision:** `research/justfile` holds the recipes: `test`, `lint`, and later ones such as `citecheck` and the paper builds. The product's justfile gets a single delegating recipe, so `just research <recipe>` works from the root. Tool versions and the license allow-list are read from the root justfile, not repeated. Formatting is checked inside the module rather than through the root's fmt-check, which walks the whole repository and would fail a research task on product files. The lint recipe uses the repository's `.golangci.yml`, which golangci-lint finds in the parent directory.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** every task with code or a build.
- **Sources:** `justfile`, `research/justfile`.

### Q23. What the product's tooling excludes

- **Question:** keep research files out of the product's image, quality gate and history where they do not belong.
- **Why it matters:** research code would otherwise land in the runtime image's build context and in the product's quality gate.
- **Decision:**
  - `.dockerignore` excludes `research`;
  - `sonar.exclusions` and `sonar.test.exclusions` add `research/**`, so the product's quality gate — which the product's plan must clear before every task — is not blocked by research code;
  - `.gitignore` ignores `research/archive/`;
  - the comment on `go.work` now explains the nested module.

  Codecov needs no change: the product's coverage profile comes from `go test ./...` at the root, which never includes the research module.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** all tasks.
- **Sources:** `.dockerignore`, `.gitignore`, `sonar-project.properties`, `codecov.yml`.

### Q24. How figures are drawn

- **Question:** which tool draws the papers' figures.
- **Why it matters:** the papers' figures must be redrawn by a script from saved data (G2), never edited by hand.
- **Options:** pgfplots from CSV; Go plotting libraries; matplotlib in a container.
- **Decision:** pgfplots from CSV by default. The paper build draws the figures, so fonts and sizes match the text and a changed CSV changes the figure with no manual step. Python only when a task needs a library Go lacks, in a pinned container.
  - R21's list governs the dependencies of the repository's code: the modules the research module imports are held to it (Q27).
  - Tools that only build a paper or run an analysis in a container, such as TeX Live, pgfplots (LPPL) or a Python library, are dependencies of no code in the repository, so R21 does not apply to them. S55 records their licences with the build.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** S37 onwards, S55.
- **Sources:** `docs/decisions.md` (R21).

### Q25. CI does not run the research checks yet

- **Question:** the product's CI builds and tests only the root module, so a product change can break the research module without anyone noticing. The product checks that do reach research files walk files rather than packages: formatting, the pre-commit hook and gitleaks (Q21).
- **Why it matters:** a product change can break an experiment silently.
- **Options:**
  - a research job in the product's CI — a broken experiment would then block product merges;
  - a separate, non-required workflow — it cannot be tried out locally tonight, and a half-working workflow marks every commit red;
  - no CI for now.
- **Decision:** no CI for now. Every research task that touches code starts with `just research test`, so a break is caught at the next research task rather than in production. Revisit at S37, when the module holds its first real experiment: a separate workflow triggered by changes to `research/**` or `internal/**`.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** S08, S37 onwards.
- **Sources:** `.github/workflows/ci.yml`, `.github/workflows/codeql.yml`.

### Q26. `research/go.sum` follows the product's go.mod by hand

- **Question:** the research module's go.sum holds hashes of the product's dependencies, and Dependabot watches only the root.
- **Why it matters:** the research module stops building after a dependency bump in the product.
- **Decision:** `just research tidy` after the product's go.mod changes. A stale go.sum fails loudly ("missing go.sum entry") at the next `just research test`, which every code task runs first (Q25).
  - A quieter drift is version selection. If a dependency of an experiment needs a newer version of a module the product also uses, Go picks the newer one, and the product's packages then run in the research module against a version the product does not ship. Nothing fails, but the experiment measures a different program.
  - `just research deps-check` catches it: it fails when a module both builds use is selected at different versions. Every task that adds a dependency runs it.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** every task with code.
- **Sources:** `research/go.mod`, `.github/dependabot.yml`.

### Q27. The research module's dependencies get the product's license and vulnerability checks

- **Question:** the root's `ci-licenses` and `ci-vuln` never see the nested module.
- **Why it matters:** the project admits only MIT-compatible dependencies into the repository (R21), and the research module is part of the repository, although it builds nothing that is distributed.
- **Decision:** `just research licenses` and `just research vuln` run the same pinned tools with the same allow-list, each read from the product's justfile by name with `just --evaluate`, which fails when the name is missing. Any task that adds a Go dependency to the research module runs both.
  - The module's own packages fall under the repository's LICENSE, which go-licenses does not look for above a module's root, so the check leaves them out; what they import is still checked.
  - `vuln` passes `-test`, so imports of test files are scanned too.
  - go-licenses reads only non-test imports. Experiments therefore live in non-test packages, and test files import only the standard library, the product and what non-test code already imports.
- **Who and when:** executor, S00, 2026-09-25.
- **Blocks:** S08 and every task that adds a dependency.
- **Sources:** `justfile` (`ci-licenses`, `ci-vuln`, `GO_LICENSES`, `ALLOWED_LICENSES`), `research/justfile`.

### Q28. Venues and the schedule

- **Superseded** for paper A by Q58 (AIED 2027, paper A first); paper B has left the plan (Q60), and with it the schedule this entry sets.
- **Question:** where each paper goes and in what order (S01; Q08 set the default).
- **Why it matters:** page limits, templates, anonymity and dates shape every writing task and the order of the whole run.
- **Options:** AIED, EDM or L@S 2027 for A, or journals; EC-TEL 2027, iTextbooks, Blue Sky or journals for B; A first or B first.
- **Decision:**
  - Paper A → AIED 2027 main track, full paper: 14 pages including references, Springer LNAI, double-blind with the system's name removed. EDM or L@S of the same season are alternatives chosen at submission, not fallbacks: they close before AIED decides. After a rejection: IJAIED or AIED 2028; EC-TEL 2027 only if its deadline falls after AIED's decisions.
  - Paper B → EC-TEL 2027. iTextbooks 2027 only for a distinct contribution, after checking dual-submission rules. Both papers can grow into IJAIED articles.
  - Order of the run: the table's order stays for now. The switch Q08 allows has a concrete trigger: if S31–S39 are not done by 31 October 2026, the executor switches to the "A first" order and records it in the log. What limits AIED 2027 most is the author's part (K01, K02, K04: the generation study and the blind review), not the order of the free tasks.

  The 2027 dates are expected ones, inferred from 2026; each writing task confirms them on the official site.
- **Who and when:** executor, S01, 2026-09-25.
- **Blocks:** S05, S06, S55 onwards.
- **Sources:** [literature/venues.md](literature/venues.md).

### Q29. Where the drafts live, per paper (Q05 applied)

- **Superseded** for paper B, which left the plan (Q60); paper A's drafts stay in `research/paper-a/`.
- **Question:** Q05 left it to S01 to decide, per venue, whether drafts may sit in the public repository during review.
- **Why it matters:** a public, identifiable draft during double-blind review can break a venue's rules.
- **Decision:**
  - Paper A: drafts in `research/paper-a/`, public. AIED and L@S allow preprints; only the anonymous build is submitted. EDM, the other alternative of Q28, states no rule on preprints, so before paper A goes to EDM, its rule is checked first.
  - Paper B: its drafts start at S67, after S55 has read EC-TEL's rule on preprints and public drafts, and S55 decides then whether they live in `research/paper-b/` or somewhere private. No task writes paper B's text before that. Ignoring the folder in git instead would keep the drafts out of review and out of every backup.
- **Who and when:** executor, S01, 2026-09-25.
- **Blocks:** S55, S67.
- **Sources:** [literature/venues.md](literature/venues.md).

### Q30. Which commit the ledger pins

- **Question:** the ledger's proofs are lines at one commit. Which commit, while the product keeps changing and the paper is not yet written?
- **Why it matters:** a reviewer must be able to open the pinned commit. The repository merges pull requests by squashing, so a commit of a working branch never reaches `main`: after the merge it survives only through the pull request's refs on GitHub and in local clones.
- **Options:**
  1. Pin the newest pushed commit now, and move the pin with `just research pin` and `just research ledger` as the product changes.
  2. Wait for a commit on `main` before starting the ledger.
- **Decision:** option 1. The ledger pins `3597c6a0633d`, the newest pushed commit on 2026-09-25, of the branch `standards-5–6-weighings-and-transfers`. S56 pins the final commit, and that one must be on `main`.
  - Moving is cheap, because the tool names exactly the claims a move breaks. Going from `e1c315303bf6` to `3597c6a0633d` broke one claim: C057, when the corpus grew from 594 to 603 reference tasks.
- **Who and when:** executor, S02, 2026-09-25.
- **Blocks:** S56.
- **Sources:** [evidence/ledger.md](evidence/ledger.md); `git log --oneline main` (the merges are squashed, "(#34)").

### Q31. How a claim about the prototype is proven

- **Question:** G1 asks that a proof from a file outside git carry a quote, the file's SHA-256 and a commit of the public prototype repository where the file matches byte for byte. Which commit is that, and what form does each proof take?
- **Why it matters:** the local `prototype/` and `docs/prototype/` go away, and a claim that points only at them could never be checked again.
- **Options:**
  1. Quote, hash and link written by hand for each claim.
  2. The product ledger's own form: `path:line@commit` in the public repository, resolved by the same tool from a clone.
- **Decision:** option 2, in [evidence/prototype.md](evidence/prototype.md). One tool and one format serve both parts of the ledger, and a proof that is gone or ambiguous fails the run.
  - The three things G1 asks for are there, one of them in another form:
    - the link is the commit;
    - the hash is in a manifest of every file at the commit;
    - the quote is the claim's `match` in the claims file: the fragment of the line, with regular-expression escapes. The published table shows the place, not the quote. The line at a commit that cannot change is the quote's source, and showing it would put the prototype's Russian into this folder's English files.
  - The commit is `02638353482e`, the head of `main` (2026-09-16). 120 of the 131 local files match it byte for byte. The rest are Python byte-code caches, which are not kept, and the product's own index of the frozen copies, which is archived.
  - Claims are numbered P001…. A number or finding a research note reports from the literature has the status *unverified* until a later task checks the original.
  - The clone lives in `research/.cache/`, which git ignores.
- **Who and when:** executor, S03, 2026-09-25.
- **Blocks:** S04, and every task that cites the prototype.
- **Sources:** [evidence/prototype-snapshot.txt](evidence/prototype-snapshot.txt), [evidence/prototype-files.sha256](evidence/prototype-files.sha256).

### Q32. Why the solver is Starlark, and why it runs twice

- **Question:** S05, block 1: why an embedded Starlark interpreter, and why a second run with the option letters moved.
- **Why it matters:** paper A's verification section has to state what the design guarantees and what it does not.
- **Decision:** answered from the repository.
  - Starlark runs inside the service's own process: no container, no network, no second service. The language has no imports, no file access, no clock and no threads, so the sandbox is the language, and what is left to configure is the dialect and the limits (SPEC 6.1).
  - The limits are a step budget, a time limit and a cap on concurrent runs (C024). Memory has no limit of its own: the step budget bounds only what the service's helpers build, and a program that doubles a list with its own operators can still exhaust memory (SPEC 6.6). Paper A states that gap.
  - The second run moves every option text to another letter (A→C, B→D, C→E, D→A, E→B), and both runs must pick the same text (SPEC 6.2, C023). A solver that returns a letter it wrote down fails the second run; one that computes a value and looks it up passes both. It does not stop a solver that writes down the value itself, and it is not meant to.
  - Paper A states both halves: what the second run removes, and what it leaves.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S58.
- **Sources:** SPEC 6.1, 6.2, 6.6; C023, C024; `docs/decisions.md` R44.

### Q33. Readability outside English

- **Question:** S05, block 1: how the readability check treats a language other than English.
- **Why it matters:** the host writes in the chat's language, and E-A2 runs in several.
- **Decision:** answered from the repository (SPEC 5.5, C030).
  - Every language is held to a longest sentence: in words where words are separated by spaces, in characters for the scripts written without them.
  - Flesch–Kincaid applies to English only.
  - The character limits are twice the word limits by analogy, and nobody has measured them.
  - E-A2 therefore reports readability refusals by script (S40), and paper A calls the character limits unmeasured.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S40, S58.
- **Sources:** SPEC 5.5; C030, C049.

### Q34. The near-duplicate threshold

- **Question:** S05, block 1: what counts as a repeat, and against what.
- **Why it matters:** novelty is the core of paper A's value, and E-A2 measures it.
- **Decision:** answered from the repository.
  - A question is a near-duplicate when the Jaccard index of its character trigrams with another question's is 0.7 or more. The trigrams are built from padded words exactly as PostgreSQL's pg_trgm builds them, so the measure is the prototype's (C027, P007).
  - A text in a script written without spaces uses character bigrams instead, at the same threshold, which nobody has measured yet (C027, SPEC 5.6). E-A2 reports novelty per script accordingly.
  - The threshold rose from the prototype's 0.6 to 0.7 when T32 measured the reference corpus: at 0.6 a third of good tasks were refused as copies (R56).
  - It is compared with the reference questions of the child's level and with the child's own past tasks, kept as MinHash sketches of 64 one-byte values (C028, C029). The sketches' error against the exact measure is measured (C050).
  - E-A2 measures novelty with the same measure and adds structural ones (S40).
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S40.
- **Sources:** C027–C029, C050, P007; `docs/decisions.md` R56.

### Q35. The blind review's rubric for "olympiad-style" (applies Q15)

- **Question:** S05, block 1: how the blind review tells an olympiad-style task from a hard exercise.
- **Why it matters:** E-A2 and the author's blind review (S44) need a rubric decided before any task is judged (G9).
- **Decision:** the working definition of RUN.md Q15, turned into a rubric. The reviewer marks each of Q15's four criteria for every task:
  - the method is not known in advance;
  - the solution rests on an idea rather than a procedure, which the reviewer names in a few words (the heuristic catalog meant for this left with paper B, Q60);
  - the statement is self-contained and needs no knowledge beyond the grade;
  - once the idea is found, the solution is short.

  A task is olympiad-style when all four hold. If S46 finds a second rater, their agreement is reported. The author's own criteria are asked in Q46.
- **Who and when:** executor, default, S05, 2026-09-25.
- **Blocks:** S40, S44.
- **Sources:** the author's problem statement (RUN.md, «Решаемая задача»); P034.

### Q36. How paper A speaks of the model knowing the answer

- **Question:** S05, block 1: the model that writes a task knows its answer. What does paper A claim about hiding it?
- **Why it matters:** "the answer stays hidden" is one of paper A's design principles, and a reviewer will ask from whom.
- **Decision:** answered from the repository (C064). Paper A says three things together:
  - the answer reaches neither the child's widget nor the model's plain text until the child answers (C033);
  - the model that wrote the task knows it;
  - an adult who reads the host's own tool-call log sees it, which is outside the threat model.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S58, S59.
- **Sources:** C033, C064.

### Q37. Why β is not updated, and why no uncertainty is kept

- **Question:** S05, block 2.
- **Why it matters:** paper A's learner model departs from both Glicko-2 and the prototype here, and has to say why.
- **Decision:** answered from the repository.
  - Every task is written for one child and never handed out again, so there is nothing to calibrate β from (C013). The prototype did update β, because it reused tasks from a bank (P009).
  - The shrinking step K = K₀/(1 + 0.05n) stands in for a deviation; no deviation and no volatility are kept (C011). The premises file explains the difference from Glicko-2.
  - S38 compares the model with Glicko-2 as a baseline.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S38, S57.
- **Sources:** C011, C013, P009; [evidence/premises.md](evidence/premises.md), the formula check.

### Q38. A floor for the step

- **Question:** S05, block 2. The step has no lower bound, and PRODUCT-V1 keeps this open as О-53.
- **Why it matters:** after many answers a child whose ability jumps is followed slowly.
- **Decision:** default for the research; the product question stays the author's.
  - S38 simulates floors K_min of 0, 0.01, 0.02 and 0.05 on θ.
  - It reports how many answers the level needs to follow a jump in ability, and how much noise each floor adds.
  - The result goes to the author as input to О-53, as a product proposal recorded here (G6).
- **Who and when:** executor, default, S05, 2026-09-25.
- **Blocks:** S38.
- **Sources:** C011; `PRODUCT-V1.md` О-53 (context only).

### Q39. Moving up a grade level

- **Superseded** by the author's answer to Q47: the product moves to a single ladder on which the grade sets only the start (T39b).
- **Question:** S05, block 2. The rating scale lives inside a level: β = d − 3 within grades 1–2, 3–4 or 5–6 (SPEC 1.1). What happens to θ and δ when a child moves up a level?
- **Why it matters:** paper A's model has to say what its numbers mean across a move.
- **Decision:** default, until the author answers Q47.
  - The repository does not specify a move between levels.
  - Paper A states that the scale is internal to a level and that v1 does not model a move.
  - S38 simulates a move as a shift in the difficulty of every task, and reports how long the ratings take to settle.
- **Who and when:** executor, default, S05, 2026-09-25.
- **Blocks:** S38, S57.
- **Sources:** SPEC 1.1; C010.

### Q40. The mastery criterion

- **Question:** S05, block 2.
- **Why it matters:** paper A formalises the chance of a false "mastered".
- **Decision:** answered from the repository (C015, SPEC 2.5).
  - An answer qualifies when it is correct, on a task whose P was at most 0.775 when handed out, and given without the hint.
  - A topic is mastered once it has at least 5 answers and at least 3 qualifying answers since its last wrong one. A correct answer that does not qualify neither adds to that count nor breaks it, so mastery can be granted on such an answer.
  - Mastery is lost after 2 wrong answers in a row; a correct answer of any kind breaks that run.
  - S57 models the run as a Markov chain and computes the false-mastery probability; S38 measures it in simulation.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S38, S57.
- **Sources:** C015; SPEC 2.5; `internal/domain/profile/answer.go`.

### Q41. Where the trap catalog comes from, and whether it is complete

- **Question:** S05, block 2.
- **Why it matters:** the catalog is both paper A's diagnosis and paper B's signal.
- **Decision:** answered from the repository, as far as it goes.
  - The 14 traps of the prototype carry over, and 6 were added for the topics of grades 5–6 (SPEC 1.3, C043).
  - The prototype made the catalog closed so that wrong answers can be counted, as Eedi ties distractors to a taxonomy of misconceptions (P030).
  - Nothing yet shows the catalog is complete. S31 compares it with the published taxonomies of pupils' mistakes in mathematics and names the mistakes no trap covers (done: Q68); S62 reports the gap among the threats to validity. Once T67.x publishes counts, the traps children actually fall into can be seen.
  - The author's experience is asked in Q48.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S31, S62.
- **Sources:** SPEC 1.3; C043; P030.

### Q42. Pace and "I don't understand" in the rule

- **Question:** S05, block 2. SPEC 2.2 says both change the next task, but the rule's code never reads either (C056).
- **Why it matters:** paper A must describe the rule the code runs.
- **Decision:** answered from the repository. Paper A says that v1 records both and that its rule does not use them. Whether the SPEC or the code is right is asked in Q49.
- **Who and when:** executor, S05, 2026-09-25.
- **Blocks:** S57.
- **Sources:** C056; SPEC 2.2.

### Q43. Which heuristics, and from which age

- **Left with paper B (Q60):** kept as a record; it binds nothing, and the tasks it names no longer exist.
- **Question:** S05, block 4.
- **Why it matters:** the heuristic catalog is the axis of the textbook.
- **Decision:** default.
  - S18 builds the catalog from Pólya's phases, Schoenfeld's heuristics and the olympiad tradition, with the age each suits taken from the literature S09 reads.
  - The author reviews it (К✓). The author's own practice is asked in Q50.
- **Who and when:** executor, default, S05, 2026-09-25.
- **Blocks:** S18.
- **Sources:** RUN.md, «Решаемая задача».

### Q44. What never changes by itself (extends Q14)

- **Left with paper B (Q60):** kept as a record; it binds nothing, and the tasks it names no longer exist.
- **Question:** S05, block 5. RUN.md Q14 settles who approves: everything that reaches children is published only after the author approves it. What else must never change without the author?
- **Why it matters:** the textbook reaches children, and paper B has to state its governance in full.
- **Decision:** default. Besides what Q14 covers, these never change without the author:
  - the heuristic catalog, which is the textbook's axis;
  - the answer or solver of a worked example, which the solver also checks;
  - anything that touches children's data.

  The loop proposes changes; the author approves each one.
- **Who and when:** executor, default, S05, 2026-09-25.
- **Blocks:** S20, S27.
- **Sources:** RUN.md, «Статья B коротко» and Q14.

### Q54. How the research reads its sources

- **Question:** S06 read fifteen papers. What did reading them show about the tools, before S08 builds `citecheck` and the literature tasks read dozens more?
- **Why it matters:** every later literature task depends on reading full texts and checking metadata, and G3 forbids citing what was not read.
- **Decision:** from what S06 found.
  - **Summaries are not quotes.** The page-reading tool summarises a page with a small model, so a quote from it is second-hand. A paper cites a passage only after reading the original and keeping it in the notes (G3).
  - **PDFs.** The container has no PDF tools, and none may be installed outside the pinned image. Two ways worked in a scratch test:
    - inflating a PDF's streams, for text set in simple fonts;
    - the Go library `github.com/ledongthuc/pdf` (BSD-3-Clause, pinned by pseudo-version), which reads more PDFs but joins words without spaces in some.

    S08 decides whether a PDF reader becomes a research tool, under the checks of Q26 and Q27.
  - **Metadata.** Crossref and DataCite answer directly, by DOI and by title. DBLP's API sits behind a check for bots from this network. `citecheck` therefore uses Crossref, DataCite and arXiv, and DBLP only where it answers. *Revised by Q61: `citecheck` uses Crossref and DataCite only — arXiv's API refused this network in S08 and arXiv's preprints are checked through DataCite; DBLP answered searches in S08 and is a search source, not a check.*
  - **Paywalls.** A paper behind one is marked as read from its abstract and metadata only, and is not cited on the strength of either.
- **Who and when:** executor, S06, 2026-09-25.
- **Blocks:** S08.
- **Sources:** [literature/exemplars.md](literature/exemplars.md).

### Q55. Where the clients of other model families live

- **Question:** S07 installs Codex CLI and Gemini CLI at exact versions. Should they go into the devcontainer's Dockerfile, or into a pinned image of their own?
- **Why it matters:** the devcontainer's image belongs to the product, and another session works on the product in parallel. A change to it means a rebuild for everyone and ties the product to tools only the research uses.
- **Decision:** an image of their own, in `research/containers/cli`.
  - Its Node base is the devcontainer's version, pinned by tag and digest.
  - The two clients come from a lockfile, with no install step run.
  - Every tool of Gemini CLI is denied by a system policy baked into the image.
  - The logins live in named volumes, one per client. Claude Code stays where it is, in the devcontainer.
- **Who and when:** executor, S07, 2026-09-25.
- **Blocks:** S41, and every task that runs these clients.
- **Sources:** [experiments/cli-spike.md](experiments/cli-spike.md); CLAUDE.md, «Exact versions only».

### Q56. How a trial call is isolated and what it records

- **Question:** what a model may see and do during a research call, and what is kept of it.
- **Why it matters:** a model that can read `content/` or run code could check its own answer, and E-A2 would then measure the tools, not the model. A client can also switch models silently, and the paper has to name the model that actually answered.
- **Decision:**
  - Every call starts in an empty working directory, with every tool off, no project instructions and no session kept. The flags for each client are in [experiments/cli-spike.md](experiments/cli-spike.md).
  - Every call keeps its raw output, and the log records the client, its version, the model asked for, the duration and the exit status. The model that answered is read from the raw output.
  - The trial's third call tests the isolation itself: it asks the model to print a file that holds a token made for that run, and passes only if the token is not in the reply.
  - Before any run, `tools/isolation` shows what a client would put before the model: it points the client at a local endpoint that records every request and refuses it, so no model is called.
- **What that check found (2026-09-25):**
  - *Gemini CLI:* the policy that denies every tool had priority 1000, outside the allowed 0–999, and Gemini CLI ignored the whole file — every trial would have run with file reading and web search on. Fixed to 999; the model is now offered no tool. Headless runs also need `--skip-trust`.
  - *Codex CLI:* no setting isolates its current models (`gpt-6-*`, `gpt-5.6-*`). With every documented feature off they keep a JavaScript tool (a V8 isolate with no files and no network) and the sub-agent tools. So a Codex call can compute, though it cannot read or reach anything.
  - *Claude Code:* not checked yet. The check runs `claude` with a temporary configuration directory, and the author runs it outside an agent session.
- **Who and when:** executor, S07, 2026-09-25; the check and its findings on the same day.
- **Blocks:** S40, S41.
- **Sources:** [experiments/cli-spike.md](experiments/cli-spike.md), [experiments/cli-trial.sh](experiments/cli-trial.sh), [tools/isolation](tools/isolation/main.go).

### Q57. Calls to other models wait; the paper comes first

- **Question:** whether to finish S07 — the trial calls of Codex and Gemini — before writing.
- **Decision:** no. The author wants paper text first (2026-09-25). S07 stops where it is. What it built and found stays in the tree for the first experiment that needs another model family. Settings, flags and tooling are settled by the executor. The author is asked only about the paper.
- **Who and when:** the author, 2026-09-25.
- **Blocks:** nothing now. S07 is closed in this scope; S41 starts with what is left of it.

### Q58. Paper A: order, venue, thesis and evidence

- **Decision (the author, 2026-09-25):**
  - paper A is written first;
  - its venue is **AIED 2027**, main track, full paper: 14 pages with references, Springer LNAI (a subseries of LNCS), double-blind;
  - its thesis: *the chat's own model writes each task, and the service checks it* — the solver brute-forces the options, and the service also checks structure, near-duplicates, readability and drawings. The service hides the answer until the child answers. Teaching thinking through non-standard problems is the motivation in the introduction, not the claim;
  - its evidence comes from offline experiments on the product's own code: defects injected into reference tasks (E-A1) and simulated learners on the real rating code (E-A3). No model generates tasks for the first version (E-A2 waits, Q57).
- **Replaces:** Q28 for paper A. Paper B has since left the plan (Q60).
- **Blocks:** nothing; S56 onward follow it.

### Q59. Every paper draft has a Russian copy

- **Decision (the author, 2026-09-25):** "всегда создавай тот же документ на русском". Beside the paper draft stands a Russian copy for the author: `paper-a/draft.ru.md` beside `paper-a/draft.md`.
  - The English draft is the source and the one a venue receives. The copy mirrors it section by section, with the same claim IDs, [TBD] marks, citation keys and formulas, and changes in the same edit.
  - The project's rule that documents are in English makes this one exception; CLAUDE.md and RUN.md say so.
- **Who and when:** the author, 2026-09-25.
- **Blocks:** nothing; every writing task S56–S66 follows it.

### Q60. Paper B leaves the plan

- **Decision (the author, 2026-09-25):** «полностью выбрасываем из плана создание второго документа — фокусируемся исключительно на первом». The research program now produces one paper, paper A.
  - Removed from RUN.md: the summary of paper B, its 37 tasks (S11–S15, S17–S30, S47–S54, S67–S76), the schedule that could put B first, and the critical questions K07 and K08.
  - Kept, because paper A needs them: S08 (the search protocol and `citecheck`), S09 and S10 (the literature behind the introduction), S16 (privacy and law, for the ethics section and S34).
  - Gone with paper B: the heuristic catalog and the annotation of the corpus by heuristics (S18, S19). Paper A's artifact is the reference tasks with their solvers and traps.
  - What was written for paper B stays as a record and binds nothing: decisions Q09–Q14, Q16 and Q17 in RUN.md and Q29, Q43, Q44 and Q50–Q52 here; paper B's skeleton in `literature/exemplars.md`; its venue in `literature/venues.md`; the check of the consensus formula in `evidence/premises.md`; the "B:" entries in the ledger's "Used in" column.
- **Who and when:** the author, 2026-09-25.
- **Blocks:** nothing; RUN.md lists paper A's tasks only.

### Q61. How the bibliography is checked, and where the literature is searched

- **Question:** S08 — what `citecheck` asks, where the literature tasks search, and whether a PDF reader becomes a research tool (Q54).
- **Decision:**
  - `citecheck` asks Crossref, then DataCite for the DOIs Crossref does not hold. arXiv's own API answers this network with HTTP 406, so a preprint is checked by the DOI arXiv registers with DataCite (`10.48550/arXiv.<id>`), which carries its title, authors and year.
  - It compares title, the authors' family names in order, year and venue, and fails a work a registry marks retracted, withdrawn, removed, partly retracted or under an expression of concern (Crossref's `updated-by`, which since 2023 carries Retraction Watch's data). A venue the registry does not record cannot be claimed: DataCite records none for preprints and deposits, so a preprint that was later published is cited by the published version's DOI.
  - Entries are written by `citecheck add` from the registry's record, never typed: that is rule G3 applied by the tool rather than by care. Words with two capitals or more are braced in titles, so a style that changes case keeps acronyms; braces and backslashes from a registry are dropped, since one unmatched brace would leave a file no BibTeX reader can read. Whether the paper's build reads UTF-8 names correctly (8-bit BibTeX does not) is for S55.
  - Search runs on OpenAlex (it also lists the works citing a work), Crossref, DBLP and arXiv's listings; Semantic Scholar refuses without a key.
  - No PDF reader joins the research module. Full texts are read as HTML where they exist (arXiv's HTML, PubMed Central, publishers); a PDF's text is extracted locally for reading and never committed. The notes keep each passage verbatim with its page, which anyone can check against the PDF.
  - Web pages without a DOI go to `literature/web.bib`, outside the check.
- **Who and when:** executor, S08, 2026-09-25.
- **Blocks:** S09, S10, S16, S31–S35 and every writing task.
- **Sources:** [literature/protocol.md](literature/protocol.md), [tools/citecheck](tools/citecheck/main.go); Crossref's documentation of Retraction Watch data (`crossref.org/documentation/retrieve-metadata/retraction-watch/`), read 2026-09-25; arXiv's API manual (`info.arxiv.org/help/api/user-manual.html`), read 2026-09-25.

### Q62. What the problem-solving literature lets paper A claim

- **Question:** S09 — which of MathTrail's choices the literature on teaching problem solving grounds, and what that changes in paper A.
- **Decision:**
  - The corridor is presented as a design choice in line with three arguments: desirable difficulties, the zone of proximal development as van de Pol et al. read it, and the eighty-five percent rule. None of them is a measurement on children, and none grounds the numbers 0.70 and 0.85.
  - Novelty is claimed as "new tasks on recurring ideas": a nonroutine problem is new to its solver by definition, and topics and traps recur. The check of novelty is described as textual.
  - The hint is called a hint, not scaffolding: a question fixed in advance is neither contingent nor faded. Only its question form is grounded; the product also allows a first step, which nothing read grounds.
  - The introduction cites Pólya and Schoenfeld by checked keys. It speaks of the uneven development of the strands that *Adding It Up* measured in American 13-year-olds, rather than of superficial procedures. Neither it nor the abstract says any longer that olympiad problems teach the skill: what such practice is known to do is S10's.
  - Productive failure is not cited in support: for grades 2–5 its meta-analysis leans towards instruction first.
  - The claim of transfer stays within mathematics; paper A promises no general thinking skill.
- **A proposal for the product (G6), not critical:** *How People Learn* warns that practice labelled by chapter leaves knowledge inert. If the task widget is to show the task's topic, it gives the child that cue; the author may want the widget to leave the topic out. The research changes nothing in the product.
- **Who and when:** executor, S09, 2026-09-25.
- **Blocks:** nothing; it affects S59 (the hint is called a hint) and S63 (the evidence missing for grades 1–6).
- **Sources:** [literature/problem-solving.md](literature/problem-solving.md).

### Q63. What paper A may say about olympiad practice, and how it names novelty

- **Question:** S10 — what the literature says olympiad-style practice does, and how "olympiad-style" and "new to the child" become operational.
- **Decision:**
  - The introduction says three things about olympiad practice: the genre comes from a long tradition of circles and competitions; its effects have seldom been studied systematically; in the one study with a comparison group found for primary school, selected third- and fourth-graders taught in a group course were ahead of the comparison group in the competition and, in grade 4, on a curriculum test, with groups that were not randomised and differed before the course. It claims no effect for a child practising alone in a chat.
  - The working definition is a default the executor set (Q15), and the author's own criteria are still asked (Q46). Its first criterion now cites PISA 2012: the method is not immediately obvious to the solver. The notes give sources for two more; "self-contained" and "short once the idea is found" have none.
  - Novelty has a surface (the text) and a structure (the idea). The service checks the surface only, and paper A says so. If E-A2 is ever run, a blind reviewer judges whether a task's idea repeats one of the child's earlier tasks (S40).
  - Paper A measures no transfer and says so.
- **Who and when:** executor, S10, 2026-09-25.
- **Blocks:** nothing; it affects S40 (a blind judgement of repeated ideas), S59 (the near-duplicate check is called a guard against repeated text) and S63 (the evidence comes from selected children taught in groups).
- **Sources:** [literature/olympiad-and-measurement.md](literature/olympiad-and-measurement.md).

### Q64. An area of the literature may hold fewer than 15 works

- **Question:** the plan asks each literature area for 15–30 works (phase 1, «Литература»). S09 holds eleven and S10 ten, with two borrowed from S09.
- **Decision:** the range is a target, not a floor. An area stops when its questions are answered from works actually read, and says in its Remarks what kept the rest out and which leads remain. Padding an area with works read from their abstracts would break rule G3 and the protocol's reading rules, which is worse than a short list.
  - In S09 and S10 the limit was access: publishers and some repositories refuse automated readers, and paywalled works cannot be read. Each notes file lists the leads it could not read.
  - For S10 the short list is also the finding: research on what olympiad practice does is thin.
  - S35, which judges paper A's novelty against the literature, may widen an area if a claim needs it.
- **Who and when:** executor, S10, 2026-09-25.
- **Blocks:** nothing; it applies to S16 and S31–S34.
- **Sources:** [literature/problem-solving.md](literature/problem-solving.md), [literature/olympiad-and-measurement.md](literature/olympiad-and-measurement.md).

### Q65. The ledger moves to `main` before the literature and the experiments

- **Question:** the ledger pinned `3597c6a0633d`, a commit of a branch from 2026-09-25 that never reached `main` (Q30). Since then the product built what the ledger and the draft called planned: the ladder and the trial series (T39b), the MCP endpoint and its tools (T40–T46), sign-in and the profile in Drive (T47–T51), the limits and the load measurements (T52–T52b), the widget in 22 languages (T54–T59), and the first runs with a real model (T46, T53.2, T58). Move now, or leave it for S56?
- **Why it matters:** every later task rests on the ledger. S35 judges novelty against what is built, S36 fixes the experiments on the checks and the learner model as they are, S37 and S38 run the product's current code, and the author reads a draft that called built things planned.
- **Options:**
  1. Move the ledger to the newest commit of `main` now, as a task of its own, and bring the draft to it; S56 moves it once more, to the final commit.
  2. Leave the ledger where it is until S56.
- **Decision:** option 1, as S02.1. The move broke 25 of the 77 claims, and 16 more still found their lines but no longer said what the product does; every one of them was read again against the code.
  - The pin is `52ce86908135`, the head of `main` on 2026-09-30, so a reviewer can open it.
  - Claims whose subject stayed keep their numbers with new text and status: C038–C041 name the tools, sign-in, the limits and the recording of answers as built; C010, C016 and C065 describe the ladder, the eleven ranks and a grade that sets only the start.
  - Two remarks the product resolved leave the table and are listed as retired: C055 and C058. No number is reused.
  - New claims C078–C090 cover what the draft now needs: the profile in Drive, the price of bytes in the sandbox, the trial series, topics within reach, mastery held at a level, the model's own choice, the languages of the cards, the planned acceptance runs, and the measurements of the first runs and of the load.
  - Four facts are new, computed at the commit: no model provider's SDK among the product's modules and no model API's host in its Go code, which turn C003 from specified into built; the rule's lines that name the pace or "I don't know", which C056 needs; and the number of the card's dictionaries.
  - The draft, in both languages, describes the built system in the present tense, and section 5.4 describes the trial series as the code runs it.
- **Who and when:** executor, S02.1, 2026-09-30.
- **Blocks:** nothing; S16 onward rest on the new pin, and S56 still pins the final commit.
- **Sources:** [evidence/ledger.md](evidence/ledger.md), [evidence/product-stats.txt](evidence/product-stats.txt); `git log --oneline 3597c6a0633d..52ce86908135`; SPEC 2–10 at `52ce86908135`; `docs/live/04-local-run.md`, `docs/live/05-login-drive.md`, `docs/live/06-drawings.md`, `docs/load.md`.

### Q66. What paper A says about children's data and the law

- **Question:** S16 — what the laws on children's data and the rules for research with children let paper A say, and what they ask it not to.
- **Decision:**
  - Paper A claims no compliance with any law and makes no legal argument. It describes what the design collects and keeps, and the notes (`literature/privacy-law.md`) say where each law bears on it, marked as not legal advice.
  - It does not argue that the law stops at the parent because the parent holds the account: COPPA looks at whom a service is directed to, the GDPR at whom it is offered to, and the UK code at who is likely to use it.
  - It calls the child's data pseudonymised personal data, never anonymous (GDPR Recital 26; Mourby et al.). What it can say is that the service keeps no copy of it: the profile lives in the parent's own Drive (C078), logs hold no content (C036), nothing trains a model or serves advertising (C003, C091). It also says what does leave or stay: the child's grade and interests and the parent's notes reach the chat's model with every task (C092), and identifiers remain — a random one for the child in the profile, a derived one for the parent's account in the logs, the requester's address in the platform's logs (C093).
  - It says that the learner model serves only the choice of the next task and the family's own view of progress. That is profiling in the GDPR's sense, and the UK code's exception for profiling intrinsic to the core service describes it; paper A does not need to argue either point.
  - The ethics paragraph names what a future study with children needs: an ethics board and, by default, the parent's permission and the child's own assent (45 CFR 46.404; BERA 2024, paragraphs 23–24), which the board may waive — on the grounds 45 CFR 46.408 lists, including those of 46.116 — or find exempt under 46.104(d)(1); and, for experiments inside a lesson, comparisons of sensible options only (Motz et al.).
- **A proposal for the product (G6), not critical: the privacy policy.** Its account of what reaches the chat's model names the parent's notes and says the rest of the profile stays out, while the package also carries the child's grade and interests; and it says the logs hold counts, not content, without naming the derived identifier of the parent's account on every signed-in line (C094). The policy may want to name both.
- **A proposal for the product (G6), not critical:** the planned public usage numbers (T67: cells of at least 10, rounded to 5, one dimension per table) protect each table on its own, as k-anonymity does. Tables released month after month can be differenced, and "overly accurate answers to too many questions will destroy privacy" (Dwork and Roth). Before T67 publishes, the author may want fewer and coarser tables, or a differentially private release. The research changes nothing in the product.
- **For the research itself (applies G5):** G5 lets the research use the product's public aggregates under that same rule. The weakness above applies to it too, so the research uses those tables only as the product publishes them and derives no new cross-tabulation or difference from them — S45 follows this.
- **Who and when:** executor, S16, 2026-09-30.
- **Blocks:** nothing; it affects S34 (the hosts' rules for children), S59 (privacy by design) and S63 (ethics).
- **Sources:** [literature/privacy-law.md](literature/privacy-law.md); `site/content/en/privacy.md`, `docs/usage-evidence-for-grants.md` §4.4 at `52ce86908135`.

### Q67. How paper A is built

- **Question:** S55 — how the paper becomes a PDF the same way on every machine, with its numbers, citations, figures and anonymity held to the rules.
- **Decision:**
  - **TeX in a container.** `texlive/texlive:TL2025-historic`, pinned by tag and digest: the frozen TeX Live 2025, which carries Springer's LNCS class (llncs 2.26), pgfplots, siunitx, latexmk and BibTeX. Nothing of TeX is installed in the devcontainer.
  - **One command.** `just research paper-a` writes the numbers, checks the citations and the names, and builds the named PDF and the anonymous one into `paper-a/build/`. The PDFs carry the pinned commit's date, so the same sources make the same file; `just research paper-a-figure-check` shows it byte for byte on a test figure.
  - **Numbers only through macros (G2).** `tools/macros` turns files of computed numbers — `key=value` lines such as `evidence/product-stats.txt`, and each experiment's results once they exist — into `paper-a/generated/numbers.tex`. The text writes `\stat{source}{key}`, and a key nothing defines stops the build.
  - **Figures from CSV.** pgfplots reads each figure's CSV itself, so a changed result changes the figure with no edit to the LaTeX.
  - **Citations.** Each build runs LaTeX alone first; its `.aux` records every key the paper cites, from whatever file LaTeX read it, with comments and verbatim text already set aside. `citecheck cited` reads it and asks no registry: every key must be in `refs.bib`, whose entries `citecheck` checks against Crossref and DataCite, or in `web.bib`, the laws and pages with no DOI. Only then does BibTeX run, so a lead from `seed.bib` that nobody has read cannot reach a PDF. The full check against the registries needs the network and stays its own recipe, which every writing task runs, as rule 6 of RUN.md says.
  - **UTF-8 names (open since S08).** The 8-bit BibTeX takes the first byte of a name that starts with a letter outside ASCII for its initial, and the PDF breaks: a test entry "Řihák, Émile" came out as bytes. The build uses `bibtexu`, which prints "Řihák, É."; `paper-a/latexmkrc` sets it.
  - **Anonymity.** The text writes `\System` and `\ArtifactURL`, which the anonymous build replaces, and the build refuses a source that names MathTrail or its repositories by hand. A computed value that would name the system — the commit, a web address, the project's name — is generated as `\statiddef`, which the anonymous build prints as a mark; a test document built both ways shows the names and the commit in the named log and nowhere in the anonymous one. The anonymous build's bibliography is read back for the project's name as well.
  - **TBD marks** print boldly in every draft. `just research paper-a-final`, the build for submission, refuses to finish while one is left.
  - **Where the English text lives.** From S56 on, `paper-a/main.tex` and `paper-a/sections/*.tex` are the English source; each writing task ports its sections from `draft.md` and takes them out of it, and `draft.ru.md` mirrors the LaTeX section by section (Q59).
  - **Licences of the build's tools** (Q24; they are part of no code of the repository): llncs CC BY 4.0, pgfplots GPL 3 or later, siunitx, booktabs, hyperref, amsmath and the LaTeX tools LPPL 1.3, latexmk GPL 2, bibtexu GPL, Latin Modern GUST Font Licence — as `tlmgr info` reports them in the pinned image.
- **Who and when:** executor, S55, 2026-09-30.
- **Blocks:** nothing; S56–S66 build with it.
- **Sources:** `research/justfile` (`paper-a`, `paper-a-final`, `paper-a-figure-check`), `research/paper-a/`, `research/tools/macros`, `research/tools/citecheck/cited.go`.

### Q68. What paper A may say about generating and checking problems

- **Question:** S31 — what the literature on generating and checking maths problems lets paper A claim, and which pupils' mistakes the trap catalog lacks (Q41).
- **Decision:**
  - **The design is supported, not proven.** Paper A cites the literature for three choices: a program checks what can be computed (PAL, Program of Thoughts); the author's own judgement never admits a task alone (Huang et al., Tyen et al., MATH², MathWiz's model judge); the self-check asks named questions rather than an open one (UMWP, QuestBench, AbstentionBench). It claims no rate at which a misreading shared by the solver and the key gets through: no work measures one, and RQ1's injected faults measure what the checks catch, not that rate.
  - **The gap is a finding of the searches.** Among the works read, none writes olympiad-style problems for primary school, and none admits a problem by running a program its author wrote against the problem's options. The two nearest were read for it: validator agents that are models (Ikram et al., AIED 2026) and a model checking its own answers with code it reads itself (Zhou et al.). The related work says so with "among the works we read"; S35 decides whether it becomes a claim of novelty.
  - **Repetition is textual.** The near-duplicate check is a guard against repeated text, and a paraphrase passes such checks (OpenMathInstruct-2). The Artificial Hivemind is cited as a risk whose strength is disputed (Schaeffer et al.), not as a rate.
  - **The trap catalog.** Seventeen of its 20 traps have counterparts in Eedi's misconception graph; the three without — the worst case, knights and liars, the first move of a game — are olympiad topics. The common primary-school mistakes no trap covers — place value, calculation slips, the order of operations, fractions, factors and multiples, clocks, perimeter and area on a grid, Venn diagrams, and Newman's decoding errors and careless slips — go to S62's threats to validity. Whether to add traps is the author's and the product's question (Q48); the research changes nothing in the product.
  - **The prototype's literature claims.** P039 is wrong: EDUMATH's 90, 77 and 66 % are the rates at which two annotators agreed, not shares of problems. P037, P038, P040–P043, P046 and P051 hold, some with the corrections in the notes; P044, P045 and P053 stay unread. Paper A cites the works, never the P-claims; `evidence/prototype.md` says so in its Remarks.
  - **Numbers from the literature** go into `literature/numbers.txt`, each with the passage it comes from, and reach the paper as `\stat{literature}{…}` (G2).
  - **The notes' names.** RUN.md named the notes of S31–S34 `a1-generation.md` to `a4-mcp.md`; the protocol (S08) named them after their topics, as S09, S10 and S16 named theirs: `generation-verification.md`, `learner-models.md`, `llm-tutors.md`, `mcp-safety.md`. The protocol's names stand, and RUN.md now uses them.
  - **What the self-check asks** is claim C095, so the draft can say it.
- **Who and when:** executor, S31, 2026-10-01.
- **Blocks:** nothing; it affects S35 (the gap), S37 (a paraphrase and a same-template task with new numbers as injected faults), S59 (the self-check described as a prompt, not a verification) and S62 (the traps' gap; the fit of a task to its topic and difficulty is not checked).
- **Sources:** [literature/generation-verification.md](literature/generation-verification.md).

### Q69. What paper A may say about its learner model

- **Question:** S32 — where MathTrail's learner model stands in the literature, and what paper A may claim about its step, its mastery rule and its target.
- **Decision:**
  - **Its lineage is named, not claimed as new.** The response model is Birnbaum's three-parameter logistic model with Rasch's slope and a floor fixed at one in five; the update is the hierarchical Elo of adaptive practice — a general level corrected by a topic level, a floor for multiple choice, a step α/(1 + βn) (Nižnan et al. 2015; Pelánek et al. 2017). Paper A cites them as the direct precedent. What MathTrail changes is item difficulty, set by design on one ladder because no task is used twice; whether that, or the trial series (C080), is a claim of novelty is S35's question.
  - **Rasch is cited from what was read.** His 1960 book has no open copy; paper A cites his 1961 paper for the model and Birnbaum's chapter 17 for the floor.
  - **The step.** K₀/(1 + 0.05n) is Papoušek et al.'s uncertainty function and a sequence "of type 1/n" in Robbins and Monro's own sense; the draft says so instead of the textbook conditions. Paper A states the idealised Robbins–Monro reading of the rule with its limits, and names two costs S38 measures: the lag behind a child who learns, and whether the low bias found for Elo with a constant step, fixed difficulties and a target above one half survives a shrinking step. It cites no number from the simulations of others for either.
  - **Mastery.** A run of correct answers is a simple rule often used in practice and, tuned, nearly as good as knowledge tracing; three in a row is a value in use, at the low end of the tuned ones; paper A also states the tension — a run assumes no slips, and the tasks are set at a 77.5 % success rate — and leaves the rate of false mastery to S38, with premature mastery and over-practice as the two errors.
  - **The target.** The 85 % rule supports a band of 75–85 % for binary choices and gradient learners; for five options it stays an analogy, as the draft says.
  - **A language model as learner model.** Hooshyar et al.'s published paper replaces the seed's arXiv record. Paper A cites it for the division of labour and says the evidence is narrow.
  - **Labels.** "Hierarchical Elo" or "a general level with topic levels" (a bifactor structure); not "testlet" or "multidimensional Rasch".
  - **Assumed, not measured:** the floor of 0.2 (traps may pull a guessing child below it) and every task's difficulty; S59 and S62 say so.
  - **Records that print badly.** Three registry records give whole names as single strings (Pavlik et al. 2009; two EDM papers registered at Zenodo); S61 decides how the typeset list names them.
  - **Numbers** taken from these works are in `literature/numbers.txt`.
- **Who and when:** executor, S32, 2026-10-01.
- **Blocks:** nothing; it affects S35 (novelty of the designed difficulty), S38 (the lag, the bias, false mastery; Glicko-2 and Urnings have no floor as published, so a baseline with one is S38's own construction), S58 (Section 5), S61 and S62.
- **Sources:** [literature/learner-models.md](literature/learner-models.md).

### Q70. What paper A may say about model tutors, study modes and competing products

- **Question:** S33 — what the evidence on language-model tutors lets paper A say, how it places MathTrail beside the assistants' study modes, and what it may say about competing products.
- **Decision:**
  - **No learning claim.** Paper A claims no learning gain. It says that unguarded chat help harmed exam results in a high-school trial, that a guarded tutor only removed the harm, and that every gain in the tutoring trials read came with correctness and structure from outside the model — two of those gains uncertain; a meta-analysis across subjects, mostly of higher education, pools a larger effect, with no significant effect for its two primary-school studies. Whether automated checks can play that role for a child is untested, and the limitations say so.
  - **Mastery is the rule's state.** A large trial found that a three-in-a-row rule raised the platform's own count of mastery but not delayed learning; paper A reports MathTrail's mastery as the rule's state, not as learning.
  - **The study modes are cited from their vendors' pages,** and described by what the pages say they do. That none describes checking its problems is written as "not publicly described", never as "absent". The difference paper A claims is structural: a checked task, a sealed key, explanations tied to named traps, a deterministic choice of task, a profile kept across sessions.
  - **The checks' scope.** MathTrail checks a task before the child sees it, not the explanation the model gives afterwards in its own words; S59 and S63 are to say so in Section 4 and the limitations.
  - **Competitors.** The prototype's list holds in substance; the product table in the notes is the record, and S35 weighs it against the claim of novelty. Claude's learning mode dates from 2 April 2025 for universities; its opening to all users is cited from no page of Anthropic's, so paper A does not date it.
  - **Retracted work.** Wang and Fan's meta-analysis (2025) was retracted in April 2026 and is not cited; Paper A cites Wu et al. (2026) for the thinness of evidence on primary pupils; their pooled effect, mostly from higher education, is context in the notes, not a figure paper A quotes.
- **Who and when:** executor, S33, 2026-10-01.
- **Blocks:** nothing; it affects S35 (the product table), S59 (the checks' scope), S62 and S63 (no learning claim; mastery is not learning).
- **Sources:** [literature/llm-tutors.md](literature/llm-tutors.md).

### Q71. What paper A may say about MCP, the safety of its tools, and the hosts' rules for children

- **Question:** S34 — what the MCP specification, MCP Apps, the hosts' documentation and the security literature let paper A say about MathTrail's delivery in chat, and what they ask of it.
- **Decision:**
  - **The protocol is cited as the requirement, not as an audit.** MathTrail's endpoint speaks the stateless revision 2026-07-28 and is its own OAuth 2.1 authorization server (C040, C062); paper A cites the specification for what it follows and claims no conformance test.
  - **Sealing is worded as "never shown".** The card is drawn by `submit_task`, and MCP Apps has a host hand a card the complete arguments of the call that drew it, so the key and the solution reach the card's memory unseen (C096). Paper A says the answer is never shown before the child answers, and names the card's memory beside the host's tool log among what sealing does not cover; the draft now says so.
  - **The conversation is outside the checks.** The parent's notes reach the model that talks to the child (C092) though the rule never reads them (C097); other servers in the same chat can steer the model's calls, and in text mode the model records the child's answer. Paper A names these channels in its limitations and cites the literature for the kind of risk, a benchmark for "often", and no rate for its hosts.
  - **The hosts' rules.** Both hosts admit only the adult; MathTrail's terms make the adult the user who runs the lesson (C098). Paper A states this and claims no compliance on a host's behalf; it does not say that the hosts permit a child's use.
- **Proposals for the product (G6) — the research changes nothing in the product; the first is raised as K10, the rest are not critical:**
  - **The key reaches the card's memory (C096, C099)** — raised to the author as K10 in RUN.md, since it touches a rule of the product's own: the answer must not appear in widget payloads, and the arguments a host forwards to the card that `submit_task` draws are such a payload. The card could be drawn by another call whose arguments hold no key.
  - **The child's own words.** The card's field "Ask a question about the task" sends the child's words to the chat (C100); how that sits with OpenAI's rule that for children under 13 "the actual interaction with ChatGPT must be conducted by an adult" is the author's to weigh, and perhaps a host's review.
  - **The notes as a lasting injection.** `save_profile`, which the model calls, rewrites the notes that return with every task; a change to the notes could need the adult's confirmation, or carry a mark that the adult wrote it.
  - **For ChatGPT's directory:** `_meta.ui.domain`, required of a listed plugin with UI, is not set; `save_profile` is marked `destructiveHint: false` although it overwrites what the parent wrote, which OpenAI counts as destructive; timestamps and handles in payloads may need a stated reason; and an app that "explicitly targets children under 13" is not admitted, which a listing has to answer.
  - **Literal conformance:** registration is kept as a fallback, which the revision deprecates while keeping it for compatibility; a CSRF nonce cookie is set with the consent page, before approval, while the Best Practices' rule speaks of the cookie that holds the `state`; refresh tokens are rotated without retiring the one presented, against a rule whose source (OAuth 2.1 §4.3.1) was not read.
  - **The privacy policy** could say that ChatGPT sends a coarse location and anonymised ids with every call and that the service does not read them, and that a host's age signals could disable a parent's account or move it to a teen's experience.
- **The author's answer to K10 (2026-10-01):** a known limitation, like the host's own tool log (О-27). The product does not change; paper A says the key is never shown on the card but lies in its memory until the child answers, as the draft already does. The product's rule on widget payloads is to be clarified on the product's side; the report of S37 proposes a wording.
- **Who and when:** executor, S34, 2026-10-01; the author, K10, 2026-10-01.
- **Blocks:** nothing now. It affects S59 (transport, authorization), S62 and S63 (limitations, ethics), and the product's directory tasks (T71.x) through the proposals.
- **Sources:** [literature/mcp-safety.md](literature/mcp-safety.md); C092, C096, C097, C098.

### Q72. What paper A claims as new

- **Question:** S35 — what in paper A is new against the literature read for S09, S10, S16 and S31–S34, and which work comes nearest to each claim.
- **Decision:**
  - **Two claims of novelty, both "among the works we found".** N1: the chat's model writes each task, and a deterministic service admits it only after running a program the same model wrote, which must single out exactly the keyed option of five, again under rotated labels; nearest: MATHWELL (a program that is the key, unchecked), Zhou et al. (a model checking its own answer with code), GSM-Symbolic (programs written by people), Ikram et al. (validator agents that are models). N2: olympiad-style problems for grades 1–6, written for one child, with the key checked before display — whether a task is olympiad-style is asked for in the brief, not checked; nearest: MathWiz and EDUMATH (routine problems for primary grades), Beast Academy and MasterOlympiad, before launch (olympiad problems written by people), Spiral Learning (olympiad problems for ages 7–13 from an adaptive bank whose origin its page does not state), MathForces (generated olympiad quizzes, audience not stated, no check described). The introduction states these two (S60).
  - **The architecture's split is not new on its own.** An open MCP tutor (Tutor MCP, 2026) also separates what the model writes from rule-driven progression; MathTrail's claim is the split together with N1 and N2. The related work now names it.
  - **The learner model is named by its lineage** (Q69): the hierarchical Elo of adaptive practice; MathTrail's own part is difficulty set by design for tasks used once. No novelty is claimed for the trial series, whose literature (adaptive testing) was not read, for the near-duplicate method, or for any learning effect.
  - **P037 holds and widens** within S31's and S35's searches in English and the forward citations of the nearest works: no work found generates olympiad-style problems for grades 1–6; among products, MathForces generates olympiad quizzes but states no audience and describes no check. The Russian tradition was not searched beyond OpenAlex, which returned nothing; S62 names that limit. One systematic review of word-problem generation (Utami and Hwang, IJSME 2026) could not be read, and the author cannot obtain it (2026-10-01); paper A names it among the limits of its search.
- **Who and when:** executor, S35, 2026-10-01.
- **Blocks:** nothing; it affects S60 (contributions), S62 (threats to validity: the search's coverage) and S63 (discussion: Tutor MCP, the lineage).
- **Sources:** [literature/positioning-a.md](literature/positioning-a.md).

### Q73. What the offline protocol fixes, and one proposal for the product

- **Question:** S36 — the choices the protocol of E-A1, E-A3 and E-A4 had to make before any data, where RUN.md left them open.
- **Decision:**
  - **E-A1 sees what the deterministic checks catch when the model's self-check misses the defect.** Every case carries a self-check that found nothing and agrees with the key, except in the classes that break the self-check itself. How often a model notices its own defect is the model's property, which only E-A2 measures.
  - **Defects go only into reference tasks the checks accept as they are,** made into submissions with fixed texts for the fields a reference task lacks, so that a refusal is the defect's doing. How many qualify at each level is a result of the sanity run.
  - **Every operator applies to every eligible host,** rather than a sample of them; randomness only picks letters, donors and types, from a fixed seed per case.
  - **Mechanism and discovery are kept apart.** Most must-catch operators break a rule a check states itself, and test that the pipeline holds it end to end; three — a second correct answer in words, a copy with new numbers, a repeat with new numbers — test what the checks find beyond their definitions. A wrong key is caught when the solver is right; the same key with a solver that shares its misreading is out of scope, and the paper states both halves.
  - **What only one check catches** is read from the outcome by removing that check's code, which is exact here because every check runs and is reported on its own; the product's code is not switched.
  - **Out-of-scope classes** add a missing condition, a task off its topic or difficulty, and the same idea in a new setting to RUN.md's list, from the S31 reading. Their cases are partly read by the executor for equivalent mutants.
  - **E-A3 runs the service through its own domain code** — the rule, the profile's request, issue and record, the real sealing — in a closed loop. A baseline replaces only the estimate, written into the profile before each choice of task, so selection and mastery are the service's for every rule.
  - **Baselines as published where they have a form, constructions named as ours where they do not:** Glicko-2 and Urnings only for one level and for levels per topic; Glicko-2 with the guessing floor derived from the Bernoulli likelihood; Urnings against items of known difficulty with an acceptance rule derived for a fixed item and checked by simulation, without the correction for adaptive selection, which a deterministic rule would make reject nearly every move.
  - **Generator parameters are set before data** — the spread of children around their start, the writing error of a model, the rates of learning and forgetting, the size of a jump, a harder host — with a sensitivity generator for each assumption the model makes.
  - **Four comparisons are primary;** everything else is labelled secondary. Intervals are bootstrap intervals over children, with no tests of significance.
  - **Freezing.** A new recipe, `timestamp`, runs the official OpenTimestamps client at exact versions in a Python image pinned by tag and digest, and refuses to stamp a protocol twice.
- **Proposal for the product (G6, not critical):** the answer event `answer_recorded` carries no predicted chance of success, though `Record` computes it. With it, rounded to two places, and a mark for the trial series, the published aggregates could show calibration on real children within the product's rule of cells of ten. The author decides whether this becomes a T-task.
- **Who and when:** executor, S36, 2026-10-01.
- **Blocks:** nothing; S37, S38 and S39 follow the protocol.
- **Sources:** [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md); `internal/transport/mcp/submitanswer.go:344` and `internal/domain/profile/answer.go:251` at `52ce86908135`.

### Q74. What the injected defects found, and two proposals for the product

- **Question:** S37 — what E-A1 measured under the frozen protocol, what it lets paper A say, and what it asks of the product.
- **Decision:**
  - **The run.** 447 of the 603 reference tasks pass every check as they are (126 of 200 for grades 1–2, 168 of 250 for 3–4, 153 of 153 for 5–6); the rest fail readability at their own level or read as near-copies of another reference task of it. On the 447, 68 operators of 27 classes made 23,453 cases. The run kept to the protocol, with one addition: beside a class's bootstrap interval over its hosts, the exact interval of its pooled cases, because the bootstrap has no width when every host has the same share, as at 0 and 100 %. The code under test was the code at the pin, and the interpreter the pin's version. The numbers reach the paper as `\stat{ea1}{…}`.
  - **Mechanism.** Every case of the 53 operators that break a rule a check states was refused by the check expected. Two kinds are refused by two checks (the right answer written twice; an explanation under the right option); every other kind by one check alone, so no check is redundant.
  - **Discovery.** A copy of a reference task with new numbers was refused in 94.6 % of cases and a repeat of the child's own task with new numbers in 96.1 %; every miss was a task of arithmetic tricks, whose short questions are mostly numbers. A second correct option written in words beside the digits was never refused (0 of 213).
  - **Out of scope.** Never refused: a hint that gives the answer, a trap that names the wrong mistake, a task of another topic or difficulty. Refused in 1–7 % of cases, only when the changed wording tripped readability or a drawing's labels: a vague quantity or a range, a removed condition, a question asking the opposite, the pseudonym. A copy with people and objects renamed was refused as a near-duplicate in 92.4 %.
  - **The reading.** 120 out-of-scope cases were read by the executor, an AI: 4 of 10 questions turned to ask the opposite kept their answer or read as nothing, and 1 of 10 removed sentences held nothing the answer needed; every other case read was its defect. `reading.csv` records each verdict with its reason; the author checks it afterwards (K04). Q86: the reading left the paper, and the author does not check it.
  - **Tooling.** `deps-check` compared every module in either build's graph, so a module no compiled package uses failed it (x/term, which only starlark-go's go.mod names); it now compares the modules whose packages both builds compile, which is what its rule protects. A new recipe, `faultinject`, reruns the experiment and records the commit and whether the code under test is still the code at the pin.
- **Proposals for the product (G6, not critical):**
  - **The same answer in other words.** Options count as one answer only when they read alike or are the same number in digits, so "6" and "six" can stand side by side as two options, one right and one "wrong". The instructions could ask for numbers in digits, and the structure check could refuse an option that is a number word of another option's number in the task's language.
  - **Repeats with new numbers.** Comparing questions with every number masked would catch a short calculation given again with new numbers, which the textual check let through in 18 of 39 copies and 14 of 39 repeats of that topic.
- **Who and when:** executor, S37, 2026-10-01.
- **Blocks:** nothing; it affects S58 (the checks, Table 3), S62 (results and threats) and S63 (discussion).
- **Sources:** [experiments/faultinject/results](experiments/faultinject/results); [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md).

### Q75. What the simulated learners showed, and three proposals for the product

- **Question:** S38 — what E-A3 measured under the frozen protocol, what it lets paper A say, and what it asks of the product.
- **Decision:**
  - **The run.** 167 cells — the fifteen rules of the sweep on the nine generators, and 32 cells of variants — of 1,000 children and 200 answers each, through the service's own rule, profile and update at the pin. The first round of review found five faults of the harness, each fixed with a test that fails without the fix, before the final run: during the trial series a step rule chose and predicted from its frozen start rather than from the series' estimate, which touched the first five answers of the step rules' calibration and the whole run of the thirteen cells whose rule chooses its own tasks; a mastery declared on the very answer at which the child was first truly mastered was left out of R5; R4b lacked the distribution of the true chances the protocol sets beside it; the error of the overall level was published for rules that keep none; and sums taken in a map's random order made a rerun equal only to the last bit. Of the numbers the draft cites, three moved: the answers from true mastery to its declaration, 9.9 to 9.8; the share never declared, 21.9 % to 21.8 %; and the constant step's bias at a target of 0.5, 0.23 to 0.22. The second round found four more, fixed the same way: R4b left out the topics whose count stopped before m attempts, which overstated it — counted for the attempts they made, it moved within five attempts from 47 % to 44 % and within ten from 88 % to 82 %, and its distribution of true chances now covers the same attempts; a comparison with no values and a placement read before its answers went unreported; the shared harness's test of a clone did not check the self-check; and the generator of linked topics drew a number per topic that it threw away, so its children changed. The numbers reach the paper as `\stat{ea3}{…}`; the file holds them at four places, and the paper rounds them as the draft does, so no key holds a rounded copy.
  - **Deviations from the protocol, reported with the results.**
    - Glickman's worked example: computed exactly, it gives a rating of 1464.05 and a volatility of 0.06000, where he prints 1464.06 and 0.05999, because his printed intermediate values are rounded. The test holds every intermediate value at his precision and the final ones to their rounding.
    - The fixed-item Urnings: the protocol's check, a binomial law within 0.01 with items of random difficulty, passes even with the acceptance step removed, so it cannot show that the step works. A stricter companion check, items drawn near the learner, holds the law within 0.003 and fails without the step; both are reported.
    - The recipe is `just research learnersim`; the protocol and the plan call it `just research-learnersim`.
    - The protocol calls every number a mean over the children of its cell, while its own definitions make R4, R5, R4b and the share left unsettled after a jump shares of declarations, of topics and levels, of attempts and of children; they are pooled as defined, and the bootstrap still draws children. R4b, the share declared within the first m attempts, is read as one less a Kaplan–Meier estimate over attempts, since a topic whose count stops before m — the run ends, or the child comes to master the topic — has no share of its own. The paper may cite Kaplan and Meier (1958) when it is typeset (S62).
    - Where the protocol leaves room: R4b's true chances are given in five bins — below 0.5, by tenths up to 0.8, and 0.8 or more; and the error R6 follows after a jump is averaged over the topics with its sign, as R6 reads the lag elsewhere, because a bound of 0.5 on the root mean square would sit at the error a child who never jumped already has, 0.49 after 200 answers.
  - **What paper A may say.** Section 5.2: at the service's target the shrinking step's estimate runs 0.08 logits low after 200 answers; at a target of 0.5 it runs 0.16 high and the constant step's 0.22 high, so the bias does not come from a target above one half alone; under a slope or a floor other than the model's, the service's step and the gradient step settle together. Section 5.3: the chain bounds the chance of being declared mastered only for a child whose true chance is the same on every eligible task, as the protocol states; the draft now says so, with the chain's numbers. Sections 5.4 and 6.3: the results, the trial's cost on children placed well among them. The abstract and the limitations say that mastery comes early and that a learner is followed slowly.
- **Proposals for the product (G6, not critical):**
  - **Mastery is declared early.** Under the service's update, 63 % of declarations came while the child's true chance on the level's middle task was below 0.775, and under every other estimator 54–71 %: the run of three decides, not the estimate. On the eligible attempts of children not truly mastered, the true chance of success was 0.8 or more in 28 % of cases, so the run of three is often earned on tasks the child finds easy. A longer run, five as Kelly et al. found better for transfer, or a gate on the estimate — declare only when the predicted chance on the level's middle task is at least 0.775 — could be simulated as variants before one is chosen.
  - **The step and a child who learns (О-53, Q38).** О-53 expects the missing floor to matter after about a thousand answers. In the simulation it mattered within 200: a child whose level grew by 0.01 logits an answer was left 1.12 logits behind over answers 101–200, and after a jump of one logit 71 % of children had not been caught up with by the 200th answer. A floor of 0.05 shortened the catch-up by 7 answers only. A constant step of 0.1 for θ and 0.2 for δ — the service's steps after 20 answers overall and 20 in a topic — lagged by 0.76 and was no less accurate on children who stay put (0.48 against 0.49). A floor at those values is a different rule, since δ's step comes down to 0.2 only after 20 answers in its topic, and would need a run of its own. How fast real children learn is not known, so the size of the lag is the simulation's, not the product's; the direction is the input to О-53.
  - **The trial's cost.** The trial places a child put a level off far better (0.94 logits off after five answers, against 2.24), but a child whose grade placed them well comes out of it further off than the update alone leaves them (0.87 logits against 0.72 after five answers), and is still 0.49 against 0.47 after 200 answers. A narrower prior than σ₀ = 2.5, or a trial that stops moving θ while the answers agree with the start, could be weighed by the same harness.
  - Q73's proposal, the predicted chance in the answer event, stands: it is what would let real children check these numbers.
- **Who and when:** executor, S38, 2026-10-01.
- **Blocks:** nothing; it affects S57 (the formal model: Sections 5.2 and 5.3), S62 (results and threats) and S63 (discussion).
- **Sources:** [experiments/learnersim/results](experiments/learnersim/results); [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md).

### Q76. What a review costs, and what stays unmeasured

- **Question:** S39 — what E-A4 measured under the frozen protocol, and how its numbers reach paper A.
- **Decision:**
  - **The run.** On the development machine (an AMD Ryzen AI 9 365 with 20 logical processors, `GOMAXPROCS` 20, Go 1.27.1, recorded with the numbers), each of the 447 reference tasks the checks accept — E-A1's — was reviewed in ten passes over all of them, one review at a time, by the service's reviewer in its sandbox. P1: a review took a median of 3.45 ms (95 % interval over tasks 3.37–3.52), 9.14 ms at the 95th percentile, 26.8 ms at the 99th and 63.2 ms at most; judging takes most of it, a median of 3.14 ms, and reading the task with its two solver runs 0.18 ms. All 8,940 solver runs ended normally; the slowest took 35 ms, and the one with the most steps took 2,413,617, 9.7 % of the step limit. P2: three benchmark passes over all 447 tasks give 2.5 MiB and 50,683 allocations a review. P3: the package at every point of the ladder each topic is taught at, 175 in all, built at every count of answers of a whole cycle of the turns its reference tasks take — 15 answers, so 2,625 packages — weighs 17.2–25.2 KiB, at most 39 % of the service's 64 KiB budget, about 5,000 tokens at the median and 6,200 at most by a quarter of its characters, rounded to hundreds as befits the rule. The numbers reach the paper as `\stat{ea4}{…}`.
  - **What the timings hold.** Writing a task out as JSON, which the service receives already written, and leaving its own question out of the comparison are the harness's work, done before the clock starts; the copy of a level's reference list the leave-out makes on every judgement stays inside the timing, a few kilobytes against the 2.5 MiB a review allocates. The interval of a median is over tasks; how much the machine itself moved shows in the median of each pass over all tasks, 3.28–3.73 ms, which the paper gives beside it. Allocations and steps do not move from run to run.
  - **The package** is built for a new child of the first grade of the point's level, with no interests and no notes, from the brief the rule writes when the model asks for that point; a parent's notes and a child's interests lengthen it by their own length.
  - **The draft** no longer opens the cost with C053's 275 µs of 204 ms: that trace was of one request the service answered with Not Found, not of a task.
  - **P5 and P6** wait for T64 and T62–T64: the cost of the deployed service under load, and a family's messages per task. The draft says they are not yet measured. No load is run against the cloud from the research.
  - **The shared harness.** E-A1's code that makes reference tasks into submissions and reviews them moved into `experiments/reviewing`, which E-A1 and E-A4 both use: one prepared review, whose two halves E-A4 times and E-A1 runs in turn, so that both review a task by one path. E-A1 rerun after the move reproduced every one of its result files exactly. C053, the trace of a request answered with Not Found, now says so in the ledger and is marked as used by no paper.
  - **Deviations from the protocol.**
    - P4 is not copied into a data file of its own: its numbers stay in the ledger, C086 for the cost of a task and the cold start and C101, added now, for the instance's pace, each proved by its lines of `docs/load.md` at the pin. A second file with the same numbers would be a second source to keep in step; the typeset paper, which needs them as macros, gets them from a file generated from the ledger itself (Q77), not from a copy.
    - The table of cost (3.2) is the paragraph of Section 6.4 for now — no inference for the operator, the free tier against P4, the family's subscription — and becomes a table in S62 if the page allows.
    - P3 covers every turn of the reference tasks as well as every point, since which tasks a package shows depends on the child's count of answers.
    - Percentiles are read at the index ⌊q·n⌋ of the sorted values, as E-A1 and E-A3 read theirs.
    - While the harness was being built it was run several times, and those runs' records were overwritten, against the rule that every run keeps its records; their medians, 3.28–3.43 ms, lie within the range of the final run's passes. Only the final run's records are kept, and the paper cites them.
    - The recipe is `just research perf`; the protocol calls it `just research-perf`.
- **Who and when:** executor, S39, 2026-10-01.
- **Blocks:** nothing; it affects S62 (results, the table of cost) and S45 (the cloud numbers, when T64 is done).
- **Sources:** [experiments/perf/results](experiments/perf/results); [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md); C086 and C101 in [evidence/ledger.md](evidence/ledger.md).

### Q77. When the sections are typeset, and the model's own numbers

- **Question:** S57 onwards — whether each section task also writes its LaTeX section, which the skeleton of S55 marks `\TBD{S57}` and so on, now.
- **Decision:**
  - The section tasks finish their sections in `paper-a/draft.md` and its Russian copy, the text the author reads and reviews. The LaTeX sections are written from the draft in one pass once the text has settled, in S63, which finishes the paper: until then every change would have to be made in three places, and the LaTeX copy would fall behind unseen.
  - That pass needs every number through a macro (G2), the model's own constants included — the guessing floor, the corridor, the steps, the ladder's shift, the trial's prior, the rating scale — and the numbers the ledger proves, such as those of C086 and C101. The constants are not exported by the product, so they are to be read from its code at the pin by `evidence/product.sh` into `product-stats.txt`, and what follows from them, such as the corridor's bounds on the difficulty scale, computed there too; the ledger's numbers are to reach the macros through a file the ledger tool writes from the claims themselves, each number checked against the line that proves it, so that no hand copy, which Q76 declined, stands beside the ledger. Both are built in that pass, when it is known which numbers the typeset paper keeps.
  - S57 checked every number of Section 5 against SPEC 2 by computing it again from the model: the corridor's bounds 1.4663 and 0.5108 on the difficulty scale and its width 0.9555, the middle at 0.9383, the reach of a topic of grades 1–2 down to −1.4892, the service's steps 4.6 % and 12 % longer than the gradient's at the corridor's ends, a third longer at P = 0.5 and 2.4 times as long at P = 0.3, and 166 rating points a corridor.
- **Who and when:** executor, S57, 2026-10-01.
- **Blocks:** nothing; it moves the typesetting of every section into S63.
- **Sources:** [paper-a/main.tex](paper-a/main.tex); [paper-a/preamble.tex](paper-a/preamble.tex); SPEC 2.3 at the pin.

### Q78. A renamed copy, and the learner model's weak points before submission

- **Question:** S61 asked the author two things. First: the near-duplicate check refuses a copy of a reference task with its people and objects renamed in 95.5 % of cases, 88.2 % with new numbers as well, while R56 says a new setting makes a new task and the protocol had put this class, X09, out of scope; is such a copy a copy or a new task? Second: E-A3 found that the run of three declares mastery early and that the shrinking step lags a child who learns; is the product changed before the paper is submitted?
- **Answer (the author, 2026-10-01):**
  - **A renamed copy is a copy, and its refusal is right.** Names and objects swapped in the same text do not make a new setting; a child would know the task. The paper reads the 95.5 % and 88.2 % as the check working. R56's new setting is a plot written anew, which passes by design. This revises Q74's reading of the class, which counted it out of scope and pooled its two operators at 92.4 %.
  - **The learner model is described as it is.** The paper reports both weak points and Q75's proposals as future work; the product is not changed before submission, and E-A3 is not run again. The author's reason: a child who errs again in the topic is given an easier task to consolidate it.
- **What the executor adds to the second answer, for the paper and for the author:** the correction the author counts on is real but not immediate. A wrong answer makes the next task reinforce its topic, the corridor slides toward easier tasks, and two wrong answers in a row take mastery away [C015, C019]; but a topic declared mastered leaves the rotation until every topic within reach is mastered [C070], so a false declaration is undone only once the topic comes back. The paper says so where it reports the early declarations.
- **Who and when:** the author, 2026-10-01, asked by the executor in S61.
- **Blocks:** nothing; it settles how S61 and S62 read X09 and how S62 and S63 frame E-A3's weak points.
- **Sources:** [experiments/faultinject/results/operators.csv](experiments/faultinject/results/operators.csv); R56 in `docs/decisions.md`; [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md) 1.4; Q75.

### Q79. What Figure 2 shows, and how the evaluation names its weak points

- **Question:** S62 — which of E-A3's measures the paper draws, what a declaration of mastery made below the gate is called, what the paper may say causes such declarations, and how it weighs its many comparisons.
- **Decision:**
  - **Figure 2** draws the fifteen rules of the sweep on children who stay put, by the share of tasks in the corridor and the share of false declarations of mastery, each with its 95 % interval; the service is the filled mark. The estimate's error stays in the text. The figure reads `corridor.csv`, which `learnersim` writes, so a new run redraws it (Q24).
  - **"False", not "early".** A declaration made while the child's true chance on the level's middle task is below 0.775 is called false. On children who stay put there is no later mastery for it to come before, so "early" misdescribed the measure; the paper defines the term once, in the simulation's measures (Section 6.3 of the extended version, 6.2 of the 14-page paper), and uses it throughout.
  - **No single cause is named.** Every rule of the sweep shares the service's mastery rule, and no variant changes its run of three, so the paper says that none of the fourteen other estimates made most declarations true, not that the run of three causes the false ones. This revises Q75's "the run of three decides, not the estimate"; Q75's proposals for the product stand.
  - **Multiplicity is reported, not corrected.** The protocol reads a difference as found when its 95 % interval excludes zero and names no correction; adding one after the data would be a choice made with the results in view. `learnersim` now writes how many primary comparisons there are and how many found a difference, 80 and 65, and the threats to validity name the borderline ones as the likeliest chance findings.
  - **The simulated children's model is a named threat.** They answer by the curve the service assumes, which favours its update; the two generators that depart from it change only its slope or its floor.
- **Who and when:** executor, S62, 2026-10-01.
- **Blocks:** nothing; S63's discussion uses the same terms.
- **Sources:** [experiments/learnersim/results/corridor.csv](experiments/learnersim/results/corridor.csv), [experiments/learnersim/results/numbers.txt](experiments/learnersim/results/numbers.txt); [experiments/PROTOCOL-A-offline.md](experiments/PROTOCOL-A-offline.md) 2.7; Q75, Q78.

### Q80. Fitting paper A into fourteen pages

- **Question:** S63 — the draft after S62 runs to about 12,500 words of body text and cites 89 works, while AIED's full paper allows 14 LNCS pages with the references, about 5,500–5,800 words of text beside the figures, tables and a reference list of about 30 entries (measured: one LNCS page of this template holds about 650 words of running text). What is cut, and where does it go?
- **Decision:**
  - **The paper is cut, not squeezed.** `paper-a/draft.md` becomes the 14-page paper, to the page budget of S06's skeleton (`literature/exemplars.md`), and its Russian copy follows it. The long text is kept unchanged as `paper-a/draft-extended.md` and its Russian copy, the extended version: every result, proof and source the paper drops stays there, for the artifact and for a journal version (`literature/venues.md`: the long versions can go to IJAIED). The paper points to it.
  - **What the paper keeps:** the gap and the split; the nine checks with the solver run twice and what they cannot see; the response model, the update's relation to maximum likelihood, the corridor, the mastery rule and the trial series; both experiments' designs and main results with their negative findings (G9); cost in a paragraph; threats and limitations by kind; discussion, ethics and conclusion.
  - **What moves to the extended version:** the derivations behind the step ratios and the chain beyond their results; the secondary variants of the simulation; the detail of the sandbox's memory pricing, the sketch's error analysis and the readability calibration; the first live runs beyond a sentence; the full product survey of the related work; works cited for context only.
  - **The reference list** keeps the works a claim of the paper rests on and the nearest works the novelty is stated against, about 35; the others stay cited in the extended version.
  - **The typeset paper is written from the cut draft,** every number through a macro (Q77), and is built until it fits.
- **Who and when:** executor, S63, 2026-10-01.
- **Blocks:** nothing; the author's checks of Sections 3–6 (K04) read the extended version, which is the text those tasks wrote.
- **Sources:** `literature/exemplars.md` (skeleton of paper A); `literature/venues.md`; `literature/writing-standard.md` §1 ("a section that does not help answer one of them is cut").

### Q81. How every number reaches the typeset paper, and what the rubric found

- **Question:** S63 — the typeset paper may state no number typed by hand (G2, Q77). Where does each number come from, how is a number the text itself chooses told apart, and does the paper pass the writing standard?
- **Decision:**
  - **The model's constants** are read from the product's code at the pin: `evidence/product.sh` writes a test into each package of the unpacked commit, after every count it makes, and the test prints the constants, unexported ones included, with what follows from them — the corridor's bounds on the difficulty scale, the step's excess over the gradient at chosen points, the ladder's shifts. Sixty-one facts were added to `product-stats.txt`; the old ones did not change.
  - **The ledger's numbers** are written by the ledger tool from a `numbers` field of a claim, each number found as a whole number in a line that proves the claim, into `evidence/ledger-numbers.txt` (C050, C084 so far).
  - **Other works' numbers** stand in `literature/numbers.txt` with the passage they come from. Two numbers the draft had without such a passage were dropped from the text: the 13-year-olds of the national assessment, and the grades of Ariyarathne et al.; Beast Academy's range is the ages its site states, 6–13, not the grades 1–5 the notes had.
  - **The experiments' design** is written by the experiments themselves: E-A1 adds the tasks accepted in all, the operators one check alone refused, and exact intervals for every out-of-scope class and operator; E-A3 adds its children, answers, generators, sweep, spreads and the window of the lag. Both were run again, and every other result came out byte for byte the same.
  - **Rounding and form** are chosen where a number is cited, its value never: `\statr`, `\statpct`, `\statabs`, `\statnum`, `\statword`, and `\statmin`/`\statmax` over a named list of keys for a range across cells.
  - **A number the text chooses** — a confidence level, the point a function is evaluated at — is marked `\given`, so it stays visible as a choice. The names that carry digits (RQ1, Glicko-2, XChaCha20-Poly1305, the defect classes) are listed in `paper-a/names-with-digits.txt`. `tools/handtyped` fails the build on any other digit in the paper's text, apart from a formula's own 0, 1 and 2, and in the words its figures print, whose coordinates are layout. A claim in words that rests on a count — every case was refused, an option was never refused — is guarded by `\statzero`, which stops the build unless the count is zero, so a rounded share of 100.0 cannot stand for it.
  - **The rubric** (`literature/writing-standard.md`): one central contribution in the title; contributions as a checkable list tied to sections, each backed by a result or the ledger, the negative results stated in the abstract (G9); limitations and threats by kind; every claim marked to the ledger in the draft; every number through a macro; every citation to a checked entry, 30 in all (29 once Q82 dropped one); formalism that defines or proves; the anonymous build hides the name, the repositories and the commit; the AI use disclosed in the evaluation's method as Springer asks, from the AI-use log, until the author withdrew the disclosure (Q83). Open: the CRediT statement and the declarations wait for K05; the release of the experiments' code and data for S65; the author's reading of sampled cases for K04, which left the paper instead (Q86). The named build runs to 15 pages because of the declarations' placeholders; the anonymous build, the one submitted, has 14.
- **Who and when:** executor, S63, 2026-10-01.
- **Blocks:** nothing.
- **Sources:** `evidence/product.sh`; `tools/ledger`; `tools/handtyped`; `paper-a/preamble.tex`; `literature/numbers.txt`; Q77, Q80.

### Q82. Where the reference tasks' ideas come from

- **Question:** reading the Russian PDF of paper A, the author asked that the paper say where the reference tasks came from: an earlier study chose the sources, materials whose copyright has expired. The records say less: the prototype's study of task sources sorted the sources by copyright and advised taking those in the public domain when the choice was otherwise equal (P059), while decision D23 of the prototype and the product's tasks T37–T39 say the tasks were written in the genre of all the sources that study surveyed, protected ones among them, and no task records its source (P013, C066). How should the paper put it?
- **Answer (the author, 2026-10-02):** the ideas of the reference tasks are taken from problem books in the public domain. The paper says so, together with what the records hold: the earlier study chose those books, Claude wrote the tasks from their ideas in wording of its own, and the author checked them.
- **What the executor adds:** the statement is the author's; the records support the list of public-domain books and the rule of taking ideas and never wording, not the origin of each task, which no task records. A reviewer who asks for the source of a given task cannot be answered from the repository. To keep the anonymous build at 14 pages with the new sentence, Section 5.2 dropped its clause on Elo's estimates running low under a constant step, with its reference (Bolsinova et al. 2025); the extended version keeps both.
- **Who and when:** the author, 2026-10-02, asked by the executor.
- **Blocks:** nothing.
- **Sources:** P059 (`research/12-task-sources.md` of the prototype); P013 (D23 of the prototype); C066.

### Q83. Whether paper A discloses the AI assistance

- **Question:** reading the Russian PDF of paper A, the author asked to remove the sentence opening Section 6 that said the paper was drafted with Claude Code from the evidence ledger and that the author reviewed every section and answers for it. Rule G7 has the papers disclose AI assistance by the venue's rules, and Springer, which publishes AIED, asks that the use of a language model be documented in the methods; only copy editing of a text a human wrote needs no declaration (`literature/venues.md`). The executor offered one plain line at the end of the paper instead.
- **Answer (the author, 2026-10-02):** remove it altogether: the paper does not disclose the assistance.
- **What the executor adds:** the sentence is gone from both copies of the draft and from the LaTeX, and the placeholder of the named build's contributions no longer points to it. The AI-use log stays in the repository as it is, and rule G7 is not changed. Paper A departs from G7's second point by the author's decision. Springer's rule applies when the paper is submitted. S66 builds the submission package and checks the venue's rule again; if the rule still asks for the disclosure, the question returns to the author there. One mention of the assistant stays, in Section 6.1: the sampled out-of-scope cases were read by the AI assistant that ran the experiment, and the author still has to check its verdicts. That sentence states who did the reading, which is part of the method, so it is not a disclosure of help with the writing. It changes once the author has read the cases (K04). On 2026-10-03 the sentence left the paper instead (Q86).
- **Who and when:** the author, 2026-10-02, asked by the executor.
- **Blocks:** nothing now; S66 checks the venue's rule.
- **Sources:** `literature/venues.md` (Springer); rule G7; `ai-use-log.md`.

### Q84. What the artifact of paper A holds, and how its anonymous copy is made

- **Question:** S65 — the paper promises that the experiments' code, data and seeds and the evidence ledger are released with the service's code and reference tasks. The research tree stays on this machine and out of the repository (`.gitignore`), so nothing of it is public yet. What goes into the artifact, how is it shown to reproduce, and what does a double-blind review link to?
- **Decision:**
  - **What it holds.** `just research release` writes one tarball. `service/` is the product's repository at the ledger's pin, `52ce86908135`, the tree every claim of the ledger points into. Only the research plan of that day, the Russian `research/RUN.md`, is taken out. `service/research/` holds the protocol with its OpenTimestamps proof, the code of E-A1, E-A3 and E-A4 with their shared harness and their results as the paper cites them, the ledger with its claims, facts and scripts, `literature/numbers.txt`, and `paper-a/numbers.tex`, the macro behind every number the paper prints. A README at the top says what is where, which section of the paper each part backs, and how to rerun each experiment. Licences: the product's MIT licence covers the research code and data as well; the Go modules the experiments add are listed in a third-party file written by go-licenses.
  - **What it leaves out, for now.** The extended version is a working draft as it stood after S62, with the ledger's markers, TBD marks and none of the corrections since; it goes in once it has been brought in step with the paper, which is still open in S65. The AI-use log is not in the artifact either: the paper does not disclose the assistance (Q83), so whether the log is published is the author's call (K11).
  - **Reproduction.** The experiments were run in the working tree, whose code under test was the pin's byte for byte (`provenance.txt`). Run from the pin's tree in the artifact, E-A1 reproduced every result file byte for byte in about two minutes, and E-A3 in about four. The one input E-A1 reads beside the code is `reading.csv`, the verdicts of the case-by-case reading, which the run merges into its tables; the README says to copy it in first. `just research release-check` unpacks the tarball into an empty directory, reruns E-A1 and E-A3 there and compares every file they write, and reruns E-A4 and compares its package sizes, the part of it a machine does not change; both copies pass. `just research release` writes the paper's numbers afresh before it copies them, and refuses a file of a kind the artifact does not ship, since the research tree is out of git and nothing else keeps a stray file out.
  - **The anonymous copy.** `just research release anonymous` keeps of the product only the packages the experiments and their tests import, as `go list` finds them, `content/` among them, and its note names them. It replaces the system's name in every case form and with any separator, in file contents and Go module paths alike, so the code still builds. It replaces the prototype repository's name, and masks with zeros every run of seven to forty hexadecimal characters, in either case and wherever it is joined to a word, that names a commit of either repository, found by asking git rather than listed by hand; it therefore needs the prototype's clone and refuses to run without it. A run of digits alone is read as a number, and refused if it begins the hash of a commit the research names. It fails if the name, the prototype repository, the author's surname in any of its spellings or a commit is left in any file, binary ones included, or in a file name. The search finds commits in the local clone, so a commit the clone does not hold would pass; none is in the artifact, as a check by hand of every hexadecimal word left in it showed. Its files carry the first day of 2000 rather than the pinned commit's time, which would point to the commit as surely as its hash; in both copies every file has one owner and one set of modes, so the same sources make the same tarball. A note in the copy says what was replaced and which checks only the named copy allows: the OpenTimestamps proof attests the protocol as written, whose SHA-256 the note gives, and the prototype's quotations point to a repository whose name is hidden.
  - **Where it is hosted** is the author's: K11.
- **Who and when:** executor, S65, 2026-10-02.
- **Blocks:** the anonymised artifact link and the DOI wait for K11; the build for submission stops until the link is set (Q86).
- **Sources:** `release/assemble.sh`, `release/check.sh`, `release/README.md`, `release/ANONYMISED.md`; the reruns of E-A1 and E-A3 from the pin's tree.

### Q85. What the internal review changed in paper A, and what waits for other model families

- **Question:** S64 — three internal reviews in the roles the plan names, the methodologist, the AIED specialist and the skeptic. Rule G10 asks for reviewers of other model families, and their calls wait for the author (K01; rule 4). What does a review by the executor's own model change, and what of the task is left?
- **Decision:**
  - **The passes.** Three sessions of the executor's model, each given one role, AIED's criteria and read-only access to the repository, reviewed the anonymous build; the reports and a meta-review are in `paper-a/reviews/`. Every finding was checked against the paper, the results, the protocol and the product at the pin before it was acted on, and the meta-review gives each its decision: 18 fixed, 1 fixed in part, 1 dropped, 1 left to the author and 7 open. Q86 has since settled the one left to the author and two of the open ones.
  - **Reporting the protocol asks for.** The protocol says every primary comparison is reported whichever way it falls (§2.7), and the paper had reported comparisons 2 and 4 in part: on learners, Glicko-2 with one level per child lagged 0.24 logits against the service's 1.12 and kept more tasks in the corridor, and a floor under the step shortened the catch-up after a jump. Both are now in §6.2, with the generators beyond the first three. D14b, refused in 94.6 % of its cases, is named as not reliably caught, by the protocol's rule for a discovery operator under 95 % (§1.9), and the classes X01 and X04, which pool two operators on the same hosts, are shown by operator, each with the exact interval the protocol prescribes for an operator.
  - **Framing.** RQ1's in-scope result is called what it is, a test that the checks do what they state, and the abstract names what passed beside it. The title drops "Verified": the checks confirm the key against the model's own formalisation, and nothing checks that a task is olympiad-style; "the Service Checks" already says what is done.
  - **New evidence.** C102: the reference tasks are in English whatever the chat's language, so a translated copy is not caught. C103: the first local run's one refusal by the solver was a false alarm, and two hints gave too much away.
  - **Room.** Table 1, the learning goals and their mechanisms, became one sentence of §3, and several details went to the extended version; the anonymous build stays at fourteen pages.
  - **G10 is not met by these passes**, since the reviewers and the writer share a model family. The passes by other families wait for K01 and the author's confirmation of the calls, so S64 stays `[~]`.
- **Who and when:** executor, S64, 2026-10-02.
- **Blocks:** nothing.
- **Sources:** `paper-a/reviews/S64-*.md`; `experiments/PROTOCOL-A-offline.md`, §1.9 and §2.7; C102, C103.

### Q86. The final text of paper A: no new evaluation, no unchecked reading, no disclosure

- **Question:** after S64 the executor asked the author whether paper A should judge real tasks a model writes (the reviewers' first reason to reject), who would check the 120 sampled out-of-scope cases whose verdicts the AI assistant recorded (K04), and whether the paper discloses the AI assistance after all.
- **Answer (the author, 2026-10-03):** the author has checked generated tasks informally for several weeks and finds them good; no disclosure; the final text is to be prepared now, and the author reads and edits it next.
- **Decision:**
  - **No new evaluation for this version.** The informal checks cannot enter the paper. The paper cites only what is recorded and can be checked (G1, G2), and these sessions are not. A child's results are a child's data, which the research uses only after an ethics review (G5), and the paper says no child took part. In a double-blind review, such details would also point to the author. False refusals and the quality of real tasks stay named as unmeasured, as before. The experience is a reason for a study with children later, under G5. Nothing about the child is recorded here.
  - **The sampled reading leaves the paper.** The author does not check the 120 cases, and keeping them would mean naming the AI assistant that read them, which the author's decision rules out. So the sentence "4 of 10 …", its TBD and the threat "one reader read the out-of-scope cases" are gone from both copies and the LaTeX. The refusal rates of Table 2 stand on their own. The verdicts stay in the experiment's results and the artifact, where `reading.csv` and the README say who recorded them. The review version links that artifact, so whether its README may name the assistant, or the artifact should go without the reading, which the paper's numbers do not need, is the author's call under K11.
  - **No disclosure** (Q83 stands). The reviewer's suggestion of a line in the camera-ready acknowledgements is declined.
  - **The artifact sentence of §8** no longer waits on S65. The review version names an anonymised artifact holding the reference tasks with their solvers, the code the experiments run, their data and seeds and the ledger. The named version keeps one TBD, the artifact's DOI, until K11. The review version's link stays a placeholder in drafts, and the build for submission stops on it until the link is set in `paper-a/preamble.tex`.
  - **What is left before submission:** K11, which supplies the anonymised link and the DOI; K05, which supplies the credits of the named version; and the checklist in `paper-a/submission.md`. `just research paper-a-final` stops on one thing only, the anonymised link; the anonymous text is otherwise complete at 14 pages and 29 references.
- **Who and when:** the author, 2026-10-03, asked by the executor; recorded by the executor.
- **Blocks:** nothing; the author reads the text next.
- **Sources:** Q83, Q85; rule G5; `paper-a/sections/evaluation.tex`, `discussion.tex`.

### Q87. The frozen experiments of paper A run the product at its commit, not the working tree

- **Question:** the module reached the product through `replace … => ../` (Q21), so the paper's frozen experiments compiled whatever the working tree held. The product's next change to the rating, a floor under θ's step, would turn `TestStepRulesAreTheServicesUpdate` red and make a rerun of E-A3 report the new product's numbers as the paper's.
- **Why it matters:** the paper's numbers must come from the code it describes, the product at the ledger's pin `52ce86908135` (the protocol, Q73). A rerun on a changed working tree is a deviation, and with the floor a silent one.
- **Decision (the author, 2026-10-04, asked by the executor of the product's T72.7):** the module requires the product at `v0.1.53`, the tag on the pin's commit, and no longer replaces it with the working tree. It revises Q21 in that alone: the module stays nested and imports `internal/` packages as before.
  - `_provenance` writes "as at the pin" when the version required is the pin's commit and nothing replaces it, rather than comparing the working tree with the pin.
  - `deps-check` compares this module's versions with those of the product at the version required, read where Go keeps it, rather than with the working tree's.
  - The artifact keeps building from the product it ships: `assemble.sh` points the module at it with a `replace` of its own.
  - What the product does now is measured by the learners' bench, `tools/learners`, which follows the product as it changes.
  - Left to T73.1: the module's checks in CI, its coverage, and how the module would hold both sets of numbers if an experiment ever needed the live product.
- **Who and when:** the author, 2026-10-04; done by the executor of T72.7.
- **Blocks:** nothing; the product's rating may change.
- **Sources:** Q21, Q73, Q84; the product's RUN.md, T72.7 and T73.1; `go.mod`; `justfile` (`_provenance`, `deps-check`); `release/assemble.sh`.

## Answer when convenient

Questions about the author's reasons and intentions. They block nothing: each has a default, or needs none.

The rest of S05's list is answered elsewhere:
- in RUN.md, «Решения по открытым вопросам»:
  - clarity against productive struggle, with transfer as the main criterion: Q09;
  - the heuristic label on each task: Q10;
  - the textbook's languages: Q11;
  - where the textbook engine runs: Q13;
  - who approves a change: Q14, extended by Q44 here;
  - what makes a task olympiad-style: Q15, applied by Q35 here;
  - what "uncertainty" means: Q16;
  - Graphiti: Q17, which left with paper B (Q60).
- in RUN.md, «Критические вопросы автору»:
  - subscriptions and hardware: K01;
  - the share of the limits: K02;
  - the author's time: K04;
  - authorship and funding: K05;
  - olympiad teachers: K06;
  - trials with children: rule G5 (K07 left with paper B, Q60);
  - arXiv: K09.

Questions for the author about reasons and intentions. Each has a default, so each blocks nothing.

### Q45. Which mistakes of the models have you seen yourself?

- **Question:** Which mistakes of the models have you seen yourself — in the prototype's runs and since?
- **Why it matters:** they become defect classes of E-A1 (S37).
- **Recommendation:** the default until the author answers: the classes S36 derives from the checks and the literature.
- **Answer (the author, 2026-09-25):** the executor decides; the default stands.
- **Blocks:** nothing; answer when convenient.

### Q46. How do you, as a coach, tell an olympiad problem from a hard textbook problem?

- **Why it matters:** the blind review judges tasks by it (S44).
- **Recommendation:** the default until the author answers: the rubric of Q35.
- **Blocks:** nothing; answer when convenient.

### Q47. What should happen to the ratings when a child moves up a grade level?

- **Why it matters:** paper A has to say what its numbers mean across a move.
- **Recommendation:** the default until the author answers: Q39.
- **Answer (the author, 2026-09-25), in two steps:**
  - First: as the product's code does it. At the pinned commit the grade chooses the catalogs and the readability thresholds, and the ratings carry over a change of grade unchanged (C065).
  - Then the author changed the product's design: **the grade sets only the starting point** (product task T39b, planned in the product's plan on 2026-09-25, not built yet):
    - every task stands on one ladder for grades 1–6: its difficulty is its level's shift plus (d − 3); the shift is 0 for grades 1–2, then 2.5 and 5 — a guess, recorded as a decision and to be checked on live runs;
    - a child starts at the shift of their grade; a trial series of 5 tasks on different topics follows, and after each answer the level is re-estimated from the start and every trial answer together;
    - topics open by ability rather than by grade; readability follows the task's level; one rating scale covers the whole ladder, with about 11 ranks; a change of grade is only a label.
  - For the research: paper A describes the model at its final pinned commit (S56). Once T39b is built, section 5 describes the ladder and the trial series. Every claim tied to the grade is re-proven or replaced at the re-pin: C010, C014, C016, C018, C019, C028, C030, C049, C065 and C070. S38 simulates the new behaviour — how well the series places a child whose true level is a step below or above the start (S38 did, Q75).
- **Blocks:** nothing; answer when convenient.

### Q48. Which mistakes do children make that the trap catalog lacks?

- **Why it matters:** the catalog is both paper A's diagnosis and paper B's signal.
- **Recommendation:** the default until the author answers: S31 looks for them (Q41).
- **What S31 found:** the list in Q68 — place value, calculation slips, the order of operations, fractions, factors and multiples, clocks, perimeter and area on a grid, Venn diagrams, and decoding errors, where a word itself is misread, and careless slips. The author's own experience is still the better source.
- **Blocks:** nothing; answer when convenient.

### Q49. Should the pace and "I don't understand" change the next task, as SPEC 2.2 says, or is the code right?

- **Why it matters:** paper A must describe the rule the product runs.
- **Recommendation:** the default until the author answers: paper A describes the code (Q42).
- **Answer (the author, 2026-09-25):** paper A leaves this detail out.
- **Blocks:** nothing; answer when convenient.

### Q50. Which heuristics do you teach your child, and from which grade?

- **Left with paper B (Q60):** kept as a record; it binds nothing, and the tasks it names no longer exist.
- **Why it matters:** the heuristic catalog is the textbook's axis.
- **Recommendation:** the default until the author answers: S18's catalog (Q43).
- **Blocks:** nothing; answer when convenient.

### Q51. Does "uncertainty" mean what RUN.md Q16 says: the method is not given, and the child looks for it?

- **Left with paper B (Q60):** kept as a record; it binds nothing, and the tasks it names no longer exist.
- **Why it matters:** paper B's objective has to measure the right thing.
- **Recommendation:** the default until the author answers: Q16.
- **Blocks:** nothing; answer when convenient.

### Q52. How does the textbook relate to the paid edition PRODUCT-V1 mentions?

- **Left with paper B (Q60):** kept as a record; it binds nothing, and the tasks it names no longer exist.
- **Why it matters:** paper B's design should not assume a feature that belongs elsewhere.
- **Recommendation:** the default until the author answers: paper B assumes the free, MIT-licensed project and names no paid feature.
- **Blocks:** nothing; answer when convenient.

### Q53. What should the ethics section say about how the project is funded?

- **Why it matters:** venues ask for funding and conflicts of interest to be stated.
- **Recommendation:** the default until the author answers: nothing beyond what K05 settles.
- **Answer (the author, 2026-09-25):** settled before submission, with K05; the review version names nobody.
- **Blocks:** nothing; answer when convenient.
