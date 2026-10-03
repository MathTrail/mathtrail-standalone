# Premises of the Gemini prompt

The research program began with a prompt written by Gemini ([RUN.md](../RUN.md), appendix A). The prompt assumes much about MathTrail that the repository does not bear out. This file checks each of its 40 premises against the evidence ledger before anything is built on it, maps the prompt's vocabulary onto terms a reviewer can check against the code, and checks its two formulas. S04 wrote it.

Evidence is cited by ledger ID:
- `C…` are claims about the product, in [ledger.md](ledger.md), at commit `3597c6a0633d`, where S04 read them; the ledger has since moved to `52ce86908135` (S02.1), and the last section of this file follows it;
- `P…` are claims about the prototype, in [prototype.md](prototype.md), at commit `02638353482e` of its public repository;
- `S…` and `T…` are tasks of the research plan and of the product plan; a premise about method is shown by the plan itself.

**Verdicts.** A premise is about one of three things, and its verdict reads accordingly. Each row has one verdict; where the two papers differ, the last column says how.

- **About MathTrail or its prototype:**
  - *true* — the evidence shows it;
  - *false* — the evidence shows otherwise;
  - *partly* — part holds and part does not, and the row says which;
  - *future* — nothing exists yet, and the named task would create it.
- **About the literature:**
  - *true* — the literature exists and bears on MathTrail as the premise says;
  - *partly* — it exists but bears differently, or on paper B only;
  - *future* — a literature task decides.
- **About the method of writing:**
  - *true* — the plan keeps it;
  - *partly* — the plan keeps part of it.

## Premises

### Role and goal

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 1 | The paper suits ACM Transactions, IEEE TPAMI, NeurIPS or AIED | partly | [literature/venues.md](../literature/venues.md); Q28 | A goes to AIED 2027 and B to EC-TEL 2027 (S01). TPAMI is off-topic. |
| 2 | One flagship paper covers the system and the textbook | false | C038–C042: much of the product is unbuilt; the textbook has no design yet, and S20–S29 make one | Two papers: A on what is built, B on the textbook's design, formal model and simulation. |
| 3 | The materials include technical benchmarks | false | C049–C054 are calibrations and single measurements; P019–P020 are two pilot runs of 10 tasks | The evaluation is built: E-A1–E-A4 (S37–S39, S43), E-B1–E-B3 (S50–S54). |
| 4 | MathTrail is a distributed, LLM-based multi-agent system | false | C002: one binary; C003: no model calls of its own; P001–P003: the prototype's agents were removed | A: one host model and deterministic checks. A panel of judge models is paper B's subject (S25). |

### Phase 1: audit of the repository

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 5 | Backend services in Go and Python | partly | C001; the ledger's "What the product does not contain"; P003 | A describes one Go binary. Python was the prototype's. |
| 6 | API contracts and MCP integrations exist | future | C004; C038–C041; C062 | Until T41–T45 are built, A describes the MCP tools as specified. S56 pins the commit A describes. |
| 7 | Database schemas: PostgreSQL and vector databases | partly | C002; P056 | The product has no database: the profile is a file in the parent's Drive (C040, planned). PostgreSQL was the prototype's and belongs to the design history. No vector database ever existed. |
| 8 | Streaming pipelines | false | C002, C062: every request stands alone; P057 | None. The only "streaming" in MathTrail is the name of the MCP transport, Streamable HTTP. |
| 9 | The research folder holds benchmarks, test scripts, evaluation logs, prompt templates and measurement history | partly | P036: the notes summarise literature; P026: no quality verdict on record; C046: the model's instructions; C049–C054, P018–P024: measurements | The ledger records what exists. Everything measured from now on comes from a saved script (G2). |
| 10 | Architecture decision records, a system topology, state machines and multi-agent loops | partly | C063: diagrams, including the task's state machine, and a decision log; P058, P028–P035: the prototype's diagrams and decision log; C003, P002: no agent loop | A cites the diagrams and the decision logs. It describes no agent loop. |

