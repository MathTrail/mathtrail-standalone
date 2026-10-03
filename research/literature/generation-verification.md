# Generating and checking maths problems with language models

The notes of task S31 (`protocol.md`, section 1), with one entry added in S61 in a section of its own at the end. The questions:
- how language models write school maths problems and their wrong options, who checks what they write, and how often it fails;
- which checks work: a program that computes the answer, a model's check of its own work, the detection of a problem that cannot be answered as posed;
- how models fare on children's olympiad problems, and what contamination means for a service that asks a model for new ones;
- how alike the texts a model writes are, how the literature measures novelty, and how near-duplicates are found;
- how readable generated text is for children;
- how the trap catalog compares with the published taxonomies of pupils' mistakes (Q41).

The notes end with what paper A may claim from this literature, the prototype's claims checked against the originals, and what the task changed in the draft.

The searches are in `search-log.tsv` (task S31). Thirty-three works with a DOI or an arXiv id and four web sources were read, all in full unless the table says otherwise. Every quotation below is verbatim, checked by script against the saved full text, with whitespace and line-end hyphens ignored. Where a converter set a formula in LaTeX (`3.23\pm 1.28`), the quotation keeps it so. The product's facts are cited by their claims in the ledger ([C0xx]), at the pinned commit `52ce86908135`.

| Key | Work | Read |
|---|---|---|
| `christ2024mathwell` | Christ, Kropko and Hartvigsen, MATHWELL (Findings of EMNLP 2024) | in full, arXiv version 5 with appendices A–G |
| `christ2026edumath` | Christ et al., EDUMATH (ACL 2026) | in full, arXiv version 2 with appendices A–J |
| `ariyarathne2025elementary` | Ariyarathne et al., elementary word problems, the MathWiz system (2025) | in full, arXiv version 2 |
| `shah2024aiassisted` | Shah et al., MATH², difficult questions from combined skills (2024) | in full, arXiv version 4 |
| `patel2025get` | Patel, Reddy and Bahdanau, CHASE (2025) | in full, arXiv version 1 |
| `huang2025keypointdriven` | Huang et al., key-point-driven data synthesis, KPDDS (AAAI 2025) | in full, the AAAI PDF; the arXiv version 3 compared |
| `mirzadeh2024gsmsymbolic` | Mirzadeh et al., GSM-Symbolic (2024; ICLR 2025, which issues no DOI) | in full, arXiv version 2 |
| `ikram2026multiagent` | Ikram et al., validator agents for personalised maths problems (AIED 2026) | in full, arXiv version 1; the published chapter not compared |
| `feng2024exploring` | Feng et al., distractors for maths multiple-choice questions (Findings of NAACL 2024) | in full, arXiv version 3 |
| `fernandez2024divert` | Fernandez et al., DiVERT (EMNLP 2024) | in full, arXiv version 2 |
| `wang2020instructions` | Wang et al., the NeurIPS 2020 Education Challenge on Eedi's diagnostic questions (2020) | in full, arXiv version 3; the starter-kit manual skimmed |
| `eedi2026misconceptions` (web.bib) | Eedi, *Eedi Misconceptions Graph v1.0* (2026) | the README and data dictionary in full; all 19,294 rows of the misconception map and 2,258 of the construct map processed by script |
| `mu2026eedis` (web.bib) | Mu, the news page announcing the graph's release (undated) | in full |
| `eedi2025from` (web.bib) | Eedi, the news page on its 2024 Kaggle competition (undated) | in full |
| `white2009revaluation` (web.bib) | White, a revaluation of Newman's Error Analysis (2009) | in full, the conference PDF; its fonts drop ligatures, so quotations are single intact lines |
| `gao2022pal` | Gao et al., PAL: program-aided language models (2022; ICML 2023, no DOI) | in full, arXiv version 2 |
| `chen2022program` | Chen et al., Program of Thoughts (2022; TMLR 2023, no DOI) | in full, arXiv version 4; the exemplar prompts are images and were not read |
| `zhou2023solving` | Zhou et al., code-based self-verification with GPT-4 Code Interpreter (2023; ICLR 2024, no DOI) | sections 1–3, 4.1 and 5 in full, arXiv version 1; the other results skimmed |
| `huang2023large` | Huang et al., models cannot self-correct reasoning yet (2023; ICLR 2024, no DOI) | in full, arXiv version 2 |
| `tyen2024llms` | Tyen et al., models cannot find reasoning errors but can correct them given the location (Findings of ACL 2024) | in full, arXiv version 3 |
| `shen2026lets` | Shen et al., MathQ-Verify and ValiMath (KDD 2026) | in full, arXiv version 2, which carries the KDD header |
| `verga2024replacing` | Verga et al., a panel of models as judges, PoLL (2024) | the abstract only, read again in S61 on 2026-10-01: the exemplar notes of S06 came from a summary, which G3 does not let a paper cite |
| `zhang2024clamber` | Zhang et al., CLAMBER (ACL 2024) | in full, arXiv version 2 |
| `li2025questbench` | Li, Kim and Wang, QuestBench (NeurIPS 2025) | in full, arXiv version 2; the prompts recovered from the page's SVG boxes |
| `kirichenko2025abstentionbench` | Kirichenko et al., AbstentionBench (NeurIPS 2025) | in full, arXiv version 1, the only one; most results per dataset are in figures only |
| `sun2024benchmarking` | Sun et al., unanswerable maths word problems, UMWP (LREC-COLING 2024) | in full, arXiv version 1; results per model are in figures only |
| `gupta2025beyond` | Gupta et al., *Beyond Final Answers*, models as maths tutors (AIED 2025) | in full, the authors' arXiv version 1 of 23 February 2025; Springer's page refused automated readers, so the published chapter was not compared |
| `cherian2024evaluating` | Cherian et al., SMART-840, models on children's olympiad problems (NeurIPS 2024) | in full, arXiv version 2 with appendices A–C |
| `balunovic2025matharena` | Balunović et al., MathArena (NeurIPS 2025) | the main text and the contamination appendix in full, arXiv version 3; result tables and prompts skimmed |
| `jiang2025artificial` | Jiang et al., the Artificial Hivemind (NeurIPS 2025) | the main text and appendices A–C in full, arXiv version 1 (marked camera-ready); appendix D skimmed |
| `schaeffer2026strong` | Schaeffer et al., a re-evaluation of the Artificial Hivemind (2026, preprint) | sections 1–5 in full, arXiv version 1; appendices not read |
| `ye2025assessing` | Ye et al., CreativeMath, the novelty of models' solutions (AAAI 2025) | in full, arXiv version 1 with its appendix; the AAAI PDF compared, same numbers |
| `toshniwal2024openmathinstruct` | Toshniwal et al., OpenMathInstruct-2 (2024; ICLR 2025, no DOI) | in full, arXiv version 2; the ICLR camera-ready read for its decontamination passages, unchanged |
| `valentini2023automatic` | Valentini et al., generating and simplifying children's stories (EMNLP 2023) | in full, arXiv version 1 |
| `broder1997resemblance` | Broder, the resemblance and containment of documents (1997) | in full, the author's 9-page copy from a Princeton course archive; the IEEE version not read |
| `broder2000minwise` | Broder, Charikar, Frieze and Mitzenmacher, min-wise independent permutations (JCSS 2000) | sections 1 and 4 in full, the authors' 36-page copy; sections 2–3 on the size of families skimmed; ScienceDirect refused automated readers |
| `lee2022deduplicating` | Lee et al., deduplicating training data (ACL 2022) | the main text and appendices A–B in full, arXiv version 2 |

## 1. Generating problems for school

### MATHWELL

**Key:** `christ2024mathwell`. Teachers judged word problems for kindergarten to grade 8 written by models, each with a Python function that computes its answer.

