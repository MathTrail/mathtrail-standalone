# S64, internal review 2 of 3: the AIED specialist

Reviewer: a Claude sub-agent of the executor's session (claude-opus-5-5), given the role of a programme-committee member for AIED 2027 with a focus on relevance, positioning, novelty, pedagogy, ethics and clarity, read-only access to the repository and the web, and the anonymous build of 2026-10-02. Same model family as the writer, so rule G10 is not met by this pass (Q85). Its report, as delivered:

---

All 14 pages read as page images, rendered into the scratchpad with a local poppler image; nothing in the repository was changed.

## 1. Summary
The paper presents a service that runs inside chat assistants over MCP.
- Division of work: the host's model writes each olympiad-style, five-option task for grades 1–6. A deterministic service that calls no model picks the topic and difficulty, admits a task only if nine checks pass, and seals the answer in a profile kept in the parent's Drive.
- Central check: a Starlark solver written by the same model must select exactly the keyed option, and must do so again with the labels rotated.
- Learner model: a hierarchical Elo with a 0.2 guessing floor on a designed cross-grade ladder, with a placement series and a 70–85 % success corridor.
- Offline results: every injected defect that breaks a rule a check states is refused, while semantic defects pass. In simulation, children who stay put are estimated well, but only 42 % of tasks land in the corridor, a learning child is trailed by 1.1 logits, and 63 % of mastery declarations are false.
- No child and no real model output is evaluated.

## 2. Scores
- **Relevance 4:** core AIED topics; the infrastructure detail is off-audience.
- **Novelty 3:** the program-checked key under label rotation is new among the works cited. The learner model is standard, and the venue's nearest work is missing.
- **Technical soundness 3:** the protocols are rigorous, but 53 of 68 operators test the checks' own rules, the simulated children answer by the service's own curve, and no real output is judged.
- **Significance 2:** nothing shows that the checks matter for real tasks, and RQ2 is mostly negative.
- **Clarity 3:** well structured, but compressed and idiosyncratic for AIED readers.
- **Overall: weak reject.** It reaches borderline with the text-only fixes (W2–W7), and weak accept is within reach with the rating proposed in W1.

## 3. Strengths
- A clear, timely split: the model writes; a deterministic service chooses, checks and seals. It is private by design and costs the operator no inference.
- The rotated-label check is cheap (median 3.45 ms), and its guarantee and blind spots are stated precisely in §4.
- Exemplary reporting: timestamped protocols, exact intervals, scripted numbers, and negative results in the abstract.
- A strong baseline sweep, including Glicko-2 and Urnings. The trial series helps children placed a grade off (0.94 against 2.24 logits after five answers).
- The code and the 603 reference tasks with their solvers are released under MIT.

## 4. Weaknesses

**W1 (MAJOR). No real model-written task is judged, and RQ1's headline is true by construction.**
- Where: the title, the end of §3, §6.1. Quotes: "Every case of the 53 operators that break a rule a check states was refused, by the check expected"; "Such runs show that the path works but are too few to measure anything".
- Problem: those 53 operators inject exactly what a check's rule names, so 100 % refused works like a test suite. The informative results are the misses: D02b at 0 %, X01–X08 at most 7.1 %. The base tasks were written by Claude and the defects are synthetic. Readers never learn how often a real task for a 7-year-old is wrong, ambiguous or not olympiad-style, or how many good tasks get refused. The checks themselves refuse 74 of the 200 vetted grade 1–2 reference tasks, 47 of them on readability. The title's "Verified Olympiad-Style" names exactly what §6.3 (Construct) says is not checked.
- Why raise it: it is known and deferred (E-A2, Q57), but the reviewer judges it the likeliest reason for rejection, because reviewers will read RQ1 as tautological.
- Checked: Q74 in `research/questions.md`; `research/experiments/faultinject/results/sanity.csv`.
- Fix: blind-rate 60–100 pipeline tasks, accepted and refused, from the model family the paper already uses (Claude); the author plus one teacher rate the key, ambiguity, the §1 olympiad definition and grade fit, reporting κ and Clopper–Pearson intervals; about 10 lines. Fix without new data: write "Checked" instead of "Verified", say in the abstract that no real output was rated, and lead RQ1 with the misses.

**W2 (MAJOR). Related work misses the venue's nearest work and the item-generation strands.**
- Where: §2, paragraphs 1 and 3. Quote: "Among the works we read, none writes olympiad-style problems for primary school, and none admits a problem by running a program its author wrote against the problem's options."
- Cut in the 14-page version: `ikram2026multiagent` (AIED 2026, LNCS pp. 233–247, doi:10.1007/978-3-032-29744-0_16) — generate, validate with model agents, revise, up to three rounds; Q72 names it the nearest work for N1. `feng2024exploring` and `fernandez2024divert` — wrong options built on named misconceptions, the closest analogue of the paper's traps. `openai2025study` and `khan2026chatgpt` — without them, "why not just use Study Mode?" goes unanswered.
- Never searched: automatic item and question generation — Kurdi et al., *A Systematic Review of Automatic Question Generation for Educational Purposes*, IJAIED 30(1):121–204 (2020); Gierl, Lai & Turner, *Using automatic item generation to create multiple-choice test items*, Medical Education 46(8):757–765 (2012), doi:10.1111/j.1365-2923.2012.04289.x, whose item models compute the key.
- Not found by the authors' searches: Jia et al., *EduAgentQG: Multi-Agent Personalized Mathematics Question Generation with Explicit Diversity and Objective-Aware Evaluation*, arXiv:2511.11635 (grades 1–9, difficulty targets, solvability judged by model agents); Mishra, Poesia, Mo & Goodman, *MathCAMPS*, NeurIPS 2024 MathAI workshop, arXiv:2407.00900 (a cycle-consistency check of the text against its formal problem — exactly the gap §4 concedes).
- Checked: `research/literature/refs.bib:641`; these keys are cited in `paper-a/draft-extended.md` but in none of `paper-a/sections/*.tex`; `research/literature/search-log.tsv` has no search for item or question generation; web search.
- Fix: one sentence per strand, about 6 lines plus 5 references.

