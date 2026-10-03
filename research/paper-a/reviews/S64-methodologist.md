# S64, internal review 1 of 3: the methodologist

Reviewer: a Claude sub-agent of the executor's session (claude-opus-5-5), given the role of a programme-committee member for AIED 2027 with a focus on design, statistics and reproducibility, read-only access to the repository, and the anonymous build of 2026-10-02. Same model family as the writer, so rule G10 is not met by this pass (Q85). Its report, as delivered:

---

**Method note.** `pdftoppm` is not installed, so I could not render the PDF pages. I reviewed the LaTeX the PDF is built from (`sections/*.tex`, `figures/*.tex`, `generated/numbers.tex`) and checked the build log, which reports 14 pages. The anonymous build does not define `\Final`, so the PDF prints bold "[TBD: K04]" and "[TBD: S65]". Every result number in the paper matches the result files.

## 1. Summary
A chat assistant's own model writes olympiad-style tasks with five options for grades 1–6. A service that calls no model picks the topic and difficulty (hierarchical Elo with a guessing floor, a trial series, a 0.70–0.85 corridor). It admits a task only if nine checks pass, the central one being a solver the model wrote, run twice under rotated labels, and it seals the key. The evaluation is offline: 68 defect operators applied to 447 curated reference tasks (23,453 cases), a closed-loop simulation of synthetic children against 14 estimators, and a cost measurement. There are no learners and no new model output.

## 2. Scores
- **Relevance 4:** checked, LLM-written practice with learner modelling is core AIED.
- **Novelty 3:** the split, the label-rotated solver and the sealed key are new together; the parts are known.
- **Technical soundness 2:** rigorous machinery, but RQ1's headline holds by construction and the E-A3 results that are reported favour the service.
- **Significance 2:** it depends on how many real defects fall in scope, which is not measured.
- **Clarity 3:** exact but dense, and two design facts that decide interpretation are missing (W2).
- **Overall: weak reject.** W1–W2 can be fixed within the page limit and would move it to borderline.

## 3. Strengths
- The protocol was frozen and timestamped before the runs: classes, operators, seeds, primary comparisons and decision rules. Negative results are reported.
- Must-catch, discovery and out-of-scope classes are kept apart, so the real findings are interpretable and actionable: D02b refused 0/213; copies with new numbers refused 94.6 % and 96.1 %, every miss in one topic.
- E-A3 runs the product's own code in a closed loop, with common random numbers, paired bootstrap differences, and baselines checked against their sources.
- Every number comes from saved data. I re-derived the step ratios and the chain value (0.7014) exactly.

## 4. Weaknesses

**W1 (MAJOR). E-A3 reports three of nine generators and leaves out a primary comparison; the left-out results go against the service.**
- Where: §6.2 lists nine generators, but Results cover only G0–G2. Abstract: "the service estimated a child who stays put as well as any baseline". §8: "a floor under the step and a stricter rule are future work."
- Checked (`learnersim/results/numbers.txt`): the service is best or tied on G5, G7 and G8. Plain Glicko-2 (one level) has lower RMSE after 200 answers on G3 (0.57 vs 0.83), G4 (0.61 vs 0.66) and G6 (0.66 vs 0.74).
- On learners (G2): Glicko-2 lags 0.24 logits against the service's 1.12, and keeps 43 % of tasks in the corridor against 26 %.
- After a one-logit jump: 71 % [68–74] of children are still not caught up at answer 200 under the service, against 5 % under Glicko-2.
- Floors: primary comparison 4 (`comparisons.csv`) already tested them. Floors of 0.01–0.02 change nothing; 0.05 saves 7 answers (children not caught up: 71 % → 56 %).
- Calibration (ECE 0.004) is shown only on G0, where children answer by the service's own curve; on G2 it is 0.079.
- Why reviewers care: nine generators in the design and three in the results reads as selection.
- Fix: replace Figure 2 (G0 only) with a nine-row table — the service against the best baseline, per generator — or cut Table 1, which restates §§3–5. Qualify the abstract and §8: "on children who answer by its own curve; Glicko-2 tracked learning, a jump and another slope better".

**W2 (MAJOR). RQ1's headline holds by construction.**
- Where: Abstract: "of 23,453 injected defects, every one that broke a rule a check states was refused". §6.1: "so switching off a check would let its defects through."
- Problem: the 53 mechanism operators come from the checks' own specification. Protocol §1.4 says a check that enforces its rule "refuses every such case". So 18,383/18,383 is a conformance test. "One check alone" follows from each operator targeting one check. The protocol's unit is the operator (§1.9), not the case, yet the abstract leads with cases.
- Missing from the paper: the self-check is held at "nothing found" (protocol §1.3), and D01–D03 always run the curated reference solver. So E-A1 cannot see the central check's real failure, a model's solver that shares the key's error. RQ1 ("in a model-written task") is answered for synthetic defects only. This is not a request for the deferred E-A2.
- Fix (net shorter): reword "all 53 operators that break a stated rule were refused, a conformance result; of the three that break none, …"; drop the switching-off clause; add one sentence on the fixed self-check and the reference solvers.