- **What it did.** The criteria were solvability, accuracy and appropriateness; a problem that "meets all criteria" is MaC. Llama-2 70B prompted with GSM8K examples reached 35 % MaC. The authors fine-tuned it twice, on GSM8K and then on 1,906 teacher-approved problems, into MATHWELL. On 250 problems per model, MATHWELL reached 74.8 % MaC, GPT-4 Turbo 78.8 %.
- **Where the key comes from.** The model's program computes it, and only the teachers check it.
- The criteria: "1) Solvability, where questions are possible to solve and have one correct answer. 2) Accuracy, where generated answers must be correct. 3) Appropriateness, where MWPs should be questions teachers would feel comfortable giving to K-8 students." (§3, "Human Evaluation Criteria").
- The baseline: "with the worst performance being in appropriateness where barely over 50% of generations are appropriate, leading to only 35% of the generations being labeled as MaC." (§3.1).
- MATHWELL's row of Table 3 (solvability, accuracy, appropriateness, MaC, then three more columns): "MATHWELL (Ours) 89.2 (1.97) 96.9 (1.17) 86.5 (2.29) 74.8* (2.75)".
- A program that encodes a misreading: "The solution adds an additional bag of food to the total assuming that the answer is a decimal and, therefore, that the shelter would need to buy another full bag since a partial bag is not possible." (appendix E.2.2, a GPT-3.5 accuracy failure).

**What paper A takes.** MATHWELL's three criteria are what MathTrail checks automatically, in part: the solver targets solvability and accuracy, and of appropriateness only readability is checked [C021, C030]. Both systems take the answer from a program the model wrote; MATHWELL uses it as the key, MathTrail runs it against a key stated separately [C023]. A program can encode a misreading, as the extra bag shows, and MathTrail cannot see one its key shares [C026].

### EDUMATH

**Key:** `christ2026edumath`. Word problems aligned with the standards of US grades 3–5, judged by teachers and then tried with pupils.

- **What it did.** Llama 3.3 70B wrote 3,012 problems with worked solutions. 1,372 Prolific workers who identified as US teachers labelled them for solvability, accuracy, appropriateness and standards alignment, two or three per problem, and a Gemma 3 27B judge demoted the problems teachers passed but it did not. EDUMATH 12B and 30B were trained on the 1,552 problems that met every criterion. MaC ran from 63.9 % (Gemma 3 12B) to 92.8 % (GPT-4o) and 94.6 % (EDUMATH 30B); most failures were in accuracy and alignment. Pupils solved generated and human-written problems at similar rates.
- The grades: "We study generating MWPs for grades 3-5 in the US." (§3).
- The agreement of the annotators, which the prototype's notes misread as shares of problems (P039 in the table of the prototype's claims below): "The first two annotators agreed on solvability 90.1\pm 1.2\% of the time, accuracy 76.6\pm 1.5\% of the time, educational appropriateness 77.5\pm 1.5\% of the time, standards alignment 75.3\pm 1.6\% of time, and MaC 65.5\pm 1.7\% of the time." (§3.2).
- The judge against experts: "The annotators agreed with each other 76% of the time on MaC and had an average agreement rate of 75% with Gemma 3 27B IT for MaC." (§4.1).
- Where errors sit: "Table 9 shows the majority of standards alignment errors for each model are for MWPs that are missing important parts of the prespecified standard(s)." (appendix E).
- With pupils: "In this experiment (see Appendix G.1 for full results), students performed similarly on human-written and LLM-generated MWPs, further suggesting EDUMATH 30B’s MWPs are human quality." (§5).

**What paper A takes.** The closest published pipeline for primary grades. A strong model's remaining errors sit in fitting the requested standard and in reasoning, not in solvability. None of MathTrail's nine checks tests whether the written task exercises the topic and difficulty the rule chose [C021]; the task's difficulty is the one asked for [C013]. EDUMATH's problems are open-answer, routine and without options; its pupils' study measured solve rates and preferences, not learning.

### MathWiz

**Key:** `ariyarathne2025elementary`. The only work read that writes problems for grades 1–6. The search for other systems named MathWiz found none.

- **What it did.** Llama-2 7B, fine-tuned on about 4,000 hand-made problems for grades 1–6 of the US Common Core, writes problems from a grade, a curriculum section and a count, with no answers. A Gemini "solvability detector" discards problems it judges unsolvable. The problems were rated by people on a 12-category taxonomy of errors.
- The input: "the only input to our system is the number of MWPs needed, the grade and the type of question (e.g. addition, subtraction)" (abstract).
- The model judge: "Accordingly, 88.24% of the MWPs that are actually solvable, have been identified as solvable by the solvability detector (TPs). Similarly, 40% of unsolvable MWPs has been identified as unsolvable (TNs)." (§7.2, "Solvability Module").
- An automatic judge against people: "However, it did not show a strong correlation with humans for more complex errors, such as “Unsolvable Problems” or “Grade/Section Requirements.”" (§7.1).
- Grade fit (columns: zero-shot with tuned decoding; fine-tuned zero-shot; fine-tuned five-shot; five-shot with the filter): "Grade relevance 65 65.38 79.62 74.52" (Table 10).

**What paper A takes.** First-hand support, at the right ages, for checking solvability with a program rather than a model judge: the model evaluator tracked people poorly on unsolvable problems. The detector's 40 % points the same way, but paper A does not cite it: the appendix's prompt for the detector asks for "TRUE" when a problem is not solvable, while §4.6 describes TRUE as solvable, so what the figure counts is uncertain. Even after fine-tuning, about one problem in five missed the requested grade, which MathTrail does not check either. MathWiz writes routine exercises without answers or options.

### MATH²

**Key:** `shah2024aiassisted`. Frontier models combined two skills of the MATH dataset into new, hard questions, validated them themselves, and graduate students checked what survived.

- **What it did.** Each question passed the generating model's own checks — an "adversarial" attempt, a rubric judged by a majority of four, a re-solve by a majority of four — before humans saw it. Of the 210 questions kept, 130 were changed by people.
- The re-solve: "If all the answers obtained in this step are unique, indicating potential ambiguity, the question is discarded." (§2, step 5).
- After people: "Out of 210 question-solution pairs included in MATH2, 130 (61.9%) underwent some form of modification by the human annotators before being included in the dataset." (§3.1).
- The authors: "Note that none of the above failure modes are completely eliminated in the pipeline described in Section 2. Thus, human verification is required." (appendix A.3). And as future work: "This could include leveraging code generation and autoformalization capabilities of LLMs to generate responses which can be compiled using compilers or interpreters." (§5).

**What paper A takes.** The strongest evidence read here that a model's validation of its own problems does not make them trustworthy. MathTrail replaces the human verifier with an executed program, the route MATH² names as future work, for small problems a program can enumerate. MATH² is high-school level and built to test models.

### CHASE

**Key:** `patel2025get`. Problems built step by step to be hard for the model that writes them, each step verified by other models.

- **What it did.** GPT-4o-mini extended GSM8K and SVAMP seed problems; every verifier model had to reproduce each step's answer, or the step was discarded. 500 problems were kept.
- The verification: "We prompt each \mathbf{V_{k}} with the generated context c_{i} and question q_{i} and check whether the prediction is the same as the generated answer a_{i} ." (§4.3).
- Without it: "For the math task, we manually examined the generated problems and found that 34 of them had some kind of error such as the problem text being ambiguous or vague or the reasoning and answer being incorrect." (§5.2, a direct-generation baseline of 100 problems).
- With it: "We found 6 errors in CHASE-QA, 3 errors in CHASE-Code and 7 errors in CHASE-Math." (appendix C.1, 100 examples per benchmark).

**What paper A takes.** Two of MathTrail's principles: check pieces that can be checked separately, and do not let the author verify its own problem alone. CHASE's verifiers are other models, which MathTrail cannot call [C003]; MathTrail's verifier is a program by the same model, deterministic but not independent of the author's reading.

### KPDDS

**Key:** `huang2025keypointdriven`. Training data synthesised by GPT-4 from the topics and key points of MATH and GSM8K problems.

- **What it did.** GPT-4 scored each new problem; a threshold kept about half. Ten sampled solutions per problem had their arithmetic checked by a Python program and were then voted. A human check of 100 problems per subset followed.
- The program check: "Next, we extract mathematical expressions from each solution and use a Python program to verify their accuracy, excluding any solutions with computational errors." (PDF p. 4 / p. 24179).
- What remained: "the KPMATH-G subset achieved a correctness rate of 95%, with observed errors primarily in problem statements or logical reasoning rather than calculation errors." (PDF p. 5 / p. 24180).

