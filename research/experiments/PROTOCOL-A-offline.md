# Protocol of paper A's offline experiments: E-A1, E-A3 and E-A4

This protocol fixes, before any data, what the offline experiments of paper A measure, on what, and how the results are read (rule G9 of [RUN.md](../RUN.md)). It covers three experiments: E-A1, defects injected into reference tasks (task S37, RQ1); E-A3, simulated learners on the service's rating code (S38, RQ2); and E-A4, the cost of the checks (S39). E-A2, tasks written by models of several families, has a protocol of its own (S40) and waits (Q57).

## 0. Ground rules

- **Freezing.** This file is frozen once its SHA-256 is recorded in the run journal of RUN.md and stamped with OpenTimestamps; the proof lies beside it as `PROTOCOL-A-offline.md.ots`. The file is never edited after that. A change of plan is a deviation: the experiment's report lists each one under "Deviations from the protocol", with its reason and what it changes in the results.
- **No data before freezing.** Nothing named in this protocol has been run. The only runs allowed before an experiment's full run are its harness's own tests (sections 1.10, 2.9, 3.4), which check the code and are not results.
- **The code under test** is the product at commit `52ce86908135`, the ledger's pin. Its `internal/domain/`, `internal/infra/starlark/`, `internal/infra/seal/` and `content/` are byte for byte the same at `775878f901fd`, the head of the tree when this protocol was written. The research module imports the product through `replace`, so an experiment runs on the working tree. Each report records the commit it ran on and whether any of those four directories differs from the pin; a difference is a deviation.
- **The product is not changed** (G6). The harnesses call the product's exported functions. What the product would need to change goes to the author as a proposal (section 4).
- **No child, no language model, no network.** E-A1 and E-A3 use reference tasks and synthetic learners only (G5). E-A4's cloud numbers are read from the reports of product tasks, not measured here.
- **Numbers** reach the paper only through data files and generated macros (G2). Every run saves its raw per-case or per-learner records.
- **Negative and null results are reported** (G9), the out-of-scope classes of E-A1 included.
- **Determinism.** All randomness comes from Go's `math/rand/v2` PCG generator. The master seed is 20261001. Each unit of work — a case of E-A1, a learner of E-A3 — gets its seeds as FNV-64a of the master seed, the experiment's name, the unit's identifier and the purpose of the stream, so any unit can be rerun alone. Section 2.2 says which streams a learner has.

## 1. E-A1: injected defects (RQ1)

### 1.1 Question

Which defects in a task the service's checks refuse, by which check, and which they let through. E-A1 does not measure false refusals — refusals of good tasks — because that needs good tasks the checks were never tuned on, which only E-A2 and its blind review (S44) provide. While E-A2 waits, the paper names the false-refusal rate as unmeasured.

### 1.2 The code under test

- The reviewer: `checks.NewReviewer(c, runner, checks.DefaultDrawingLimits())`, both halves, `Examine` then `Judge`, exactly as `submit_task` uses them [C021–C023].
- The content `c`: the product's embedded content, `content.Load()`, behind one wrapper that changes nothing but `ReferenceQuestions` (section 1.5).
- The sandbox: `starlark.New` with the service's defaults at the pin — 25,000,000 steps, 2 s of wall clock per run, one run at a time, 3 s of wait for a slot [C024]. The harness passes these four numbers itself, as written here, rather than reading them from `internal/config`; a later change of the service's defaults is reported and does not change the experiment.
- What a case records of the outcome: every problem's code, not only the first; the checks reported as not run, with their reason; the solver's runs, with their status, steps and duration.

Reviews run one at a time, so that no solver run shares a processor with another and the wall clock judges each alone.

### 1.3 Hosts: reference tasks made into submissions

The hosts are the service's reference tasks at the pin: 603, of which 200 are of grades 1–2, 250 of grades 3–4 and 153 of grades 5–6 [C044, C066]. Each comes with its solver. A reference task lacks some fields a submission must have; the harness fills them with the same fixed texts for every host:

| Field | Value |
|---|---|
| `task.core_idea` | "Reference task {id}." |
| `task.design_thought_process` | "Taken from the reference set without change." |
| `task.hint` | the task's own hint where it has one (the 153 tasks of grades 5–6); otherwise "Read the question again and check each option against it." |
| `brief.pedagogical_goal` | `new_topic` |
| `brief.target_concept`, `grade_level`, `difficulty` | the task's own |
| `brief.setting`, `brief.rationale` | "A reference task." |
| `brief.traps_to_use` | the distinct traps of the task's wrong options, in the order of their letters |
| `brief.excluded_skills`, `brief.constraints` | empty lists |
| `self_check.issues` | an empty list |
| `self_check.option_check` | "Checked against the solution." under each of the five letters |
| `self_check.final_answer` | the task's key |
| the request the task is held against | the same brief; language `en`; no fingerprints, except in class D15 |

The self-check of every case says the model found nothing wrong and agrees with the key, except in the classes that inject a fault into the self-check itself (D07, D08, D09). E-A1 therefore measures what the deterministic checks catch when the model's own self-check misses the defect. How often a model's self-check notices its own defect is a property of the model, which only E-A2 can measure.

