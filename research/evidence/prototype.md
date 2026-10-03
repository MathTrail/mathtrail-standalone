# Evidence: the prototype

Rule G1 of [RUN.md](../RUN.md), for the prototype: every claim the papers make about the prototype `github.com/MathTrail/llm-taskgen-prototype`, or take from its research notes, is recorded here with its proof. The product's claims are in [ledger.md](ledger.md). S03 wrote this file.

## Snapshot

The local copies `prototype/` and `docs/prototype/` are outside git, and the product deletes them once it has ported what it needs. So no proof here points at them: every proof points at a commit of the prototype's public repository, which anyone can open (Q07, Q31).

- **The commit:** `02638353482e` (2026-09-16), the head of `main`, where the repository's only pull request was merged. All 59 commits of the repository lead to it. Its parent `eafbe8abdac4` has the same tree.
- **How it was chosen:** `just research prototype-snapshot` compared every local file with the commit, byte for byte, by git blob hash. The result is in [prototype-snapshot.txt](prototype-snapshot.txt):
  - 120 of the 131 local files match. They cover every source file, data file, document and research note of the prototype, and the five frozen copies in `docs/prototype/`: the SPEC, the decision log and research notes 10, 12 and 13. `CLAUDE.prototype.md` matches `CLAUDE.md`, of which it is a renamed copy.
  - 10 are Python byte-code caches (`__pycache__/*.pyc`) of sources that do match. The commit does not have them, nothing can rest on them, and they are not kept.
  - 1 is not the prototype's own file: `docs/prototype/README.md`, the product's index of the frozen copies. It is kept in `research/archive/` with its SHA-256 in `research/archive/MANIFEST.sha256`, as context.
  - Every file at the commit has a local copy that matches it.
- **SHA-256 of every file at the commit:** [prototype-files.sha256](prototype-files.sha256), 115 files, computed from the commit itself.

## What counts as a proof

- **A file proof** `path:line@02638353482e` is that line of that file at a commit of a public repository, which cannot change. G1 asks three things of a file outside git, and here is where each one is:
  - the link is the commit of the public repository;
  - the hash is the file's SHA-256 in [prototype-files.sha256](prototype-files.sha256);
  - the quote is the claim's `match` in [claims-prototype.json](claims-prototype.json): the fragment of the line, with a backslash before each character a regular expression treats specially. The table does not show it. Most of the prototype's lines are Russian while the files of this folder are English, and the line itself, at a commit that cannot change, is the quote's source (Q31).
- **A fact** `key = value` was computed at the commit by [prototype.sh](prototype.sh), from a clone of the public repository. It was not read from the local copy.
- **Statuses** are those of [ledger.md](ledger.md), and for the prototype they say how far a claim is checked:
  - *built* — checked: the prototype's code or data at the commit show it;
  - *specified* — checked as a statement: the prototype's SPEC, decision log or notes say it about themselves;
  - *measured* — the prototype's own record of a run. The database the runs wrote to was never published, so nothing here can reproduce the number;
  - *unverified* — what a research note reports from a source. The notes were compiled by web search, and an auxiliary model retold the pages (P036). Every such number stays unchecked until a task of phase 1 or 2 reads the original. That task states its verdict in its own notes, and the papers cite the source, never the claim; the claim stays as the prototype's record (Remark 4).

## Commands

```sh
just research prototype-fetch              # clone the public repository into research/.cache/, or update the clone
just research prototype-pin <commit>       # recompute prototype-stats.txt at a commit
just research ledger                       # re-resolve every claim, the product's and the prototype's
just research ledger-check                 # the same without writing; fails if a table is out of date
just research prototype-snapshot <commit>  # only while the local copies exist
just research seed-check                   # every link of the research notes is in literature/seed.bib
```

## Claims

Generated from [claims-prototype.json](claims-prototype.json) by `just research ledger`; edit the claims file, not the table.

<!-- ledger:prototype:begin -->
Pinned commit: `02638353482e` of `github.com/MathTrail/llm-taskgen-prototype`, 2026-09-16. Every fact below was computed at that commit.