**What paper A takes.** After a model's score, an arithmetic check and a vote over ten solutions, one problem in twenty of the grade-school part was still wrong, and the errors were "in problem statements or logical reasoning", which an arithmetic check cannot see. MathTrail's solver sees such an error only when it does not share it.

### GSM-Symbolic

**Key:** `mirzadeh2024gsmsymbolic`. People turned 100 GSM8K questions into templates whose program computes each instance's answer.

- **What it did.** Templates name variables, their domains and conditions that keep answers whole; automated checks and a manual review of ten instances each guard them. Across 25 models, accuracy varies between instances of one question, drops more when numbers change than when names do, and drops by up to 65 % when an irrelevant clause is added (GSM-NoOp).
- The numbers: "Specifically, the performance of models declines when only the numerical values in the question are altered in the GSM-Symbolic benchmark." (abstract).
- The clause: "When we add a single clause that appears relevant to the question, we observe significant performance drops (up to 65%) across all state-of-the-art models, even though the added clause does not contribute to the reasoning chain needed to reach the final answer." (abstract).
- The authors' reading: "Overall, we find that models tend to convert statements to operations without truly understanding their meaning." (§4.4).

**What paper A takes.** The only work read in which a program computes the key, and people write and check the program. MathTrail's key is checked by a program the task's own author wrote. NoOp bears on MathTrail twice: the host model that writes and checks a task is itself prone to turning an irrelevant statement into an operation, and the children's traps `number_from_text` and `ignored_condition` have a machine analogue. Appendix B answers a statistical critique (Ivanova et al., 2025, not read).

### Validator agents for personalised problems

**Key:** `ikram2026multiagent`. A model rewrites existing problems around a pupil's interest, and four model agents validate each rewrite until it passes or three rounds of revision are spent. Read in its arXiv version 1 (6 April 2026); the AIED 2026 chapter was not compared.

- The task: "We study the task of personalizing a math problem to a target topic that reflects a learner’s interest, while preserving the underlying mathematical structure and solution and grade-appropriate language and context." (§3.1). The problems: "We experiment on 600 math problems drawn from an online mathematics homework platform, ASSISTments [8]." (§4.1), for middle school, with GPT-5.2 in every role.
- Solvability is judged by a model that solves: "Solvability Agent. The solvability validator agent solves the problem to verify mathematical consistency with the original and checks answer-set validity (i.e., the correct option is present in multiple-choice items)." (§3.2).
- It seldom failed: "For solvability, all strategies start with a low failure count and rapidly converges to near zero after a single iteration." (§4.1). People did not check it: "We excluded solvability since its labels were highly skewed (few failures), making agreement estimates less informative." (§4.1). On the other criteria the agents agreed with three annotators about as weakly as the annotators agreed with each other: "Therefore, the agreement between humans and LLM-based validators are roughly in the same range." (§4.1).

**What paper A takes.** The closest system at paper A's own venue: generate, validate, revise, as MathTrail refuses a task with every problem named and lets the model try again [C032, C071]. Its validators are models, solvability included, and nobody checked the solvability verdicts; MathTrail's is a program the service runs [C023]. It rewrites existing middle-school problems rather than writing new ones for grades 1–6.

## 2. Wrong options and the mistakes behind them

### Distractors from models: Feng et al. and DiVERT

**Keys:** `feng2024exploring`, `fernandez2024divert`. Both add distractors to multiple-choice questions that teachers wrote for Eedi, a maths platform, for pupils aged 10–13.

**Feng et al.**
- Five ways of writing three distractors were compared against the teachers' own: similar questions as examples (kNN) did best; choosing from 444 hand-written error explanations (RB) next; sampling a solver model's wrong answers worst.
- The list: "we approximately follow the baseline approach in Dave et al. (2021) and manually construct 444 distinct error explanations, such as “confuses factor and multiples” for question-distractor pairs that correspond to common errors or misconceptions among real students." (§2.2).
- Why it lost: "This result is likely due to the fact that despite extensive effort in labeling error explanations, we cannot come up with a comprehensive list of them; as a result, many target MCQs are not matched with error explanations for GPT-4 to select from." (§3.4).
- Two raters on 20 questions (columns: agreement for model and for human distractors; mean rating 1–5 for each): "Validity 0.34 0.23 3.28 3.99∗ Plausibility 0.54 0.54 2.68 3.72∗" (Table 7).
- Teachers' own options: "This result suggests that many MCQs have 1 or 2 highly plausible distractors while the others are placeholders." (§3.5).

**DiVERT.**
- Writes the error behind each distractor as text first, with three fine-tuned 7B models; on 1,434 Eedi questions it matches GPT-4o with examples.
- What Eedi's labels leave out: "Among these pairs, 3601 pairs have human annotated error descriptions; other pairs are not labeled due to the non-mathematical nature of errors, e.g., careless slipping, reading the question incorrectly, etc." (appendix D).
- The commonest failure: "By far, the most frequent failure pattern we observe is on the controllable distractor generation model p(d|s,e) , where the generated distractor is not faithful/consistent to the error." (§5.2).
- Ratings of error texts (Human, DiVERT, GPT-4o): "Rating 3.23\pm 1.28 3.07\pm 1.39 2.56\pm 1.25" (Table 5). GPT-4o: "Qualitatively, we find that GPT-4o’s errors are often not what real students are likely to make, and even occasionally confuse the correct solution approach with an error." (§6.2).

**What paper A takes.**
- MathTrail asks the host model to do what Feng et al.'s RB did: name a mistake from a closed list and write the option it leads to [C020, C072]. RB's list of 444 was never complete and lost to examples; MathTrail's has 20 [C043].
- Raters put model distractors a point below teachers' on plausibility.
- DiVERT's commonest failure points at what no MathTrail check covers. The explanations check is mechanical, and its code says it does not judge whether an explanation is a good one (`internal/domain/checks/explanations.go` at the pin); nothing checks that a wrong option is the answer its named trap would give [C021].
- Whether the trapped options are plausible can be measured only by which options children choose.

### Eedi's diagnostic questions and its misconception graph

**Keys:** `wang2020instructions`, `eedi2026misconceptions`, `mu2026eedis`, `eedi2025from`.

- **The design.** "A diagnostic question is a multiple-choice question with four answers, exactly one of which is correct and where each of the three incorrect answers is chosen to highlight a common misconception." (`wang2020instructions`, §1.1). In 2020 the links were missing: "When teachers create diagnostic questions, each incorrect answer should be chosen to highlight a common misconception, but these misconceptions are not labelled or linked between questions." (§1.1). The platform: "This platform offers crowd-sourced diagnostic questions to students from primary to high school (roughly between 7 and 18 years old)." (§2).
- **The competition.** "But here’s the thing: while we had all this rich data, we didn’t have labels linking distractors to the misconceptions they revealed." (`eedi2025from`, "Data"). The task: "Given a distractor and a list of misconception descriptions, predict which ones match." (`eedi2025from`).
- **The graph.** Released in 2026 under CC BY 4.0: "Coverage: **8,117 misconceptions**, **2,258 constructs**, 236 sub-topics, 50 topics, 4 subjects (Number, Geometry and Measure, Algebra, Data and Statistics)." (`eedi2026misconceptions`, README, "Contents"). Without the questions: "The Diagnostic Questions themselves, their distractors, question images, and any student-level response data are **not** part of this release." (README). The news page: "The roots of our Misconception Graph can be found in our diagnostic questions: four multi-choice options, one right, three wrong, with every wrong option tied to a specific misconception." (`mu2026eedis`).
- **What the graph does not give:** year levels, or how often pupils hold a misconception. Its weights say only how a misconception's tags spread over constructs.

**What paper A takes.** Eedi's diagnostic question is the closest existing design to MathTrail's options: every wrong option stands for a mistake. MathTrail's differ in having five options, being written by a model, and serving multi-step non-routine tasks rather than single-skill items. A trap id behind every wrong option [C077] is the link across questions Eedi lacked in 2020. The graph is the published catalogue the trap catalog is compared with below; drawing on it would oblige MathTrail to credit Eedi under CC BY 4.0.

### Newman's error analysis

**Key:** `white2009revaluation` (web.bib). White reports Newman's error analysis (NEA) and its use in a New South Wales numeracy programme.

