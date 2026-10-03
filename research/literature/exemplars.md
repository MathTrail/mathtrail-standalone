# Exemplar papers

Task S06 of [RUN.md](../RUN.md): read the best papers of the same kind before writing, and build the skeletons of both papers from them.

**How they were read, and what that allows.**
- **The texts.** Everything was read on 2026-09-25 through a tool that summarises a page with a small model: the full text on arXiv or PubMed Central, or the text of a PDF extracted locally. Three entries are known only from their abstract or metadata — E4 and E7 because the full text sits behind a paywall, E6 because its text could not be extracted — and each says so.
- **Structure and lengths** come from those summaries and are approximate. They give proportions, not page budgets.
- **Quotes** are as the summaries gave them. They are not citations: any paper that uses one re-reads the original and keeps the passage in its notes (G3).
- **Metadata** comes from arXiv, Crossref or DataCite, and page ranges Crossref lacks from the article's own first page. No entry has passed `citecheck` yet (S08).
- **Numbers** quoted from an exemplar are not findings of the literature review. They are checked where a paper relies on them (S09–S16, S31–S34).

## The exemplars

| # | Work | Venue | Kind | What it shows us |
|---|---|---|---|---|
| E1 | Gupta, Reddig, Calo, Weitekamp, MacLellan. *Beyond Final Answers: Evaluating Large Language Models for Math Tutoring* | AIED 2025, LNCS, pp. 323–337, doi:10.1007/978-3-031-98414-3_23; arXiv:2503.16460 | evaluation of LLMs in maths education | The target venue and genre of paper A |
| E2 | Christ, Kropko, Hartvigsen. *MATHWELL: Generating Educational Math Word Problems Using Teacher Annotations* | Findings of EMNLP 2024; arXiv:2402.15861 | generation with expert annotation | How generated problems are judged: solvability, accuracy, appropriateness, "meets all criteria" |
| E3 | Christ, Molitz, LeBlond, Gottesman, Kropko, Hartvigsen. *EDUMATH: Generating Standards-aligned Educational Math Word Problems* | ACL 2026; arXiv:2510.06965 | generation with teachers and a study with pupils | A study with children in the same genre, and what it reports on ethics |
| E4 | Klinkenberg, Straatemeier, van der Maas. *Computer adaptive practice of Maths ability using a new item response model for on the fly ability and difficulty estimation* | Computers & Education 57(2), 1813–1824, 2011, doi:10.1016/j.compedu.2011.02.003 | adaptive practice on an Elo-style model | Abstract only (paywall): the model and the 0.75 target |
| E5 | Brinkhuis, Savi, Hofman, Coomans, van der Maas, Maris. *Learning as It Happens: A Decade of Analyzing and Shaping a Large-Scale Online Learning System* | Journal of Learning Analytics 5(2), 29–46, 2018, doi:10.18608/jla.2018.52.3 | a system at scale | How an adaptive system reports fit and misfit honestly |
| E6 | Pelánek. *Applications of the Elo Rating System in Adaptive Educational Systems* | Computers & Education 98, 169–179, 2016, doi:10.1016/j.compedu.2016.03.017 | review with a case study | Abstract only: its text could not be extracted |
| E7 | Koedinger, Stamper, McLaughlin, Nixon. *Using Data-Driven Discovery of Better Student Models to Improve Student Learning* | AIED 2013, LNCS, pp. 421–430, doi:10.1007/978-3-642-39112-5_43 | closing the loop | Metadata only (paywall); E8 tells the same loop at length |
| E8 | Liu, Koedinger. *Closing the Loop: Automated Data-Driven Cognitive Model Discoveries Lead to Improved Instruction and Learning Gains* | Journal of Educational Data Mining 9(1), from p. 25, 2017; doi:10.5281/zenodo.3554625 | closing the loop | The loop paper B formalises, with its randomised test |
| E9 | Williams, Kim, Rafferty, Maldonado, Gajos, Lasecki, Heffernan. *AXIS: Generating Explanations at Scale with Learnersourcing and Machine Learning* | L@S 2016, pp. 379–388, doi:10.1145/2876034.2876042 | learning content improved by learners' data | A bandit choosing among explanations, tested against an instructor's |
| E10 | LearnLM Team, Google. *Towards an AI-Augmented Textbook* | arXiv:2509.13348, 2025 | an AI textbook with a trial | Expert rubrics plus a randomised trial, and what it leaves untested |
| E11 | Romera-Paredes et al. *Mathematical discoveries from program search with large language models* (FunSearch) | Nature 625, 468–475, 2024, doi:10.1038/s41586-023-06924-6 | evolutionary search with an LLM | Why an evaluator is the heart of such a loop, and what the method needs |
| E12 | Novikov et al. *AlphaEvolve: A Coding Agent for Scientific and Algorithmic Discovery* | white paper, arXiv:2506.13131, 2025 | evolutionary search with an LLM | Ablations of each part of the loop; a cascade of evaluations |
| E13 | Verga et al. *Replacing Judges with Juries: Evaluating LLM Generations with a Panel of Diverse Models* | arXiv:2404.18796, 2024 | a panel of LLM judges | How a panel is compared with a single judge and with people |
| E14 | Kim, Garg, Peng, Garg. *Correlated Errors in Large Language Models* | ICML 2025; arXiv:2506.07962 | correlated errors of LLMs | The measure of correlated errors, and what it does to an LLM judge |
| E15 | Sosnovsky, Brusilovsky, Lan. *Intelligent Textbooks* | IJAIED 35(3), 967–986, 2025, doi:10.1007/s40593-024-00451-9 | editorial of a special issue | The field paper B joins, and its generations |

