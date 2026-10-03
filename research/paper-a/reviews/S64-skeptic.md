# S64, internal review 3 of 3: the skeptic

Reviewer: a Claude sub-agent of the executor's session (claude-opus-5-5), given the role of the programme-committee member who looks hardest for a reason to reject, read-only access to the repository, and the anonymous build of 2026-10-02. Same model family as the writer, so rule G10 is not met by this pass (Q85). Its report, as delivered:

---

## 1. Summary
The paper presents an MCP service. The chat assistant's own model writes five-option maths tasks for grades 1–6, and a deterministic service admits a task only when nine checks pass. The central check runs the model's own Starlark solver twice, the second time with the option labels rotated. The answer is sealed in a profile in the parent's Drive, and a hierarchical Elo model with a guessing floor sets the difficulty.

RQ1 injects 23,453 hand-made defects into 447 reference tasks. RQ2 simulates 9×1,000 learners against 14 baselines. Apart from 15 anecdotal tasks, the evaluation has no newly generated task, no expert judgement and no learner. §4 and §7 are candid, but the title, abstract and conclusion promise more verification than the checks give.

## 2. Scores
- **Relevance 4.** LLM-written practice and learner modelling are core AIED topics.
- **Novelty 2.** Checking an answer with code the model wrote (PoT, code-based self-verification) already exists, as do generate–validate–revise loops (AIED 2026). The rotated second run is a small twist.
- **Soundness 2.** RQ1's in-scope result holds by construction. The simulated children answer by the service's own curve. Comparisons are reported selectively.
- **Significance 2.** There is no evidence on real model output or on learners.
- **Clarity 3.** The text is compact, but TBD markers are visible and the framing contradicts §4.
- **Overall: reject.** It would be borderline if reframed and given even a small audit of real tasks.

## 3. Strengths
- Open about blind spots: §4 "What the checks cannot see" names them, Table 3 measures the out-of-scope defect classes, and the paper says the answer key reaches the task card.
- Sound procedure: the protocol was frozen before the runs and its SHA-256 matches the run journal; the design is seeded and paired; the baselines reproduce their source papers (`research/experiments/learnersim/results/checks.txt`); every number is generated from data.
- Clean separation of roles: the service calls no model, the answer is sealed, and the profile lives in the parent's Drive.

## 4. Weaknesses

**W1 — MAJOR. RQ1 is mostly true by construction, and the abstract leads with that part.**
- Where: abstract; §6.1 Results; §8 ¶1. Quote: "Offline, of 23,453 injected defects, every one that broke a rule a check states was refused".
- The 53 "mechanism" operators are written from the checks' own rules; the protocol says so: "A check that enforces its rule refuses every such case … a miss is a defect of the implementation" (`research/experiments/PROTOCOL-A-offline.md:63`). D01's wrong keys are caught only because each host keeps its trusted reference solver. The rotated second run matters for exactly one operator, D06a (`return ["K"]`), which was built to trigger it. No injected defect is a mistake a model made, yet RQ1 asks about "defects in a model-written task".
- What the abstract hides: its denominator includes 5,070 cases that break no stated rule. Most of those semantic defects passed: refusal rates of 7.1 % for a removed condition, 1.0 % for an inverted question, 0 % for an answer-revealing hint and 0 % for a wrong trap (`faultinject/results/numbers.txt`).
- Fix, at equal length: in §6.1, "refused, as a correct implementation must: an end-to-end test of the code"; in the abstract, "…while semantic defects (a removed condition, an inverted question, a hint giving the answer away) passed in 93–100 % of cases".

**W2 — MAJOR. "Verified", "A key to trust" and "confirmed the key twice" contradict §4.**
- Where: title; Table 1; §8. Quote: "Only the self-check is the model's opinion of its own work; the solver is the model's program, but the service runs it and judges the result."
- The solver is not independent: it is the same model's formalisation of the task, as §4 itself admits. Running it adds determinism, not independence. The second run adds almost nothing as verification: the model is told the solver "runs twice … never write a letter yourself" (`content/instructions/task_writing.md:66–68` at the pin), and the code calls the second run a guard against "the laziest failure" (`internal/domain/solver/verdict.go:56–57` at the pin).
- The only real-model run points the other way (`docs/live/04-local-run.md:124,176–177,187` at the pin): of 15 tasks, none had a wrong key; the solver check fired once, and that was a false alarm — the solver said "Mail" where the option read "Mail car"; two hints nearly gave the answer away, and no check saw them. The paper reports neither the false alarm nor the hints.
- "Olympiad-style" is also unmeasured (§6.3). The released tasks include "Tom is 6 years old. How old will he be in 3 years?" (`cal-12-d1-2`), which fails the paper's own definition in §1.
- Fix, about zero net lines: drop "Verified" from the title and say "olympiad-inspired"; in Table 1, "a key the model's own program agrees with"; in the conclusion, "agrees with the key under relabelled options"; replace the quoted sentence with the live run's refusals.