- NEA places a pupil's error on a written word problem at the first hurdle the pupil fails: "then that person had to be able to pass over a number of successive hurdles: Level 1 Reading" "(or Decoding), 2 Comprehension, 3 Transformation, 4 Process Skills, and 5 Encoding." (p. 251; the passage is two printed lines). Careless errors and a lack of motivation sit outside the hierarchy (p. 251).
- White's figures on primary pupils, and Clements's (1980) on 634 children, are reported here only through White; Clements was not read.

**What paper A takes.** NEA classifies where a solution of a word problem breaks down, not which concept is wrong, which suits traps that are mostly about process.

### The trap catalog against Eedi and Newman

The catalog: 20 traps at the pin [C043]. Each was compared with (a) the names of the Eedi graph, searched with keyword patterns and every hit read; (b) Newman's stages; (c) what DiVERT says Eedi leaves unlabelled. The Eedi names below are verbatim from the graph's CSV, checked by script; "none" means no hit for the patterns tried.

| Trap | Eedi counterparts (id: name) | Fit |
|---|---|---|
| `off_by_one` | 15304: When finding the difference between two marks on a scale, counts lines instead of gaps; 21563: Believes that counting on includes the starting number as part of the count | close |
| `missed_case` | 28412: Believes there is only one solution to a problem with multiple solutions; 19524: Does not consider all possible lines of symmetry | partial |
| `double_count` | 27094: Believes the product rule for counting needs to be doubled to account for both orderings; 18622: When calculating perimeter, adds one side twice | close |
| `ignored_not` | 25160: Confuses finding a false statement with finding a true statement; 15658: Thinks they are finding a possible value, not impossible | partial |
| `answered_other_question` | 15027: Finds the shaded section rather than the specified section; 16778: When asked to shade a fraction of a shape, gives the total number of parts rather than the number to be shaded | topic-bound instances only |
| `stopped_early` | 15251: In a worded problem, does not complete all calculations; 21202: Forgets to subtract from the total after adding | close |
| `wrong_operation` | 13926: Adds rather than subtracts when answering worded problems; 13917: Multiplies rather than adds when given the command word 'more' | close |
| `reversed_relation` | 18930: Confuses taller and shorter when comparing heights; 21198: Mixes up 'less than' and 'more than' in word problems | close |
| `best_case_not_worst` | none | none |
| `trusted_statement` | none | none |
| `time_unit_mixup` | 15726: Answers as if there are 100 minutes in an hour; 18302: Confuses hours and minutes | exact |
| `number_from_text` | 19197: When given a word problem, just writes the information given as the answer rather than working it out; 13816: When asked for a total, just gives one value | exact |
| `wrong_parity` | 19479: Believes that the sum of two even numbers can be odd; 19481: Believes that adding an even and an odd number could give an even number | partial: no item for odd plus odd |
| `ignored_condition` | 24838: Thinks an answer only needs to satisfy one of the properties given in a multi-constraint problem; 28761: When solving more than one inequality, believes the solution only needs to satisfy one of the inequalities | close |
| `percent_wrong_base` | 17160: Believes that they can reverse a percentage increase by decreasing the new value by the same percentage, and vice versa | close |
| `part_whole_swap` | 15831: Confuses the part and the whole in fraction of an amount problems; 20365: Confuses part and whole when using a bar model | close |
| `ratio_total_confusion` | 22118: Assumes a ratio is part to whole instead of part to part; 15800: Divides total amount by how many numbers are in ratio instead of dividing by the sum of the parts | close |
| `remainder_vs_quotient` | 14241: When dividing, confuses the remainder with the quotient | exact |
| `area_perimeter_swap` | 13901: Calculates area when asked for perimeter | exact |
| `first_move_assumed` | none | none |

- **No counterpart in Eedi:** `best_case_not_worst` (the worst case, pigeonholes), `trusted_statement` (knights and liars) and `first_move_assumed` (strategy games). They are errors of olympiad topics, outside the school curriculum the graph is built on.
- **Weak in Eedi:** the misreading traps `ignored_not` and `answered_other_question` have only scattered, topic-bound instances, which fits DiVERT's note that Eedi leaves "reading the question incorrectly" unlabelled.
- **In Newman's stages** every trap has a place. This placement is ours, not White's: comprehension (`ignored_not`, `ignored_condition`, `reversed_relation`, `trusted_statement`); transformation (`wrong_operation`, `part_whole_swap`, `ratio_total_confusion`, `percent_wrong_base`, `area_perimeter_swap`, `best_case_not_worst`, `first_move_assumed`); process skills (`off_by_one`, `missed_case`, `double_count`, `time_unit_mixup`, `wrong_parity`, `remainder_vs_quotient`); encoding (`stopped_early`, `number_from_text`); comprehension or encoding, depending on whether the child misunderstood the question or wrote down the wrong quantity (`answered_other_question`). The misreading traps are failures of comprehension: the child read the words and took the wrong meaning. No trap names a decoding error, where a word itself is misread, or a careless slip.
- **Common primary-level mistakes no trap covers**, with the number of distinct Eedi names in the sub-topic, a measure of how finely Eedi splits it, not of how often pupils err. Where one touches a MathTrail topic, the topic is named.
  - Place value (528 names), for example 13708: When asked for the value of a digit, just gives the digit.
  - A plain slip in mental or written calculation (284, 295 and 260 names), which Newman counts as a process or careless error.
  - The order of operations (62; touches `arithmetic.tricks`), for example 14402: Adds from left to right.
  - Comparing and adding fractions (95, 96 and 134; touches `fractions.parts`), for example 14090: Believes that fractions with larger denominators are greater in size.
  - Factors against multiples (touches `number.divisibility`): 13957: Confuses factors and multiples.
  - Reading a clock, and am against pm (touches `time.clocks`), for example 18058: Confuses the hour and minute hand on a clock; `time_unit_mixup` covers hours against minutes only.
  - Perimeter and area on a grid (touches `geometry.grid`), for example 14017: Does not count both edges when turning a corner when counting perimeter.
  - Venn diagrams (touches `logic.sets`), for example 15276: Forgets to include the intersection value when asked for total value of a set in a Venn diagram; `double_count` and `missed_case` cover these only by being general.
  - Outside MathTrail's topics: decimals, negative numbers, rounding, money, units of measure, reading scales, shapes and angles, data.
  - Newman's decoding errors and careless slips, which sit outside any topic.
- **Limits of the comparison.** The graph has no year levels, so "primary" is a choice of topics; it has no prevalence. The mapping rests on keyword search over 8,117 names and reading the hits, not on reading every name. Newman's scheme was read through White.

## 3. Programs that compute the answer

**Keys:** `gao2022pal`, `chen2022program`, `zhou2023solving`. PAL and Program of Thoughts are the same idea, published at the same time: the model writes a Python program for a word problem, and an interpreter runs it.

**PAL.**
- With Codex on grade-school sets, PAL is ahead of chain-of-thought (CoT) everywhere: GSM8K 72.0 against 65.6, and widely on GSM-Hard, where numbers are large. Row CoT with Codex, then PAL (columns gsm8k, gsm-hard, svamp, asdiv, …): "65.6 23.1 74.8 76.9 89.1 91.9 86.0 95.9" and "PaL 72.0 61.2 79.4 79.6 96.1 94.6 92.5 99.2" (Table 1). The text of §5.1 gives other GSM-Hard figures, 20.1 and 61.5.
- The gain is the interpreter's: "This resulted in a 23.2 solve rate on gsm8k, much lower than PaL (72.0), and only 4.5 points higher than Direct." (§6, the model "executing" its own program in text).
- Correct "given the right program": "PaL avoids these problems by offloading the calculation and some of the reasoning to a Python interpreter, which is correct by construction, given the right program." (§7).
- Agreement taken as proof: "Notably, this annotation process assumes that a program that produces a correct answer to a gsm8k question indicates the correctness of the program itself. While this is not guaranteed due to possible spurious correlations, we manually checked 25 programs and found all of them are correct." (appendix H.1).