**Eligible hosts.** A host is eligible when the reviewer accepts its completed submission with no problem and no check left unrun, under the comparison rule of section 1.5. Only eligible hosts receive defects, so that a refusal after an injection is the injection's doing. The sanity run of section 1.8 reports how many hosts are eligible at each level and why the others are not.

### 1.4 Defect classes

Each class comes from a source outside this experiment and is marked before any run: **must catch**, with the check expected to refuse it, or **out of scope**, with the reason the service cannot see it. The sources:
- the prototype's research note on verification, `research/05-verification.md` at commit `02638353482e` of its public repository, written in Russian: the defects of generated problems it lists — a missing condition, a condition that can be read two ways, conditions that contradict each other, a wrong answer — its demand that exactly one of the five options be right, and its warning that code checks the formalisation, not the text;
- the product's SPEC, section 5, which states what each of the nine checks refuses;
- the literature read for S31 (`literature/generation-verification.md`): MATHWELL's criteria of solvability, accuracy and appropriateness; MathQ-Verify's incomplete and conflicting conditions; UMWP's missing and ambiguous key information; the finding that a program checks a formalisation that can share the misreading of its author;
- SPEC 5.2 for the pseudonym, which no check looks for by design, and the working definition of an olympiad-style task (Q15), which the brief asks for and no check tests [C021].

The operators of the must-catch classes (section 1.6) fall into two groups, and the paper keeps them apart.
- **Mechanism operators** break a rule a check states in its own terms: a sentence longer than the level allows, a field missing, a solver that fails, a key the reference solver contradicts. A check that enforces its rule refuses every such case, so these test that the pipeline holds its rules end to end; a miss is a defect of the implementation, and goes to the author as one. Every must-catch operator is one, except the three below.
- **Discovery operators** make a defect a child would see whose refusal does not follow from any check's definition: D02b, a second correct answer written in words; D14b, a copy of a reference task with new numbers; D15b, a repeat of the child's own task with new numbers.

D01 and X03 are read together. A wrong key is refused when the solver works the task out correctly, which the reference solvers do, so D01 is a mechanism class; a wrong key whose solver shares the misreading behind it is X03, outside what the service can see. The paper states both halves.

| Class | Defect | Source | Expected |
|---|---|---|---|
| D01 | Wrong key, consistent: the key names a wrong option, and the explanations and the self-check agree with it | research/05; MATHWELL accuracy | must catch: `solver_disagrees` |
| D02 | Two correct options | research/05; MATHWELL solvability; SPEC 5.2 | must catch: `bad_structure` or `solver_disagrees` |
| D03 | No correct option | research/05; SPEC 5.7 | must catch: `solver_disagrees` |
| D04 | The solver fails: an error, no parse, no `solve`, an output that is not a list of letters, a bad `%` format | SPEC 5.7, 6.3 | must catch: `solver_error` |
| D05 | The solver exceeds its budget | SPEC 6.6 | must catch: `solver_error` |
| D06 | The solver writes out the letter instead of working out the answer | SPEC 5.7 | must catch: `solver_disagrees` |
| D07 | The self-check disagrees with the key, or calls the task unsolvable | SPEC 5.7 | must catch: `solver_disagrees` |
| D08 | The self-check reports a blocking issue | SPEC 5.8 | must catch: `self_check_blocking` |
| D09 | Broken structure: a required text empty, four options, a key that is no letter, the key's option explained, a text past its limit, a self-check entry missing or of an unknown type | SPEC 5.2 | must catch: `bad_structure` |
| D10 | The brief handed back does not match the request: another topic, level or difficulty, or a skill the profile excludes dropped | SPEC 5.2 | must catch: `bad_structure` |
| D11 | A trap outside the catalog | SPEC 5.2 | must catch: `bad_structure` |
| D12 | An explanation of a wrong option that repeats the solution or the hint, repeats another, is the catalog's description, is too short, or is empty | SPEC 5.3 | must catch: `distractor_explanations`; `bad_structure` when empty |
| D13 | A question too hard to read for its level: a sentence over the level's limit, or English over the Flesch–Kincaid limit | SPEC 5.5 | must catch: `readability` |
| D14 | A copy of a reference task of the level: verbatim, with new numbers, or with its sentences reordered | SPEC 5.6 | must catch: `near_duplicate` |
| D15 | A repeat of the child's own earlier task: verbatim, or with new numbers | SPEC 5.6 | must catch: `near_duplicate` |
| D16 | A drawing out of size: too tall, too wide, a long run of spaces, trailing spaces | SPEC 5.4 | must catch: `drawing_format` |
| D17 | A drawing with characters outside the allowed set | SPEC 5.4 | must catch: `drawing_format` |
| D18 | A drawing whose labels disagree with its structure or the wording | SPEC 5.4 | must catch: `drawing_mismatch` |
| X01 | Ambiguous wording: an exact quantity made vague, or a number made a range | research/05; UMWP; MathQ-Verify | out of scope: the solver and the key keep the intended reading |
| X02 | A missing condition: a sentence with a number removed | research/05; UMWP | out of scope: the same |
| X03 | A misreading the key and the solver share: the question asks the opposite, the options, key and solver unchanged | research/05; [C026] | out of scope: the solver checks the formalisation, not the text |
| X04 | A trap that names the wrong mistake | SPEC 5.2 checks only that a trap exists | out of scope: no check reads what a trap means |
| X05 | A hint that gives the answer away | — | out of scope: no check reads the hint against the answer |
| X06 | The pseudonym in the task | SPEC 5.2 | out of scope by design |
| X07 | A task off its topic: a well-formed task of another topic under the brief's topic | Q15; EDUMATH's failures in fitting a standard (S31) | out of scope: no check tests the topic of a task |
| X08 | A task off its difficulty: a well-formed task of the same topic two or more difficulties away | the same | out of scope: no check tests a task's difficulty [C013, C021] |
| X09 | The same idea in a new setting: a copy with its names and objects changed, with or without new numbers | SPEC 5.6 (the threshold is set above a new setting); S31 on textual near-duplicates | out of scope: the check compares text |