### Architecture: from API agents to an MCP server

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P001 | The prototype was first designed as four roles on the Claude API — Methodist, Generator, then Analyst and Skeptic in parallel — with the two verifiers seeing only the question and the options. | specified | `docs/decisions.md:31@02638353482e` | not in paper A |
| P002 | The code that called the Claude API was removed before the first live call: no task was ever written by the agent pipeline. | specified | `docs/decisions.md:65@02638353482e` | not in paper A |
| P003 | The prototype became an MCP server: the chat client's model writes each task within the user's own subscription, and the server makes no LLM API calls of its own. | built | `docs/decisions.md:23@02638353482e`<br>`pyproject.toml:9@02638353482e`<br>`src/taskgen/mcp_server.py:18@02638353482e` | not in paper A |
| P004 | The prototype's decision log records that claude.ai, Claude Desktop and Claude Code did not support MCP sampling, so the task is written in the conversation rather than requested by the server. | unverified | `docs/decisions.md:23@02638353482e` | not in paper A; S34: revision 2026-07-28 deprecates sampling (SEP-2577), and no page read says which hosts support it |
| P005 | With the move to MCP there is no independent verifier: the same model writes and checks the task, the solver program still fails for other reasons than reasoning does, and the author's manual review is the main guard against false accepts. | built | `SPEC.md:434@02638353482e`<br>`docs/decisions.md:23@02638353482e` | not in paper A |
| P006 | A task was accepted when its structure passed, its solver ran in a network-less Docker sandbox and returned exactly the correct option, the self-check had no blocking issue, the readability filter passed and no near duplicate was found; the model had three attempts per request. | built | `SPEC.md:426@02638353482e`<br>`src/taskgen/sandbox.py:80@02638353482e`<br>`config.yaml:3@02638353482e` | not in paper A |
| P007 | Near duplicates were PostgreSQL pg_trgm similarities of 0.6 or more against the bank and the reference tasks. | built | `config.yaml:27@02638353482e` | not in paper A |
| P008 | The prototype's ratings are the product's: P = 0.2 + 0.8·σ(θ + δ − β), K = K₀/(1 + 0.05·n) with K₀ = 0.2 for θ and 0.4 for δ, and the corridor [0.70, 0.85]. | built | `src/taskgen/rating.py:45@02638353482e`<br>`config.yaml:10@02638353482e`<br>`config.yaml:11@02638353482e`<br>`config.yaml:13@02638353482e`<br>`config.yaml:14@02638353482e` | not in paper A |
| P009 | Unlike the product, the prototype also moved a task's difficulty β after every answer (K₀ = 0.4), because accepted tasks went into a bank and were handed to other children. | built | `SPEC.md:375@02638353482e`<br>`config.yaml:12@02638353482e`<br>`SPEC.md:340@02638353482e` | not in paper A |
| P010 | The prototype knew that the answer sat in the model's context and in the arguments of submit_task, where a curious child could read it; hiding it was left to an MCP Apps card after the prototype. | specified | `SPEC.md:69@02638353482e` | not in paper A |
| P056 | The prototype kept everything in PostgreSQL — profiles, history, ratings, the task bank and the logs — and used no vector search: exact fields such as grade, topic and skill must never be vectors. | built | `docs/decisions.md:73@02638353482e`<br>`docs/decisions.md:77@02638353482e` | not in paper A; S04 |
| P057 | The prototype left out Kubernetes, Kafka, Go, workers and queues. | specified | `SPEC.md:30@02638353482e` | not in paper A; S04 (premises 8, 10) |
| P058 | The prototype drew its request flow and the states of an attempt for the API pipeline; after the move to MCP both diagrams were kept only as history. | specified | `RUN.md:62@02638353482e` | not in paper A; S04 (premise 10) |

### Content

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P011 | The prototype's catalogs held 10 topics, 14 traps and 15 skills. | built | `topics = 10`<br>`traps = 14`<br>`skills = 15` | not in paper A |
| P012 | The prototype had 450 reference tasks: five for each of the 90 combinations of topic, grade level (1–2 or 3–4) and difficulty 1–5. | built | `reference_tasks = 450`<br>`reference_cells = 90`<br>`reference_cells_not_five = 0` | A: §3 system |
| P013 | Claude wrote the reference tasks at the author's request, in the genre of Soviet collections of olympiad and graded problems; every answer was checked by a brute-force function and a blind solve by another model, and the author reviewed them by hand. | specified | `docs/decisions.md:81@02638353482e`<br>`reference_check_modules = 10` | A: §3 system |
| P014 | The decision log records a risk: reference tasks in Claude's own style may make generation less diverse; the near-duplicate check and the author's review were to compensate. | specified | `docs/decisions.md:81@02638353482e` | A, extended version only: §6 evaluation |
| P015 | Only topics whose tasks a program can brute-force entered the catalog; sequence patterns were excluded because code cannot prove a continuation unique. | specified | `docs/decisions.md:19@02638353482e` | not in paper A |
| P016 | Tasks are text only, because most problems of a major international multiple-choice contest for schoolchildren use a picture and models are much weaker on them. | specified | `docs/decisions.md:17@02638353482e` | not in paper A |
| P017 | Difficulty is a level from 1 to 5 within the grade level, with the starting β = difficulty − 3; the contest the project started from is never named. | specified | `docs/decisions.md:21@02638353482e` | not in paper A |
| P059 | The prototype's study of task sources sorted the sources of task ideas by copyright, named the problem books in the public domain — Perelman, Ignatiev, Dudeney, Loyd and Carroll — and advised taking those when the choice was otherwise equal; only a task's idea and answer were to be taken, never its wording. | specified | `research/12-task-sources.md:13@02638353482e`<br>`research/12-task-sources.md:125@02638353482e`<br>`research/12-task-sources.md:16@02638353482e` | A: §3 system; Q82 |