**W3 (MINOR). Corridor and lag numbers are driven mostly by assumed parameters.**
- Where: "Under every rule most tasks missed the corridor"; "fell 1.1 logits behind a child who learns."
- Problem: with a writing error of SD 0.5 and a ladder step of 1, a rule that knew the child's true level would reach only about 59–64 % (the reviewer's calculation, not in the repository; with no writing error, 93 %). The lag assumes growth of 2 logits per 200 answers, a rate the paper does not justify.
- Fix: give this ceiling in one clause, and state the lag as conditional on the assumed rate. Pay for it by cutting the test-function counts in §3.

**W4 (MINOR). False mastery is reported as a share of declarations, with no base rate or sensitivity.**
- Where: "Of the service's declarations of mastery, 63 % were false."
- Problem: the share depends on how many child-topics are truly mastered, and on a threshold equal to the rule's own target. Computed but not reported: the preregistered 0.70 reading (52.6 % false), and R5 (21.8 % of truly mastered topics never declared). Every estimator shares the same run-of-three rule, so the 14 R4 comparisons say little.
- Fix: one sentence with both numbers. Pay with the 30-second live-run sentence in §6.3.

**W5 (MINOR). The trial series is judged on the scenario it was tuned on, against a weak comparator.**
- Where: "the trial series did what it was built for … 0.94 … against 2.24."
- Problem: σ₀ = 2.5 was tuned on simulated children placed a level off who answer by the model's formula (R79, `docs/decisions.md:134` at the pin) — exactly G1. Glicko-2 starting from the first answer reaches 1.07 after 5 answers and 0.78 after 10; the service is at 0.87 after 10.
- Fix: one clause naming the tuning and the Glicko-2 figures. Pay by dropping the ECE sentence.

**W6 (MINOR). Statistical reporting.**
- §6.2 promises "a 95 % bootstrap interval", but the text prints none.
- "No difference found" (0.495 vs 0.493) becomes "as well as", an equivalence claim. Print the paired interval, [−0.004, 0.008].
- With 1,000 children and common random numbers, tiny differences count as "found" (0.7 points of corridor share).
- Multiplicity: by the reviewer's approximation from `comparisons.csv`, about 60 of the 65 found differences survive a Bonferroni correction; say that instead of "a few perhaps by chance".
- Table 3: rows X01 and X04 pool two operators on the same hosts, yet print exact intervals as if the cases were independent. X04 shows 0.0–0.4, against 0.8–0.9 per operator. Protocol §1.9 prescribes a bootstrap over hosts.

**W7 (MINOR). Out-of-scope validity rests on an unverified AI reading.**
- Where: "read by the AI assistant that ran the experiment, its verdicts still to be checked by us [TBD: K04]".
- Problem: the reader built the operators and was not blind. Four of ten has a 95 % interval of 12–74 %. In `reading.md`, only 2 of the 10 X03a cases actually turn the question around (`par-12-d4-5`, `par-12-d3-4`); the rest are contradictions, a degenerate question, an ambiguity, or equivalent mutants.
- Fix: have a person check the readings and report agreement. Relabel X03 as "a word swapped for its opposite".

**W8 (MINOR). The preregistration cannot be checked at review.**
- Where: "code, data and seeds … are to be released [TBD: S65]"; the link prints "[anonymised artifact link]".
- Undisclosed deviations (recorded in `research/questions.md` Q74–Q76): R4b switched to a Kaplan–Meier estimate after review moved it from 47 % to 44 %; R6 read as a signed error; nine harness faults fixed after the first runs; perf records overwritten. None of this is mentioned beside "frozen and timestamped before their first runs".
- Fix: link the anonymous artifact the release recipe already builds, and point to its list of deviations.

**W9 (MINOR). E-A4 times only the cheap path.**
- It measures accepted reference tasks only; the largest solver used 9.7 % of the step limit. The worst case, two solver runs at 2 s or 25M steps, is not measured. "39 % of its budget" hides what a family pays, about 5,000 tokens per package.
- Fix: give tokens and the worst-case bound instead of the percentage.

## 5. Questions to the authors
1. In the live run, was the one `solver_disagrees` refusal a real wrong key or a solver bug?
2. Why keep the shrinking step, given how much better Glicko-2 tracks learners and jumps?
3. The sanity run refuses 156 of 603 vetted reference tasks (37 % at grades 1–2; 115 as near-copies of other reference tasks). Would a teacher accept them? This is your only evidence on false refusals.
4. Do other equivalent forms, such as ½ and 0.5 or 3 m and 300 cm, pass the way D02b does?