**Program of Thoughts.**
- With GPT-4 (columns: parameters, GSM8K, AQuA, SVAMP, TabMWP, FinQA): "CoT-GPT4 175B 92.0 72.4 97.0 - 58.2" and "PoT-GPT4 175B 97.2 84.4 97.4 - 74.0" (Table 2).
- Where programs help: "The largest improvements of PoT are in the categories ‘linear/polynomial equation’, ‘iterative’, ‘symbolic’, and ‘combinatorics’. These questions require more complex arithmetic or symbolic skills to solve. In contrast, on ‘arithmetic’, ‘probability’, and ‘geometric’ questions, PoT and CoT perform similarly." (§3.3).
- What remains wrong: "The first type indicates that the model fails to assign correct values to the variables relevant to the question. The second type indicates that the model fails to generate the correct computation process to answer the question based on the defined variables." "Among the 198 failure cases of numerical reasoning questions with the PoT (greedy) method, 47% have value grounding errors and 33% have logic errors. In 15% both types of errors occurred and in 5% we believe the answer is actually correct." (§3.3, on financial tables).

**Code-based self-verification.** `zhou2023solving`, read in its arXiv version 1. It was presented at ICLR 2024 (OpenReview), which issues no DOI, so the registries hold only the arXiv record.
- A model checks its own solution with code it writes: "This method employs a zero-shot prompt on GPT-4 Code Interpreter to encourage it to use code to self-verify its answers. In instances where the verification state registers as “False”, the model shall automatically amend its solution, analogous to our approach of rectifying errors during a mathematics examination." (abstract). On MATH, from GPT-4 Code Interpreter's own 69.69 %: "On top of GPT4-Code, our method further improves its accuracy, raising the result to 73.54% after adding explicit code-based self-verification, and 84.32% after adding both explicit code-based self-verification and verification-guided weighted majority voting (the number of sampled paths is 16)." (§4.1). The abstract's two figures, 53.9 % before and 84.3 % after, compare the method with an earlier state of the art, not with the same model.
- The model also decides when it cannot check: "An Uncertain classification indicates that GPT4-Code encountered difficulties in identifying an effective method for answer verification, thereby abstaining from delivering a definitive verification result." (§3.2).
- It solves given problems; nobody outside the model reads the verification's result.

**What paper A takes.** A runtime removes a model's arithmetic errors, the class of error MathTrail's solver catches. In PoT's analysis of failures on financial questions, most of what the programs still got wrong was a wrong value bound to a quantity or wrong logic. The first kind is a misreading, and a solver written by the task's author shares it with the author's key [C026]; the analysis is of another domain, so paper A cites it for the kind of error, not for its shares. Zhou et al. come nearest to MathTrail's solver: the model writes code to check its own answer. The difference is who reads the result. There the model reads it and decides; in MathTrail the service runs the program and compares its result with the key [C023]. With a strong model, execution adds little on simple arithmetic problems (SVAMP), so the solver's worth for MathTrail lies less in computing better than the model and more in making one check deterministic and able to try all five options [C023]. MathTrail uses the program the other way round: it does not produce the answer but must single out the keyed option, and the key is the same model's own, not a gold label. Neither paper measures how often a wrong program still returns the expected answer.

## 4. A model checking its own work

**Keys:** `huang2023large`, `tyen2024llms`, and `gupta2025beyond`, which the introduction cites.

**Huang et al.**
- Without oracle labels, self-correction lowers accuracy: "Upon closer examination, we observe that the improvements in these studies result from using oracle labels to guide the self-correction process, and the improvements vanish when oracle labels are not available." (§1).
- GPT-4 (columns: calls, GSM8K, CommonSenseQA, HotpotQA): "GPT-4 Standard Prompting 1 95.5 82.0 49.0 Self-Correct (round 1) 3 91.5 79.5 49.0 Self-Correct (round 2) 5 89.0 80.0 43.0" (Table 3).
- Why: "For GSM8K, 74.7% of the time, GPT-3.5 retains its initial answer. Among the remaining instances, the model is more likely to modify a correct answer to an incorrect one than to revise an incorrect answer to a correct one. The fundamental issue is that LLMs cannot properly judge the correctness of their reasoning." (§3.3).
- What works: "In particular, when the problem description clearly specifies the intended code execution behavior, e.g., with unit tests, the code executor serves as the perfect verifier to judge the correctness of predicted programs, while the error messages also provide informative feedback that guides the LLMs to improve their responses." (§6).

**Tyen et al.**
- The bottleneck: "In this paper, we show that poor self-correction performance stems from LLMs’ inability to find logical mistakes, rather than their ability to correct a known mistake." (abstract). "GPT-4 attains the best results but only reaches an overall accuracy of 52.87 with direct step-level prompting." (§3.1).
- On one's own traces: "Note that the traces in our dataset are generated using PaLM 2 Unicorn and are sampled according to whether the final answer was correct or not. Therefore, we expect that using PaLM 2 itself to do mistake finding will produce different and likely biased results." (§3.1, footnote 6). PaLM 2 Unicorn located the mistake in 5.88 % of its own multistep-arithmetic traces that contained one, with direct trace-level prompting (appendix E, Table 9(b)).
- Given the location, the repair works: on multistep arithmetic, +18.04 points on traces first wrong, no loss on those first right (Table 6). "This suggests that the main bottleneck in self-correction methods is the identification of mistakes, rather than the correction process." (§4.2).

**Gupta et al.** Models as tutors in college algebra, judged by two reviewers.
- The setting: "The evaluators prompted each of the five models to provide tutoring support on the same set of 30 problems, resulting in a total of 150 LLM dialogues." (§4.2). The models: GPT-3.5 Turbo, GPT-4, GPT-4o, o1-mini and o1-preview.
- How a dialogue counted as wrong: "If there were any errors, then the dialogue was coded as incorrect, otherwise it was recorded as correct." (§3.2). The reviewers agreed with κ about 0.82 on correctness (§4.2).
- The finding: "Across all five of the LLMs only 56.6% of the tutor dialogues were entirely correct." (§5).

**What paper A takes.** The self-check is the least independent of the nine checks. It rests on what Huang et al. find missing — judging whether one's own reasoning is right — and Tyen et al. find a model checking its own traces a poor mistake finder. MathTrail never lets the self-check admit a task alone: a task with a blocking issue is refused, and a self-check that reaches another answer than the key is a disagreement [C021]. The literature supports that design and does not support calling the self-check a verification. Its use is as a prompt to look for named defects (section 5). MathTrail's refusals name every failed check at once [C071]: that tells the model which check failed, not which step of its reasoning went wrong, so Tyen et al.'s repair from a known location does not carry over to them. No work read measures how often a model's check finds the defects of a problem it wrote itself. Gupta et al. measure tutoring dialogues in college algebra, not generated problems, and the paper read is the authors' copy.

## 5. Problems that cannot be answered as posed

**Keys:** `shen2026lets`, `sun2024benchmarking`, `li2025questbench`, `kirichenko2025abstentionbench`, `zhang2024clamber`.

**MathQ-Verify** checks a question rather than its answer: five stages, each a prompt to a model, for contaminated instructions, linguistic errors, a condition that breaks its domain's rules, conditions in conflict, and incompleteness.
- The benchmark: "we present ValiMath, a benchmark of 2,147 questions (1,299 correct, 848 incorrect) covering five diverse error types" (§3.1). Multiple-choice questions were excluded: "To ensure the uniqueness and determinacy of the final answers, we applied a GPT-4o-based prompt filtering process to exclude all multiple-choice questions." (§3.2).
- The best single model (columns: voters, precision, recall, F1, true positives, false positives; the positive class is a valid question): "(1, 1) ① 86.72 68.90 76.79 895 137" (Table 4). It accepts 137 of the 848 flawed questions. Three different models voting unanimously accept 75: "(3, 3) ① ② ④ 91.42 61.51 73.54 799 75" (Table 4).
- The paper claims "a 15% absolute F1 gain over direct baselines on ValiMath" (§1); every difference per model in its Table 3 is far smaller, and the agreement of its annotators is announced but not reported.

**UMWP** pairs answerable elementary word problems with versions made unanswerable.
- "UMWP comprises a total of 5,200 questions with half answerable questions and half unanswerable questions." (§1). Unanswerable means no solution or no unique one, in five kinds: key information missing; ambiguous key information (ranges, vague terms, negations); unrealistic conditions; an object absent from the text; a missing question.
- "GPT-4 demonstrates the best performance achieving an impressive F1 score of 85.24%. However, it still shows a difference when compared to the human benchmark result of 93.16%." (§5.4).
- A model computing through a range: "Q: The Razorback t-shirt shop sells each t-shirt for $ 51 dollars. During the Arkansas and Texas tech game they offered a discount of more than $ 8 per t-shirt and sold roughly 130 t-shirts. How much money did they make from selling the t-shirts? A: The t-shirts were sold for $51 - $8 = $43 each after the discount." (appendix A.6, Table 6, GPT-4).

