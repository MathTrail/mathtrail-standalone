# Writing standard

Task S01 of [RUN.md](../RUN.md). This is the rubric every draft section is judged against before it is handed to the author (S56–S76), and again in the internal reviews (S64, S75). It combines three things:

- the criteria of the target venues ([venues.md](venues.md));
- a small set of widely used guides to writing research papers;
- the program's own rules, G1–G10.

Every guide cited here was read on 2026-09-25, and its metadata was taken from Crossref. `citecheck` re-checks them once S08 exists.

## 1. What the reviewers of the target venues ask

- **AIED** (Springer):
  - three PC reviews and a meta-review, double-blind;
  - criteria: relevance, novelty, technical soundness, significance, clarity of presentation;
  - "responsible reporting": the composition of human-sourced data, demographic imbalances and their impact, awareness of ethical issues ([call](https://aied-conference.org/2026/call-for-paper)).
- **L@S** (ACM): "originality, research quality, potential impact, and value to the development of future learning at scale" ([call](https://learningatscale.hosting.acm.org/las2026/call-for-papers/)).
- **EDM:** the call publishes no formal criteria; a full paper must be "original, substantive, mature, and unpublished work" ([call](https://educationaldatamining.org/edm2026/call-for-papers/)).
- **IJAIED:** not checked; confirmed before a long version targets it.
- **EC-TEL** (Springer LNCS): double-blind, no URLs to one's own systems; the call encourages pre-registration and shared data and code ([call](https://ea-tel.eu/ectel2026/cfp)). Its review criteria are on a separate page, confirmed before a draft of paper B is judged.

A draft is checked against the criteria of the venue it targets (Q28); a section that does not help answer one of them is cut.

## 2. How a paper is built

Rules 1, 3 and 4 follow Mensh and Kording, *Ten simple rules for structuring papers*, PLOS Computational Biology 13(9), 2017, doi:10.1371/journal.pcbi.1005619. Rule 5 follows Lipton and Steinhardt, *Troubling trends in machine learning scholarship*, ACM Queue 17(1), 2019, doi:10.1145/3317287.3328534 (also arXiv:1807.03341). Rules 2 and 6 are the program's own.

The rules:

1. **One central contribution, communicated in the title.** "Papers that simultaneously focus on multiple contributions tend to be less convincing about each." Everything else supports it or goes.
2. **Contributions stated as a list** in the introduction, each one checkable and each tied to the section that backs it up.
3. **Context, content, conclusion** — at every scale. The introduction sets the context, the results are the content, the discussion brings home the conclusion; each paragraph does the same in small. The abstract tells the complete story: context narrowing to a gap, what was done and found, and why it matters. The results are a sequence of statements, each supported by a figure or table. The discussion says how the gap was filled, the limits of the interpretation, and the relevance to the field.
4. **Written for readers who do not know the work.** Terms are defined where first used; an example comes before the general statement.
5. **The four troubling trends to avoid** (Lipton and Steinhardt):
   - explanation mixed with speculation — speculation is labelled as such;
   - gains without their source — the ablations show which part does the work;
   - mathiness — formulas that decorate rather than define or prove (rule G4);
   - misused language — suggestive names, overloaded technical terms, suitcase words. This is why Gemini's lexicon is filtered through the term map of S04.
6. **Limitations are a section, not a sentence.** Threats to validity are named by kind:
   - internal: could something other than the claimed cause explain the result?
   - external: does it hold beyond these tasks, models and settings?
   - construct: does the measure capture what it claims?
   - conclusion: are the statistics adequate?

## 3. Reporting checklist

Adapted from the NeurIPS paper checklist, read on 2026-09-25 ([checklist](https://neurips.cc/public/guides/PaperChecklist)), and AIED's responsible-reporting rules. The checklist's sixteen headings are claims, limitations, theory, reproducibility, open data and code, experimental details, statistical significance, compute, code of ethics, broader impacts, safeguards, licenses, assets, human subjects, IRB approvals, and declaration of LLM usage. Each item is answered in the paper or its appendix, or marked as not applicable with a reason.

- **Claims.** The abstract and introduction claim only what the results show.
- **Limitations** are stated (see §2.6).
- **Theory:** every proposition has its assumptions and a proof or a proof sketch.
- **Reproducibility:**
  - the code, data and seeds that produce every number;
  - the exact models and versions used, including the model each CLI actually served (S07).
- **Statistics:**
  - intervals, not bare point estimates;
  - how they were computed;
  - what the unit of analysis is.
- **Compute and cost:** what was run and on what, including subscription calls.
- **Human participants:** who took part (the author, teachers), what they did, and consent. No children (rule G5).
- **Licenses** of the data, code and models used, and the licence of what is released (MIT).
- **Ethics and broader impact:** children, privacy, the use of chat assistants by children, who pays for inference.
- **AI assistance:**
  - disclosed in the form the publisher asks for (§5);
  - consistent with [ai-use-log.md](../ai-use-log.md).

## 4. The program's own rules, as a check

For every section, before it goes to the author:

- **Claims:** every claim about MathTrail or the prototype has a ledger entry (G1). Nothing *specified* or *planned* is described as *built*.
- **Numbers:** every number comes from a generated macro (G2).
- **Citations:**
  - every reference passes `citecheck`, and the notes keep the passage it rests on (G3);
  - no reference is cited that was not read.
- **Formalism:** every formula, definition and proposition carries content (G4).
- **Anonymity (G8):** the anonymous build contains no "MathTrail", no link to `github.com/MathTrail/…` (product or prototype), the author's own work only in the third person, and an anonymised artifact link.
- **Negative and null results** from the protocols are reported where they belong (G9).
- **Blinding:** analysis choices were fixed before the model families were unblinded (G10).
- **Contributions:** a CRediT statement names what each author did (G7), and the AI disclosure matches the log.
- **Product boundary (G6):** a change the research proposes to the product is described as a proposal, never as done.

## 5. Disclosure of AI assistance: wording

The final text is written from the log. The templates below are starting points.

- **Springer venues** (in Methods): "Large language models were used in preparing this work: [tool, model], for [literature search, drafting of sections X and Y, code for experiments Z]. All text, code and results were checked by the author, who takes responsibility for the content. The prompts and the log of AI use are available in the artifact."
- **ACM venues** (in Acknowledgements): "Generative AI tools ([tool, model]) were used to [draft sections, write code, search the literature]. The author reviewed and takes responsibility for all content, including every reference."
- **IEEE venues** (in Acknowledgements): "[Tool, model] generated content in sections [X, Y] ([what: text, code, figures]) at the level of [drafting / editing]. The author reviewed all of it and is responsible for the article's content."

In the review version, "the artifact" is an anonymised link; the public repository is named only in the final version.