### Measured by the prototype

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P018 | Readability calibration on the 450 reference tasks: a Flesch–Kincaid margin of +1 passed only 48 % of the grades 1–2 tasks for a first-grader and 66 % for a second-grader; +3 passed 83 % and 96 %, and at least 94 % of the grades 3–4 tasks for grades 3 and 4; the sentence limits passed 90 % of the tasks at each level. | measured | `docs/decisions.md:59@02638353482e`<br>`docs/decisions.md:59@02638353482e` | A: §6 evaluation |
| P019 | First live run (2026-09-15), one seed profile: 2 tasks, both accepted at the first attempt, in 68 s and 87 s; the solver agreed with the model, and the author found both correct and unambiguous. | measured | `RUN.md:318@02638353482e` | not in paper A |
| P020 | Second live run, the same day: all five scenarios, the child played by Claude in headless sessions; 9 requests, 8 tasks generated and all accepted at the first attempt in 44–109 s, 1 served from the bank; the maths of every task was correct and unambiguous. | measured | `RUN.md:318@02638353482e` | not in paper A |
| P021 | The live runs found that the model turned "I don't understand" into a hint without recording it, that a seed pseudonym leaked into the text of a task bound for the shared bank, and that Russian trap explanations used the masculine gender; the last two were fixed in the writing guide. | measured | `RUN.md:318@02638353482e`<br>`RUN.md:318@02638353482e` | not in paper A |
| P022 | The market note sums the live runs as 8 of 8 tasks accepted at the first attempt with a median of 69 s, and calls its sample of 10 tasks small. | measured | `research/13-competitors.md:115@02638353482e`<br>`research/13-competitors.md:203@02638353482e` | not in paper A |
| P023 | By the metrics report the model had changed only three of the rule's briefs, so the blind comparison of the rule's and the model's briefs was deferred. | measured | `docs/decisions.md:69@02638353482e` | not in paper A |
| P024 | In a live run, after a wrong answer in ordering, the child asked for combinatorics and the brief read reinforce combinatorics.enumeration: the goal describes the child's situation, not the topic. | measured | `docs/decisions.md:67@02638353482e` | not in paper A |

### Set as targets, never measured

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P025 | The prototype set targets it never reached a verdict on: over 90 % of 20–30 accepted tasks good by the author's review; no false accepts (at most 2 %); the model's briefs better than the rule's; over 80 % unique plots among 20 tasks of a topic. | specified | `SPEC.md:615@02638353482e`<br>`SPEC.md:616@02638353482e`<br>`SPEC.md:621@02638353482e`<br>`SPEC.md:622@02638353482e` | not in paper A |
| P026 | The file for the author's verdicts holds none: no task of the prototype was ever given a quality verdict on record. | built | `SPEC.md:628@02638353482e`<br>`eval_reviews = 0` | not in paper A |
| P027 | The prototype's other targets were first-attempt acceptance over 50 %, acceptance within three attempts over 80 % and under 90 s per accepted task; only the live runs above speak to them, on 10 tasks. | specified | `SPEC.md:613@02638353482e`<br>`SPEC.md:614@02638353482e`<br>`SPEC.md:623@02638353482e` | not in paper A |