**QuestBench** casts a problem as constraints with exactly one needed value missing.
- The distinction: "This formalization helps pinpoint the difference between semantic ambiguity (where multiple valid interpretations exist, but each yields a solvable answer) and underspecification (where the problem is unsolvable without additional information)." (§1).
- Told that something is missing, GPT-4o picks the right question for 86.81 % of the grade-school items (Table 2). Asked to answer or say "not sure", it reaches an F1 of 60.87 against 50.0 for a proportional random guess (Table 7). On planning: "we found that Gemini Flash Thinking, Claude 3.5, and GPT-4o predicted “not sure” on only 1.4%, 4.6%, and 0.7% of cases in the zero-shot (no chain-of-thought) setting on Planning-Q, when the ground-truth ratio is 41.8%." (§5.4).

**AbstentionBench** gathers questions a model should not answer outright.
- "we find that reasoning fine-tuning degrades abstention (by 24% on average), even for math and science domains on which reasoning models are explicitly trained." (abstract).
- "Instead of clarifying, expressing uncertainty, or pointing out incorrect assumptions, models inappropriately respond definitively (see appendix F for qualitative examples)." (§4.1).
- Deleting a problem's whole context is easy to notice: o1's recall on GSM8K with its context removed is 0.95 (appendix D.4, Table 5).

**CLAMBER** sorts ambiguous queries into eight kinds, not about mathematics.
- "Compared to small-scale LLMs, ChatGPT stands out as the superior model. However, it only reaches an accuracy of 54.25% and an F1 score of 52.77%." (§5.1), on a balanced test.
- "We observe that 81.97% of errors are false negatives, indicating that ChatGPT often misidentifies queries with self-contradictions as unambiguous." (§5.2).

**What paper A takes.**
- MathTrail's self-check asks the model, as a strict critic, the questions these benchmarks ask: is a condition missing, can the question be read two ways, is a negation easy to miss, is a word or a range vague, do conditions contradict each other, does the right option answer another question [C095]. Its list covers UMWP's missing information and its ranges, vague words and negations, and CLAMBER's contradictions. UMWP's unrealistic conditions have no question of their own; the nearest issue type is `factual_error`.
- The benchmarks predict how the author's own check will err: towards calling a flawed problem fine (CLAMBER, QuestBench, AbstentionBench), and by computing through a vague condition as if it were exact (UMWP's example). A problem whose whole context is missing is easy to notice (AbstentionBench); a missing or vague detail is not. A solver written under the same reading agrees with the key, so the solver cannot close that gap either; the author knows the value it meant and writes it into the program.
- MathQ-Verify is the closest work to MathTrail's admission step, and its best precision needs several different models voting, which a service that calls no model cannot do [C003]. It excludes multiple-choice questions; none of these works covers five-option tasks for grades 1–6.

## 6. Children's olympiad problems and fresh competitions

### SMART-840

**Key:** `cherian2024evaluating`. Models on every Math Kangaroo problem of 2020–2024, against children's answers.

- The data: "we consider only MK competitions from 2020-2024, that amounts to 840 problems in our dataset, dubbed SMART-840, and consisting of 240 questions all together from grades 1–4 and 600 questions from grades 5–12" (§3.2). Most need a picture: "we plot the distribution of the number of problems that need both text and image reasoning (∼69%) against those that only have text questions" (§3.2).
- Against children: "Specifically, the best accuracy of LVLMs are in the range of 40-50% while the children’s average performance is consistently near 60% or above." (§3.4). And: "the performance gap is larger (nearly 30-40%) on the tests for younger children and improves to 10-20% for higher-schoolers, iii) there appears to be a lack of any significant correlation between the difficulty of a puzzle to children against that to an AI model" (§1).
- Text only: o1-preview solved every text-only problem of grades 1–2 and 3–4 and 84.6 % of grades 5–6, but those are 10.83 %, 15.00 % and 17.33 % of the grades' problems (Table 3).
- The authors suspect memory: "But here we are seeing signs that similarity to the large mass of training examples appears to be what is driving performance across all levels of these problems." (§5).

### MathArena

**Key:** `balunovic2025matharena`. Models on competitions held after their release.

- Kangaroo was left out: "We experimented with other well-known competitions, such as Kangaroo, and excluded them as they are already saturated by existing models." (§3.1). One sentence, no figures.
- Contamination: "Most models lie above this line on AIME with a margin of 10%−20%, suggesting inflated performance on AIME 2024 due to data contamination." (§4.1). Changed problems keep it: "While this strategy reduces overlap, it does not fully eliminate contamination: perturbed problems rely on the same underlying reasoning patterns." (§2).
- Even new ones: "and find that 8 problems from AIME 2025 and 1 problem from HMMT 2025 can be found online in a similar form." (§4.1).

**What paper A takes.** Models of 2024 were weakest against children in the youngest grades, largely because of pictures, and MathTrail's tasks are text with optional text drawings [C031]. Text problems at this level look nearly solved by a reasoning model, though on few problems, which supports the premise that the chat's model can solve what it writes; it says nothing about problems it writes itself. A model's sense of what is hard need not be a child's, and MathTrail's task difficulty is the model's reading of the brief, never recalibrated from children's answers [C013]. For MathTrail contamination is a risk to novelty, not to measurement: a memorised olympiad problem would come back as a "new" task and pass, because the check compares a task only with the reference tasks and the child's own history [C028].

## 7. How alike generated texts are, and what counts as new

### The Artificial Hivemind and its re-evaluation

**Keys:** `jiang2025artificial`, `schaeffer2026strong`.

- **The claim.** Many models answering real open-ended requests, with similarity measured as the cosine of sentence embeddings of whole answers. Within one model: "Despite using high-stochasticity decoding parameters (top-p =0.9, t =1.0), responses from the same model remain highly repetitive, as shown in Figure 4: in 79% of cases, the average similarity exceeds 0.8." (§3). Across models: "the average pairwise similarity between responses from different models ranges from 71% to 82%, with some pairs notably higher." (§3). Rewording the request helps little: "Our results across all 42 models show that the within-prompt similarity averaged 0.821, while the cross-paraphrase similarity averaged 0.781." (appendix C.4).
- **The dispute.** "Under a more demanding null (same-prompt responses expressing genuinely different ideas), 20%–32% of such pairs already exceed the paper’s 0.8 convergence threshold." (`schaeffer2026strong`, abstract). Concentration is not disputed: "We do not dispute concentration. river holds 80.5% of our responses and 90.2% of the original authors’." (§2). Nor is the effect refuted: "We do not resolve whether the Artificial Hivemind is real. We show that the published evidence does not establish it." (abstract). And: "Rewording moves the embedding metrics and the modal-idea share, not the number of countable ideas." (§4).

### CreativeMath

**Key:** `ye2025assessing`. Model judges rated whether models' solutions to competition problems were new against known human solutions.

- Novelty by method: "Methodological Differences: If the methods used to arrive at the solutions are fundamentally different (e.g., algebraic manipulation versus geometric reasoning), the solutions are considered distinct." (§4.1).
- Judges unchecked against people, and wrong once in the paper's own appendix: a solution that concludes "Since the difference is divisible by 16 but not by 32, the largest power of 2 that divides 134−114 is (C) 32." (appendix A.1.2, where 134−114 is 13⁴ − 11⁴) was accepted: "The solution by Gemini-1.5-Pro is correct, as verified by all LLM Evaluators." (appendix A.1.2).

### OpenMathInstruct-2

**Key:** `toshniwal2024openmathinstruct`. New problems written by few-shot prompting from seed problems, checked against test sets.

- Why a model judge: "The most commonly used methods, such as n-gram overlap and embedding similarity search, are susceptible to simple variations in test data (e.g., paraphrasing, translation), allowing rephrased samples to bypass these basic detection techniques easily." (§3.1).
- How many: "Overall, our decontamination pipeline removes about 50K questions out of the 569K new questions synthesized (569K ⟶ 519K)." (§3.1).
- What stays: "Our dataset does have questions that are similar (but not equivalent) to MATH test set questions with sample pairs shown in Table 11." (appendix C.2).