**W3 — MAJOR. RQ2 reports the favourable comparisons and leaves out the unfavourable ones, including a pre-registered primary comparison.**
- Where: §6.2; abstract; §8. Quotes: "…against 0.48 for a constant step"; "the trial series did what it was built for".
- Checked in `learnersim/results/numbers.txt` and `comparisons.csv`: on learners (G2), Glicko-2 with one level per child lagged 0.24 logits against the service's 1.12, its RMSE was 0.61 against 1.44, and it kept 42.6 % of tasks in the corridor against the service's 25.7 %. Jump (G3), not reported: 71 % of children were never caught up within the run, against 9 % under a constant step. Step floors (primary comparison 4) are missing: a floor of 0.05 shortened catch-up by 7 answers at no measurable cost to children who stay put, yet §8 calls a floor "future work". Mastery: five per-topic baselines made significantly fewer false declarations of mastery (Glicko-2 with the floor: 53.9 % against 63.2 %). Trial series: its stated purpose, a child who would "fail task after task", shows no change in the longest run of wrong answers (2.02 against 2.03); its prior spread σ0 was tuned on the same response model (`docs/decisions.md` R79), so the G1 result is in-sample. "As well as any baseline": the paper's own variant without the trial series beats the service on children who stay put (0.470 against 0.495). Calibration: the ECE of 0.004 holds only for children who answer by the service's own curve; for learners it is 0.079.
- Fix: report the lag range (0.24–0.48 against 1.12) and the corridor shares (26 % against 37–43 %); add one sentence each on G3 and on the floor; pay for it by cutting §5.2 to two sentences, about 8 lines.

**W4 — MAJOR. Unfinished, AI-judged material is in the review PDF.**
- Where: end of §6.1; §8. Quotes: "read by the AI assistant that ran the experiment, its verdicts still to be checked by us [TBD: K04]"; "[TBD: S65]".
- These readings decide whether the out-of-scope mutants are real defects, and they already flag some that are not (X03: 4 of 10).
- Fix: have a person read the 130 sampled cases, and say who read them; remove both TBD markers; add an anonymised artifact link. Naming the assistant here, with no methods disclosure (Q83), invites the question.

**W5 — MAJOR. The checks refuse a third of the authors' own grade 1–4 tasks, yet the paper says false refusals are "not measured".**
- Where: §6.1 ¶1. Quote: "the others read too hard for their level or as near-copies of another."
- Refusals by level: 74 of 200 at grades 1–2, 82 of 250 at grades 3–4, and 0 of 153 at grades 5–6. Whole topics fail at grades 3–4: knights and liars 23 of 25 (as near-duplicates) and weighing 22 of 25 (`faultinject/results/sanity.csv`). The knights-and-liars tasks are distinct problems that share one stock opening ("On an island, knights always tell the truth…"), and the same tasks are shown to the model as examples.
- Hidden assumption: the reference tasks are in English "whatever the language of the chat" (`content/example.go:17–21` at the pin). For a task in any other language, the check against copies of reference tasks does nothing. Three of the four live scenarios were in Russian.
- Fix: one sentence in §6.1 and one clause in §7; pay by cutting the MinHash sentence about 62,653 pairs.

**W6 — MINOR. The safety framing claims more than is checked.**
- Quote: "Letting a chat model write for a child is safe only as far as what it writes is checked." The product tells the model "you talk with the child" (`content/instructions/mcp_instructions.md:3,41` at the pin), yet nothing checks the chat itself, and both hosts bar children.
- Fix: state in the abstract that an adult runs the lesson.

**W7 — MINOR. Protocol deviations go unmentioned.**
- The claim that the protocols were frozen holds, but deviations exist (Q75, Q76). X09 was re-read after the results (Q78). D14b, refused in 94.6 % of cases, falls under the protocol's own rule that anything under 95 % is named as not reliably caught (line 168); §6.1 instead says "mostly refused". Table 3 pools X01 and X04, so 53 operators plus 13 rows does not make the stated 68. One added line covers all of this.

**W8 — MINOR. The closest related work is missing.**
- The repository's own notes call Ikram et al. (AIED 2026) and MathQ-Verify the closest work (`research/literature/generation-verification.md:146,310`). Distractor generation from catalogs of mistakes (Feng et al.; DiVERT) is also uncited. Cut the "1,099 test functions…" sentence to make room for one sentence.

**W9 — MINOR. Anonymity and provenance.**
- "22 languages but English and Russian" hints at who the authors are; write "all but two".
- "An earlier study … chose problem books in the public domain; Claude wrote the tasks from their ideas" goes beyond the records: Q82 notes this, P013 says "genre of Soviet collections", and no task records its source.

## 5. Questions to the authors
1. On real model output, how often does the solver disagree with the key, and how often is that a wrong key rather than a text mismatch? What are the checks' precision and recall on 100 model-written tasks with blinded human keys?
2. How often do real solvers hard-code a letter or a value?
3. What share of tasks would an oracle that knows the child's true level keep in the corridor, given the 1-logit grid and the 0.5 writing error? (The reviewer's rough estimate: about 0.6; the answer separates a design problem from an estimator problem.)
4. Why is the 0.05 floor, which was pre-registered and helped, presented as future work?
5. What happens to a new knights-and-liars task that uses the stock opening, and to one written in Russian?

Process note from the reviewer: nothing in the repository was changed. The devcontainer has no `pdftoppm`, so the PDF was read through a local poppler image, with page renders and extracted text written to the session scratchpad only. The PDF's metadata and link annotations name neither the system nor the repository.