### Decisions that matter to the papers

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P028 | A solver program was adopted because models of one provider make correlated errors and code fails for other reasons; its known limit is that it shares the writer's reading of the problem. | specified | `docs/decisions.md:41@02638353482e` | not in paper A |
| P029 | The self-check carries an explicit ambiguity checklist, because models recognise ambiguity but rarely report it unless asked. | specified | `docs/decisions.md:43@02638353482e` | not in paper A |
| P030 | Traps form a closed catalog referred to by id, because free text cannot be aggregated; Eedi ties distractors to a misconception taxonomy. | specified | `docs/decisions.md:45@02638353482e` | not in paper A |
| P031 | Ratings are computed by code rather than by the language model, because language models are worse learner models than classic knowledge tracing. | specified | `docs/decisions.md:53@02638353482e` | not in paper A |
| P032 | Separate K₀ for θ and δ: with one K₀ = 0.4 two failures in a row moved the topic level by nearly a whole difficulty level; the corridor is 0.96 wide on the β scale, so it holds at most one level and can be empty. | specified | `docs/decisions.md:57@02638353482e` | not in paper A |
| P033 | The collected tasks are never used for fine-tuning: the providers' terms forbid using outputs as training targets without written permission. | specified | `docs/decisions.md:79@02638353482e` | not in paper A |
| P034 | Quality is reviewed by the author alone, an experienced olympiad solver and coach, and a false accept means a mathematical failure only: a wrong answer, an ambiguity, no solution or several correct options. | specified | `docs/decisions.md:95@02638353482e`<br>`docs/decisions.md:97@02638353482e` | not in paper A |
| P035 | The rule: after a failure it reinforces the last topic; otherwise the unmastered topic given longest ago, never-given ones first; two traps, the child's most frequent topped up from the reference tasks'. | specified | `docs/decisions.md:63@02638353482e` | not in paper A |

### What the research notes report (unchecked)

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| P036 | The notes were compiled by web search, with pages retold by an auxiliary model; their numbers are to be checked against the originals. | specified | `research/00-README.md:44@02638353482e` | every task that cites the notes |
| P037 | The notes found no work on generating olympiad problems for grades 1–4; their conclusions on generation come from school word problems and synthetic datasets. | unverified | `research/00-README.md:48@02638353482e` | not in paper A; S31, S35 |
| P038 | SMART-840: 840 problems of the contest from 2020–2024 for grades 1–12, 240 of them for grades 1–4, 69 % with a picture; models did worst on the youngest grades, and difficulty for children did not correlate with difficulty for models. | unverified | `research/01-feasibility.md:20@02638353482e`<br>`research/01-feasibility.md:32@02638353482e` | not in paper A; S31 |
| P039 | EDUMATH: of generated school word problems for grades 3–5, about 90 % were solvable, 77 % had a correct answer and 66 % met every criterion. | unverified | `research/01-feasibility.md:51@02638353482e` | not in paper A; S31 |
| P040 | GSM-Symbolic: changing only the numbers of a problem lowers every model's accuracy, and one irrelevant but plausible clause drops it by up to 65 %. | unverified | `research/01-feasibility.md:41@02638353482e` | not in paper A; S31 |
| P041 | Different language models are homogeneous in similar ways (Artificial Hivemind), so switching models does not restore diversity. | unverified | `research/03-problem-generation.md:42@02638353482e` | not in paper A; E-A2 |
| P042 | OpenMathInstruct-2 removed about 50 thousand of 569 thousand new problems (about 9 %) as paraphrases of test problems. | unverified | `research/03-problem-generation.md:38@02638353482e` | not in paper A; S31 |
| P043 | Eedi (NAACL 2024 Findings): people rated the plausibility of LLM distractors 2.68 of 5 against 3.72 for distractors written by people; prompting with three similar problems worked best. | unverified | `research/04-distractors-misconceptions.md:11@02638353482e`<br>`research/04-distractors-misconceptions.md:25@02638353482e` | not in paper A |
| P044 | When two language models err they often err the same way, more so within one provider, and judges favour answers of their own family. | unverified | `research/05-verification.md:32@02638353482e`<br>`research/05-verification.md:33@02638353482e` | not in paper A |
| P045 | Models often recognise that a problem is ambiguous but do not say so unless asked (Knowing but Not Showing). | unverified | `research/05-verification.md:38@02638353482e` | not in paper A; S31 |
| P046 | PAL: letting the model write a program that an interpreter runs beat chain-of-thought PaLM-540B on GSM8K by 15 percentage points. | unverified | `research/05-verification.md:42@02638353482e` | not in paper A; S31 |
| P047 | Math Garden chooses items so that the expected success rate is 0.75, and the Eighty Five Percent Rule puts the optimal error rate of a class of learning algorithms at 15.87 %. | unverified | `research/06-adaptive-learning.md:23@02638353482e`<br>`research/06-adaptive-learning.md:30@02638353482e` | not in paper A; S32 |
| P048 | Language models, even fine-tuned, fell short of Deep Knowledge Tracing as learner models. | unverified | `research/06-adaptive-learning.md:35@02638353482e` | not in paper A; S32 |
| P049 | Anthropic's terms forbid using outputs as training targets for models without written permission. | unverified | `research/08-finetuning-data.md:37@02638353482e` | not in paper A |
| P050 | The amended COPPA Rule was adopted on 16 January 2025 and took effect on 23 June 2025, with full compliance by 22 April 2026; it asks for a separate verifiable parental consent before a child's data is disclosed to third parties to train AI. | unverified | `research/10-risks-children.md:21@02638353482e`<br>`research/10-risks-children.md:23@02638353482e` | not in paper A; S16 |
| P051 | Texts language models wrote for children scored more than 17 % worse on every readability metric than texts people wrote for the same audience. | unverified | `research/10-risks-children.md:38@02638353482e` | not in paper A; S31 |
| P052 | Claude admits users from 18; ChatGPT from 13 with a parent's consent, and apps in its directory may not target children under 13; so the user of the app is a parent or a coach. | unverified | `research/13-competitors.md:17@02638353482e` | not in paper A; S34 |
| P053 | On USAMO 2025 proofs, models strong at problems with an answer scored in the single digits (Proof or Bluff?), so a brute-force solver does not reach proof-based olympiads. | unverified | `research/13-competitors.md:145@02638353482e` | not in paper A; S31 |
| P054 | The assistants' own study modes appeared in the summer of 2025: ChatGPT Study Mode on 29 July, Gemini Guided Learning on 6 August, and Claude's learning mode generally available on 14 August. | unverified | `research/13-competitors.md:113@02638353482e` | not in paper A; S33 |
| P055 | The only strong olympiad product for young children the market note found is Beast Academy, a finite, English-only bank written by people; Khan Academy already serves checked problems from its bank inside ChatGPT. | unverified | `research/13-competitors.md:14@02638353482e`<br>`research/13-competitors.md:67@02638353482e` | not in paper A; S33 |
<!-- ledger:prototype:end -->