### Phase 2: literature

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 11 | A literature of LLM task generation and verification: "deterministic execution subgraphs", hallucination prevention in maths | partly | P037: no work found on olympiad problems for grades 1–4; P038–P046: leads, unchecked | S31 reviews it. The prompt's term is replaced (term map). |
| 12 | A literature of IRT and Glicko-2 or Elo, applied to learners and items | partly | It exists, but the product uses a Rasch model with a guessing floor and Elo-style updates, no Glicko (C009–C011) | S32. Glicko-2 becomes a baseline in simulation (S38). |
| 13 | A literature of distributed multi-agent memory and consensus: MemGPT/Letta, AutoGen, Zep, Byzantine fault tolerance, majority voting | partly | It bears on paper B only: the product has no agents and no memory of its own (C003); P044, unchecked | S14 and S15 for paper B. S14 examines the Byzantine analogy rather than assuming it. |
| 14 | The work can be benchmarked against the state of the art, searched in arXiv, the ACM Digital Library, IEEE Xplore and Springer | partly | No comparable system or benchmark is known yet (P037, unchecked) | S08 sets the search protocol over those databases. A compares with baselines it builds (the ablations of E-A1, the baselines of S38), not with a leaderboard. |

### Phase 3: questionnaire

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 15 | A verification suite of more than 120 edge cases exists | false | The ledger's "What the product does not contain"; C047 counts the product's tests | S37 builds one: at least 120 injected defects in at least 12 classes. |
| 16 | There are latency trade-offs to report | partly | C024–C025: the solver's limits and costliest run; E-A4: a review's time, its solver's runs and the package, measured on a development machine (S39, Q76); C086, C101: an instance's cost, cold start and pace; P020: 44–109 s per task in the prototype | S39 measured a review locally. The deployed service waits for T64; generation time comes from E-A2 (S43). |
| 17 | The pipeline is a generator–solver–verifier loop of agents | false | C003, C021–C023, C032, P005 | A describes one model that writes the task and its solver, and a service that checks both, with up to three attempts. |
| 18 | There is a Glicko-2 or Elo adaptation algorithm to detail | partly | C009–C016 | A formalises the Rasch model with a guessing floor and its Elo-style updates (S57). |
| 19 | There is an error taxonomy | true | C020, C043, P030 | A: the closed trap catalog. |
| 20 | There is a diagnostic memory scheme | partly | C029, C034: a window of 20 answers and fingerprints; C040: the profile in Drive, planned | Until T50, A describes the profile as specified. |
| 21 | A fine-tuning cluster | false | C003; P033: the prototype never fine-tuned; P049: the terms, unchecked | None. |
| 22 | Local models compared with API latency | false | C003: the service calls no model | Open models appear in paper B only, as judges (S07, S53). |
| 23 | Throughput figures | future | C005: the Cloud Run limits | S39. |
| 24 | A future architecture of distributed memory, voting thresholds and textbook mutation rules | future | Nothing built | Paper B designs them (S23–S25) and defines them formally (S29). |

### Phase 4: lexicon

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 25 | Its eight terms name real mechanisms | partly | See the term map | Only mapped terms are used, where the map allows. |
| 26 | The paper uses formal LaTeX notation throughout | partly | Rule G4 of the plan | Formulas only where they carry meaning (G4). Both papers share one notation (S29, S57). |

### Phase 5: formulas

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 27 | P(correct) = σ(γ(θ_i − β_j)) with vectors θ_i and β_j is the rating model | partly | C009–C013, C016 | See the formula check. S57 formalises the real model. |
| 28 | There is an update rule for task difficulty β_j | false | C013: β is fixed per task and never updated; P009: the prototype had one, for its bank | A says so and why (S57). |
| 29 | A revision is committed when Σ w_m v_m ≥ τ | future | Nothing built | The starting point of S25. See the formula check. |