The out-of-scope classes are injected to measure the limits paper A states in its Section 4.5, not to be counted as misses. A refusal in an out-of-scope class is reported with its code, as an incidental catch.

### 1.5 The comparison rule without the source

The near-duplicate check compares a question with the reference questions of its level (`internal/domain/checks/review.go:226–232` at the pin). A host is one of those references, so its own question would refuse it as a copy of itself. The wrapper over the content therefore leaves out of `ReferenceQuestions` every reference question whose text the case was built from:
- the host's own question, in every case;
- the donor's question as well in X07 and X08, where the donor's whole task replaces the host's and the class is about something else;
- but not the donor in D14 and X09, where the donor is what the case copies.

A question is left out when its text equals the source's exactly. The wrapper has a test of its own (section 1.10).

### 1.6 Operators

Each operator turns one eligible host into one case. An operator that needs a choice — a letter, a donor, a type — makes it with the case's seed. An operator that does not apply to a host — no drawing, no number, no word from its list — produces no case for it, and the report counts how many hosts each operator applied to. "The solver's body" is the block of `def solve(options):`; a line inserted into it takes the indentation of the body's first line.

| Operator | What it does |
|---|---|
| D01a | The key moves to a wrong option chosen by seed. The explanation that stood under that letter moves to the old key's letter, trap and text unchanged; `final_answer` follows the key. |
| D02a | A wrong option, chosen by seed, is replaced by the correct option's text written another way with the same key in the sense of `solver.Key`: a whole number gains ".0", a decimal a trailing zero, a text with a letter the case of its first letter swapped, and any other text a space added at its end. |
| D02b | Applies where the correct option is a whole number from 0 to 20 written in digits and nothing else. A wrong option, chosen by seed, is replaced by that number in English words ("six"). |
| D03a | The correct option's text is replaced: a number by the nearest of n + 1, n − 1, n + 2, n − 2, n + 3 whose key no other option has; any other text by a wrong option of another eligible task of the same topic whose key no option of the host has. The key is unchanged. |
| D04a–e | (a) `injected = 1 // 0` inserted as the first statement of the solver's body, an error only a run meets; (b) a line `solve(` appended to the source; (c) `def solve(` renamed `def solve_injected(`; (d) `return "A"` inserted as the body's first statement; (e) `injected = "%5d" % 1` inserted as the body's first statement. |
| D05a | `for injected in range(100000000):` with `pass` under it inserted as the body's first statement. Only the step budget is targeted; a wall-clock limit depends on the machine, and both end in the same code. |
| D06a | `return ["K"]` inserted as the body's first statement, K being the key. |
| D07a, b | (a) `final_answer` set to a wrong letter chosen by seed; (b) `final_answer` set to `UNSOLVABLE`. |
| D08a | One issue added to the self-check: a type chosen by seed among the seven the format has, severity `blocking`, comment "Injected.". |
| D09a–g | (a) one operator per required text — `task.question`, `task.hint`, `task.solution`, `task.core_idea`, `task.design_thought_process`, `brief.setting`, `brief.rationale` — that text set to ""; (b) the option of a wrong letter chosen by seed removed, with its explanation; (c) the key set to "F"; (d) an explanation added under the key's letter, copied from a wrong option's; (e) the solution repeated until it is longer than 2,000 characters; (f) one entry of `option_check` removed; (g) an issue of type "injected", severity `minor`, added to the self-check. |
| D10a–d | (a) `brief.target_concept` set to another topic of the catalog chosen by seed; (b) `brief.grade_level` set to another level; (c) `brief.difficulty` set to another difficulty chosen by seed; (d) the request's excluded skills set to one skill of the catalog chosen by seed, while the brief keeps an empty list. |
| D11a | The trap of one explanation, chosen by seed, set to "injected_trap". |
| D12a–f | One explanation chosen by seed is replaced by: (a) the first six words of the solution, or all of it when shorter; (b) the first six words of the hint, or all of it; (c) another explanation's text; (d) its trap's description from the catalog; (e) its own first two words; (f) "". |
| D13a | Applies where the question has three sentences or more. The first two sentences are joined with ", and" — the first one's final mark dropped, the second one's first letter lowered — and the next sentence joined on in the same way, never the last one, until the joined sentence holds at least three words more than the level allows, a word being a piece between spaces that holds a letter or a digit. Applies only where that happens. |
| D13b | The sentence "Consider every possibility systematically and comprehensively." inserted before the question's last sentence. A case counts only where the changed question's Flesch–Kincaid grade, as `checks.Readability` computes it on the question alone, is over its level's limit; the report gives the share of hosts where it is. |
| D14a–c | Applies to hosts without a drawing. A donor is chosen by seed among the other eligible hosts without a drawing, of the same level and topic, or of the same level when the topic has no other. The host's question is replaced by: (a) the donor's question; (b) the donor's question with every whole number n in it written n + 1; (c) the donor's question with its first two sentences swapped, where it has three or more. The host's options, key and solver stay. A host with a drawing is left out because the donor's wording would not name the host's labels, and the drawing check would refuse the case for that. |
| D15a, b | The request's fingerprints are set to one sketch, `checks.Fingerprint(q, "en")`, of: (a) the host's own question; (b) the host's question with every whole number n written n + 1. The task is the host's, unchanged. |
| D16a–d | Applies to hosts with a drawing. (a) lines of "-" added at the end until the drawing has 13; (b) the longest line extended with "-" to 31 cells; (c) 21 spaces inserted after the first character of the first line; (d) two spaces added to the end of the first line. |
| D17a–d | Applies to hosts with a drawing. Inserted after the first character of the first line: (a) a zero-width space, U+200B; (b) a tab; (c) a no-break space, U+00A0; (d) ╭, U+256D, which is not in the allowed set. |
| D18a–c | Applies to hosts with a drawing. (a) the sentence "Point Q is marked." added before the question's last sentence, Q being declared nowhere; (b) an object with id "injected" and label "Z" added to the drawing's structure, drawn nowhere; (c) the label of one declared object, chosen by seed, changed to "Z" in the structure only. |
| X01a, b | (a) The first "exactly" in the question becomes "about"; where there is none, the first whole number n becomes "about n". (b) The first whole number n becomes "n or n + 1", written out ("3 or 4"). |
| X02a | Applies where a sentence other than the last holds a digit. The first such sentence is removed. |
| X03a | The first word of the question found in this list, in either direction and with its capital kept, is swapped: smallest–largest, least–most, fewest–most, fewer–more, minimum–maximum, first–last, before–after, left–right, cannot–can; "NOT" or "not" removed. |
| X04a, b | (a) The traps of two explanations with different traps swapped; (b) the trap of one explanation set to another trap of the catalog chosen by seed. |
| X05a | The hint set to "The answer is {the correct option's text}." |
| X06a | "Bright-Owl, " put before the question, its first letter lowered. |
| X07a | A donor of the same level and another topic chosen by seed: its question, options, key, solution, explanations, hint, drawing, drawing structure and solver replace the host's, and the self-check's `final_answer` and the brief's `traps_to_use` follow the donor's key and traps; the brief and the request keep the host's topic, level and difficulty. |
| X08a | The same with a donor of the same topic and level whose difficulty differs from the host's by two or more. |
| X09a, b | A host and a donor as in D14, neither with a drawing. The host's question is replaced by the donor's with (a) its names and objects changed by two fixed lists in the harness — every person's name in the first list to the next one in it, every noun of the second list to the next one in it; (b) the same, and every whole number n written n + 1. The lists are fixed in the harness's code before the full run, and the report prints them. |