## Remarks

Disagreements inside the prototype's own record, for the author. The research does not edit the prototype.

1. **Two accounts of the live runs.** They differ:
   - RUN.md T23 reports 2 tasks in the first run and 8 generated in the second. That is 10 in all, each accepted at the first attempt, in 44–109 s (P019, P020).
   - The market note reports "8 of 8" with a median of 69 s for "T23 and T24", and in another place a sample of 10 (P022).

   The median appears nowhere else, and the output of the metrics report was not saved. Until the author supplies that output, a paper cites RUN.md's counts and range and not the median.
2. **An earlier PRODUCT-V1.md is public.** The prototype's repository holds `PRODUCT-V1.md` as of 2026-09-16. The product's current version is a different file and stays outside git ([ledger.md](ledger.md), "Context outside git"). Neither is evidence for a paper.
3. **The quality targets were never measured.** The prototype's verdict file is `{}` (P025, P026). No false-accept rate exists for the prototype, only the author's statement that the maths of its 10 live tasks was correct (P019, P020).
4. **What the notes reported, read in the original.** The claims marked *unverified* keep the prototype's record of what its notes said. A literature task that reads a source states the verdict in its own notes, and the papers cite the source, never the claim. S31 found P039 wrong — EDUMATH's 90, 77 and 66 % are the rates at which two annotators agreed, not shares of problems — and P037, P038, P040–P043, P046 and P051 holding, some with corrections; P044, P045 and P053 were not read ([literature/generation-verification.md](../literature/generation-verification.md), "The prototype's claims, checked"). S32 found P047 and P048 holding, within the limits its notes state ([literature/learner-models.md](../literature/learner-models.md)). S33 found P054 holding for ChatGPT and Gemini, with Claude's learning mode first released in April 2025 for universities and its opening to all users unconfirmed by Anthropic's pages, and P055 holding in substance ([literature/llm-tutors.md](../literature/llm-tutors.md)). S34 found P052 holding; P004 was not checked ([literature/mcp-safety.md](../literature/mcp-safety.md)).

## Literature

Every source the research notes cite is in [literature/seed.bib](../literature/seed.bib), marked unverified: 148 sources at 167 addresses. Titles are in English, and each entry names the lines of the notes that cite it. `just research seed-check` fails when a link of the notes has no entry. Checking the entries against DOI, arXiv or DBLP metadata is S08's `citecheck`.