### Phase 6: writing

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 30 | Write section by section, each approved before the next | true | The plan, phases 5 and 6 | Kept: the author reviews each section. |
| 31 | The title and abstract stress multi-agent deterministic verification, adaptive skill tracking and consensus-based textbook evolution | partly | C003, C021–C023, C009–C011 | A: deterministic verification and the learner model. B: the textbook. |
| 32 | The introduction rests on equal access to olympiad maths and on base LLMs failing in education | partly | The author's problem statement (RUN.md); P038–P040, P051, unchecked | A starts from procedures against non-standard problems (S60). It cites failures of base LLMs from checked literature only. |
| 33 | Related work covers AIED, LLM verification and multi-agent memory | partly | The plan: S31–S35 for A, S15 for B | S31–S35 for A. Memory belongs to B. |
| 34 | The architecture section covers a Go backend, Python agents, a verifier of 120 edge cases and MCP integration | partly | C001 for Go; nothing for Python agents or the 120 cases; C038: MCP planned | A's system section describes what is built and marks what is specified or planned (S59). |
| 35 | The modelling section formalises an Elo/Glicko-2 integration and tracks the error taxonomy | partly | C009–C011; C020: the child's own traps shape the next brief | S57 formalises the model the product has. |
| 36 | The evaluation reports empirical data, error reduction, generation latency and throughput | future | P025–P027: the prototype measured little; C049–C054 | E-A1–E-A4 (S37–S39, S43–S44). |
| 37 | The vision: an autonomous consensus-driven textbook engine, optimised in real time from learners' error telemetry | future | C035, C042: the telemetry is aggregate only, and not built | Changed, not only postponed. Paper B decides each revision offline, and the only online test compares the evolving book with a frozen one. No trial on children before an ethics review (Q12, K07). |
| 38 | The ethics section covers COPPA and FERPA, the democratisation of open-weight models and a socially oriented funding model | partly | P050: COPPA, unchecked; C003: the service runs no model, so it democratises none; no claim on funding yet | S16 checks the law. S63 writes the ethics section. How the project is funded is the author's to state (Q53). |
| 39 | A conclusion with future work | true | The plan: S63, S74 | Kept. |

### Protocol

| # | The prompt assumes | Verdict | Evidence | What the plan does |
|---|---|---|---|---|
| 40 | Confirm, summarise the codebase, ask the questionnaire, and wait for the answers before any outline | true | The plan: S02–S03 summarise, S05 asks, S56 and S67 outline | Kept, with the questionnaire rebuilt on facts. |

The verdicts were given at the ledger's first pin, `3597c6a0633d`, and the rows keep what S04 found. At `52ce86908135` (S02.1) several no longer describe the product:
- premises 2, 6, 7, 20 and 34 waited for the product: the MCP tools, the profile in the parent's Drive and the widget are built (C038–C041, C068, C078), and paper A describes them as built. Rows 7 and 20 cite the profile in Drive as C040, which then named sign-in and Drive together; C040 now names sign-in, and the profile in Drive is C078;
- premise 37 cited the lesson's log lines as not built: they are (C035), and still aggregate only;
- premises 3, 16, 23 and 36 found no measurement of latency or throughput: the first runs of a real model and the load measurements of one instance now give some (C084–C086), though still no benchmark in the prompt's sense, which the experiments of S37–S39 build.
## Term map

A term stays in a paper only where it names a mechanism a reviewer can find in the code or in the paper's own definitions. The author approves this map (S04).