**W3 (MAJOR). The mastery result comes without its literature or its mechanism.**
- Where: §5.3 and §6.2, paragraph 2. Quote: "Mastery was the weak part."
- Missing literature: a run of correct answers is known to be weakest when answers are noisy (`pelanek2018conceptual`, `doroudi2020mastery`); read and planned to be cited (Q69), none cited.
- Mechanism not named: mastery is held "at the level of the task that completed the run" (`internal/domain/profile/answer.go:353–365` at the pin), while the simulation's truth is the child's chance at difficulty 3 of that level (`research/experiments/learnersim/child.go:210`); 28 % of the eligible attempts by children not truly mastered were on tasks the child truly finds easy, with a true chance of 0.8 or more (`learnersim/results/numbers.txt:6136`).
- Dropped number: 53 % false even with a 0.70 threshold (`numbers.txt:6097`) would head off a challenge to the yardstick.
- Fix: about 4 lines plus 2 references: add the citations and both figures, and name the missing gate on the estimate (Q75) as the design lesson.

**W4 (MAJOR, ethics). Children use hosts whose terms exclude them.**
- Where: §8, Ethics. Quotes: "though on the card the child answers and may type into the chat directly"; "we claim no compliance".
- Problem: both hosts forbid making the account available to anyone else, and OpenAI asks that an adult conduct the interaction (`research/literature/mcp-safety.md:76–84`). The child's typed words go to a model whose replies no check covers (C100). The child's interests and the parent's notes reach the host (C092).
- Why it matters: an ethics-minded reviewer may read this as a design that invites terms-of-service violations with 6–12-year-olds.
- Fix: a rewording that costs no space — state adult-conducted use as the design condition, name the two flows as risks, and replace "we claim no compliance" with "we make no legal assessment".

**W5 (MINOR). Table 1 promises more than the checks deliver.**
- Quote: "Each wrong option embodies a trap from a closed catalog and explains its mistake". A trap naming the wrong mistake (X04, 0 of 873) and a hint that gives the answer away (X05, 0 of 447) always pass, yet the brief personalises by these trap labels. The 0.70–0.85 band rests on Math Garden alone, because `wilson2019eighty` was cut.
- Fix: add a "checked?" column set ragged-right, which also fixes the stretched "Answer not given away" cell, and cite Wilson et al.

**W6 (MINOR). AI involvement in the method.**
- Quote: "read by the AI assistant that ran the experiment, its verdicts still to be checked by us [TBD: K04]". The reported "4 of 10" rests on LLM verdicts nobody has validated. Springer Nature asks that LLM use beyond copy-editing human-written text be documented, in Methods or in the chapter's acknowledgements; Q83 removed the paper's disclosure.
- Fix: have the author read the 90 sampled cases (K04). Question for the author on Q83: a disclosure in the camera-ready acknowledgements would cost no space in the review version.

**W7 (MINOR). Clarity.**
- The abstract is about 270 words; Springer asks for 150–250.
- §6.2 says "Fourteen alternatives" but names only 11: it never says Glicko-2 and Urnings each run per child and per topic.
- Fig. 2's "shrinking step" is never tied to the service's own step.
- "at P = 0.5 33 % further" runs two numbers together.
- "trap", "card" and "stays put" need a gloss at first use.
- Huang, Zhou, Chen and GSM-Symbolic are cited as arXiv preprints although published ICLR and TMLR versions exist.

Page budget proposed for these fixes (about 35 lines): §6.3 cost paragraph down to one sentence (6); product test counts and MinHash error figures (4); Utami & Hwang, cited unread against the plan's own rule (5); §5.2 Robbins–Monro passage (7); abstract trim (3); engineering detail in §3 (4); merge X07/X08 and X09a/X09b in Table 3 (2); overlap between §7 and the threats to validity (4).

## 5. Questions to the authors
1. What share of real pipeline tasks is faulty, before and after the checks?
2. How many tasks land in the corridor with no writing error, and what supports the assumed 0.5-logit writing error? No reported cell isolates it.
3. Would a gate on the estimate remove most false mastery declarations?
4. With a 6–8-year-old, who reads the chat and who types?
5. Is the readability threshold too strict at grades 1–2, given 47 refusals among 200 vetted tasks and the live run where the model cut the word the answer rested on?
