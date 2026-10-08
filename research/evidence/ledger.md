# Evidence ledger

Rule G1 of [RUN.md](../RUN.md): every claim the papers make about MathTrail or its prototype is recorded with its proof — the product's here, the prototype's in [prototype.md](prototype.md). S02 filled the product part and S03 the prototype part; S02.1 moved the product part to a commit of `main` once the product had grown its endpoint, sign-in, storage and widget. The service's real use began after the product's pin, so its claims are a part of their own, pinned to a later commit (Q88). Later tasks add the claims they need.

## Format

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|

- **ID** — `C001`, `C002`, … for the product, `U001`, `U002`, … for its real use, `P001`, `P002`, … for the prototype. Numbers are never reused.
- **Claim** — one fact, stated the way a paper would state it.
- **Status:**
  - *built* — the code, tests or data at the pinned commit show it;
  - *specified* — only SPEC, the diagrams or a decision log describe it;
  - *planned* — a product task T… will build it;
  - *measured* — a number a report or SPEC recorded; no script here reproduces it, so a paper cites it as that document's measurement;
  - *unverified* — what a research note reports from a source nobody here has checked yet (the prototype's notes, [prototype.md](prototype.md));
  - *remark* — two documents, or a document and the code, disagree; it goes to the author and into no paper as a fact;
  - *context* — it rests on a file no public repository holds (see below).

  Nothing planned is ever presented as a result.
- **Evidence:**
  - For a file in git: `path:line` at a pinned commit.
  - For a prototype file outside git: a short quote, the file's SHA-256, and the commit of the public repository `github.com/MathTrail/llm-taskgen-prototype` where the file matches byte for byte (Q07).
  - For a fact a script computed: `key = value` from [product-stats.txt](product-stats.txt). Where the claim states the number, or rests on lines a count finds and someone read, the claims file records the value as `expect`: a commit with another value fails the run, and the claim is read again.
  - A file that is in no public repository, such as `PRODUCT-V1.md`, is context, not evidence. A claim that rests only on it has the status *context*.
- **Used in** — the paper and section, or the task, that relies on the claim.

## Pinned commit

The commit that paper A describes. S02 records it and S56 fixes it for good, on `main` (Q30). Numbers that depend on the commit, such as the size of the reference corpus, are computed by a script rather than copied.

- **Which commit:** the one [product-stats.txt](product-stats.txt) names; the table below states it.
- **Facts computed at it:** [product-stats.txt](product-stats.txt), by [product.sh](product.sh). The script unpacks the commit with `git archive` into a temporary directory, so no other work in the shared tree can change the numbers.
- **Moving to another commit:**

  ```sh
  just research pin <commit>   # recompute the facts at the new commit
  just research ledger         # re-resolve every claim; fails on a proof that is gone or ambiguous, or a fact that changed
  just research ledger-check   # the same without writing; also fails if a table is out of date
  ```

  Then read the diff of the table. A pattern that still matches exactly one line is taken to be the right line, so a proof that slid onto an unrelated line shows only there.

## Product

Generated from [claims-product.json](claims-product.json) by `just research ledger`; edit the claims file, not the table. Every file proof is a line at the pinned commit; every fact comes from the stats.

<!-- ledger:product:begin -->
Pinned commit: `52ce86908135` of `github.com/MathTrail/mathtrail-standalone`, 2026-09-30. Every fact below was computed at that commit.

### Architecture

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C001 | The product is one Go module, built with Go 1.27.1. | built | `go.mod:3@52ce86908135` | A: §3 system |
| C002 | The service is one stateless Go binary on Cloud Run with no database: the child's profile is kept in the parent's Drive, or in the process's memory for development. | built | `CLAUDE.md:28@52ce86908135`<br>`infra/terraform/service.tf:47@52ce86908135`<br>`internal/store/store.go:12@52ce86908135` | A: §3 system, §8 discussion |
| C003 | The service makes no LLM API calls of its own; the chat's model writes every task. No module of a model provider's SDK is among the product's dependencies, and no model API's host is named in its code. | built | `CLAUDE.md:34@52ce86908135`<br>`README.md:61@52ce86908135`<br>`llm_sdk_modules = 0`<br>`llm_api_hosts_in_code = 0` | A: §3 system, §6 evaluation, §8 discussion |
| C004 | The HTTP server answers the MCP endpoint /mcp, the endpoints of the service's own authorization server — discovery, registration, authorization, consent, Google's callback, tokens and revocation — and /health. | built | `internal/transport/http/router.go:234@52ce86908135`<br>`internal/transport/http/router.go:256@52ce86908135`<br>`internal/transport/http/router.go:262@52ce86908135`<br>`internal/transport/http/router.go:280@52ce86908135`<br>`internal/transport/http/router.go:228@52ce86908135` | A, extended version only: §3 system |
| C005 | Cloud Run is configured for at most 3 instances, 80 concurrent requests each, 1 vCPU, 1 GiB and a 60 s request timeout. | built | `cloud_run_max_instances = 3`<br>`cloud_run_concurrency = 80`<br>`cloud_run_cpu = 1`<br>`cloud_run_memory = 1Gi`<br>`cloud_run_request_timeout = 60s`<br>`infra/terraform/service.tf:73@52ce86908135`<br>`infra/terraform/service.tf:67@52ce86908135`<br>`infra/terraform/service.tf:81@52ce86908135`<br>`infra/terraform/service.tf:82@52ce86908135`<br>`infra/terraform/service.tf:66@52ce86908135` | A: §3 system |
| C006 | The content (catalogs, reference tasks, instructions, schemas, solver templates, drawing frames) is embedded in the binary. | built | `content/content.go:27@52ce86908135` | A, extended version only: §3 system |
| C007 | The instructions version is a prefix of a SHA-256 over the instructions, and it travels with every package. | built | `content/instructions.go:86@52ce86908135`<br>`content/package.go:98@52ce86908135` | A, extended version only: §3 system |
| C008 | A test holds every package a profile allows to a budget of 64 KB, so nothing has to be dropped from one to meet it; the service does not enforce the budget at run time. | built | `content/package.go:23@52ce86908135`<br>`content/package.go:21@52ce86908135`<br>`SPEC.md:484@52ce86908135` | A: §3 system |
| C067 | The product is released under the MIT license. | built | `LICENSE:1@52ce86908135` | A: §1 introduction, §8 discussion |
| C068 | Every card is drawn by one MCP Apps resource, ui://mathtrail/app.html: a single HTML page with its dictionaries inlined, whose content security policy names no outside domain, which MCP Apps reads as no external connections. | built | `internal/transport/mcp/widget.go:11@52ce86908135`<br>`internal/transport/mcp/widget.go:65@52ce86908135`<br>`SPEC.md:1158@52ce86908135` | A: §3 system |
| C062 | The MCP endpoint speaks the protocol over Streamable HTTP without sessions, 2026-07-28 being the newest version it speaks: every request carries what the server needs, so any instance can answer any request. | built | `internal/transport/mcp/server.go:142@52ce86908135`<br>`internal/transport/mcp/server.go:88@52ce86908135`<br>`internal/transport/mcp/endpoint_test.go:126@52ce86908135` | A: §3 system, §8 discussion; S04 (premise 8) |
| C063 | The product's design is drawn as Mermaid diagrams — its context, sign-in and tokens, the tool flows with the task's state machine, the profile file and its storage in Drive — and its implementation decisions are logged from R01 on. | specified | `docs/architecture/01-context.md:1@52ce86908135`<br>`docs/architecture/02-auth.md:1@52ce86908135`<br>`docs/architecture/03-flows.md:1@52ce86908135`<br>`docs/architecture/04-profile.md:1@52ce86908135`<br>`docs/architecture/05-storage.md:1@52ce86908135`<br>`docs/decisions.md:11@52ce86908135` | not in paper A; S04 (premise 10) |

### Learner model

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C009 | The chance of a correct answer is P = 0.2 + 0.8·σ(θ + δ_topic − β): a Rasch model with a guessing floor of 0.2 for five options. | built | `SPEC.md:192@52ce86908135`<br>`internal/domain/rating/rating.go:37@52ce86908135`<br>`internal/domain/rating/rating.go:60@52ce86908135`<br>`internal/domain/solver/options.go:17@52ce86908135` | A: §5 learner model |
| C010 | Every task stands on one ladder for grades 1–6: its β is the shift of its grade level plus (d − 3), for a difficulty d of 1–5, with shifts 0, 2.5 and 5 for grades 1–2, 3–4 and 5–6, so the two hardest difficulties of a level overlap the two easiest of the next. The step of 2.5 is a guess nothing has measured yet. | built | `internal/domain/rating/ladder.go:32@52ce86908135`<br>`internal/domain/rating/ladder.go:66@52ce86908135`<br>`internal/domain/rating/ladder.go:96@52ce86908135`<br>`internal/domain/rating/rating.go:54@52ce86908135`<br>`internal/domain/rating/ladder.go:31@52ce86908135` | A: §5 learner model, §6 evaluation |
| C011 | After the trial series, each answer moves θ by K_θ·(S − P) and δ by K_δ·(S − P), Elo-style, with K = K₀/(1 + 0.05·n): for θ, K₀ = 0.2 and n is the child's answers so far, the trial answers included; for δ, K₀ = 0.4 and n is the answers so far in that topic. | built | `internal/domain/rating/rating.go:45@52ce86908135`<br>`internal/domain/rating/rating.go:46@52ce86908135`<br>`internal/domain/rating/rating.go:49@52ce86908135`<br>`internal/domain/rating/rating.go:102@52ce86908135`<br>`internal/domain/rating/rating.go:114@52ce86908135`<br>`internal/domain/profile/answer.go:224@52ce86908135` | A: §5 learner model |
| C012 | Only correctness enters the update: "I don't know" counts as a wrong answer, and neither the hint nor the pace enters. | built | `internal/domain/rating/rating.go:102@52ce86908135`<br>`internal/domain/profile/answer.go:29@52ce86908135` | A: §5 learner model |
| C013 | Task difficulty β enters each update as an argument and is neither updated nor stored: every task is written for one child and never handed out again. | built | `internal/domain/rating/rating.go:102@52ce86908135`<br>`internal/domain/rating/rating.go:67@52ce86908135`<br>`SPEC.md:240@52ce86908135` | A: §2 related work, §5 learner model |
| C014 | The target corridor is P ∈ [0.70, 0.85]. The recommended point is the one of the topic's points — every difficulty of every level the topic is taught at — whose P is closest to the corridor's middle, 0.775, the easier of two equally close on either side of the middle, marked inside, harder or easier than the corridor. | built | `internal/domain/rating/corridor.go:13@52ce86908135`<br>`internal/domain/rating/corridor.go:14@52ce86908135`<br>`internal/domain/rating/corridor.go:19@52ce86908135`<br>`internal/domain/rating/corridor.go:75@52ce86908135`<br>`internal/domain/rating/corridor.go:111@52ce86908135`<br>`internal/domain/tutor/tutor.go:378@52ce86908135` | A: §3 system, §5 learner model, §6 evaluation |
| C015 | A topic is mastered once it has at least 5 answers and at least 3 qualifying answers since its last wrong one: correct, on a task whose P was at most 0.775, and without the hint. A correct answer that does not qualify neither adds to the run nor breaks it. Mastery is lost after 2 wrong answers in a row. | built | `internal/domain/profile/answer.go:48@52ce86908135`<br>`internal/domain/profile/answer.go:51@52ce86908135`<br>`internal/domain/profile/answer.go:54@52ce86908135`<br>`internal/domain/profile/answer.go:350@52ce86908135` | A: §5 learner model, §7 limitations |
| C016 | The child sees a chess-style number, R = 1500 + (400/ln 10)·level, on one scale for every child, and one of eleven ranks, each a corridor — 166 points — wide. Neither is shown during the trial series. | built | `internal/domain/rating/scale.go:11@52ce86908135`<br>`internal/domain/rating/scale.go:12@52ce86908135`<br>`internal/domain/rating/scale.go:17@52ce86908135`<br>`internal/domain/rating/scale.go:28@52ce86908135`<br>`SPEC.md:301@52ce86908135` | A, extended version only: §5 learner model, §8 discussion |
| C017 | An answer is tagged fast under one minute and slow over three minutes. | built | `internal/domain/profile/answer.go:39@52ce86908135`<br>`internal/domain/profile/answer.go:40@52ce86908135` | no paper: paper A leaves the pace out (Q49); paper B, which used it, left the plan (Q60) |
| C065 | The grade sets only the start: a child begins at the shift of the level the grade falls into, and a grade changed later moves no rating. The rule never reads the grade. | built | `internal/domain/rating/ladder.go:120@52ce86908135`<br>`internal/domain/profile/student.go:18@52ce86908135`<br>`internal/domain/tutor/tutor.go:18@52ce86908135`<br>`internal/transport/mcp/profile_test.go:372@52ce86908135` | A: §5 learner model |
| C080 | The first five answers are a trial series. After each of them θ is set to the most likely level given the start and every trial answer at once — a normal prior around the start with a spread of one level shift, 2.5, chosen by simulation, times the chance of each answer — and the topic corrections do not move. | built | `internal/domain/rating/trial.go:11@52ce86908135`<br>`internal/domain/rating/trial.go:20@52ce86908135`<br>`internal/domain/rating/trial.go:44@52ce86908135`<br>`internal/domain/profile/answer.go:314@52ce86908135`<br>`internal/domain/profile/answer.go:315@52ce86908135`<br>`internal/domain/rating/trial.go:17@52ce86908135`<br>`SPEC.md:251@52ce86908135` | A: §5 learner model, §6 evaluation; E-A3 |
| C082 | Mastery is held at a level: a topic counts as mastered only while its recommended point stays at the level of the task that completed the run, or below; once the child's tasks in it come from the level above, it is back in the rotation. | built | `internal/domain/tutor/tutor.go:309@52ce86908135`<br>`internal/domain/tutor/tutor.go:315@52ce86908135`<br>`internal/domain/profile/answer.go:353@52ce86908135` | A, extended version only: §5 learner model |

### Task selection

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C018 | A deterministic rule suggests the brief — goal, topic, grade level, difficulty, setting, traps — from the profile alone. | built | `internal/domain/tutor/tutor.go:179@52ce86908135`<br>`internal/domain/tutor/tutor.go:9@52ce86908135` | A: §3 system |
| C019 | Outside the trial series, after a failure the goal is to reinforce the topic of the last answer while it stays within reach; otherwise it is a new topic within reach and not mastered, one never given first, then the one given longest ago, ties going to the topic first in the catalog's order. In the trial series every task takes a new topic. | built | `internal/domain/tutor/tutor.go:263@52ce86908135`<br>`internal/domain/tutor/tutor.go:254@52ce86908135`<br>`internal/domain/tutor/tutor.go:280@52ce86908135`<br>`internal/domain/tutor/tutor.go:324@52ce86908135`<br>`internal/domain/tutor/tutor.go:335@52ce86908135`<br>`internal/domain/tutor/tutor.go:219@52ce86908135` | A: §5 learner model |
| C020 | The brief asks for two traps: the child's most frequent in the topic, topped up from the most frequent of the topic's reference tasks at the brief's level, then at the levels nearest it. | built | `internal/domain/tutor/traps.go:22@52ce86908135`<br>`internal/domain/tutor/traps.go:78@52ce86908135`<br>`internal/domain/tutor/tutor.go:36@52ce86908135` | A: §3 system |
| C070 | A mastered topic leaves the rotation; once every topic within reach is mastered, all of them come back into it. | built | `internal/domain/tutor/tutor.go:273@52ce86908135`<br>`internal/domain/tutor/tutor.go:277@52ce86908135` | A: §5 learner model, §7 limitations |
| C073 | The package the model writes from holds the brief, the corridor, the topic, every trap, the skills the task may not use, three reference tasks, the topic's solver templates and drawing frames, the limits the task is held to, the guide to writing it and the version of the instructions — and no pseudonym. | built | `content/package.go:58@52ce86908135`<br>`content/package.go:59@52ce86908135`<br>`content/package.go:63@52ce86908135` | A: §3 system |
| C081 | A topic is within reach when its easiest point, difficulty 1 of the lowest level it is taught at, has P ≥ 0.70; the rule sets only topics within reach, and a child below all of them gets the topics taught from grades 1–2. | built | `internal/domain/rating/ladder.go:129@52ce86908135`<br>`internal/domain/tutor/tutor.go:216@52ce86908135`<br>`internal/domain/tutor/tutor.go:225@52ce86908135` | A: §5 learner model |
| C083 | The model may ask for another topic, level or difficulty than the rule's, with a reason of at most 300 characters; the brief then keeps the rule's goal, the model's choice and both reasons, and the request records that the model chose. | built | `internal/domain/tutor/tutor.go:59@52ce86908135`<br>`internal/domain/tutor/tutor.go:372@52ce86908135`<br>`SPEC.md:441@52ce86908135` | A: §3 system, §5 learner model, §7 limitations |
| C097 | The rule that sets the next task never reads the parent's notes, which reach the model as tone and are not allowed to change what the child is set. | built | `internal/domain/tutor/tutor.go:16@52ce86908135`<br>`internal/domain/tutor/tutor.go:17@52ce86908135`<br>`internal/domain/tutor/property_test.go:263@52ce86908135` | A, extended version only: §7 limitations; S34 |

### Verification

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C021 | A submitted task is refused with one of nine codes: bad_structure, distractor_explanations, drawing_format, drawing_mismatch, readability, solver_error, solver_disagrees, self_check_blocking, near_duplicate. | built | `internal/domain/checks/checks.go:25@52ce86908135`<br>`internal/domain/checks/checks.go:53@52ce86908135`<br>`check_codes = 9` | A: §4 verification |
| C022 | The review runs in two halves: Examine decodes the task and runs the solver before the profile is read; Judge runs every check that has its input, and notes a check whose input is missing as not run rather than passed. A task is accepted when the review finds no problem; every input a check can lack — a part that can be read, the question, the structure of a drawing — is one the structure check requires, so a task with a check not run is refused in any case. | built | `internal/domain/checks/review.go:121@52ce86908135`<br>`internal/domain/checks/review.go:141@52ce86908135`<br>`internal/domain/checks/review.go:138@52ce86908135`<br>`internal/domain/checks/outcome.go:33@52ce86908135`<br>`internal/domain/checks/decode.go:40@52ce86908135`<br>`internal/domain/checks/structure.go:158@52ce86908135`<br>`internal/domain/checks/structure.go:282@52ce86908135` | A: §4 verification |
| C023 | The model's Starlark solver runs twice, the second time with the option labels rotated by two, and must pick the same option both times. | built | `internal/domain/solver/options.go:59@52ce86908135`<br>`internal/domain/solver/verdict.go:63@52ce86908135`<br>`internal/domain/solver/options.go:50@52ce86908135` | A: §3 system, §4 verification |
| C024 | A solver run is limited to 25,000,000 steps and 2 seconds. An instance runs one solver per vCPU, and a run that waits 3 seconds for its slot does not start: the task is told it was not checked, and no attempt is spent. | built | `internal/config/config.go:56@52ce86908135`<br>`internal/config/config.go:57@52ce86908135`<br>`internal/config/config.go:58@52ce86908135`<br>`internal/config/config.go:59@52ce86908135`<br>`infra/terraform/service.tf:114@52ce86908135`<br>`infra/terraform/service.tf:115@52ce86908135` | A: §4 verification |
| C025 | Under the sandbox's price of bytes, the costliest of the 603 reference solvers takes 2,413,617 steps, under a tenth of the ceiling. When the limits were first set, over the 450 solvers of grades 1–4 and before that price, the costliest took 2,081,362 steps and the slowest ran in about 35 ms; no timing of all 603 is recorded. | measured | `SPEC.md:892@52ce86908135`<br>`SPEC.md:1660@52ce86908135` | A, extended version only: §4 verification |
| C026 | The verification checks the model's formalisation, not the meaning of its text: the solver is written by the same model that wrote the task. | specified | `SPEC.md:771@52ce86908135` | A: §4 verification |
| C027 | A question is a near-duplicate when the Jaccard index of its character trigrams with another question's is 0.7 or more. The trigrams are built from padded words exactly as PostgreSQL's pg_trgm builds them, so the measure is the prototype's; scripts written without spaces use character bigrams instead. | built | `internal/domain/checks/duplicate.go:12@52ce86908135`<br>`SPEC.md:699@52ce86908135`<br>`SPEC.md:703@52ce86908135` | A: §3 system, §4 verification |
| C028 | Near-duplicates are sought among the reference questions of the task's level and the child's own past tasks. | built | `internal/domain/checks/review.go:230@52ce86908135`<br>`internal/domain/checks/review.go:231@52ce86908135` | A: §3 system, §4 verification, §7 limitations; E-A1 (leave-one-out) |
| C029 | Past tasks are remembered only as MinHash sketches of 192 positions of four bits, 96 bytes each, at most 200 of them, with no text. | built | `internal/domain/checks/fingerprint.go:20@52ce86908135`<br>`internal/domain/checks/fingerprint.go:28@52ce86908135`<br>`internal/domain/profile/validate.go:49@52ce86908135` | A: §4 verification, §7 limitations |
| C030 | Readability is held to the task's level, not the child's grade: the longest sentence may have at most 20, 25 or 30 words — 40, 50 or 60 characters in scripts written without spaces — for levels 1–2, 3–4 and 5–6, and in English the Flesch–Kincaid grade may exceed the youngest grade of the level by at most 3. | built | `internal/domain/checks/readability.go:16@52ce86908135`<br>`internal/domain/checks/readability.go:17@52ce86908135`<br>`internal/domain/checks/readability.go:18@52ce86908135`<br>`internal/domain/checks/readability.go:25@52ce86908135`<br>`internal/domain/checks/readability.go:86@52ce86908135` | A: §4 verification |
| C031 | A text drawing may be at most 30 cells wide and 12 lines high. | built | `internal/domain/checks/drawing.go:31@52ce86908135` | A: §4 verification |
| C032 | The model gets at most three attempts per request: after the third refusal the request closes and the child is handed nothing. Requests closed in failure count toward a daily ceiling of their own. | built | `internal/domain/profile/validate.go:52@52ce86908135`<br>`internal/domain/profile/request.go:56@52ce86908135`<br>`internal/transport/mcp/task_test.go:433@52ce86908135`<br>`internal/transport/mcp/limits.go:21@52ce86908135` | A: §3 system |
| C071 | A refusal names every problem found at once, so the model can fix them all in one more attempt, and never quotes an option's letter or text: it reaches the child's widget with everything else the tool returns. | built | `internal/domain/checks/checks.go:8@52ce86908135`<br>`internal/domain/checks/checks.go:9@52ce86908135`<br>`internal/domain/checks/checks.go:57@52ce86908135` | A: §4 verification |
| C072 | A submission holds the brief, the task, the model's self-check and the source of its solver. The task gives the question, five options, the key, a hint, a step-by-step solution, and for each of the four wrong options the trap it comes from and what the child reads after choosing it. | built | `internal/domain/checks/review.go:33@52ce86908135`<br>`internal/domain/checks/review.go:32@52ce86908135`<br>`content/schemas/task.json:67@52ce86908135`<br>`content/schemas/task.json:91@52ce86908135` | A: §3 system |
| C061 | Every task the model writes carries a hint: one leading question or the first step. | built | `content/schemas/task.json:13@52ce86908135` | A: §3 system; S04, Q62 |
| C074 | The hint arrives with the task and the card keeps it folded until the child opens it; whether it was opened travels with the answer. The checks require only that a hint is present: that it must not give the answer away is a rule of the instructions. | built | `internal/transport/mcp/submittask.go:119@52ce86908135`<br>`SPEC.md:627@52ce86908135`<br>`content/instructions/task_writing.md:12@52ce86908135` | A, extended version only: §3 system |
| C075 | Memory the language takes by itself is not priced: an operator such as xs = xs + xs, repeated, reaches about 480 MiB in under 300 steps, and a loop of empty dictionary literals killed an instance of 1 GiB for its memory. A solver written to exhaust memory can take its instance down. | measured | `SPEC.md:896@52ce86908135`<br>`docs/load.md:109@52ce86908135` | A, extended version only: §4 verification, §7 limitations; Q32 |
| C076 | Two limits are reasoned, not measured: the sentence limits in characters for scripts written without spaces, set at twice the word limits, and the near-duplicate threshold of 0.7 on character bigrams. | specified | `SPEC.md:685@52ce86908135`<br>`SPEC.md:701@52ce86908135` | A, extended version only: §4 verification, §7 limitations; Q33, Q34 |
| C079 | Everything the sandbox predeclares pays a step for every sixteen bytes it builds, charged before it builds anything, so at the step ceiling the built-ins of one run keep at most 381 MiB. | built | `SPEC.md:848@52ce86908135`<br>`docs/load.md:30@52ce86908135` | A: §4 verification |
| C095 | The self-check asks the model to read its task as a strict critic before handing it in: is a condition missing, can the question be read two ways, is a negation easy to miss, is a word or a range vague, do conditions contradict each other, does the right option answer another question. It records each problem with one of seven types and a severity, blocking or minor, and gives the answer it reaches solving the task afresh. | built | `content/instructions/task_writing.md:84@52ce86908135`<br>`content/instructions/task_writing.md:86@52ce86908135`<br>`internal/domain/checks/draft.go:94@52ce86908135` | A: §4 verification |
| C102 | The reference tasks are in English whatever the language of the chat, because what they carry is the structure and not the wording; since the near-duplicate check compares character trigrams (C027, C028), a reference task translated into another language is not caught as a copy. | built | `content/example.go:19@52ce86908135`<br>`content/example.go:20@52ce86908135` | A: §7 limitations |

### Answer sealing and privacy

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C033 | The answer, the explanations of the wrong options, the solution and the solver are sealed with XChaCha20-Poly1305 under an HKDF subkey of their own purpose, bound to this child and this task — and, once answered, to the letter given — until the child answers. | built | `internal/infra/seal/key.go:87@52ce86908135`<br>`internal/infra/seal/key.go:83@52ce86908135`<br>`internal/domain/profile/sealed.go:49@52ce86908135`<br>`internal/domain/profile/sealed.go:129@52ce86908135` | A: §3 system, §4 verification |
| C034 | The history window holds the last 20 entries, answers and skipped tasks together, and never lets go of one of the 5 latest answers. | built | `internal/domain/profile/validate.go:42@52ce86908135`<br>`internal/domain/profile/validate.go:47@52ce86908135` | not in paper A |
| C035 | The service writes one log line for each event of a lesson — a task requested, skipped, submitted or accepted, an answer recorded, a solver run — each with the version of the instructions the task was written to, and builds it from fields named one by one, never from a profile or a task. | built | `internal/transport/mcp/events.go:15@52ce86908135`<br>`internal/transport/mcp/events.go:19@52ce86908135`<br>`internal/transport/mcp/events.go:27@52ce86908135` | not in paper A |
| C036 | Task text, options, answer letters and the pseudonym stay out of logs, traces and metrics; tests check that nothing of the answer, the child or what the parent typed reaches a span, a log line or a label. | built | `SPEC.md:1558@52ce86908135`<br>`internal/transport/mcp/answer_test.go:641@52ce86908135`<br>`internal/transport/mcp/endpoint_test.go:410@52ce86908135`<br>`internal/transport/mcp/profile_test.go:661@52ce86908135` | A: §4 verification, §8 discussion |
| C037 | Telemetry replaces attributes the service may not keep with a fixed 'redacted' value. | built | `internal/telemetry/redact.go:13@52ce86908135` | A, extended version only: §8 discussion |
| C069 | The child is known to the service by a pseudonym alone: the profile holds no name, birth date or school. | built | `internal/domain/profile/student.go:12@52ce86908135`<br>`internal/domain/profile/student.go:32@52ce86908135` | A: §8 discussion |
| C064 | The card's tool results carry the task's question, drawing, options and hint and nothing that gives the answer away; the answer, the explanations and the solution come in a result only once the child has answered. The model that wrote the task knows the answer, and an adult who reads the host's own tool-call log sees the submitted task with it: that is outside the threat model. | built | `internal/transport/mcp/submittask.go:120@52ce86908135`<br>`internal/transport/mcp/task_test.go:775@52ce86908135`<br>`docs/architecture/04-profile.md:198@52ce86908135` | A: §3 system, §4 verification, §7 limitations; Q36 |
| C091 | The privacy policy states that MathTrail shows no advertising, builds no advertising profile, sells nothing and uses no child's data to train any model. | specified | `site/content/en/privacy.md:17@52ce86908135`<br>`site/content/en/privacy.md:91@52ce86908135` | A: §8 discussion |
| C092 | The package carries the child's grade and interests and the parent's free-form notes, up to 500 characters: all three reach the chat's model with every task. The privacy policy names the notes. | built | `content/package.go:200@52ce86908135`<br>`internal/domain/profile/validate.go:35@52ce86908135`<br>`site/content/en/privacy.md:77@52ce86908135` | A: §3 system, §8 discussion |
| C093 | The profile holds a permanent random identifier of the child; every log line of a signed-in call carries user, a 16-character HMAC-SHA256 of the parent's Google subject under a key of the service's, never the subject itself; and Cloud Run's own request logs keep the address each request came from. | built | `internal/domain/profile/profile.go:64@52ce86908135`<br>`internal/transport/mcp/boundary.go:289@52ce86908135`<br>`SPEC.md:1372@52ce86908135`<br>`site/content/en/privacy.md:59@52ce86908135` | A: §8 discussion |
| C096 | The task's card is drawn by submit_task, whose arguments hold the task with its key and its solution. | built | `internal/transport/mcp/submittask.go:154@52ce86908135`<br>`internal/transport/mcp/submittask.go:38@52ce86908135`<br>`content/schemas/task.json:52@52ce86908135` | A: §4 verification, §7 limitations |
| C098 | The service's terms make an adult — a parent or a tutor old enough for the chat platform — its user: the child does not sign in and has no account, and the adult runs the lesson. | specified | `site/content/en/terms.md:20@52ce86908135`<br>`site/content/en/terms.md:22@52ce86908135` | A: §8 discussion; S34 |
| C100 | The task's card has a field, "Ask a question about the task", that sends the child's own words to the chat as a message. | built | `web/locales/en.json:15@52ce86908135`<br>`web/src/widget/TaskCard.tsx:95@52ce86908135` | A: §8 discussion; S34 |

### Tools, sign-in and storage

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C038 | The service offers seven MCP tools: get_profile, save_profile, get_progress, read_progress (called by the card alone), next_task, submit_task and submit_answer. | built | `internal/transport/mcp/profile.go:95@52ce86908135`<br>`internal/transport/mcp/profile.go:110@52ce86908135`<br>`internal/transport/mcp/progress.go:79@52ce86908135`<br>`internal/transport/mcp/progress.go:95@52ce86908135`<br>`internal/transport/mcp/nexttask.go:79@52ce86908135`<br>`internal/transport/mcp/submittask.go:146@52ce86908135`<br>`internal/transport/mcp/submitanswer.go:85@52ce86908135` | A: §3 system |
| C039 | The child's answer is recorded once, by submit_answer, which the card calls itself or the model calls when the child answers in the chat, card or none; the same answer sent again is told what was recorded. | built | `internal/transport/mcp/submitanswer.go:85@52ce86908135`<br>`internal/transport/mcp/answer_test.go:313@52ce86908135`<br>`content/instructions/mcp_instructions.md:11@52ce86908135` | A: §3 system, §8 discussion |
| C040 | The service is its own OAuth 2.1 authorization server: a host signs the parent in through Google, which is asked for the parent's identity and the drive.file scope alone. It keeps nothing between requests: a sign-in under way travels sealed with the parent, and a finished one in the host's tokens. | built | `CLAUDE.md:30@52ce86908135`<br>`internal/transport/oauth/server.go:1@52ce86908135`<br>`internal/transport/oauth/server.go:4@52ce86908135`<br>`internal/infra/googleauth/googleauth.go:31@52ce86908135` | A: §3 system |
| C041 | Requests are held to paces per account, per address and per instance, counted in each instance's own memory, and next_task opens no request once the day's ceiling of accepted tasks, or of requests closed in failure, is reached. | built | `internal/ratelimit/ratelimit.go:1@52ce86908135`<br>`internal/transport/mcp/limits.go:20@52ce86908135`<br>`internal/transport/mcp/limits_test.go:224@52ce86908135`<br>`internal/transport/mcp/limits_test.go:107@52ce86908135` | A: §3 system |
| C078 | The child's profile is one JSON file in the parent's own Google Drive, reached with the parent's token and none of the service's. Drive has no conditional write, so a save reads the file again before it writes and refuses when it has changed. | built | `internal/store/drive/drivestore.go:1@52ce86908135`<br>`internal/store/drive/drivestore.go:13@52ce86908135` | A: §3 system, §8 discussion |
| C087 | The cards speak 22 languages, a dictionary each; English and Russian were written by people, and the other twenty are machine translations nobody who speaks them has read yet. | built | `widget_languages = 22`<br>`SPEC.md:1243@52ce86908135` | A: §7 limitations |

### Not built yet

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C042 | Usage evidence with a monthly counting pseudonym and daily aggregates is an open task (T67.1). | planned | `RUN.md:166@52ce86908135` | no paper: paper B, which used it, left the plan (Q60) |
| C090 | Acceptance in Claude and in ChatGPT, and the load, logs and message limits measured on the deployed service, are open product tasks (T62–T64). | planned | `RUN.md:157@52ce86908135`<br>`RUN.md:158@52ce86908135`<br>`RUN.md:159@52ce86908135` | A, extended version only: §3 system |

### Content

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C043 | The topic catalog holds 17 topics, the trap catalog 20 traps, the skill catalog 25 skills. | built | `topics = 17`<br>`traps = 20`<br>`skills = 25` | A: §3 system |
| C044 | The reference tasks at the pinned commit, by grade level, each with its own solver. | built | `reference_tasks = 603`<br>`reference_tasks_1_2 = 200`<br>`reference_tasks_3_4 = 250`<br>`reference_tasks_5_6 = 153`<br>`reference_solvers = 603`<br>`reference_tasks_without_solver = 0` | A: §1 introduction, §3 system |
| C077 | Every reference task names, behind each of its four wrong options, a trap of the catalog. | built | `reference_tasks_with_a_trap_per_wrong_option = 603`<br>`reference_tasks = 603` | A: §1 introduction, §3 system |
| C045 | The model is given 33 solver templates and 11 drawing frames to start from. | built | `solver_templates = 33`<br>`drawing_frames = 11` | A, extended version only: §3 system |
| C046 | The instructions for the model: the task-writing guide and the MCP instructions, with their word counts. | built | `words_task_writing = 1631`<br>`words_mcp_instructions = 1271` | A, extended version only: §3 system |
| C066 | The reference tasks of grades 5–6 were written by Claude in ten batches of the product's own plan, each batch checked by the author; those of grades 1–4 are the prototype's 450 (P012, P013). | built | `RUN.md:105@52ce86908135`<br>`RUN.md:106@52ce86908135`<br>`RUN.md:107@52ce86908135`<br>`RUN.md:108@52ce86908135`<br>`RUN.md:110@52ce86908135`<br>`RUN.md:111@52ce86908135`<br>`RUN.md:112@52ce86908135`<br>`RUN.md:113@52ce86908135`<br>`RUN.md:114@52ce86908135`<br>`RUN.md:115@52ce86908135`<br>`reference_tasks_5_6 = 153` | A: §3 system |

### Quality

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C047 | The product's tests: test, fuzz and benchmark functions, and gopter properties, at the pinned commit. | built | `test_functions = 1099`<br>`fuzz_functions = 32`<br>`benchmark_functions = 3`<br>`gopter_properties = 99`<br>`test_files = 201` | A, extended version only: §3 system |
| C048 | Statement coverage of the product's tests at the pinned commit. | built | `coverage_statements = 82.1%` | A, extended version only: §3 system |

### Measurements already in the repository

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C049 | On the reference tasks, the whole readability check passes 76 % of the grades 1–2 tasks and 91 % of the grades 3–4 tasks at the limits of their own levels. | measured | `SPEC.md:693@52ce86908135` | A, extended version only: §4 verification |
| C050 | Over every pair of reference questions of one level — 62,653 pairs among all 603 — the sketches miss the exact trigram Jaccard index by 0.031 on average and 0.139 at worst, and decide 33 pairs differently from it at the 0.7 threshold, all within 0.1 of it. | measured | `SPEC.md:712@52ce86908135`<br>`SPEC.md:712@52ce86908135` | A, extended version only: §4 verification |
| C051 | /health took a median 196 ms through the domain (minimum 150, maximum 218, 10 requests) after the first deploy. | measured | `docs/live/03-first-deploy.md:53@52ce86908135` | not in paper A |
| C052 | The runtime image took 6.8 MB in Artifact Registry. | measured | `docs/live/03-first-deploy.md:60@52ce86908135` | A, extended version only: §6 evaluation |
| C053 | In one traced request, a request for a path the service does not serve, answered with Not Found, the service's own span took 274.734 µs of the 204.011 ms Cloud Run measured for the whole request. | measured | `docs/live/03a-telemetry.md:35@52ce86908135`<br>`docs/live/03a-telemetry.md:36@52ce86908135`<br>`docs/live/03a-telemetry.md:49@52ce86908135` | no paper: the trace is of a request answered with Not Found, not of a task, and Section 6.3 (6.4 of the extended version) measures reviews instead (S39, Q76) |
| C054 | With the solver templates, the drawing frames and the reference tasks' drawings in, a child with an ordinary profile gets a package of 19.7 KB on average across the catalog and 25.2 KB at most. | measured | `SPEC.md:484@52ce86908135` | no paper: Section 6.3 (6.4 of the extended version) measures the packages at every point and turn (S39, Q76) |
| C084 | In a first local run of an earlier build (2026-09-27), through Claude Code with claude-sonnet-5 in text mode, the model wrote 15 tasks over 17 hand-ins: 13 were accepted at the first attempt and 2 at the second, one after solver_disagrees and one after readability; from the request to acceptance took a median of 30 s and at most 57 s. Solved by hand, each of the 15 had the keyed answer as its intended one; one could also be read another way (C089). | measured | `docs/live/04-local-run.md:10@52ce86908135`<br>`docs/live/04-local-run.md:121@52ce86908135`<br>`docs/live/04-local-run.md:122@52ce86908135`<br>`docs/live/04-local-run.md:3@52ce86908135` | A: §3 system |
| C089 | In that run one accepted task became ambiguous when the model shortened its question to pass the readability check and dropped the word "different", on which the answer rested; no check saw it. | measured | `docs/live/04-local-run.md:187@52ce86908135` | A: §4 verification |
| C085 | On the deployed service, in Claude on the web (2026-09-30), a child's whole trial series of five tasks went through: each task was accepted at its first attempt, 16 to 31 s after next_task. The build that ran, 51b78cc, was a branch commit; the drawing checks, the drawing frames and the guide for the model changed after it and before the pinned commit. | measured | `docs/live/05-login-drive.md:127@52ce86908135`<br>`docs/live/05-login-drive.md:126@52ce86908135` | A: §3 system |
| C086 | Measured on the development machine, in a container of one instance's size with the profile kept in memory, a lesson's task is billed 0.36 vCPU-seconds and 0.36 GiB-seconds, which would let the platform's free tier cover about 500,000 tasks a month, and an instance starts in a median of 254 ms. Drive's round trips and the extra calls a host makes are not in these numbers; the deployed service is measured in T64. | measured | `docs/load.md:43@52ce86908135`<br>`docs/load.md:46@52ce86908135`<br>`docs/load.md:127@52ce86908135`<br>`docs/load.md:42@52ce86908135`<br>`docs/load.md:48@52ce86908135` | A, extended version only: §6 evaluation |
| C088 | The drawing limits were calibrated on the task card in Chromium with three monospaced fonts: 30 cells and 12 lines stay, now measured, and the allowed characters were narrowed to 65. | measured | `docs/live/06-drawings.md:3@52ce86908135` | not in paper A |
| C101 | Measured on the development machine, an instance lets in about three tasks handed in a second, the pace its limit of 200 messages a minute allows, and its one solver slot kept up with them: of a costly solver handed in at that pace for thirty seconds, all 90 hand-ins were examined, with a median of 134 ms. | measured | `docs/load.md:76@52ce86908135` | A, extended version only: §6 evaluation |
| C103 | In that run the one refusal by the solver check was a false alarm: the solver returned «Mail» where the option read «Mail car», so it matched no option, and the model added « car» to its solver. Two accepted tasks had hints that gave too much away, one naming the whole method and one nearly giving the answer; no check looks at what a hint gives away. | measured | `docs/live/04-local-run.md:124@52ce86908135`<br>`docs/live/04-local-run.md:176@52ce86908135`<br>`docs/live/04-local-run.md:177@52ce86908135` | A: §4 verification |

### Continuity with the prototype

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C059 | The prototype's own code exported golden vectors into the product: the readability of all 450 of its reference questions, pg_trgm similarities, the rating formulas' worked example with the replayed seed profiles, and the rule's briefs for the five seed profiles; no Go test may rewrite them. | built | `testdata/golden/export/README.md:3@52ce86908135`<br>`testdata/golden/export/README.md:3@52ce86908135` | A, extended version only: §3 system |
| C060 | The product's Go implementations of the ratings, the readability check, the near-duplicate measure and the rule are tested against those vectors. | built | `internal/domain/rating/golden_test.go:99@52ce86908135`<br>`internal/domain/checks/readability_golden_test.go:36@52ce86908135`<br>`internal/domain/checks/similarity_test.go:162@52ce86908135`<br>`internal/domain/tutor/golden_test.go:116@52ce86908135` | A, extended version only: §3 system |

### Remarks for the author

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| C056 | SPEC says the pace changes what is chosen next, yet the rule's code names neither the pace nor "I don't know"; SPEC's own remark 43 records the gap. "I don't know" reaches the rule only as a wrong answer, through the run of failures that makes it reinforce the topic. | remark | `SPEC.md:242@52ce86908135`<br>`SPEC.md:1726@52ce86908135`<br>`pace_or_confused_lines_in_rule = 0`<br>`internal/domain/profile/answer.go:234@52ce86908135`<br>`internal/domain/profile/answer.go:298@52ce86908135` | the author; it tells A's learner model and rule what not to claim |
| C057 | docs/usage-evidence-for-grants.md counts 513 reference tasks; the pinned commit has 603. | remark | `docs/usage-evidence-for-grants.md:195@52ce86908135`<br>`reference_tasks = 603` | the author; no paper |
| C094 | The privacy policy says the package brings the notes and that the rest of the profile stays out, but the package also carries the child's grade and interests; and it says the logs hold counts, not content, without naming the derived identifier of the parent's account that every signed-in log line carries. | remark | `site/content/en/privacy.md:79@52ce86908135`<br>`content/package.go:200@52ce86908135`<br>`site/content/en/privacy.md:55@52ce86908135`<br>`internal/transport/mcp/boundary.go:289@52ce86908135` | the author; no paper |
| C099 | The product's rule is that the answer must not appear in widget payloads. MCP Apps has a host hand a card the complete arguments of the call that drew it, and the task's card is drawn by submit_task, whose arguments hold the key (C096); so the key reaches the card's memory before the child answers, although the card never shows it. | remark | `CLAUDE.md:91@52ce86908135`<br>`internal/transport/mcp/submittask.go:154@52ce86908135` | the author decided (K10, 2026-10-01): a known limitation, like the host's tool log; the paper says the key is never shown on the card [C096]; no paper as a fact |
<!-- ledger:product:end -->

### What the product does not contain

Paper A describes none of these as part of MathTrail. Each is contradicted by a claim above.
- No database (C002).
- No LLM calls of its own, so no agents, no fine-tuning and no model hosting (C003).
- No Python in the product: one Go module (C001); Python and PostgreSQL were the prototype's (S03).
- No Glicko or rating deviation: a Rasch model with a guessing floor and Elo-style updates (C009, C011).
- No vector database: near-duplicates use trigram sets and MinHash sketches (C027, C029).
- No streaming pipeline or message queue: every request stands alone (C002, C062); the prototype had none either (P057).
- No suite of "120+ edge cases": the checks' tests are counted in C047, and S37 builds a real fault-injection suite.

### Context outside git

Disagreements with the code are claims with the status *remark* in the table above (group "Remarks for the author"). One more rests on a file outside git, so no tool here can check it:

- `PRODUCT-V1.md` still lists two of the prototype's catalog sizes in section 4.6 — 10 topics for grades 1–4 and 14 traps — against 17 and 20 at the pinned commit (C043); its constraint skills now read 25, as the catalog has. The copy read on 2026-09-30 is `research/archive/PRODUCT-V1.2026-09-30.md`, SHA-256 `70d53036575a3ff4e684b9381ac315a60bd55a5a8855451b671f169503349c0a`; the copy S02 read, which said 15 skills, stays beside it as `research/archive/PRODUCT-V1.md` (Q07).

### Retired claims

A claim that stops being true of the product leaves the table; its number is never given to another claim.

- **C055** — README described the widget, the answer button and the explanation of the mistake in the present tense while they were open tasks. Retired at `52ce86908135`: all three are built (C038, C039, C068).
- **C058** — a profile could exclude at most 15 skills while the catalog held 25. Retired at `52ce86908135`: the limit is 25, the size of the catalog, and a test of the content holds the two to one number.

## Use

The service's real use: families' lessons in the deployed service, and the totals it publishes of them. Real lessons began after the commit paper A describes, so these claims are proven at a later commit of their own, which [use-stats.txt](use-stats.txt) names (`just research use-pin <commit>`); nothing is counted there. Generated from [claims-use.json](claims-use.json) by `just research ledger`. The paper prints no number of real use from these claims: what it prints comes from the service's monthly snapshot of public totals, shown only behind ten children or more (U002).

<!-- ledger:use:begin -->
Pinned commit: `443b06faaa6b` of `github.com/MathTrail/mathtrail-standalone`, 2026-10-07. Every fact below was computed at that commit.

### Real use

| ID | Claim | Status | Evidence | Used in |
|---|---|---|---|---|
| U001 | Families have used the deployed service in real lessons since its first weeks in production: in Claude their tasks reached the card, where the child answers, and the first parents' reading of the progress changed how it names a child's rank. | measured | `docs/live/16-task-time.md:11@443b06faaa6b`<br>`docs/live/16-task-time.md:339@443b06faaa6b`<br>`docs/decisions.md:643@443b06faaa6b` | A: §7 limitations, §8 discussion |
| U002 | Since 4 October 2026 the service counts its use into totals it keeps for years, and shows a month's totals publicly, on the site's page "Research" and in applications for grants, only once ten children and thirty answers or more stand behind them. | built | `docs/grants/README.md:20@443b06faaa6b`<br>`infra/analytics/views/public/chances_total.sql:12@443b06faaa6b`<br>`site/content/en/privacy.md:95@443b06faaa6b` | A: §7 limitations |
<!-- ledger:use:end -->

## Prototype

In [prototype.md](prototype.md): the snapshot of the prototype's public repository, its claims and its remarks.