| The prompt's term | Precise term | Where it may be used |
|---|---|---|
| Deterministic Hallucination Mitigation Subgraph | deterministic verification pipeline: the service's checks of structure, a sandboxed Starlark solver run twice, the self-check, readability, near-duplicates and text drawings (C021–C023) | A. It never claims to mitigate hallucination: the checks test the model's formalisation of the task, not the meaning of its text (C026). |
| Dynamic Knowledge Tracing via Multi-Dimensional Item Response Theory / Glicko-2 Rating Vector | a Rasch model with a guessing floor: an overall level plus per-topic offsets, updated online in the Elo style with decaying steps (C009–C011) | A. Never "Glicko". "Multidimensional" only as "an overall level plus topic offsets". "Knowledge tracing" only in the broad sense of tracking a level from answers. |
| Formative Diagnostic Error Taxonomy & Scaffolding | a closed trap catalog, where every wrong option names the mistake it catches (C020, C043), and a hint on request: one leading question or the first step (C061) | A. The catalog may be called diagnostic, since each wrong option names the mistake it catches. "Formative" is not used: no task defines it. "Scaffolding" is not used either: a hint written in advance is neither adapted to the child nor withdrawn as the child learns, so it is called a hint (Q62). |
| Zero-Friction Ubiquitous Pedagogical Integration | in-chat delivery through MCP Apps widgets (C038, C068) | A. The tools, the widget build and the task widget are built at `52ce86908135` (S02.1). Never "zero-friction" or "ubiquitous". |
| Closed-Loop Epistemic Memory Pipeline | none yet; working name "a closed loop from aggregate learning signals to textbook revisions" | B, once S29 defines it. |
| Heterogeneous Multi-Agent Consensus Protocol | a panel of judge models from different families, with weighted votes and a calibrated threshold | B, once S25 and S29 define it. |
| Distributed Agentic Long-Term Memory Network | a versioned, content-addressed store of the textbook, with its lineage and an event log | B, once S23 and S29 define it. "Distributed" and "network" only if the design turns out to be either. |
| Adaptive Curriculum Mutation Subsystem | the variation step: one named, size-bounded operation per revision of a textbook section | B, once S24 and S29 define it. The order of topics, the curriculum, is not what changes. |
| multi-agent system; agents | A: the host's model. B: judge models. | A: nowhere. B: only for the judge panel. |
| generator–solver–verifier loop | the model writes the task and its solver, and the service checks both, with up to three attempts (C032) | A |
| Byzantine fault tolerance | none | Nowhere as a claim. S14 examines the analogy. Byzantine agreement makes processes agree although some are faulty, under stated assumptions about how many fail and how messages travel. Judges do not need to agree, and what makes them wrong together is correlation, which no such assumption describes. |
| 120+ edge cases | a suite of N injected defects in K classes | A, after S37, with the measured N and K. |
| hallucination prevention | detection of some wrong tasks | A, with the rates E-A1 and E-A2 measure. |
| real-time curriculum optimisation | none | Nowhere: B decides offline (S26). |
| Python agents | none | Only in the prototype's history: its first design had four agents on the Claude API, removed before their first live call (P001, P002). |
| fine-tuning cluster; vector database; streaming pipelines | none | Nowhere: neither the product nor the prototype had any (C003, C062, P033, P056, P057). |
| flagship; high-impact; SOTA | none | Nowhere. |

## Formula check

### The rating model

The prompt writes

$$P(\text{correct}\mid\theta_i,\beta_j)=\sigma\bigl(\gamma(\theta_i-\beta_j)\bigr),\qquad \theta_i,\beta_j\in\mathbb{R}^k .$$

The prompt calls the dimension d; it is k here, because d is the product's name for a task's difficulty. The product computes (C009–C011)

$$P=c+(1-c)\,\sigma(\theta+\delta_t-\beta),\qquad c=0.2,\quad \beta=d-3,$$

where:
- θ is the child's overall level and δ_t the offset for the task's topic t;
- d is the task's difficulty, from 1 to 5;
- c is the guessing floor of five options with no penalty.

After an answer S ∈ {0, 1}:
- θ moves by K_θ(S − P), with K_θ = 0.2/(1 + 0.05 n) and n the child's answers so far;
- δ_t moves by K_δ(S − P), with K_δ = 0.4/(1 + 0.05 n_t) and n_t the answers so far in topic t.

This is the model at the ledger's first pin, `3597c6a0633d`. At `52ce86908135` (S02.1) a task stands on one ladder for grades 1–6, β = s_ℓ + (d − 3) with level shifts s_ℓ of 0, 2.5 and 5 (C010); the first five answers are a trial series, which sets θ to the most likely level given the start and all its answers and leaves δ_t alone; and the steps above apply after it, n counting the trial answers too (C011, C080). Nothing below depends on how β is made: the six differences and the check of the step are about x = θ + δ_t − β, and hold on the ladder as they did within one level. S57 formalises the model at the final pin.

The two formulas differ in six ways:

1. **The guessing floor.** The prompt has none. Five options and no penalty put a correct answer by chance at 0.2.
2. **γ = 1.** The model works on the logit scale. The chess number the child sees is only a display: Elo's expected score 1/(1 + 10^(−ΔR/400)) equals σ(ΔR · ln 10/400), so a logit x shows as R = 1500 + (400/ln 10)·x, about 1500 + 173.7·x (C016). "Elo" names this display and the style of the update, not a rating system with a model of its own.
3. **Dimensions.** For k > 1, σ(γ(θ_i − β_j)) is not defined: a vector difference has to become a number first. Multidimensional IRT does this with a loading vector, σ(a_j·θ_i − b_j). The product's θ + δ_t − β is the case of one overall dimension and one dimension per topic, each task loading 1 on the overall dimension and 1 on its own topic. That is the form of the Rasch testlet model of Wang and Wilson (2005) — a lead for S32, not yet checked.
4. **β is not a rating.** A task's difficulty is fixed and never updated, because every task is written for one child and never reused (C013). The prototype did update β, because it had a bank (P009). The prompt's "task difficulty rating" does not exist in v1.
5. **No rating deviation.** Besides the rating, Glicko-2 keeps two numbers for each player:
   - a deviation, which shrinks with the information each game carries and grows again while the player is away;
   - a volatility, which tracks how fast the true level moves.

   The product keeps neither. Its steps shrink with the number of answers and never grow again, so after a long break they cannot widen. S05 asks the author about a floor for the step (О-53).
6. **The step is not the likelihood gradient, but it aims at the same level.** The log-likelihood of one answer is ℓ = S ln P + (1 − S) ln(1 − P). With x = θ + δ_t − β:

   $$\frac{\partial\ell}{\partial x}=\frac{S-P}{P(1-P)}\cdot\frac{\partial P}{\partial x},\qquad \frac{\partial P}{\partial x}=(1-c)\,\sigma(1-\sigma)=\frac{(P-c)(1-P)}{1-c},$$

   so

   $$\frac{\partial\ell}{\partial x}=(S-P)\cdot w(P),\qquad w(P)=\frac{P-c}{P(1-c)} .$$

   The product's step K(S − P) is the gradient step divided by w(P), for a right answer and a wrong one alike.
   - Its expected value, K(P_true − P), vanishes exactly where the gradient's does: both aim at the same level, and dropping w adds no bias.
   - What changes is the size of the step. It is 4.6 % larger than the gradient step at P = 0.85 and 12 % larger at P = 0.70, the ends of the corridor; 33 % larger at P = 0.5, and 2.4 times at P = 0.3. Above the corridor the difference fades: 1.3 % at P = 0.95.
   - So the level moves further, and more noisily, after tasks harder than the corridor, which the rule seldom gives.

   This is our derivation. S57 derives it again for the paper, and S38 measures the noise it adds.

### The consensus rule

The prompt commits a revision ΔT_k when Σ_m w_m v_m(ΔT_k) ≥ τ, with votes v_m ∈ {0, 1}.

Suppose each judge m votes independently once the truth is fixed. Let s_m be its sensitivity, the chance it approves a good revision, and t_m its specificity, the chance it rejects a bad one. The log-likelihood ratio of a set of votes is then

$$\Lambda=\sum_m\Bigl[v_m\ln\frac{s_m}{1-t_m}+(1-v_m)\ln\frac{1-s_m}{t_m}\Bigr]=\sum_m v_m\ln\frac{s_m t_m}{(1-s_m)(1-t_m)}+\sum_m\ln\frac{1-s_m}{t_m},$$

and the decision that minimises the expected cost commits when Λ exceeds a constant set by the prior and by the costs of a false accept and a false reject. That is exactly the prompt's form, with:

- $w_m=\ln\frac{s_m t_m}{(1-s_m)(1-t_m)}$;
- τ equal to that constant minus $\sum_m\ln\frac{1-s_m}{t_m}$.

A judge that is right with the same probability p_m either way gets $w_m=2\ln\frac{p_m}{1-p_m}$. That is the weighting of Nitzan and Paroush (1982) up to a constant factor, a lead S14 checks.