**What paper A takes.**
- A model asked again and again for a new problem is expected to repeat itself; that is the risk the near-duplicate check exists for. The Hivemind is the work to cite for it, as a documented risk whose strength is disputed, not as a measured property of task writing. Its rates are embedding cosines of whole answers to open requests and do not transfer to trigram Jaccard on maths questions. MathTrail's own repetition has to be measured on its own output.
- Two hosts need not diversify each other, and all 603 reference tasks come from one family of models (P013, [C066]).
- A trigram check sees repeated text, not a repeated idea. Rewording changes the measure without adding ideas (`schaeffer2026strong`), and a paraphrase slips past n-gram checks (`toshniwal2024openmathinstruct`). CreativeMath's criteria show what judging at the level of the idea would need, and its unchecked judges show why a model's judgement is not enough. Paper A claims that a task is no near-copy of a reference task or of the child's past tasks, not that its idea is new.

## 8. Readability for children

**Key:** `valentini2023automatic`. Stories for preschoolers by InstructGPT, ChatGPT and Vicuna, against human preschool books.

- "in comparison to our dataset of human-generated stories for the same demographic, the models’ stories exhibit scores that are over 17% worse for all readability metrics tested (Flesch Reading Ease, Flesch-Kincaid Level, Gunning-Fog Index, and Automated Readability Index)." (§1).
- Naming the age: "we find that age-specific prompt-tuning has little effect on the simplicity of children’s stories generated by LLMs." (§4.5).
- Words: "Further, we find that none of the 750 total generated stories stayed within the age range of 6 or younger." (§4.5), measured by the age at which children learn a word.
- The authors on their measures: "These measures, although well-established and widely used, are coarse oversimplifications of language use." (§4.4).

**What paper A takes.** Readability is to be checked, not trusted to the grade named in the brief, and MathTrail checks it [C030]. Its check has no word-level part, where this study's sharpest signal was. The evidence is thin: models of 2023, English stories for ages 2–5, no maths questions; no work read validates sentence length or Flesch–Kincaid on short maths questions, or in another language.

## 9. Finding near-duplicates

**Keys:** `broder1997resemblance`, `broder2000minwise`, `lee2022deduplicating`.

- **The measure.** Broder defines the resemblance of two documents as the Jaccard index of their sets of shingles, runs of w tokens (p. 3). Shingle size trades two risks: "on the other hand a large size is possibly over-sensitive to small alterations since the change in one token affects w shingles." (p. 4).
- **The estimate.** The chance that two sets share their minimum under a random permutation is their resemblance (`broder2000minwise`, p. 4, equation 5), so a sketch of minima estimates it: "Hence, we can choose say, 100 independent random permutations π1, . . . , π100." (p. 4). Real hash families are only approximately min-wise, and one used in practice "is not min-wise independent, its performance should be sufficient in many practical situations" (p. 22).
- **Its error.** "For a fixed s our estimate is likely to be worse if r is close to 0.5, but often we are interested only in the case where r is larger. For practical applications, 100 samples seems reasonable and 200 seems more than enough." (`broder1997resemblance`, p. 7).
- **Thresholds in practice.** AltaVista clustered web pages: "We calculated our clusters based on a 50% resemblance." (p. 2). Deduplicating training data: "In our implementation, we use 5-grams and a signature of size 9,000." (`lee2022deduplicating`, §4.2), and "For document pairs that were identified as potential matches, we computed their actual Jaccard index, and if that was above 0.8, we computed their edit similarity. Document pairs with edit similarity higher than 0.8 were identified as duplicates." (appendix A). What they found: "The text is identical except for the names of places, businesses, products, dates, and so on." (§5.2).

**What paper A takes.** MathTrail's measure is Broder's resemblance over the character trigrams of padded words [C027]. Its sketch keeps one minimum for each of 192 hash functions, the scheme of the min-wise paper, but only four bits of each [C029]; truncation is not treated in these papers (b-bit minwise hashing, a lead not read). MathTrail's hash family is not proven min-wise independent, so the error the draft measures on real question pairs [C050] is the right evidence for it. Its threshold of 0.7 is its own: the classic thresholds belong to whole documents and to word shingles and do not carry across. Lee et al.'s near-duplicates that differ only in names and dates are what MathTrail's check is meant to catch, as is the same task with new numbers: by the author's decision (Q78), a copy with its people and objects renamed is a copy, while a plot written anew is a new task and passes by design.

## Answers

**1. Who writes, who checks, and how often generated problems fail.**
- Models write; people, model judges or other models check. Teachers judged MATHWELL's and EDUMATH's problems, a model judge filtered MathWiz's, model agents Ikram et al.'s, the generators themselves and then people MATH²'s, other models CHASE's, a model's score and an arithmetic program KPDDS's. Two works take their keys from programs: MATHWELL from the model's own, GSM-Symbolic from programs people wrote.
- Failure rates are not comparable across these works, but none is near zero: a quarter of MATHWELL's problems failed some criterion; EDUMATH's models from 5 % to 36 %; MathWiz missed the grade one time in five; KPDDS kept one wrong problem in twenty after every filter; CHASE seven in a hundred after verification by other models, against 34 in a hundred without it.
- No work read runs a program its author wrote against a fixed set of options, twice with the labels moved, or tries every option. MATHWELL is closest in mechanism, its program being the key; GSM-Symbolic in purpose, its program written by people; Zhou et al. in the check, a model verifying its own answer with code, but the model itself reads the result, and the problems are given, not written.

**2. Which checks work.**
- *A program* fixes arithmetic, not reading. Its remaining errors are wrong values and wrong logic (PoT), and output agreement is not proof of a right program (PAL).
- *A model's check of its own work* is weak: without outside feedback it does not improve answers (Huang et al.), and finding the mistake, not fixing it, is the bottleneck (Tyen et al.). A model evaluator tracked people poorly on unsolvable elementary problems (MathWiz); nobody checked the solvability verdicts of Ikram et al.'s model agents; humans still rejected or edited many problems frontier models had validated (MATH²).
- *Detecting an ill-posed problem* is better with an explicit list and with several models: MathQ-Verify's best single model still accepted 137 of 848 flawed questions. Models lean towards treating a flawed question as fine and answering it as if it were complete or exact (CLAMBER, QuestBench, AbstentionBench, UMWP).
- So MathTrail's division of labour is supported as far as the literature goes: the program checks what can be computed, the self-check is a prompt, not a gate on its own, and neither sees a misreading the author shares with its program [C026]. How often that happens in MathTrail is unknown; RQ1 measures what the checks catch by injecting faults, not this rate.

**3. Children's olympiad problems and contamination.** Models of 2024 were below children on Kangaroo problems, most for the youngest grades and the pictures; text problems of the early grades look solved by a reasoning model, on few items. Kangaroo is called saturated by a benchmark that gives no figures. For MathTrail contamination means a memorised problem could return as new and pass, since the check never compares tasks with the public olympiad corpus [C028].

**4. Homogeneity, novelty and near-duplicates.** Repetition by one model is a documented risk, its strength contested, its rate for maths problems unmeasured. Novelty is measured at three levels: text (shingles, n-grams), meaning (embeddings) and idea (model or expert judges). No validated automatic measure of "the same idea" exists for problems among the works read. MathTrail measures text only, by Broder's resemblance with a min-wise sketch, at a threshold of its own.

**5. Readability.** Generated text for young children reads harder than human text, and naming the age changes little; this supports an explicit check. The formulas are coarse, unvalidated for short maths questions, and English.

**6. The trap catalog.** Seventeen of the 20 traps have counterparts in Eedi's graph, close or exact for most. Three — the worst case, knights and liars, the first move of a game — have none: they belong to olympiad topics. The common primary-school mistakes no trap covers are listed above: place value, calculation slips, the order of operations, fractions, factors and multiples, clocks, perimeter and area on a grid, Venn diagrams, and Newman's decoding errors and careless slips. S62 reports the gap among the threats to validity (Q41).