## What each shows

Each entry gives, where the text was read: the shape, the contributions, the figures and tables, the evaluation, and the limitations or threats to validity.

### E1. Beyond Final Answers (AIED 2025) — the genre of paper A

- **Shape:** pp. 323–337 in the proceedings. AIED 2025's own page limit was not checked; AIED 2026's is 14 pages including references ([venues.md](venues.md)). The summary's section lengths are rough and sum to about 10 pages, so they give proportions:
  - introduction, with two research questions and four contributions: 1.5;
  - related work: 1;
  - methodology: 2.5;
  - results: 1.5;
  - discussion: 1.5;
  - limitations and future work: 0.75, as a section of its own;
  - conclusion: 0.5;
  - references: 1.
- **Contributions,** stated as a list of four: two evaluation methods, one finding, one set of guidelines. The finding is phrased as a limit: LLMs "frequently contain mistakes and inaccuracies, suggesting they are not yet ready for direct in-class deployment".
- **Figures and tables:** two workflow figures; a rubric table; two results tables; a table of themes.
- **Evaluation:**
  - 110 problems per model across 22 problem types;
  - 150 tutoring dialogues;
  - two independent reviewers, with Cohen's κ of about 0.85 for quality and 0.82 for correctness;
  - an automated judge whose disagreements were checked by hand.
- **Limitations:** no real students, one platform, one family of models, public models only.
- **What paper A takes:** the shape, research questions in the introduction, a rubric table, agreement statistics for every human judgement, and limitations as a section. Its headline — only 56.6 % of dialogues were entirely correct — is the gap paper A's verification answers; S31 checks it before A cites it.

### E2. MATHWELL (Findings of EMNLP 2024)