Two consequences:
- **Nothing is free.** The weights and the threshold follow from each judge's measured quality and from the cost of a false accept. They are not chosen by hand.
- **Correlation breaks the rule.** Judges whose errors are correlated (P044, unchecked) break the independence it rests on. The plan measures that correlation and calibrates τ on held-out data (S25, S53).

## What can be claimed, and when

Every claim of the ledger falls in one of these groups, by its status. The product's groups follow the ledger at `52ce86908135` (S02.1, 2026-09-30): the endpoint, the tools, sign-in, the profile in Drive, the limits and the widget, which S04 found planned, are built there, and so are the ladder and the trial series.

- **Now, as built at the pinned commit of the product:**
  - the architecture: one stateless binary with no database, no model calls, the MCP endpoint and the widget's resource (C001–C008, C062, C067, C068);
  - the learner model and the rule: the ladder, the start, the trial series, topics within reach, mastery held at a level, the model's own choice, the parent's notes the rule never reads, and the profile's history window (C009–C016, C018–C020, C034, C065, C070, C073, C080–C083, C097);
  - the checks and the sandbox (C021–C024, C027–C032, C061, C071, C072, C074, C079, C095);
  - answer sealing, the lesson's log lines and redaction, what reaches the chat's model, which identifiers remain, the card the task is drawn by, and the card's field for the child's own question (C033, C035–C037, C064, C069, C092, C093, C096, C100);
  - the tools, sign-in, the profile in Drive, the limits and the languages of the cards (C038–C041, C078, C087);
  - the content (C043–C046, C066, C077);
  - the product's tests (C047–C048);
  - the golden vectors from the prototype (C059–C060).
- **Now, as another document's measurement:** C025, C049–C052, C075, C084–C086, C088, C089, C101 — among them the first runs of a real model through the service (C084, C085, C089) and the pace of one instance (C101).
- **Built, and used by no paper:** C017, the pace of an answer, which paper A leaves out (Q49).
- **Measured, and used by no paper:** C053, the trace of one request the service answered with Not Found, which says nothing of a task, and C054, the packages' sizes for an ordinary profile, which E-A4 now measures at every point and turn (S39, Q76).
- **Now, as design only ("is specified"):** C026, C063, C076; C091, the privacy policy's promise of no advertising and no training, which paper A states as the policy's promise; and C098, the terms that make an adult the service's user.
- **After a product task:** usage evidence (C042, T67.1); acceptance in Claude and ChatGPT and the measurements of the deployed service (C090, T62–T64).
- **In no paper as a fact:** the remarks for the author (C056, C057, C094, C099; C055 and C058 were retired when the product caught up with them). C099, the key in the arguments a card is handed, is a known limitation by the author's decision (K10). C056 tells A's learner-model section what not to claim: that the pace changes the next task, or that "I don't know" does so in any way but as a wrong answer — it raises the run of failures, which makes the rule reinforce the topic, exactly as a wrong letter does.
- **Now, as the prototype's design history:**
  - what its code or data show (P003, P005–P009, P011, P012, P026, P056);
  - what its documents state (P001, P002, P010, P013–P017, P025, P027–P036, P057, P058).
- **Now, as the prototype's own record of a run:** P018–P024.
- **After the literature task that checks the original** (P004, P037–P055, now unverified):
  - S31: P037–P042, P045, P046, P051, P053;
  - S32: P047, P048;
  - S33: P054, P055;
  - S34: P004, P052;
  - S12: P043;
  - S14: P044;
  - S16: P050;
  - S07: P049.
- **After a research task:**
  - acceptance and false-accept rates (S37, S43, S44);
  - the learner model against baselines (S38);
  - performance (S39);
  - novelty and diversity (S43);
  - anything about the textbook (S29, S50–S54).
- **Never:**
  - Glicko or a rating deviation in the product;
  - multi-agent verification;
  - fine-tuning;
  - a vector database;
  - a suite of more than 120 edge cases as something that already exists;
  - textbook changes in real time from children's data.
