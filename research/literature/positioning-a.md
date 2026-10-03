# Where paper A stands: what is new, and what is not

The notes of task S35. The question: what, in paper A, is new against the works read for S09, S10, S16 and S31–S34, and, for each claim of novelty, which work comes nearest. The notes also check systematically the prototype's finding that olympiad-style problems for grades 1–4 are barely studied (P037).

Every cell below rests on the notes of an area — `generation-verification.md` (S31), `learner-models.md` (S32), `llm-tutors.md` (S33), `mcp-safety.md` (S34), `problem-solving.md` (S09), `olympiad-and-measurement.md` (S10), `privacy-law.md` (S16) — where the passages are quoted; this file adds no source of its own beyond the searches of section 4. MathTrail's side rests on the ledger at the pinned commit `52ce86908135` ([C0xx]). "Not described" means the work's text or page does not describe the property; it does not mean the property is absent.

## 1. Works against the properties MathTrail combines

The columns are the properties paper A describes:
- **Writes:** a model writes each task anew (MathTrail: [C018, C073]).
- **Program-checked key:** before a learner sees it, the key is checked by running a program, and something other than the model reads the result (MathTrail: the service runs the model's solver against the five options, twice with the labels rotated [C023]).
- **Olympiad:** the problems are olympiad-style, non-routine (MathTrail: [Q15], the working definition).
- **Grades 1–6:** the learners are of primary age (MathTrail: [C010]).
- **In a chat:** delivered inside a general chat assistant (MathTrail: Claude and ChatGPT over MCP [C062, C068]).
- **Learner model:** a model of the learner chooses what comes next (MathTrail: [C009, C018]).
- **Answer kept back:** in a system where a model talks with the learner, the answer is kept from the learner until they answer (MathTrail: sealed in the file and in the tools' results, and not shown on the card [C033, C064, C096]). For a bank of practice items, which naturally shows no answer first, the column is "—".

| Work | Writes | Program-checked key | Olympiad | Grades 1–6 | In a chat | Learner model | Answer kept back |
|---|---|---|---|---|---|---|---|
| MATHWELL (S31) | yes | no: the model's program computes the key, and nothing checks it | no | partly: K–8 | no | no | — |
| EDUMATH (S31) | yes | no: worked text solutions, judged by people and a model | no | partly: 3–5 | no | no | — |
| MathWiz (S31) | yes | no answers written | no | yes | no | no | — |
| MATH², CHASE (S31) | yes | no: the generator's or other models' votes | partly: MATH² is high-school competition | no | no | no | — |
| KPDDS (S31) | yes | partly: a program checks the arithmetic of solutions | no | no | no | no | — |
| GSM-Symbolic (S31) | instances from templates | yes, but the program is written by people | no | no | no | no | — |
| Ikram et al., AIED 2026 (S31) | rewrites existing problems | no: solvability judged by a model | no | partly: middle school | no | no | — |
| Zhou et al. (S31) | no: solves given problems | no: the model checks its own answer with code and reads the result itself | no | no | no | no | — |
| Feng et al.; DiVERT (S31) | wrong options only | no | no | partly: ages 10–13 | no | no | — |
| Khan Academy in ChatGPT (S33) | no: a bank written by people | — (vetted by people) | no | not described | yes | no | no: answers on request, for teachers |
| Beast Academy (S33) | no: written by people | — (edited by people) | yes | yes: ages 6–13 | no | not described | — |
| MathForces (S33) | yes: by Gemini | not described | yes | not described | no | partly: an 80 % gate | — |
| Spiral Learning (S33) | not described: "a vast, adaptive question bank" | not described | yes | partly: ages 7–13 | no | partly: a personalised plan, not described further | — |
| MasterOlympiad, before launch (S33) | no: written by the team | — ("carefully vetted", by whom not stated) | yes | partly: classes 1–12 | no | not described | — |
| Synthesis Tutor (S33) | not described | not described | no | partly: ages 5–11, K–5 | no: its own app | not described | not described |
| Vimi (S33) | not described: curriculum practice | not described | no | partly: grades 1–9 | no: its own app | partly: it finds the level and builds a plan, not described further | not described |
| Strive PDF Generator (S33) | yes: ChatGPT writes worksheets on request | no check described | no | not described | yes: a ChatGPT app | no | no: an answer key on request |
| Tutor MCP (S33) | yes | no task check described | no | no: any subject | yes | yes: a rule-driven engine | not described |
| KMath (S33) | not described | not described | no | no | planned | not described | not described |
| The assistants' study modes (S33) | partly: quizzes in the chat | not described | no | no: college, universities, 18+ for Gemini's practice | yes | no: memory only | partly: by instruction, and it can be got round |
| Math Garden (S32) | no: a bank | — | no | yes | no | yes: Elo, items rated too | — |
| Pelánek et al. 2017 (S32) | no: a bank | — | no | no: geography facts | no | yes: hierarchical Elo | — |
| LearnLM with Eedi (S33) | no: items by people | — (a tutor approves each message) | no | no: ages 13–15 | no | not described | not described |
| Rori (S33) | no: lessons by people | — | no | partly: 3–8 | yes, WhatsApp | not described | partly: hint, then the solution |
| Oreopoulos et al. (S33) | no: problems by people | — (a critic model checks replies) | no | partly: 6–8 | no | partly: a mastery rule | yes: guard-railed |
| **MathTrail** | yes | yes | asked for in the brief, not checked | yes | yes | yes | partly: sealed in the file and in the tools' results [C033]; never shown on the card, though it reaches the card's memory [C096]; in the chat, the model that wrote the key is told to keep it back |

No other row has a model writing the task and the task's key checked by an executed program that something other than the model reads: KPDDS checks only the arithmetic inside solutions, and MATHWELL's program is the key, unchecked. No row has a model writing olympiad-style problems for grades 1–6: MathForces generates olympiad quizzes for an audience its page does not state, and Spiral Learning's page does not say whether its bank is written by people or generated.

## 2. Claims of novelty, each with its nearest competitor

**N1. A chat model writes each task, and a deterministic service admits it only after running a program the same model wrote, which must single out exactly the keyed option of five, again when the labels are rotated.** [C021, C023]
- *Nearest:* MATHWELL — the model writes a program, but the program *is* the key and nothing checks it; Zhou et al. — the model checks its own answer with code and reads the result itself, on given problems; GSM-Symbolic — a program computes the key, but people write it; Ikram et al. — validator agents for generated problems, but solvability is judged by a model.
- *Status:* new among the works read (S31, answer 7). The check sees the model's formalisation, not the meaning of its text [C026]; paper A says so beside the claim.

**N2. Olympiad-style problems for grades 1–6, written for one child, with the key checked before the child sees them.** Whether a task is olympiad-style, and whether it fits its grade, topic and difficulty, is asked for in the brief and not checked [C021].
- *Nearest:* MathWiz writes for grades 1–6, but routine exercises without answers; EDUMATH writes for grades 3–5, routine, with teachers judging; Beast Academy is olympiad-style for ages 6–13, written by people; MasterOlympiad, before launch, is olympiad-style for classes 1–12, written by its team; Spiral Learning serves olympiad problems to ages 7–13 from an adaptive bank whose origin its page does not state; MathForces generates olympiad quizzes with Gemini, for an audience its page does not state, with no check described.
- *Status:* new among the works read and the searches made (section 4). P037 holds within those searches, for grades 1–6, not only 1–4.

**N3. A learner model kept outside the chat's language model, inside the chat, and a rule that proposes each task.** The model writes and talks. A deterministic rule proposes each task [C018, C097], and the model may depart from it only with a stated reason, which the service records [C083]. The service seals the answer in the file and in the tools' results until the child answers [C033] and never shows it on the card, though it reaches the card's memory [C096]. The model wrote the key and knows it [C064]; in the conversation only its instructions keep it back.
- *Nearest:* Tutor MCP — the same split of AI-written content and rule-driven progression over MCP, for any subject, with no task check and no sealing described; Khan Academy in ChatGPT — checked practice inside the chat, from a bank written by people, for teachers, answers on request; the assistants' study modes — in the chat, with neither a learner model nor a check described. Hooshyar et al. recommend this division of labour.
- *Status:* the split itself is not new (Tutor MCP, 2026); its combination with N1 and N2 is.

**N4. A hierarchical Elo learner model on a designed ladder of difficulty across grades 1–6, for tasks used once.** [C009, C010, C013]
- *Nearest:* Pelánek et al. (2017) — the same hierarchical Elo with a guessing floor and a shrinking step, but item difficulty learned from many learners; Math Garden — Elo for children with both ratings learned.
- *Status:* the model is not new and paper A names its lineage (Q69). What is MathTrail's own is setting difficulty by design because no task recurs; paper A presents that as a consequence of the design, not as a claim of novelty in modelling. The trial series [C080] places a learner by the most likely level given a prior, an ordinary estimate whose literature (computerised adaptive testing) was not read here, so no novelty is claimed for it.

**N5. Near-duplicates against each child's history without keeping the text.** [C028, C029]
- *Nearest:* Broder's resemblance with min-wise sketches; Lee et al.'s deduplication of corpora.
- *Status:* not new as a method. Keeping only sketches in the parent's own file is a detail of privacy by design, not a claim of novelty.

**What paper A does not claim as new:** the response model, Elo, the corridor, a run of correct answers as mastery, MinHash, sandboxed interpretation, MCP Apps cards, OAuth, a closed list of named mistakes behind wrong options (Eedi's diagnostic questions), or a learning effect — none was measured. The contributions of the introduction describe what the paper presents, including the trial series and the near-duplicate check; the claims of novelty are N1 and N2.

## 3. Learning goal → mechanism, against how others do it

| Goal | MathTrail's mechanism | How others do it |
|---|---|---|
| A task the child has not seen | Writes each task anew; refuses a near-copy of a reference task or of the child's recent tasks, by text [C027–C029] | A finite bank written by people (Beast Academy, Khan, Math Garden, MasterOlympiad); an adaptive bank of unstated origin (Spiral Learning); generated without a described check (MathForces, Strive). No work read checks novelty of the idea for a learner |
| A little harder than the last | A corridor of 70–85 % predicted success, aimed at 77.5 %, on a hierarchical Elo model [C014] | Math Garden aims at .75 with item ratings learned; Pelánek et al. found 65 % suitable in geography, while users in schools preferred about 75 %; the 85 % rule supports a band |
| Still solvable, with a right key | The model's own program, run by the service against the five options, twice [C023] | Teachers judge (MATHWELL, EDUMATH); a tutor approves each message (LearnLM with Eedi); a critic model checks replies (Oreopoulos et al.); other models vote (CHASE) |
| The answer not given away | Sealed in the file and in the tools' results until the child answers [C033]; never shown on the card, though it reaches the card's memory [C064, C096]. In the conversation the model, which wrote the key, is told to keep it back — an instruction, as elsewhere | An instruction to the model (the study modes; Bastani et al.'s tutor), which can be got round (Gupta et al.) |
| Learning from the mistake | Every wrong option names a trap, explained after the answer [C072, C077] | Eedi's diagnostic questions, written by people, each wrong option a named misconception; DiVERT writes error texts with a model |
| Knowing when a topic is mastered | At least five answers in the topic and at least three qualifying answers since the last wrong one — correct, unaided, on tasks at P ≤ 0.775; lost after two wrong in a row; held at a level [C015, C082] | N in a row (three is a value in use), a BKT probability of 0.95; a streak alone did not improve delayed learning (Oreopoulos et al.) |
| Readable for the child's level | Sentence length, and Flesch–Kincaid in English, by the task's level [C030] | Flesch–Kincaid as a target when re-levelling text (Learn Your Way); grade relevance judged by people (MathWiz) |
| Nothing about the child kept | A pseudonym, a profile in the parent's Drive, no content in logs [C069, C078, C036] | Accounts on the vendor's platform (Khanmigo, Beast Academy), not compared further here |

## 4. The systematic check of "barely studied" (P037)

The prototype's research note found no work on generating olympiad problems for grades 1–4. Checked here in three ways, all logged in `search-log.tsv`:
- **S31's searches** for the generation of olympiad-style or non-routine problems for primary school (two in OpenAlex, one on the web) found none; the olympiad-level generators they surfaced synthesise training data for models.
- **S35's searches** (five in OpenAlex in English: olympiad problems and primary school, mathematical olympiad and elementary students, non-routine problem generation, generated problems verified by executing code, Kangaroo problems and generation) found benchmarks that test models on olympiad problems, surveys, and code-verification methods for solving, but no work that writes olympiad-style problems for primary school. A sixth, in Russian, returned nothing: OpenAlex covers Russian methodical literature poorly, so the Russian tradition was not searched here (S10's Russian searches found programme descriptions without a controlled evaluation; eLibrary and CyberLeninka were not searched).
- **Forward citations** of the nearest works, in OpenAlex: MATHWELL is cited by five works, and its arXiv record by one more, a multi-objective problem generator with retrieval (2025), left as a lead because its record names no problems for children; EDUMATH is cited by one, MathWiz by none. Among them only a systematic review of the automatic generation of word problems (Utami and Hwang, IJSME 2026) could bear on the claim, and it could not be read: it is closed, and the publisher refused automated readers.
- **S31's open leads** were read from their abstracts: PromptCoT synthesises olympiad-level problems to train models; an EMNLP 2025 framework improves models at generating problems by synthetic data; a study of interestingness compares models' and people's judgements of problems. None writes problems for children.

**Verdict.** Within these searches, P037 holds and widens: no work read or found generates olympiad-style problems for grades 1–6, checked or not. Among products, MathForces generates olympiad quizzes but does not state its audience, and the topics its page lists run from linear to functional equations; Spiral Learning serves ages 7–13 from a bank whose origin its page does not state. Neither describes a check of its keys. The check is bounded by the search engines' coverage — OpenAlex finds few citing works of papers from 2025–2026 — by the one review not read, and by the Russian tradition, not searched beyond OpenAlex. Paper A states the gap as "among the works we found", not as a fact about the field.

## 5. What this changes in paper A

Made in the draft with this task, in both languages:
- The related work names Tutor MCP as the nearest open MCP tutor that separates written content from rule-driven progression, so the novelty is stated against it.
- The learner model's contribution names its lineage, the hierarchical Elo of adaptive practice, and what is the paper's own: difficulty set by design for tasks used once.
- The related work no longer says the answer is kept out of the language model's hands: the model wrote it. The service keeps the learner model out of its hands; a rule proposes each task, which the model may override only with a stated reason; and the answer is never shown on the child's card before the child answers.

Left to the tasks that write those sections:
- S60 (introduction) states N1 and N2 as the paper's contribution, with "among the works we found"; S63 (discussion) names N3's nearest neighbour and N4's lineage.

## Remarks

- **Not read.** The systematic review of word-problem generation (Utami and Hwang, IJSME 2026, 10.1007/s10763-026-10715-6) is closed, and the author cannot obtain it (2026-10-01). Paper A keeps "among the works we found" and names the unread review among the limits of its search (S62).
- **Not covered.** Computerised adaptive testing, whose placement procedures the trial series resembles; commercial products whose pages do not describe how their content is made or checked.
- **Dates.** Tutor MCP's repository was created on 2026-03-30 and KMath had not launched on 2026-09-30; the product landscape moves faster than the literature. S62 checked both again on 2026-10-01: KMath still announced its opening for October 2026, and Tutor MCP, pushed that day, still described no check of a written task and kept both quoted passages (`search-log.tsv`). Both are to be checked once more before submission.