**7. What paper A may claim.**
- Among the works read, none generates olympiad-style problems for primary school, and none admits a problem by running a program its author wrote against fixed options. The two nearest were read for this: Ikram et al. validate generated problems with model agents, and Zhou et al. let a model check its own answers with code. This is a finding of these searches, which S35 weighs as a claim of novelty.
- The design follows the evidence: programs for what can be computed, no reliance on the author's own judgement, an explicit checklist rather than an open question.
- It claims no rate of missed misreadings, no novelty of ideas and no learning.

## The prototype's claims, checked

The prototype's research notes reported these from sources an auxiliary model had retold (P036). Read in the original, they stand as follows. Paper A cites the works, never these claims.

| Claim | Verdict | Where |
|---|---|---|
| P037: no work generates olympiad problems for grades 1–4 | Holds within the searches of this task, for grades 1–6 | answer 1; `search-log.tsv` |
| P038: SMART-840, 840 problems, 240 for grades 1–4, 69 % with a picture, weakest on the youngest grades, no difficulty correlation | Holds. Models doing worst on the youngest grades is true of their gap to children; some models score about the same across grades | section 6 |
| P039: EDUMATH, about 90 % solvable, 77 % with a correct answer, 66 % meeting every criterion | **Wrong.** These are the rates at which two annotators agreed, not shares of problems; EDUMATH's MaC rates per model run from 63.9 % to 94.6 % | section 1, EDUMATH |
| P040: GSM-Symbolic, changing only the numbers lowers accuracy; one clause drops it by up to 65 % | Holds as the authors state it; "every model" is the note's phrase | section 1 |
| P041: models are homogeneous alike, so switching models does not restore diversity | Supported by the source, contested by a re-evaluation | section 7 |
| P042: OpenMathInstruct-2 removed about 50 of 569 thousand new problems (about 9 %) as paraphrases of test problems | Holds; 8.8 %, as judged by a model over each problem's nearest test problems | section 7 |
| P043: Eedi distractors, plausibility 2.68 against 3.72; three similar problems as examples worked best | Holds (Feng et al., Table 7; kNN best) | section 2 |
| P044: two models err alike, more within a provider; judges favour their own family | Not read here (leads: Kim et al. 2025; Panickssery et al. 2024) | Remarks |
| P045: models recognise ambiguity but do not say so unless asked | Not read here; AbstentionBench reports the related behaviour of answering definitively | section 5 |
| P046: PAL beat chain-of-thought PaLM-540B on GSM8K by 15 points | Holds: "surpassing PaLM-540b which uses chain-of-thought by absolute 15% top-1" (`gao2022pal`, abstract); PAL ran on Codex | section 3 |
| P051: texts models wrote for children scored more than 17 % worse on every readability metric | Holds, for preschool stories by models of 2023 | section 8 |
| P053: on USAMO 2025 proofs, strong models scored in the single digits | Not read here; MathArena reports proofs below the IMO bronze level only in passing | Remarks |

## What this changes in paper A

Made in the draft with this task, in both languages:
- The introduction cites Gupta et al.'s 56.6 % as checked, with its setting: 150 tutoring dialogues in college algebra.
- The related work names the works it cites, with keys checked by `citecheck`, and says what each found that bears on the design: a model evaluator tracked people poorly on whether problems for grades 1–6 could be solved; validator agents for generated problems are models; self-review finds few mistakes; models answer underspecified problems; repetition is a risk, contested. It states the gap found here, within these searches.
- Section 4.3 cites Broder for the measure and the min-wise paper for the sketch; section 4.4 cites Valentini et al. for why readability is checked; section 4.5 says what the literature predicts the checks cannot see, with UMWP's example.
- The discussion no longer says that every check is independent of the model's judgement of its own work: the self-check is that judgement, and the solver, though the service runs it and judges its result, shares the model's reading of the task.
- The provisional reference list follows `refs.bib`; EDUMATH is cited as the ACL 2026 paper, `christ2026edumath`.

Left to the tasks that write those sections:
- S59 describes the self-check as a prompt, not a verification, and the trap check as naming, not computing, the trapped answer.
- S62 lists the traps' gap against Eedi and Newman among the threats to validity (Q41), and says that difficulty and topic fit are not checked.
- S35 weighs the gap of answer 7 as a claim of novelty.
- S37 may inject a paraphrase of a reference task and a same-template task with new numbers, to show the near-duplicate check passes the first and refuses the second.

## Remarks

- **How the notes were made.** The reading was split three ways, each part read by a separate model session that saved the full texts and checked its quotations by script; the executor read Gupta et al. itself, and Ikram et al. and Zhou et al. after the first review found them missing, read the three reports in full and checked every quotation of these notes again against the saved texts. Percentages derived from a paper's counts were recomputed here.
- **Two numbers left out.** While reading, the near-duplicate measure was re-implemented in Perl and run on the papers' example pairs: OpenMathInstruct-2's paraphrase pair scored 0.534 and its same-template pair 0.74–0.79; Lee et al.'s template pair 0.654 by trigrams against 0.288 by word 5-grams. The binomial error of a 192-position sketch at a true index of 0.7 was put at about 0.035. They would support section 4.3, but they come from a copy of the measure, not the product's code, so no sentence of the paper rests on them; S37 can recompute them with the product's own packages.
- **Not checked.** The novelty of ideas, the plausibility of MathTrail's options to children, the fit of a task to its topic and difficulty, and how often the author's check finds a defect in its own task: no work read measures them for problems like MathTrail's.
- **Version and record mismatches.**
  - Crossref's record of the Artificial Hivemind's NeurIPS version names nine authors and leaves out Alon Albalak, the ninth of the ten on arXiv. S61 cites the arXiv version, 2510.22954, the one read, whose DataCite record names all ten.
  - Crossref records no year for Broder 1997; the entry carries the year of the conference its proceedings name.
  - The arXiv id the reading was given for SMART-840, 2406.12849, is another paper; `seed.bib` has the right one, 2406.15736.
  - PAL's table and text give different GSM-Hard figures; CLAMBER's table and text disagree on whether chain-of-thought was used; MathWiz's prompt inverts its detector's answer; MathQ-Verify's 15 % gain is not in its tables.
- **Copies, not the published versions.** Gupta et al., Ikram et al., Broder 1997 and the min-wise paper were read in the authors' copies; their pages may differ from the published ones. Most arXiv versions read were not compared with the published ones, as the table says.
- **Leads not followed.** The ones paper A may need first:
  - Li and König, b-bit minwise hashing (WWW 2010): the estimator for minima kept to a few bits, before paper A states the sketch's error formally.
  - Kim et al., *Correlated Errors in Large Language Models* (ICML 2025), and Panickssery et al., *LLM Evaluators Recognize and Favor Their Own Generations*: for P044.
  - Zheng et al., *Large Language Models Are Not Robust Multiple Choice Selectors* (ICLR 2024): option bias, which bears on the rotated labels.
  - TreeCut (ACL 2025) and *When Not to Answer* (2024, 2026): unanswerable word problems and prompts for abstention.
  - Clements (1980) and Newman (1977), the primary sources of the error analysis; NCETM's catalogue of primary misconceptions, found only in secondary copies.
  - On homogeneity: *Large language models are homogeneously creative* (PNAS Nexus 2026, in `seed.bib`), and Doshi and Hauser (Science Advances 2024).
  - *Beyond Benchmarks: MathArena as an Evaluation Platform* (arXiv 2605.00674, in `seed.bib`), which may give figures on Kangaroo.
  - Vajjala (2022) on the limits of automatic readability assessment.

## Added in S61

- **A jury of models** (`verga2024replacing`, the abstract, arXiv:2404.18796, read 2026-10-01). Paper A cites it for one thing: that the output of a model can be judged by a panel of other models rather than one. The abstract: "We propose instead to evaluate models using a Panel of LLm evaluators (PoLL). Across three distinct judge settings and spanning six different datasets, we find that using a PoLL composed of a larger number of smaller models outperforms a single large judge, exhibits less intra-model bias due to its composition of disjoint model families, and does so while being over seven times less expensive." It judges models' answers, not the problems a model writes, so the draft no longer says it judges questions. The quotation was taken from the abstract page as fetched that day; no full text was saved, so the script that checks S31's quotations against the saved texts did not check it.
