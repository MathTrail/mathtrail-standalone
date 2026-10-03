# Venues, deadlines and policies

Task S01 of [RUN.md](../RUN.md). Every fact below carries its source; all were checked on **2026-09-25**. None of the 2027 editions had published a call for papers by then, so their dates are **expected** ones, inferred from the 2026 edition of the same venue and marked that way. A cell marked "not checked" was not confirmed from a source. Each writing task confirms its venue's facts on the official site before relying on them.

## Venues recommended for each paper

### Paper A (the system)

| Venue | Format | Review | Other rules | 2026 dates (source) | 2027 (expected) |
|---|---|---|---|---|---|
| **AIED**, main track, full paper | 14 pages including references; Springer LNAI, a subseries of LNCS | Double-blind: three programme-committee reviews and a senior meta-review. Criteria: relevance, novelty, technical soundness, significance, clarity. Anonymisation removes "names of approaches, frameworks, projects and/or systems" | Preprints allowed "provided that they follow Springer's policies"; work must not be under consideration elsewhere; responsible reporting of human data and ethics | Abstract 26 Jan 2026, paper 2 Feb 2026, notification 23 Mar 2026, all AoE; Seoul ([dates](https://www.aied-conference.org/2026/call-for-paper/important-dates), [call](https://aied-conference.org/2026/call-for-paper)) | Abstract ≈ late January, paper ≈ early February 2027 |
| **EDM**, full paper | 10 pages; references and acknowledgements not counted ([instructions](https://educationaldatamining.org/edm2026/instructions-for-authors/)) | Double-blind | "All papers must not have been submitted for publication at other venues" | Abstract 2 Feb 2026, paper 9 Feb 2026; JEDM journal track with rolling deadlines ([dates](https://educationaldatamining.org/edm2026/important-dates/), [call](https://educationaldatamining.org/edm2026/call-for-papers/)) | ≈ early February 2027 |
| **L@S**, research paper | Up to 10 pages, references excluded; ACM two-column | Double-blind; "originality, research quality, potential impact, and value to the development of future learning at scale" | "Authors are permitted to post their work to public archives (e.g., arXiv, EdArXiv)" | Abstract 9 Feb 2026, paper 16 Feb 2026; no extensions ([call](https://learningatscale.hosting.acm.org/las2026/call-for-papers/)) | ≈ mid-February 2027 |

EDM and L@S close in the same weeks as AIED, **before** AIED announces its decisions. They are therefore alternatives, chosen at submission time — not fallbacks after a rejection. If AIED rejects paper A (notification ≈ late March 2027), the routes are:
- **IJAIED**, a journal;
- **AIED 2028**;
- **EC-TEL 2027**, only if its 2027 abstract deadline falls after AIED's notification. In 2026 it did not: EC-TEL abstracts closed 13–20 March, AIED notified on 23 March. Registering an EC-TEL abstract while the paper is under AIED review would break AIED's rule that the work is not "under consideration elsewhere".

### Paper B (the evolving textbook)

| Venue | Format and review | Other rules | 2026 dates (source) | 2027 (expected) |
|---|---|---|---|---|
| **EC-TEL**, full research paper | 8–15 pages including references; Springer LNCS. Double-blind, with "no reference to themselves and their institutions; without any URLs to projects, products or self-developed systems". Review criteria are on a separate page, not checked | Recommends Open Science practice: pre-registration of studies and shared data and code — which rule G9 already does | Abstract 13–20 Mar 2026, paper 3–12 Apr 2026; Valencia, 14–18 Sep 2026 ([call](https://ea-tel.eu/ectel2026/cfp)) | Abstract ≈ March, paper ≈ April 2027 |
| **IJAIED** (journal) | Springer; review model not checked (the [guidelines](https://link.springer.com/journal/40593/submission-guidelines) only say what to do *if* it is double-anonymous) | — | Rolling | Rolling |

**The iTextbooks workshop at AIED** (seventh edition in 2026) is the closest community. Its topics include "LLM-driven extraction/generation of learning material" and "mining learner interaction logs". Full papers are up to 12 pages, CEUR, one column; the 2026 deadline was 22 May, the workshop 28 June ([workshop](https://intextbooks.science.uu.nl/workshop2026/)).

In 2027 it falls after the EC-TEL deadline, so it cannot improve the EC-TEL submission. For the same work it risks a dual submission. The rules that would decide are EC-TEL's (not found in the call) and the workshop's own call (not checked). It suits a **distinct, non-overlapping** contribution — for example the heuristic catalog and the trap → heuristic map (S18) as a resource — once those two rules are checked.

### Considered and set aside

Named in the plan; no facts about them were checked, because none is a target.

| Venue | Why it is not a target now |
|---|---|
| AIED Practitioners, Industry and Policy track (2026 deadline 15 Feb — [dates](https://www.aied-conference.org/2026/call-for-paper/important-dates)) | A system paper with experiments belongs in the main track |
| AIED Blue Sky (2026: abstract 27 Feb, paper 5 Mar — [dates](https://www.aied-conference.org/2026/call-for-paper/important-dates)) | For visions; paper B carries results, as the author decided when the plan was made (design, formal model and experiments) |
| AIED, EDM, L@S main tracks for paper B | Their deadlines (February) come before B's experiments can finish |
| ITS, LAK | Not checked; ITS is close in topic to A and is the next alternative if the AIED season is missed; LAK is about learning analytics rather than the system |
| IEEE TLT, Computers & Education: AI, ACM TOCE | Journals: routes for extended versions after the conferences, not first targets |
| NeurIPS Evaluations & Datasets track (2026: abstract 4 May, paper 6 May — [call](https://neurips.cc/Conferences/2026/CallForEvaluationsDatasets)) | Only if the reference corpus is released as a benchmark; decided after S65 |
| NeurIPS and ICLR workshops | Not checked; a workshop alone does not satisfy arXiv's rule for position papers, and both papers aim at archival venues |

## Policies that apply to every paper

- **Use of AI in writing.**
  - *Springer* (AIED, EC-TEL, IJAIED): large language models do not meet the authorship criteria. Their use is documented in the Methods section. "AI assisted copy editing" of human-written text needs no declaration. Humans are accountable for the final text ([Springer Nature AI guidance](https://group.springernature.com/gp/group/ai/ai-guidance-for-our-researchers-and-communities)).
  - *ACM* (L@S): generative AI may be used but "must be fully disclosed", for example in the Acknowledgements. AI tools may not be authors. Authors answer for fabricated content or references. ACM's own policy page returned HTTP 403 on the check date, so the policy was taken from its text as quoted by an ACM conference ([SIGCSE Virtual 2026](https://sigcsevirtual2026.acm.org/info/policies-on-generative-ai,-llms,-and-related-tools)). Re-check before submitting to an ACM venue.
  - *IEEE* (TLT): AI-generated content is disclosed in the Acknowledgements, naming the system and the sections that use it, with how it was used. Editing and grammar help falls outside the rule, but disclosure is still recommended ([IEEE author guidelines](https://open.ieee.org/author-guidelines-for-artificial-intelligence-ai-generated-text/)).

  The papers disclose from [ai-use-log.md](../ai-use-log.md) (rule G7).
- **arXiv.**
  - Since 31 Oct 2025, review and position papers in the CS category must first be accepted after peer review, and a workshop alone is generally not enough ([arXiv blog](https://blog.arxiv.org/2025/10/31/attention-authors-updated-practice-for-review-articles-and-position-papers-in-arxiv-cs-category/)). Both papers carry new results, so both are research papers.
  - Since 21 Jan 2026, a first-time author needs personal endorsement from an established arXiv author; an institutional address no longer suffices ([arXiv blog](https://blog.arxiv.org/2026/01/21/attention-authors-updated-endorsement-policy/)). This is critical question K09.
- **Anonymity and the public repositories.** AIED and EC-TEL both forbid naming or linking one's own system in the review version. The anonymous build therefore contains:
  - no "MathTrail";
  - no link to `github.com/MathTrail/…` — neither the product nor the prototype;
  - the author's own work cited in the third person;
  - an anonymised artifact link.

  Public preprints are allowed by AIED and L@S, so for paper A drafts in the public repository are no worse than a preprint: Q05's default stands. EC-TEL's preprint rule was not found. Until S55 finds it, paper B's drafts stay under `.gitignore` (Q29).
- **Responsible reporting** (AIED): describe any human-sourced data, its demographics and ethical issues. Paper A has no child data (G5); its human data is the author's own blind review (S44) and, optionally, teachers (S46).
- **Tools Competition 2027:** abstract due 13 Oct 2026, Phase II due 21 Jan 2027, pitches in April, winners notified in early May and announced publicly in June 2027 ([tools-competition.org](https://tools-competition.org/)). A preprint of paper A before 21 Jan 2027 would strengthen the application, but only if A is ready by mid-January and K09 (arXiv endorsement) is settled. It is an option, not a plan.

## Recommendation and schedules

**Recommendation (Q28):**
- **Paper A → AIED 2027**, main track, full paper. EDM or L@S of the same season are alternatives, chosen at submission. After a rejection: IJAIED or AIED 2028; EC-TEL 2027 only if its deadline falls after AIED's decisions.
- **Paper B → EC-TEL 2027**, full paper. The long versions of both papers can go to IJAIED.
- **iTextbooks 2027** only for a distinct contribution, after checking dual-submission rules.
- **The corpus as a benchmark** is decided after S65.

**Schedule "A first"** (the default, Q08). The task sets come from «Почему такой порядок» in RUN.md. One executor runs tasks one after another, so "in parallel" means interleaved. The run follows the table's order. If S31–S39 are not done by **31 October 2026**, the executor switches to this schedule's order, as Q08 allows, and records it in the log (Q28).

| When (expected) | Milestone | Tasks |
|---|---|---|
| October 2026 | Phase 0; what paper A needs from phase 1; literature for A; offline experiments | S00–S10, S16, S18–S19, S31–S39 |
| October–November 2026 | Generation study; needs the author (K01, K02, K04). Teachers' review, if K06 finds them | S40–S44 (S46 optional) |
| December 2026 – mid-January 2027 | Paper A sections, internal review, data release | S55–S65 |
| Late January – early February 2027 | AIED 2027 abstract and paper | S66 |
| After T62–T64 | Live evidence, if it arrives in time | S45 (optional) |
| In parallel from October 2026 | Textbook design, then the experiments for B | S11–S15, S17, S20–S30, S47–S54 |
| March–April 2027 | EC-TEL 2027 | S67–S76 |

**Schedule "B first"** — only if the generation study for A (K01, K02) cannot run before December 2026:

| When (expected) | Milestone | Tasks |
|---|---|---|
| October–November 2026 | Phase 0; the textbook's literature and design; the offline part of A's experiments | S00–S30, S36, S38 |
| December 2026 – February 2027 | Experiments for B; the judge panel needs the author (K01, K02) | S47–S54 |
| February – early April 2027 | Paper B | S55, S67–S76 |
| Early April 2027 | EC-TEL 2027 | S76 |
| From April 2027 | Paper A: the rest of its experiments and sections | S37, S39–S44, S56–S66 |
| Summer 2027 onwards | Paper A to IJAIED, or AIED 2028 (≈ February 2028) | S66 |