### 1.7 Validity of a case

An operator can fail to make a defect. A case enters the denominators only when it is valid:
- **Automatically**, for every operator: the submission differs from the host's. For D01a and D03a, the host's own solver — trusted, as every reference solver was checked when the task was written — confirms the defect: run on the case's options, it does not arrive at the case's key (D01a), or finds no option (D03a). A D03a case whose replacement the host's solver accepts is a second correct text, not a missing one, and is set aside. D02b is valid by construction: the word names the correct number, and its cases are not read.
- **By reading**, for every out-of-scope class: ten cases of each operator, drawn by seed, are read by the executor, who records for each whether it is the defect it is named after. A case found not to be — a removed sentence that held nothing the answer needs, a swapped word that leaves the answer as it was — is an equivalent mutant. The report gives, per operator, the share of equivalent mutants among those read, and the paper reports the out-of-scope rates over all cases with that share beside them. The author checks the reading afterwards (K04).

### 1.8 The sanity run

Before any defect, the completed hosts are reviewed as they are. The run reports, per level, how many are accepted and, for the rest, which checks refused them. Many reference tasks of grades 1–2 are expected to fail the readability check at their own level, which SPEC 5.5 reports (76 % of them passed). This run is context for the calibration of the checks, not a false-refusal rate: the thresholds were set on these tasks.

### 1.9 Measures and analysis