- **Shape:** introduction, related work, methods, evaluation, controllability, conclusions, then separate sections for limitations and ethics, and seven appendices.
- **Contributions,** stated as a list: a finding about existing data, over 5,000 teacher annotations, a dataset, and a model released with the whole human study.
- **Figures and tables:** four figures (the criteria, a baseline's results, the training pipeline, reading levels) and eight tables (dataset characteristics, a comparison of models, results by operation, errors of appropriateness, automatic metrics, controllability).
- **Evaluation:** teachers with an average of 5.75 years of experience; 250 problems per model; the criteria solvability, accuracy and appropriateness, and "meets all criteria".
- **Limitations:** no control of grade level; grades K–8 only; text only; the cost of human evaluation; the subjectivity of appropriateness.
- **What paper A takes:** "meets all criteria" as the unit of acceptance, reported with its interval; an ethics statement even without children.

### E3. EDUMATH (ACL 2026)

- **Shape:** as E2, plus a section with pupils: roughly 1,200 words of introduction, 800 of related work, 2,500 of methods, 1,500 of results and 600 of the study with pupils, then limitations and ethics as separate sections.
- **Contributions:** four bullets, the last "the first study of custom LLM-generated MWPs with grade school students".
- **Figures and tables:** two figures (the criteria, the pipeline) and three tables in the body (datasets, a comparison of nine models, results by topic).
- **Evaluation:**
  - over 11,000 problems judged by an LLM against a human-labelled subset;
  - 1,372 teachers doubly coded 3,012 problems, with κ reported;
  - 94 pupils in grades 3–5, whose accuracy and preferences were compared between human-written and generated problems.
- **Limitations:** grades 3–5 only; text only; one fixed prompt; the cost of annotation.
- **Ethics:** an institutional review board approved the annotation study with teachers; the text read here does not say how the pupils' consent was obtained.
- **What paper A takes:** the honesty of reporting modest agreement (κ = 0.34 between people) rather than hiding it.
- **To check:** the prototype's notes quote EDUMATH's 90 %, 77 % and 66 %, and the summary read here calls the same numbers agreement rates. S31 settles which they are (P039).

### E4. Math Garden, 2011 (abstract only)

- An item response model built on the Elo rating system, with a scoring rule that uses both accuracy and response time.
- Difficulty is chosen for "a mean success probability of 0.75".
- 3,648 children answered over 3.5 million problems in ten months.
- Structure, figures and limitations: not read. S32 reads the full text; its abstract supports P047's first half.

### E5. Math Garden, a decade on (JLA 2018)

- **Shape:** 18 pages.
  - 1 Introduction.
  - 2 The system: the scoring rule, adaptive item selection.
  - 3 Evaluation of fit: prediction accuracy, reliability of parallel items; then diagnosis of misfit: global response processes, local strategies, other sources.
  - 4 Discussion. 5 Conclusions.
- **Contributions:** stated as the purpose of presenting "a decade of experience with analyzing and improving an online practice environment for math", in prose rather than a list.
- **Figures and tables:** eleven figures, among them the garden's landing page, an item, the development of fit for one child, and item ratings over time; one contingency table of item pairs.
- **Limitations:** a whole section diagnoses where the model misfits and why — violations of unidimensionality, strategies, other sources — in place of a list of caveats.
- **What stands out:** children choose their difficulty — easy, about 90 % correct; medium, about 75 %; hard, about 60 %.
- **What paper A takes:** a subsection on misfit and its sources. S32 considers the three levels beside the product's single corridor.

### E6 and E7 (not read in full)

- **Pelánek 2016** is a systematic review of Elo variants for education with one case study, an adaptive practice of geography facts. S32 reads it in full.
- **Koedinger et al. 2013** is the AIED paper whose loop E8 reports at length. It is not cited unless S11 reads it in full.

### E8. Closing the loop (JEDM 2017)

- **Shape:** about 8,000 words.
  - An introduction on knowledge-component models.
  - The data-driven discovery.
  - Closing the loop: the design of the redesigned tutor, a classroom experiment, results.
  - Discussion.
- **Contributions:** in prose: to "interpret the automated improvements ..., validate the interpretation on novel data, use it to make changes to classroom-deployed educational technology, and show that the changes lead to significant learning gains relative to a control condition".
- **Figures and tables:** item performance forward and backward, pre- and post-test results by condition in a figure and a table.
- **The test:**
  - 115 students, randomised individually; 91 finished;
  - differential attrition checked with a chi-square test;
  - pre- and post-test over five school days;
  - the post-test differed by condition, F(1,89) = 5.04, p = 0.027.
- **Limitations:** in the discussion, which the extracted text did not reach.
- **What paper B takes:** the loop's steps as its skeleton — discover in data, interpret, validate on new data, redesign, test — and the attrition check.

### E9. AXIS (L@S 2016)

- **Shape:** about 7,900 words in the ACM format: the system, a case study with 150 learners recruited online, and a randomised experiment.
- **Contributions:** in prose, in the abstract: the system, the helpfulness learners find in its explanations, and learning compared with answers alone and with an instructor's explanations.
- **Figures and tables:** six figures — a problem, an explanation shown for rating, the prompt to write one, examples, the policy over the pool of explanations, and the experiment's means.
- **Mechanism:** learners write, revise and rate explanations; Thompson sampling chooses which one the next learner sees; learners' ratings are the reward.
- **Limitations:** not reached by the extracted text.
- **What paper B takes:**
  - the comparison against an expert's version, not only against nothing;
  - the risk AXIS accepts and B must not: optimising for how helpful learners rate a text is optimising for clarity alone (Q09).

### E10. Towards an AI-augmented textbook (2025)

- **Shape:** introduction; the transformations; practice and assessment; pedagogical evaluations; the efficacy study; discussion.
- **Contributions:** three sentences in prose: an approach to transforming textbooks while keeping their content, pedagogical evaluations of each transformation, and a randomised trial.
- **Figures and tables:** ten figures (the interface, the two-step generation, examples of each transformation, expert ratings, the trial's protocol and results) and one table of rubrics.
- **Evaluation:**
  - experts rated each transformation on a three-level rubric of eight criteria;
  - a randomised trial of 60 students aged 15–18: 58 took the retention test three days later;
  - the tool against a PDF reader, compared with Mann–Whitney tests.
- **Limitations it names:** which component helps is unknown; there was one chapter.
- **What it lacks:** no parental consent procedure and no review board are reported for the minors.
- **What paper B takes:** an expert rubric as the cheap first gate; a delayed test for anything claimed to be learning; an ethics procedure that is stated, not skipped.

### E11. FunSearch (Nature 2024)

- **Shape:** the method, two domains of results, discussion, then methods and related work after it.
- **Contributions:** in prose: discoveries in open problems "by searching for programs describing how to solve a problem, rather than what the solution is".
- **Figures and tables:** six figures (the loop, the specifications, results and discovered functions) and one table (bin packing), with extended-data figures on prompting and clustering.
- **Mechanism:** an evaluator scores every program and incorrect ones are discarded; an island-based database keeps diversity; prompts show several earlier programs, sorted by score.
- **Limitations,** in the discussion: it "works best for problems having ... an efficient evaluator; ... a 'rich' scoring feedback ... and ... a skeleton with an isolated part to be evolved".
- **What paper B takes:** the central question. A textbook has no efficient evaluator — children's learning is a slow and noisy score — and B's gates, panel and simulation are the proxies it builds instead (S17).

### E12. AlphaEvolve (2025)

- **Shape:** the system; results in three domains; ablations; related work; discussion.
- **Contributions:** in prose, as results: a new procedure for 4 × 4 complex matrices, better constructions on about a fifth of 50 mathematical problems, and gains in Google's computing.
- **Figures and tables:** eight figures (the system, its workflow, the API, results, the ablations) and two tables (against FunSearch, matrix multiplication).
- **Ablations:** no evolution, no context in the prompt, no evolved meta-prompts, no full-file evolution, a small model only; each averaged over three seeds with standard deviations.
- **Limitations:** it needs an automated evaluator.
- **What paper B takes:** the ablation design of E-B3 (S54), and the cascade — a candidate goes on only if it does well on easier tests, as B's deterministic gates come before its panel.

### E13. Juries of LLM judges (2024)

- **Shape:** introduction; methods; experimental settings; results; conclusions and limitations together; a long appendix.
- **Contributions:** four sentences: the panel, its better agreement with people at a seventh of the cost, the variance of a single large judge, and less bias within a model family.
- **Figures and tables:** four figures (rankings by judge, deltas of accuracy) and seven tables (κ by dataset, correlations, variants of prompts).
- **Evaluation:** three models from different families, pooled by majority on binary judgements and by average on scores, against GPT-4 alone and against professional annotators, on six datasets, with Cohen's κ.
- **Limitations:** three settings and few panels; the choice of the panel left to future work. It does not measure how the panel's errors correlate.
- **What paper B takes:** the comparison design of E-B2 (S53), and the gap B fills: correlation and calibrated weights.

### E14. Correlated errors (ICML 2025)

- **Shape:** introduction; related work; correlated errors (data and methods, results); a case study of LLM judges; a case study of labour markets; conclusion.
- **Contributions:** three: the size of the correlation, what drives it, and its effects downstream.
- **Figures and tables:** heatmaps of agreement, the inflation of accuracy by judge, effects in hiring; a table of regressions on the models' characteristics.
- **The measure:** how often two models' wrong answers coincide when both are wrong, against a chance baseline, over 349 and 71 models.
- **The findings:**
  - errors agree 60 % of the time on one dataset;
  - more accurate models are more correlated, even across providers;
  - a judge inflates the accuracy of weaker models, especially those of its own provider.
- **Limitations:** all wrong answers count as equally wrong; multiple-choice benchmarks rather than open generation.
- **What paper B takes:** the measure, for S14 and S53, and the reason τ is calibrated rather than set. This bears out P044, which S14 verifies.

### E15. Intelligent textbooks (IJAIED 2025)

- The editorial of a special issue.
- It describes intelligent textbooks in five generations, from closed-box adaptive systems to those built on generative AI, and counts 54 contributions to the workshop from 2019 to 2023.
- **What paper B takes:** the field's own vocabulary for its related work (S69). It is not a model of structure.

## What the exemplars share

1. **Contributions are stated where the reader looks first.** Three to five checkable items, each tied to the section that backs it (rule 2 of the [writing standard](writing-standard.md)). E1, E2 and E3 give a bulleted list; the others state them in prose, E13 and E14 as a run of "we find" and "we show" sentences. Paper A and paper B use a list.
2. **Human judgement comes with its agreement statistic.** κ, or at least raw agreement, stands next to every human label (E1, E3, E13). Modest agreement is reported, not hidden (E3).
3. **Limitations are named, but not always in a section of their own.**
   - Three give them a section of their own (E1, E2, E3), and E13 one shared with its conclusions.
   - E10, E11 and E12 name them in the discussion, and E14 in its conclusion. E5 spends a whole section on where its model misfits instead. Where E8 and E9 put theirs, the extracted text did not reach.
   - What the good ones share is that they name what the study did not test. Paper A and paper B give limitations a section of their own, as rule 6 of the writing standard asks.
4. **Ablations say which part does the work.** This is clearest in E12, and E-A1 and E-B3 follow it.
5. **A loop is judged by a randomised comparison with a real alternative,** not only with nothing (E8, E9, E10). The attrition is checked (E8), and learning is tested after a delay (E10).
6. **Ethics is stated.** Even papers with no child in them do it (E2). With minors it has to cover consent and review, and as read here both E3 and E10 leave the minors' consent unreported. Paper B must not.

## Skeleton of paper A

**Target:** AIED 2027, main track, full paper — 14 pages including references, Springer LNAI, double-blind with the system's name removed (Q28). Pages below are LNAI pages. The name MathTrail appears nowhere in the submitted version.

| § | Section | Pages | What goes in | Written in |
|---|---|---|---|---|
| — | Title and abstract | 0.5 | One central contribution in the title; context, gap, what was done and found, why it matters | S56, S63 |
| 1 | Introduction | 1.5 | Procedures against non-standard problems; LLMs write problems but err (E1); research questions; three to five contributions | S60 |
| 2 | Related work | 1.25 | Generation and verification of problems; learner models (Rasch, Elo, Math Garden); LLM tutors; in-chat delivery | S61 |
| 3 | The system | 2.0 | Design principles from learning goals, as a table; what the host's model does and what the service does; what is built, specified or planned; the open corpus of reference tasks with their solvers, traps and heuristic labels, as the released artifact | S59, S65 |
| 4 | Verification | 1.5 | The checks; the solver run twice; what they catch and what they cannot (the formalisation limit, the memory gap); the sealed answer | S58 |
| 5 | Learner model and rule | 1.5 | P = c + (1 − c)σ(θ + δ − β); the updates as stochastic approximation; β fixed per task, and why; the corridor; the chance of false mastery | S57 |
| 6 | Evaluation, with threats to validity | 3.0 | The methods, including how AI was used, as Springer asks; E-A1 injected defects: which check catches what, and the ablations; E-A2 generation by model family: acceptance, false accepts in a blind review, novelty; E-A3 simulation against Glicko-2, Urnings and fixed-step Elo; E-A4 cost and latency, in text; a last subsection of threats to validity by kind (internal, external, construct, conclusion). E-A5 (live runs) and E-A6 (teachers' blind ratings), if they exist, take the spare half page, the teachers with their demographics and consent | S62 |
| 7 | Limitations | 0.5 | What the design cannot do, apart from what the experiments did not test: no data on children; the checks test the formalisation, not the text; the model that writes a task knows its answer; the memory gap; the scale inside a grade level; the text-only format | S63 |
| 8 | Discussion, ethics and conclusion | 0.75 | What the design shows; children's data; one sentence on the textbook as future work | S63 |
| — | References | 1.0 | About 25 entries | S61, S64 |
| | **Total** | **13.5** | Half a page spare: for E-A5 and E-A6 if they exist, otherwise for figures that run long | |

**Figures and tables (five):**
- Fig. 1: who writes and who checks — the flow of one task, with the answer sealed;
- Table 1: learning goal → mechanism;
- Table 2: defect class × check, what each check catches (E-A1);
- Table 3: generation by model family (E-A2), with intervals;
- Fig. 2: the learner model in simulation against the baselines (E-A3).

Anything more goes into an anonymous appendix or the artifact.

## Skeleton of paper B

**Target:** EC-TEL 2027 — a full research paper of 8 to 15 pages including references, Springer LNCS, double-blind, with no URL to the system (Q28).

| § | Section | Pages | What goes in | Written in |
|---|---|---|---|---|
| — | Title and abstract | 0.5 | A textbook of thinking, not of topics, improved by a loop that must not optimise for recipes | S67 |
| 1 | Introduction | 1.5 | Why problem-solving heuristics; why the textbook must change with its readers; why that is hard: no cheap evaluator (E11); contributions | S68 |
| 2 | Related work | 1.25 | Closing the loop (E8), learnersourcing (E9), intelligent and AI-augmented textbooks (E10, E15), LLM evolution (E11, E12), LLM juries and correlated errors (E13, E14) | S69 |
| 3 | Problem | 1.25 | The textbook as heuristics × topics; the objective, transfer; the constraints, clarity as a threshold and no recipes; the formal statement | S70 |
| 4 | The loop | 2.25 | The three representations from one source: the reference for the chat's model, the child's mini-lessons, the readable book; signals, diagnosis, variation, gates, panel, trial, memory; the algorithm of one generation; the governance | S71 |
| 5 | The consensus filter | 2.0 | The methods, including how AI was used, as Springer asks; likelihood-ratio weights, a calibrated τ, measured correlation; E-B2 against single judges and one-family panels | S72 |
| 6 | Selection and guarantees | 2.5 | Offline decisions and one online comparison; E-B1 detection; E-B3 the closed loop against its ablations and in the world of recipes; threats to validity of the simulation | S73 |
| 7 | Integration with the product | 0.5 | What the product must change for the loop to run, the proposals of S21 and S26, and in which order | S74 |
| 8 | Limitations | 0.5 | What simulation cannot show; what the panel's calibration assumes | S74 |
| 9 | Discussion, ethics and conclusion | 0.75 | No trial on children without review and consent | S74 |
| — | References | 1.5 | About 30 entries | S69, S75 |
| | **Total** | **14.5** | Within 15 | |

**Figures and tables (five):**
- Fig. 1: the loop;
- Table 1: operations of variation by Pólya's phases;
- Fig. 2: ROC of the panel against its baselines (E-B2);
- Table 2: the ablations of E-B3;
- Fig. 3: transfer over generations in the three simulated worlds.

## The Gemini prompt's sections, placed

| The prompt's section | Where it goes | Why |
|---|---|---|
| 1 Title and abstract | A and B | Each paper states one central contribution; "multi-agent" goes where it is true (the term map of [premises.md](../evidence/premises.md)) |
| 2 Introduction and motivation | A §1, B §1 | "Equal access" is claimed only as far as the evidence goes: free, any language, in the family's own chat |
| 3 Related work | A §2, B §2 | Split by paper; multi-agent memory goes to B |
| 4 Architecture and deterministic verification | A §3 and §4 | Without Python agents or 120 edge cases, which do not exist (premises 34 and 15) |
| 5 Mathematical modelling and knowledge tracing | A §5 | The real model, not Glicko-2 (premises 27–28) |
| 6 Evaluation and benchmarks | A §6; B §5 and §6 | Experiments built by S37–S54, not benchmarks that do not exist (premise 3) |
| 7 Vision: the consensus-driven textbook engine | Removed from A; it is the whole of B | A describes what is built; a vision next to it would blur that line (premise 2) |
| 8 Discussion, ethics and broader impact | A §7–8, B §8–9 | COPPA and GDPR as S16 finds them; FERPA only if S16 shows it applies. The democratisation of open-weight models is removed: the service runs no model (premise 38) |
| 9 Conclusion and future work | A §8, B §9 | Kept |

## Leads for later tasks

- **S31:** E1's 56.6 % of fully correct tutoring dialogues; what EDUMATH's 90/77/66 % measure (E3).
- **S32:** Math Garden's 0.75 target (E4); the three difficulty levels children choose from, 90/75/60 % (E5); Pelánek's review in full (E6).
- **S11:** Koedinger et al. 2013 in full, if paper B is to cite it (E7).
- **S13 and S17:** FunSearch's three conditions for the method to work (E11), and which of them the textbook lacks.
- **S14 and S53:** the agreement-when-both-wrong measure and the inflation of weaker models (E14); PoLL's pooling rules (E13).
- **S47 and S54:** AlphaEvolve's ablation list (E12); E8's attrition check and E10's delayed test.
- **S74:** consent of minors, which neither E3 nor E10 reports as read here.
