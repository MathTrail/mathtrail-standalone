# S64: meta-review and what was done with each finding

Task S64 of [RUN.md](../../RUN.md). Three internal reviews of the anonymous build of 2026-10-02, each by a separate session of the executor's own model in the role of an AIED programme-committee member: [the methodologist](S64-methodologist.md), [the AIED specialist](S64-aied-specialist.md) and [the skeptic](S64-skeptic.md). Rule G10 wants reviewers of other model families; those passes wait for the author (K01, Q85).

## Verdict

Weak reject, weak reject and reject. The three agree on what would sink the paper. Every finding was checked against the paper's text, the results files, the protocol and the product at the pin before it was acted on.

1. **RQ1's headline holds by construction.** The operators that break a rule a check states were written from the checks' own rules, so their refusal is a test that the code does what it says, and the abstract led with it.
2. **RQ2 was reported selectively.** Of nine generators the results named three, and the protocol's primary comparisons 2 and 4 were not reported in full, although the protocol says every comparison is reported whichever way it falls (§2.7). Where the child changes — learning, a jump, a harder host, another slope — Glicko-2 with one level per child was more accurate than the service, and the paper did not say so.
3. **"Verified" promises more than the checks give.** The solver is the same model's formalisation of the task; the title, Table 1 and the conclusion read as if the key were checked independently.

## What was changed, finding by finding

| Finding | Reviews | Decision | Where |
|---|---|---|---|
| RQ1 by construction | M-W2, A-W1, S-W1 | Fixed: the abstract calls the refusals what a correct implementation must do and names what passed beside them; §6.1 calls the 100 % a test of conformance, and says that the self-check was held at "nothing found" and the solvers were the reference ones, so E-A1 cannot see a model's solver that shares its key's error | abstract, §6.1 |
| RQ2 selective; comparisons 2 and 4 | M-W1, S-W3 | Fixed: comparison 2 against every baseline, comparison 4 in full, the other generators summarised, calibration on learners, and the abstract and §8 say where the baselines did better | abstract, §6.2, §8 |
| "Verified" in the title, "A key to trust", "confirmed the key twice" | S-W2, A-W1 | Fixed: the title drops "Verified"; the goals say "a checked key"; §8 says the program agreed with the key under both labellings, and that the solver is the model's own formalisation | title, §3, §8 |
| The live run's refusal was a false alarm; two hints nearly gave the answer away | S-W2 | Fixed: one sentence in §4 (new ledger claim C103) | §4 |
| The checks refuse 156 of the 603 reference tasks; the reference copies are English only | S-W5, A-W1, M-Q3 | Fixed: the counts by level and the stock openings in §6.1; the English-only reference check in §7 (new ledger claim C102) | §6.1, §7 |
| D14b under 95 % | S-W7 | Fixed: named as not reliably caught, as the protocol's rule asks (§1.9) | §6.1 |
| Table 3's pooled rows X01 and X04 with exact intervals | M-W6 | Fixed: each of the two classes is shown by its operators, whose cases are independent across hosts, each with its exact interval, as the protocol prescribes for an operator | Table 2, the former Table 3 |
| Deviations from the protocol unmentioned | M-W8, S-W7 | Fixed: one sentence says that the deviations are recorded with the artifact | §6 |
| X03's label | M-W7 | Fixed: "a word swapped for its opposite", which is what the operator does | §6.1, Table 2 |
| The trial series judged where its spread was tuned | M-W5, S-W3 | Fixed: one clause says so, with Glicko-2 from the first answer beside it, and that the longest run of wrong answers did not change | §6.2 |
| False mastery without its base rate | M-W4, A-W3 | Fixed: the 0.70 reading and the share never declared | §6.2 |
| "No difference found" read as "as well as" | M-W6 | Fixed: the paired interval is printed | §6.2 |
| Nearest work cut: Ikram et al. (AIED 2026), MathQ-Verify | A-W2, S-W8 | Fixed for the venue's own nearest work: one sentence and Ikram et al. back. MathQ-Verify stays in the extended version: its reference costs five lines the page limit does not have | §2 |
| Ethics: hosts that exclude children | A-W4, S-W6 | Fixed: the adult beside the child as the design's condition, the two flows named, and whether COPPA or the GDPR applies left open as a legal question | §8 |
| Utami and Hwang cited unread | A (budget) | Fixed: the clause and the reference are cut; the search threat stays | §6.3 |
| "22 languages but English and Russian" | S-W9 | Fixed: "most of the card's 22 languages", which names neither and needs no number typed by hand | §7 |
| Fourteen alternatives, eleven named | A-W7 | Fixed: Glicko-2, its floored variant and Urnings each run per child and per topic | §6.2 |
| Abstract over 250 words | A-W7, venue | Fixed: trimmed to the LNCS range | abstract |
| Provenance of the reference tasks | S-W9 | Dropped: the author's decision, Q82, made on the same records | — |
| AI assistance not disclosed | A-W6, S-W4 | Decided by the author (Q86): no disclosure; the reviewer's suggestion of a line in the camera-ready acknowledgements is declined | — |
| Unread AI verdicts on the sampled cases | M-W7, S-W4, A-W6 | Resolved (Q86): the author does not read them, so the sentence and its TBD are cut; the refusal rates of Table 2 stand on their own, and the threat that one reader read the cases goes with them | §6.1, §6.3 |
| No real model output judged; a blind rating of real tasks | A-W1, S-Q1 | Declined for this version by the author (Q86); the paper names it as unmeasured | — |
| Automatic item generation, EduAgentQG, MathCAMPS | A-W2 | Open: none is read yet, and G3 cites nothing unread; a reading task decides whether they enter | — |
| Bonferroni count of the found differences | M-W6 | Open: needs a computation the simulation does not yet write | — |
| The corridor an oracle would reach | M-W3, S-Q3, A-Q2 | Open: needs a cell the protocol did not run | — |
| E-A4's worst case and tokens | M-W9 | Open, minor | — |
| Published versions of arXiv citations | A-W7 | Open, minor: the conference versions have no Crossref record to check against | — |
| Replace Figure 2 by a table; cut Table 1 or §5.2 | M-W1, A, S-W3 | Partly: Table 1 is folded into one sentence of §3, which pays for the fixes; Figure 2 and §5.2 stay, since they carry contributions, and the generators are summarised in the text | §3 |

## What made room

The fixes added about a page, and the paper has to stay at fourteen. Besides Table 1, these went: the product's counts of tests, the sketch's agreement with the exact index, the step's excess at P = 0.5, the review's passes and allocation, the package's share of its budget, the 30 seconds of the first local run, the clause on the review that could not be obtained with its reference, and a sentence of §2 on MCP Apps, whose reference moved into §3. The extended version keeps them all. The anonymous build is fourteen pages with 29 references and no overfull box.