The operator is the primary unit: each applies once to a host, so its cases are independent. A class pools its operators for Table 3.
- **Refused:** per operator, the share of valid cases refused by any check, and the share refused by the expected check, each with a 95 % Clopper–Pearson interval. Per class, the same shares pooled, with an interval from a bootstrap over hosts, 2,000 resamples, percentile, because several operators of one class share hosts.
- **Class × check:** for every class, the share of valid cases on which each of the nine codes fired — every code reported, not only the first. This is Table 3 of the paper.
- **First code:** the distribution of the first code in the order of SPEC 5.1, which is what the service logs.
- **Unrun checks:** the share of cases in which a check did not run for want of its input.
- **What only one check catches.** For every check X and class, the share of valid cases refused by X and by no other check: what would be lost if X were switched off. `Judge` runs every check on its own and reports them all, and the solver runs whatever the other checks find, so removing X's code from a case's outcome is exactly the outcome without X. No check is switched off in the product's code.
- **Similarity.** For D14, D15 and X09, the exact similarity of the case's question to its source (`checks.Similarity`) and, for D15, the sketch's estimate, so that the paper can show where the threshold of 0.7 falls among the variants.
- **Out of scope:** the share refused, with each refusal's code.

The unit of analysis is the case. Every number is reported, including classes with few cases; the drawing classes can only use the hosts of grades 5–6 that have a drawing, at most 24.

**What would change the paper's wording.** A mechanism operator with a valid case not refused by its expected check is a defect of the implementation: the paper reports it and the author receives it as a product question. A discovery operator refused in less than 95 % of its valid cases is named in the paper as a defect the checks do not reliably catch, with the rate. An out-of-scope class refused in more than 5 % of its cases is reported with the codes that refused it, and the paper's Section 4.5 says which part of it the checks do see.

### 1.10 What the harness's tests must show before the full run

- Every operator, applied to a fixed host, changes the submission, and changes it as the table says.
- The wrapper of section 1.5 leaves out exactly the source questions, and nothing else.
- The completion of section 1.3 builds a submission the reviewer reads with no problem for at least one host of each level.
- The checks judge independently. `Judge` calls every check on the submission alone and only gathers their problems (`internal/domain/checks/review.go:141–175` at the pin), and the harness shows it: for one pair of operators per pair of checks that both have one, a case carrying both defects gets exactly the problems of the two single-defect cases together. This is what lets "what only one check catches" be read off the outcomes, in place of switching each check off in the product.
- Each test fails when what it checks is broken (the project's rule on tests).

`just research-faultinject` (S37) reproduces every number.

## 2. E-A3: simulated learners (RQ2)

### 2.1 Questions

How well the learner model estimates a child's level, how well it keeps the child's tasks in the corridor, how often it declares a topic mastered that is not, how far it lags behind a child who learns, and how well its trial series places a child whose grade put them a level off — each against baselines.

### 2.2 The code under test and the loop

The service is simulated through its own domain code: the rule `tutor.Next`, the profile's `Ask`, `Issue` and `Record`, the rating package beneath them (`rating.Update`, `rating.Estimate`, `rating.NewCorridor`), and the shipped catalog. Sealing is the product's own `seal.KeyRing` with a key made for the run. One step of the loop:
1. The rule writes the brief from the profile: `tutor.Next(p, catalog, tutor.Choice{})`, with no choice of the model's except in section 2.6.
2. The request is opened (`Ask`), and a task is issued at the brief's point (`Issue`), with a key chosen by seed and a placeholder text.
3. The simulated child answers (section 2.4), and `Record` records the answer: the trial series, the update, the runs of mastery, the window and the failures, all as the service does them.
4. A task is answered two minutes after it is issued, and the next is issued at once; every tenth answer starts a new day.

A run ends after 200 answers. Every learner is run under every rule with the same true parameters and the same random draws, so that rules are compared on the same children. A learner has two streams: one that draws its true parameters once, and one for its answers, from which every answer takes the same draws in the same order whatever happens — the writing error, the uniform number its correctness is decided by, the wrong letter, the uniform number for the hint and the task's key — used or not. The k-th answer of a learner therefore meets the same draws under every rule. The selection depends on the rule's estimates, so the tasks themselves differ between rules: the loop is closed, as in the service.

### 2.3 Rules and structures

A rule is an estimator of the child's level. The service's rule is the product's own path, untouched. For every other rule the harness keeps the rule's state; before each call to `tutor.Next` it writes the rule's current level into the profile — `Ratings.Theta` and each topic's `Delta` — and after `Record` it replaces what the service computed by the rule's own update. Selection, the trial series' choice of topics, failures and mastery are therefore the service's for every rule; only the estimate differs. Mastery reads the chance of success the service's formula gives at the rule's level, so the gate of 0.775 means the same for every rule.

Every estimator sees the difficulty the brief asked for, β = s_ℓ + (d − 3), never the task's true difficulty.

**Structures.**
- *General:* one level per child; topic offsets held at zero.
- *Topics:* one level per topic, each starting at the child's start; the overall level held at the start. Where the trial series runs first (below), its estimate becomes the starting level of every topic, and the overall level returns to the start.
- *Both:* the overall level plus a topic offset, as the service.

**Update rules.**
- *Shrinking step* — the service: K = K₀/(1 + 0.05n), K₀ = 0.2 for θ and 0.4 for δ, after the trial series [C011, C080].
- *Constant step:* K = K₀, the service's first step, with no decay.
- *Gradient step:* the service's step multiplied by w(P) = (P − c)/(P(1 − c)), c = 0.2 — the step of maximum likelihood (paper A, Section 5.2).
- *Glicko-2,* as Glickman's description of 2022 gives it, on its own scale: one answer is one rating period; the task is an opponent of known strength β, so φ_j = 0 and g = 1; the child starts at μ = θ₀ and φ = 350/173.7178 with a volatility of 0.06, and τ = 0.5 and ε = 0.000001.
- *Glicko-2 with the guessing floor,* this experiment's own construction (Q69). With λ = 1/(1 + e^−(μ − β)), the expected score is E = c + (1 − c)λ, and v and Δ are those of the Bernoulli likelihood with that E: v = [E′²/(E(1 − E))]⁻¹ and Δ = v E′(s − E)/(E(1 − E)), with E′ = (1 − c)λ(1 − λ); the rating moves by μ′ = μ + φ′² E′(s − E)/(E(1 − E)). With c = 0 and g = 1 these are Glickman's own v, Δ and μ′. The step of the volatility and the update of φ are Glickman's with this v and Δ.
- *Urnings,* after Bolsinova et al. (2022), for a learner against items of known difficulty: the learner's urn holds n = 20 balls, the size their analysis of Math Garden gave learners, starting at 10; the item is an urn of fixed proportion π_j = σ(β − θ₀), never updated. After an answer X, a simulated outcome X* is drawn from the two urns' game, P(X* = 1) = r(1 − π_j)/q(r) with q(r) = r(1 − π_j) + (n − r)π_j; the proposal r̃ = r + X − X* is accepted with probability min(1, q(r)/q(r̃)). For a learner who answers with the urn game's own chance σ(θ − β), this makes Binomial(n, σ(θ − θ₀)) the urn's stationary law, the same for every item, so items drawn independently of the urn keep it — our derivation, checked by simulation (section 2.8). The level is θ₀ + ln((r + 1)/(n − r + 1)). The correction of Metropolis–Hastings for adaptive selection is left out, because the service selects deterministically and the correction would then reject almost every move; Urnings' variance is therefore expected to inflate under the service's selection, as its authors describe for adaptive matching.

The shrinking, constant and gradient steps start after the service's trial series, as the service does: for the first five answers the harness leaves the service's estimate as it is, and from the sixth on applies the rule's own step to it. Glicko-2, in both forms, and Urnings start from the first answer: each has its own way of moving fast at first, a wide φ or a small urn.

**The sweep.** The three step rules run in all three structures. Glicko-2, in both forms, and Urnings have, as published, no form for a level split into an overall part and a topic part, so they run in the general and topic structures only. That makes fifteen cells.

**Variants of the service's rule,** run on the generators named in section 2.6:
- floors under θ's step, K_θ = max(0.2/(1 + 0.05n), K_min), K_min ∈ {0, 0.01, 0.02, 0.05} (Q38);
- K₀ halved and doubled, for θ and δ together;
- the constant step at the service's step after 20 answers, 0.1 for θ and 0.2 for δ;
- the trial's estimate replaced by the update: `rating.Update` from the first answer. The first five tasks are still chosen as the trial chooses them — a new topic each time — because the rule tells the series by the count of answers, which also sets the step and is left as it is. What this variant removes is the trial's estimate, the part that places the child;
- the corridor's target moved: every task chosen at the point whose predicted chance is nearest 0.70, or nearest 0.85, instead of the service's 0.775, through the model's choice of section 2.6; the gate of mastery stays at 0.775;
- the model's own slope and floor changed: a ∈ {0.5, 2} and c ∈ {0, 0.3} in place of 1 and 0.2, in the predicted chance the step uses and in the choice of each task's point, which goes through the model's choice as above; the gate of mastery reads the service's chance.

Generators G6 and G7 measure the same assumptions from the other side: children whose slope and floor differ from the model's.

The variants and the baselines are written in the research module. The harness's own copy of the step rules must give exactly what `rating.Update` gives at the service's constants (section 2.9).

### 2.4 Simulated children

Every child is synthetic. The **base population**:
- a grade drawn uniformly from 1 to 6; the start θ₀ is the service's, `rating.Start(grade)`;
- a true overall level θ = θ₀ + u, u ~ N(0, 1);
- a true offset per topic of the catalog, δ_t ~ N(0, 0.5²), independent;
- a task's true difficulty, the asked β plus a writing error ε ~ N(0, 0.5²) drawn per task: a model does not hit the difficulty it is asked for;
- an answer correct with probability c* + (1 − c*)σ(a*(θ + δ_t − β_true)), with the true slope a* = 1 and floor c* = 0.2;
- a wrong answer chooses one of the four wrong letters uniformly; the hint is never used;
- no learning.

The **generators** change the base population in one respect each:

| Generator | Change |
|---|---|
| G0 static | none: the model's own assumptions, apart from the writing error |
| G1 misplaced | u = −2.5 for half the children and +2.5 for the other half: a grade a level off |
| G2 learning and forgetting | θ grows by 0.01 per answer; δ_t grows by 0.02 per answer in topic t and, after every answer in another topic, falls back toward its first value by 1 % of the distance |
| G3 a jump | θ grows by 1.0 at once, at an answer drawn uniformly from 50 to 150 |
| G4 one harder host | tasks are written alternately on two hosts, and the second writes them harder: its ε ~ N(0.75, 0.5²) |
| G5 linked topics | the topics fall into two groups: the first ⌈T/2⌉ in the catalog's order, T being the number of topics (nine of seventeen at the pin), and the rest. A sign s = ±1 is drawn per child; the first group's offsets are 0.75s and the second's −0.75s, each plus N(0, 0.25²) per topic, so the child is strong in one group and weak in the other |
| G6 slope | a* = 0.5 for half the children, 2 for the other half |
| G7 floor | c* = 0 for half the children, 0.3 for the other half |
| G8 hints | the hint is used on 20 % of the answers, which bars a correct answer from counting toward mastery |

One thousand children per generator, in every cell, the same thousand for every rule.

### 2.5 Measures

The **truth** at a moment is the child's true θ + δ_t. A topic is **truly mastered at a level ℓ** when the child's true chance of success on a task of that level and topic at difficulty 3, with no writing error, is at least 0.775, the middle of the corridor; a secondary reading uses 0.70.

| Measure | Definition |
|---|---|
| R1 error of the level | over the topics the child has answered, the root mean square and the mean of (estimated − true) θ + δ_t, after 5, 10, 20, 50, 100 and 200 answers |
| R2 calibration | the Brier score of the rule's own predicted chance against the answers; the mean predicted chance minus the share correct; the expected calibration error over ten bins of width 0.1 on [0, 1], so that the rules with no floor, whose chances fall below 0.2, are binned like the rest |
| R3 corridor | the share of tasks whose true chance of success lies in [0.70, 0.85]; the shares below 0.50 and above 0.95 |
| R4 false "mastered" | of all the moments mastery was declared, the share at which the topic was not truly mastered at the level it was declared at |
| R5 late mastery | for each topic and level the child truly masters, the answers in that topic from the first one at which it is truly mastered until the rule declares it; the share never declared within 200 answers |
| R6 lag | G2: the mean of (estimated − true) over answers 101–200. G3: the answers from the jump until the error, averaged over the topics answered since the jump, falls below 0.5 and stays below it for ten answers; a child for whom that never happens is counted at the horizon and reported apart |
| R7 placement | G1, the service and the variant with the trial's estimate replaced by the update: the error of θ after 5 and after 10 answers; the longest run of wrong answers among the first 15; the share of the first 10 tasks whose true chance is below 0.50 |

**The chance of a false "mastered", computed.** Separately from the simulation, a script computes the Markov chain of paper A's Section 5.3: the run of qualifying answers on eligible tasks — those whose predicted chance was at most 0.775 — with states 0, 1, 2 and an absorbing 3, for a child whose true chance on an eligible task is p and who uses the hint on a share h of them. A correct answer without the hint, with chance p(1 − h), moves the state up by one; a correct answer with the hint, chance ph, leaves it where it is; a wrong answer, chance 1 − p, returns it to 0. These are the moves of `Topic.master` (`internal/domain/profile/answer.go:360`) on an eligible task. The script gives the chance of reaching 3 within m eligible attempts for p from 0.05 to 0.95 in steps of 0.05, h ∈ {0, 0.1, 0.2} and m ∈ {3, 5, 10, 20, 50}.

The chain leaves out what only lowers that chance in the service: a wrong answer on a task that is not eligible also returns the run to 0, and mastery needs five answers in the topic besides. For a child whose true chance is p on every eligible task, it therefore bounds from above the chance of being declared mastered within m eligible attempts. In the simulation the true chance is not the same on every eligible task — it moves with the task's point and its writing error — so the chain bounds nothing there. The report gives the chain's numbers as what the rule does under its assumption, and beside them **R4b**, on G0 with the service's rule: among the children and topics not truly mastered at the level of their tasks, the share declared mastered within their first m eligible attempts, with the distribution of the true chances on those attempts. Neither is used to test the other.

### 2.6 Which cells run

- The fifteen cells of the sweep, on every generator from G0 to G8.
- The floors, on G0 and G3. K₀ halved and doubled, on G0 and G2. The slower constant step, on G0 and G2. The trial's estimate replaced by the update, on G0 and G1. The corridor's target at 0.70 and at 0.85, on G0 and G2. The model's own slope and floor, on G0, G6 and G7.
- **The bias of the corridor.** On G0, the shrinking and the constant step in the structure of both run twice: with the service's selection, and with every task chosen at the point whose predicted chance is nearest 0.5. That choice is passed to `tutor.Next` as the model's own choice of level and difficulty, with the reason "simulation", which is how the service lets a model choose. The mean of (estimated − true) after 200 answers, compared between the two, shows whether a target above one half biases the estimate, as Bolsinova et al. (2025) found for a constant step.

### 2.7 Primary comparisons

Four comparisons are primary; every other number is secondary and is labelled so in the paper.
1. On G0, the service against each baseline: R1 after 200 answers, R3, R4.
2. On G2, the service against each baseline: R6, R3.
3. On G1, the service against the variant with the trial's estimate replaced by the update: R7.
4. On G3 and G0, the floors against no floor: R6 on G3, R1 after 200 answers on G0.

Every number is a mean over the children of its cell with a 95 % percentile bootstrap interval, 2,000 resamples over children. A comparison is a mean paired difference over the same children with its interval. There are no tests of significance: a difference is read as found when its interval excludes zero, and every comparison is reported whichever way it falls.

### 2.8 What is checked against the papers

- **Glicko-2** reproduces Glickman's worked example — a player at 1500 with RD 200 and volatility 0.06 against players at 1400, 1550 and 1700, with RDs 30, 100 and 300, winning the first game and losing the other two, τ = 0.5 — to the figures he prints: a rating of 1464.06, an RD of 151.52 and a volatility of 0.05999.
- **Urnings** reproduces the simulated example of its paper at a smaller scale, the same design otherwise: players whose logits are equally spaced quantiles of the standard normal, urns of 100 balls starting at 50, matching proportional to exp(−2(ℓ_i − ℓ_j)²) on the smoothed logits, with the Metropolis–Hastings correction. The average of each player's scaled urning after the first tenth of the games lies within 0.02 of the true proportion on average over players. The scale is set by the harness's run time before the full run, and the report states it.
- **The fixed-item Urnings of section 2.3**, for a learner of fixed ability θ who answers with the urn game's own chance σ(θ − β), with no floor, items of random difficulty drawn independently of the urn: the share of time the urn spends at each count matches Binomial(20, σ(θ − θ₀)) within 0.01 at every count.

### 2.9 What the harness's tests must show before the full run

- The step rules written in the research module give the same θ and δ as `rating.Update`, to 1e-12, on 10,000 random states and answers at the service's constants.
- On G0 children with no writing error, under the service's rule, the root mean square error of R1 after 200 answers is smaller than after 20: the estimate converges on the trivial learner.
- The three checks of section 2.8 pass.
- The Markov chain script agrees with a simulation of the same chain to within 0.005 at three points of the grid.

`just research-learnersim` (S38) reproduces every number.

## 3. E-A4: cost of the checks

### 3.1 What is measured, and where

Locally, in the devcontainer, with the processor, the number of cores, the Go version and `GOMAXPROCS` recorded:
- **P1, latency of a review.** `Examine` and `Judge` separately, on every eligible host of E-A1, ten times each, one at a time: the median, the 95th and 99th percentiles and the maximum. Each run of each solver with its steps and its duration, as the sandbox reports them.
- **P2, memory.** Go benchmarks of a whole review: bytes and allocations per review.
- **P3, the package.** The package `next_task` hands the model, built by `content.Package` for every point of the ladder each topic is taught at: its size in bytes, and an estimate in tokens of a quarter of its characters, stated as a rough rule rather than a tokenizer's count.

From the product's own reports, each number carried into a data file with the line of the report it comes from (G2):
- **P4, an instance.** Throughput, cost per task in vCPU-seconds and GiB-seconds, and cold start, from the load measurements of T52b [C086].
- **P5, in the cloud.** What T64 measures against the deployed service, once T64 is done. Until then the paper names the cloud numbers it does not have. No load is run against the cloud from the research: that spends money and waits for the author.
- **P6, a family's cost.** Messages of a subscription per task, from the live runs of T62–T64 once they are done; until then, not measured.

### 3.2 The table of cost

For the operator: no model inference, since the service makes no model calls [C003]; Cloud Run's free tier, from the platform's published limits, against P4. For the family: the subscription it already has, in messages per task (P6).

### 3.3 Analysis

Medians and percentiles; for P1, 95 % bootstrap intervals of the median over hosts.

### 3.4 What the harness's tests must show

The benchmarks run the same review as E-A1 on the same completed hosts, and one host's review is checked to give the sanity run's outcome.

`just research-perf` (S39) reproduces the local numbers.

## 4. For the product (G6)

The answer event, `answer_recorded`, carries the topic, the level, the difficulty, whether the answer was correct, the trap, the hint, "I don't know" and the pace (`internal/transport/mcp/submitanswer.go:344` at the pin), but not the chance of success the service predicted when it handed the task out, although `Record` computes it (`internal/domain/profile/answer.go:251`). With that chance in the event, rounded to two places, and a mark for the answers of the trial series, the published aggregates could show calibration on real children — the predicted chance against the share correct, bin by bin — under the product's own rule of cells of at least ten. This goes to the author as a proposal (Q73); E-A3 does not depend on it.

## 5. What is not covered

- The non-English paths: the reference tasks are in English, so the bigrams of the languages written without spaces, their sentence limits and the readability of other languages are not exercised by E-A1.
- A model's own self-check: E-A1 holds it fixed at "nothing found".
- The model's choice of a topic, level or difficulty of its own, except as the tool of section 2.6.
- Children: E-A3's learners are a model of children, chosen before data, and its generators are the ways this protocol expects children to depart from the model, not a measurement of how they do.
