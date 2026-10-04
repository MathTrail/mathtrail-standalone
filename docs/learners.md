# Learners

How the service's student model does with children whose true level is known: simulated children answer tasks through the service's own code, and the bench measures how each way of estimating a child places them, keeps their tasks in the corridor, declares topics mastered, follows a child who changes and moves what the child is shown. A change to the rating, the rule or mastery is measured here before it ships, and a new rule is chosen here by the criterion below, written before any candidate ran (R163).

## What it is, and where it comes from

The bench in `tools/learners` is a Go module of its own that imports the product, as the load tool does (R124, R155). It was carried over from the research program's offline experiment with simulated learners, `research/experiments/learnersim`, whose protocol, [`research/experiments/PROTOCOL-A-offline.md`](../research/experiments/PROTOCOL-A-offline.md) (section 2), defines the measures R1–R7 in full. The research copy stays as it is: it describes the rule the paper was written about. The bench follows the product as it changes.

Every step of a lesson goes through the product's domain code, the same as in the service:

1. a new profile at the start the child's grade gives (`profile.New`);
2. the rule's brief for the next task (`tutor.Next`);
3. the request opened and a task issued at the brief's point (`Profile.Ask`, `Profile.Issue`), sealed with a key made for the run;
4. the child's answer recorded (`Profile.Record`): the trial series, the step, the runs of mastery and the failures, as the service does them.

A task is answered two minutes after it is issued, the next is issued at once, and every tenth answer starts a new day. A child gives 200 answers.

## The children

Every child is synthetic and drawn from a seed of its own, and the same children are run under every rule, meeting the same random draws answer by answer. The tasks themselves differ between rules, since each rule chooses from its own estimate.

The **base population**: a grade from 1 to 6, and the start the service gives that grade; a true level of the start plus N(0, 1); a true offset per topic of N(0, 0.5²); and a task that is truly harder or easier than the difficulty asked for, by N(0, 0.5²), since a chat's model does not hit the difficulty it is asked for — the model's **miss**. A child answers correctly with chance 0.2 + 0.8·σ(level + offset − difficulty), and never asks for the hint.

The **generators** each change the base population in one respect. The first nine are those the bench was carried over with; the rest move one assumption, or one way of changing, by a degree. A generator's children are drawn from seeds named after it, so a generator added leaves every child of the others as they were.

| Generator | The children |
|---|---|
| G0 | stay put: the model's own assumptions, but for the model's misses |
| G1 | were placed a level off: their level is 2.5 below the start for half of them, 2.5 above it for the rest |
| G2 | learn fast: their level grows by 0.01 an answer, two logits over the run, the topic answered by 0.02 more, and the other topics slip back by 1 % of the way to where they began |
| G3 | jump: their level grows by 1.0 at once, at an answer from the 50th to the 150th |
| G4 | meet two hosts in turn, the second writing its tasks harder by 0.75 |
| G5 | are strong in one half of the topics and weak in the other, by 0.75 each way |
| G6 | answer with another slope than the model's: 0.5 for half of them, 2 for the rest |
| G7 | guess otherwise than the model: never for half of them, three times in ten for the rest |
| G8 | ask for the hint on one answer in five, which keeps a right answer from counting towards mastery |
| G0-exact | stay put, and get tasks the model wrote exactly as asked: no miss |
| G0-miss0.25, G0-miss1 | stay put, under a model that misses by N(0, 0.25²) or N(0, 1) |
| **G2-half** | **learn at half G2's speed: one logit over the run.** The main learner, the speed the author takes as plausible (2026-10-03); G2 stays as a test of stress |
| G2-fading | learn at G2's speed at first, slowing as 50/(50 + k) at answer k — half as fast by the 50th answer, a fifth by the 200th, about 0.8 logit over the run; they forget as G2's do |
| G3-drop | drop: their level falls by 1.0 at once, at an answer from the 50th to the 150th, as after a long break |
| G0-start0.5, G0-start2 | stay put, spread around their start by N(0, 0.5²) or N(0, 2²) |
| G0-topics0.3, G0-topics1 | stay put, their topics spread around their level by N(0, 0.3²) or N(0, 1) |

How fast real children learn is not known: the speeds are the bench's, not a measurement.

## The rules

Each rule estimates where a child stands. The bench writes a rule's estimate into the profile before the service chooses the next task, so that the choice of task, the trial series and mastery are the service's for every rule, and only the estimate differs.

| Rule | Cells | What it is |
|---|---|---|
| The service | `shrinking/both` | The product's own path, untouched: the step K = K₀/(1 + 0.05n) after the trial series, K₀ = 0.2 for the overall level and 0.4 for a topic's offset |
| Constant step | `constant/both` | The service's first step, never shrinking |
| Slow constant step | `constant_slow/both` | Half the service's first step, never shrinking: 0.1 and 0.2 |
| Floor under the step | `floor_0.05/both` | The service's step, the overall level's never below 0.05 |
| No trial series | `no_trial/both` | The service's step from the first answer, with no estimate of the trial series |
| Glicko-2 with guessing | `glicko2_floor/general`, `glicko2_floor/topics` | Glicko-2 over a chance that a child can guess, with one level, or with a level per topic |
| The oracle | `oracle/both` | Stands where the child truly stands, in every topic, at every moment, and knows the child's true chance: **the ceiling** of every rule that learns of a child from answers |

The step rules start from the trial series' estimate, as the service does; Glicko-2 and the oracle start from the first answer. These are the bench's own set of rules, `bench`, which a run is given unless it names another; the step chosen joins them once one is, unless it is among them already, as the floor of 0.05 the choice came to is. Every rule runs on all nineteen generators: 152 cells. The other sets, which choose the step, are described under [The candidates of the step](#the-candidates-of-the-step).

The oracle is the ceiling of the corridor at each miss of the model, and of mastery under the mastery rule the service has now, given a perfect estimate. It knows a child's level but not their slope or their floor (G6, G7). It is not a ceiling for a new rule of mastery: the oracle keeps the service's, while a rule of mastery takes the service's place under the step chosen (see [The candidates of mastery](#the-candidates-of-mastery)). It is compared with no rule as one, and the measures of what the child is shown leave it out: its rating moves only when the child does.

## The measures

A topic is **truly mastered** at a level when the child's true chance on a task of that level and topic at difficulty 3, written as asked, is at least 0.775, the corridor's middle. The names are those of `cells.csv`.

| Measure | Names | What it reads |
|---|---|---|
| R1 error of the level | `r1_rms_N`, `r1_mean_N` | Over the topics the child has answered, the root mean square and the mean of the estimate less the truth, after 5, 10, 20, 50, 100 and 200 answers |
| R2 calibration | `r2_brier`, `r2_bias`, `r2_ece` | The Brier score of the rule's own chance against the answers, the chance it predicted less the share right, and the calibration error over ten bins |
| R3 corridor | `r3_inside`, `r3_below_0.5`, `r3_above_0.95`, `r3_reachable` | The share of tasks whose true chance lies in 0.70–0.85, below 0.50 and above 0.95; and the share of tasks whose topic's ladder held a point at a true chance in 0.70–0.85 at all — the ceiling of the ladder itself, which no rule passes when the model misses nothing |
| R4 false "mastered" | `r4_false`, `r4_false_0.70`, `r4_declared`, `r4_below`, `r4_false_at_task` | Of the masteries declared, the share whose topic was not truly mastered at the level it is held mastered at, or not at 0.70; how many were declared; the share declared at a level below the task's, which the child is not shown; and the false share of those declared at the task's level |
| R4b false "mastered" within m attempts | `r4b_3` … `r4b_50`, `r4b_chance_*` | The chance that a topic not truly mastered is declared mastered within its first m tasks set at a chance of at most 0.775, the ones mastery counts, and the true chances on those tasks |
| R5 late mastery | `r5_late_answers`, `r5_never` | For a topic and level the child truly masters, the answers in the topic until the rule declares it, and the share never declared |
| R6 lag | `r6_lag`, `r6_jump_answers`, `r6_jump_unsettled` | The estimate less the truth from the 101st answer on, below zero when the estimate lags behind; after a jump or a drop, the answers until the error stays under 0.5 for ten answers, and the share of children it never does for |
| R7 placement | `r7_error_5`, `r7_error_10`, `r7_longest_wrong`, `r7_hard_first` | The error of the overall level after 5 and 10 answers, the longest run of wrong answers among the first 15, and the share of the first 10 tasks whose true chance is below 0.50 |
| R8 the screen | `r8_move_p95_W`, `r8_rank_W`, `r8_topic_rank_W` | What the child is shown, in two windows W of answers, 6–20 (`6_20`), the first after the trial series, which shows no rating, and 150–200 (`150_200`): how far the topic's rating on the card moves after an answer, `|before − after|` in rating points as the card shows them, at the 95th percentile of every answer in the window of every child of the cell; and how many times in a hundred answers the overall rank, on the progress screen, and the rank of the topic answered change |

R8 is read off the rule's own estimate, through the product's own `rating.Elo`, `rating.Shown` and `rating.Rank`, as the card and the progress screen would show it.

Every number is read off all the children of its cell, with a 95 % interval from 2,000 resamples of the children. A comparison is the difference between two cells over the same children, with an interval from the same resamples taken in both. Four groups of comparisons are made of the rules a run has, in `comparisons.csv`, the last two when the run has their rules:

1. on G0, the service against every other rule but the oracle: R1 after 200 answers, R3, R4;
2. on G2, the service against every other rule but the oracle: R6, R3;
3. on G1, the service against no trial series: R7;
4. the floor of 0.05 under the step against the service: R6 on G3, R1 after 200 answers on G0.

A comparison is read as finding a difference when its interval leaves zero out. None is corrected for the others.

## The criterion

A new step for the rating, and then a new rule of mastery, is chosen by these rules, written down before any candidate ran, so that what the candidates' numbers turn out to be cannot shape what they are judged by. The bench evaluates them for every rule of a run, in `results/criterion.md` and `results/criterion.csv`. One number of the criterion, the share c*, was read off the run that wrote it down and fixed; the tolerances every run reads off itself, by a formula fixed here, since a decision run of more children resolves finer than a rough look. Both come from the service's rule and the rules already on the bench, never from a candidate: they are calibration, not a choice.

"The service" below is always the service's cell of the same run, on the same children. A number the run cannot read leaves its constraint unmet and its score unread.

### Choosing the step

A candidate is chosen on the working seeds, in a decision run of **4,000 children a cell**, and confirmed once on the held-out seeds. A run of fewer children is a rough look: its criterion is headed so, and decides nothing.

**1. Constraints.** Every one must hold.

- **The goals,** read on point estimates:
  - on G2-half, the lag, `|r6_lag|`, is at most 45 % of the service's — the goal of 0.5 instead of 1.12 that was set on G2, as the same share of the way to the oracle, whose lag is 0;
  - on G2-half, the corridor, `r3_inside`, is at least the service's plus c* of the way from it to the oracle's. c* is the share of the way the goal of 35 % instead of 26 % closes on G2: c* = (0.35 − the service's corridor on G2) / (the oracle's on G2 − the service's), read once, on the run that wrote this criterion down, and fixed: c* = (0.35 − 0.257) / (0.605 − 0.257) = **0.27**. With the service at 36 % on G2-half and the oracle at 61 %, that asks for 42 % — where 1.35 times the service would ask for 48 %, a stricter goal than the one it carries over;
  - on G3, the share of children not caught up after the jump, `r6_jump_unsettled`, is at most 25 %.
- **Not worse than the service,** read on the paired difference with the service on the same children: the worse end of its 95 % interval lies within the tolerance.
  - On every generator, the error after 200 answers, `r1_rms_200`, and the corridor, `r3_inside`. On G0 they are the goals of no worse than now.
  - False masteries are not among them (R166). Under the service's rule of mastery they grow the closer an estimate follows the child: on G2-half from the service's 52 % to 54 % under the floor of 0.05, 56 % under the slow constant step, 59 % under the constant step and 73 % under the oracle. At the decision size such a check would fail every step that follows a learner better, the floor of 0.05 of the exit among them, for a fault of the rule of mastery, which is chosen next. They are shown for every rule in `scenarios.md`, and the choice of mastery holds them on every generator.
  - The tolerance of a check is the larger of the author's — 0.01 logit for the error, 1 percentage point for the corridor — and the bench's resolution on that check: 4.3 standard errors of the paired difference between the slow constant step and the service there, in the same run. A candidate exactly as good as the service passes such a check 99 times in 100: a tolerance finer than the bench resolves would fail good candidates by chance across some forty checks.
  - A rough look reads them as no clear harm: a check fails only when the better end of the interval is beyond the tolerance too.
- **The screen,** on G0 and G2-half, read on point estimates: in both windows, the card's move at the 95th percentile and the changes of the overall rank are no more than the service's in answers 6–20. Late in a run the child sees no more movement than a child sees at the start now, and a candidate's first cards after the trial series move no more than the service's do.

**2. The score.** The share of the way from the service to the oracle a candidate closes: averaged over a generator's measures, then weighted across generators.

| Group | Weight | Generators | Measures |
|---|---|---|---|
| Stays put | 1 | G0 | the error after 200 answers; the corridor |
| Learns | 2 | G2-half 1.25, G2-fading 0.5, G2 0.25 | the lag, `|r6_lag|`; the corridor |
| Changes at once | 1 | G3 0.5, G3-drop 0.5 | the answers until the estimate catches up, `r6_jump_answers`; the corridor |
| Placed a level off | 0.5 | G1 | the error of the overall level after 10 answers, `r7_error_10`; the share of hard first tasks, `r7_hard_first` |

- For each measure, the way is the service's value less the oracle's, with the corridor counted as its shortfall, both read on every child of the run and fixed; the share a candidate closes is the service's value less its own, over the way. The service scores 0 and the oracle 1. A way of zero or less leaves the score unread.
- With the way fixed, the score is a weighted sum of paired differences with the service. Its 95 % interval comes from resampling each generator's children apart, the candidate and the service on the same children.
- The oracle's error, lag and placement are 0. On those measures the share of the way is the share of the service's own value the candidate saves, and so smaller than on the corridor: a static child is held by the constraints rather than the score.

**3. Better.** A candidate is better than the service when its score's interval lies above zero, and better than another candidate when the interval of the paired difference of their scores does.

**4. The choice.** Among the candidates that meet every constraint and are better than the service, A* is the one with the highest score, and its equals are those whose difference from A* has an interval holding zero. The chosen one is the simplest of A* and its equals: the one with the fewest parameters added to the service's rule, then the one with no field added to the profile, then the one nearest the service's rule. The backup candidate, with its uncertainty kept in the profile, is chosen over the main one only when the interval of its score less the main one's lies above 0.05: a gain smaller than that does not pay for new fields in the profile.

**5. Confirmation.** The chosen candidate is run once more, on the held-out seeds at the decision size, beside the service and the oracle. It must meet every constraint and be better than the service there too. If it is not, the exit is taken; no second candidate is tried on the held-out seeds, since a set tried twice is held out no longer.

**The exit.** If no candidate meets the constraints and is better than the service, or the chosen one fails its confirmation, the step gets only a floor (О-53): 0.02, 0.05 or 0.1 under the overall level's step, the one with the highest score among those that meet the constraints of not worse and of the screen; the goals do not apply to it.

### Choosing mastery

The rule of mastery is chosen in the same way, over the step already chosen, with these constraints and this score. **"The service" here is the baseline: the step chosen, the floor of 0.05, under the service's own rule of mastery** (R175). That measures what a rule of mastery gives by itself, rather than charging every candidate with what the floor already moves under the old rule — two points more masteries never declared to a child who learns, three more declared falsely. The baseline is also the exit.

- **Constraints,** on point estimates:
  - false masteries are at most 20 %, on G0 and on G0-topics1, the widest spread of topics, on which a margin of mastery is chosen; they are read at the level a topic is held mastered at, the baseline's included, and never against the share the old reading gave;
  - the answers until a mastery the child has is declared, `r5_late_answers`, are at most 1.5 times the baseline's, on the same generators;
  - the screen's, as for the step, against the baseline's numbers in answers 6–20, which are the service's: the floor does not act before the 60th answer.
- **Not worse than the baseline:** false masteries, `r4_false`, on every generator, with a tolerance of 2 percentage points or the bench's resolution, whichever is larger, as for the step (R166); the share of masteries never declared, `r5_never`, with the same tolerance — or a rule would wait less only by declaring less — read on G0 and G0-topics1 alone, where the goals of mastery are (R175, decided by the author after the pilot): there whatever is left undeclared is the rule's own doing, while for a child who learns the step chosen trails the child, and a margin on top of the trail leaves masteries reached late in a run undeclared by its end, which is the step's to answer for; and the step's checks of not worse, the error after 200 answers and the corridor, which a rule of mastery moves by the topics it takes out of the rotation.
- **The score:** the share of the way from the baseline to perfect — no false mastery and no wait — closed on false masteries and on the answers until mastery, with the weights of the step.
- **Better, the choice and the confirmation** as for the step. **The exit** is the baseline, kept as it is: the service's rule of mastery over the step chosen, which no check is asked of.

**The pilot.** Before the decision run, the candidates are run on the sweep's children, a thousand a cell (`-rules mastery-pilot`). A candidate is stopped there when, on some generator, the whole 95 % interval of its difference from the baseline in masteries never declared lies above 2 percentage points: the author's tolerance, not widened by the bench's resolution. If every candidate is stopped, the decision run waits for the author (R175).

## The candidates of mastery

The rules of mastery put forward, written down before any of them ran. A rule of mastery takes the service's place on the bench as an estimate does: after each answer, once the service has recorded it and the rule's estimate has taken it in, the rule judges it, and what it holds — the level each topic is mastered at and since when, both or neither — is written into the profile over what the service's own rule wrote there, since the profile is all the service's choice of tasks reads mastery from. What is measured is the rule's own verdict. The counts of answers and the topic's run of wrong answers are read off the profile, which keeps them alike under every rule. A test holds the service's own run of three, put in its place this way, to the service's rule on every answer of every generator, and one at an infinite margin holds the profile to no mastery at all.

Every rule declares only after a right answer without the hint, with at least five answers in the topic, and only at a level above the one the topic is held at; every rule loses mastery after two wrong answers in a row, as the service's does. None is tuned: each is run as written here.

- **A run of five** (`run5`): the service's rule with a run of five instead of three, with all its ways — the run grows before the topic's fifth answer, may complete on a right answer to an easy task, and is spent when it completes, whatever it earns. It adds no number.
- **The cautious estimate** at z = 1.0, 1.28 and 1.64 (`cautious_z…`): the topic is mastered at the highest of its levels, at or below the task's, on whose middle task — difficulty 3 — a child at the estimate less z·√v answers at the corridor's middle, 0.775, or better. v is the uncertainty of the level in the topic from the counts the profile keeps: v = 1/(1/2.5² + I·n) + 1/(3 + I·m), with n the answers in all and m in the topic and I = 0.15 — the trial series' spread narrowed by every answer, plus the topic's uncertainty the service's step reads its first step from, narrowed by the topic's answers. These are the product's own numbers, so the rule adds z alone. The sum overstates v a little, since a topic's answers tell of the overall level too, which suits a cautious rule. The cautious estimate at no margin, z = 0, is run beside them and is no candidate: it shows what reading at the middle task costs apart from the margin.
- **Wald's test at each level** (`wald`): for every level of the topic, the chance on its middle task at most 0.70 against at least 0.85, the corridor's bounds, with α = 0.05 and β = 0.2. Every wrong answer and every right and unaided one adds to each level's sum how much likelier it was for a child at the top of the corridor there than at its bottom, the task's own difficulty taken in; a sum that falls under ln(β / (1 − α)) starts again; past ln((1 − β) / α) = ln 16 the topic is mastered at the highest such level at or below the task's, whose sums at and below it are then spent. The sums have to be kept, so the rule adds fields to the profile, and two numbers: it is the backup, chosen over the main rule only by the margin of 0.05. A test holds a single test of it to its errors: at the corridor's bottom it lets a child through at most α / (1 − β) of the time, at its top it turns one away at most β / (1 − α) of it.

The simplest of equals is the one adding the fewest numbers — the run of five none, the cautious estimate one, Wald's test two — then the one adding no field, then the nearest the service's rule: the run of five, then the cautious estimate, the smaller margin the nearer.

A rule that declares at a level below the task's declares a mastery the child is not shown and the rotation does not act on, which still counts among the masteries read for falseness; `r4_below` is the share of those, so that a rule that keeps its false share down that way is seen to, and `r4_false_at_task` the false share of the masteries declared at the task's level, the ones the child is shown.

### What the pilot came to

The pilot, `results/mastery-pilot/`, a thousand of the sweep's children a cell, stopped every candidate on masteries never declared, read then on every generator. Against the baseline:

| Rule | False, G0 / G0-topics1 | False among those shown, G0 | Answers until mastery, G0 | Never declared, G0 / G2-half | Generators clearly worse on never declared | Score |
|---|---:|---:|---:|---:|---:|---:|
| The baseline, the floor of 0.05 under the service's mastery | 65 % / 59 % | 65 % | 9.9 | 24 % / 19 % | — | 0 |
| A run of five | 70 % / 57 % | 70 % | 11.9 | 76 % / 72 % | 19 | −0.13 |
| The cautious estimate, z = 1.0 | 4 % / 6 % | 4 % | 4.7 | 18 % / 34 % | 6 | 0.69 |
| The cautious estimate, z = 1.28 | 2.5 % / 4 % | 3 % | 4.9 | 23 % / 38 % | 7 | 0.68 |
| The cautious estimate, z = 1.64 | 1 % / 3 % | 1.5 % | 5.0 | 30 % / 43 % | 12 | 0.68 |
| Wald's test | 0.5 % / 0.5 % | 0.4 % | 11.0 | 61 % / 64 % | 19 | 0.38 |
| The cautious estimate at no margin, no candidate | 16 % / 16 % | 20 % | 4.2 | 7 % / 18 % | 1 | 0.65 |

The cautious estimate met every other constraint, and was stopped only on children who learn, jump or meet two hosts, and on the widest spread of topics at z = 1.28 and above: there the step chosen trails the child, and the margin comes on top of the trail. On children who stay put it leaves fewer masteries undeclared than the baseline. The author then had masteries never declared read on G0 and G0-topics1 alone, where the goals of mastery are (R175), and the pilot was run again with that reading before the decision run; its numbers are those above, but for what the criterion reads of them. A run of five, which reads the task's level as the service's rule does, comes nowhere near 20 % false; Wald's test is as slow as it was expected to be.

## The candidates of the step

The candidates and the way they are searched and chosen, written down before any of them ran.

### The main candidate: a step from an uncertainty counted by answers

Each level — the overall one and each topic's offset — moves by its step times the surprise of the answer, and the step is κ·v, where v is how uncertain the level is. An answer narrows v as an answer at the middle of the corridor would, and adds q back: v ← v / (1 + I·v) + q. I = 0.15 is what an answer at a chance of 0.775 tells of the level, and κ = 0.927 how far a surprise of one moves the level for each unit of uncertainty there: the information and the gain of a Kalman filter at the corridor's middle, both read off the service's own chance. v is worked out from the counts of answers the profile keeps already, `ratings.answers` and `topics[t].answers`, so the step adds no field to the profile:

- the overall level's v counts from the end of the trial series, where it stands at v_T, one number for every child;
- a topic's counts from its first answer, where it stands at the spread of topics squared, s²;
- with q > 0 the step never dies away but settles at κ·v*, where v* = q/2 + √(q²/4 + q/I) is where an answer takes away exactly q;
- the limit L holds how far one answer may move the level in a topic, θ + δ: past it both steps are scaled down together, so that two failures in a row do not move a child a whole difficulty (D37 of the prototype).

The service's step is the case q = 0 with gains of 0.6 and 1.2 and an uncertainty of a third before any answer, and the floor of О-53 is near the case of a small q. With q = 0 the bench works the step out by the service's own formula, K₀/(1 + a·n), rather than carrying v answer by answer, which rounds otherwise: the service's numbers give the service's step to the last bit, which a test holds on every generator, answer by answer. The step takes v before the answer; a filter's exact gain is κ·v/(1 + I·v), 15 % less at v = 1, so v_T is a number of the search, not the trial series' own spread.

What the candidate adds to the service's rule is counted as q for the overall level, q for a topic and the limit, each where it is used; v_T and s stand where the service's first steps and its decay stand, and add none.

### The backup: an uncertainty kept in the profile

Each child keeps an uncertainty for the overall level and one for every topic, as the profile would have to store them. The overall level's starts at the end of the trial series from what the series' five answers tell at the level it found, on top of the start's spread of 2.5 — the expected information, not the curvature of the answers, which a guess can turn the wrong way; a topic's starts at s². An answer narrows them by what it tells at the chance the child had, adds q to each, and shares its surprise between the overall level and the topic by their uncertainties, as a Kalman filter shares it between two levels whose errors are independent. The same limit applies. It adds fields to the profile.

### The comparisons, which the product does not take

- **The full filter**: a Kalman filter over the overall level and every topic's offset at once, keeping how their errors go together, at the best main candidate's numbers: what that candidate's simplicity costs.
- **The estimate over the whole history**: the most likely overall level and offsets given every answer at once — θ ~ N(start, 2.5²), as the trial series has it, δ ~ N(0, s²) — and nothing added for a child who changes: the least error a rule of answers can have on a child who stays put.
- **Two speeds with a switch**: the service's step and the constant step 0.2/0.4 side by side, the slow one's levels the estimate's; a running sum of how much better the fast one has foretold the answers, never let below zero, moves the slow one onto the fast one's levels when it passes a threshold. A sum that forgets would not do: a jump of a level makes an answer only some hundredths of a nat likelier under the fast step, and a sum that forgets a tenth of itself an answer settles near half a nat. This is close to the "recent results" D17 of the prototype turned down, and is on the bench only to compare.
- **Floors** of 0.02, 0.05 and 0.1 under the service's overall step, for the exit.

Every candidate and comparison follows the trial series as the service does, and steps from the sixth answer.

### How they are searched and chosen

1. **The sweep** (`-rules sweep`) runs every point of a grid of the main candidate on children of its own — experiment `sweep`, so that the decision run measures what it was given on children it was not picked on — 300 children a cell, all nineteen generators, beside the service, the slow constant step, the oracle, and the two speeds at thresholds of 2, 4 and 8:
   - v_T: 0.1, 0.17, 0.4, 1.0 — a first overall step after the series of 0.09, 0.16 (the service's), 0.37, 0.93;
   - s: 0.35, 0.5, 0.7;
   - q for the overall level: 0, 0.003, 0.01 — steady steps of none, 0.13 and 0.24;
   - q for a topic: 0, 0.006, 0.03 — none, 0.19 and 0.43;
   - L: none; 0.42, the card's move of 73 points, the service's at the start; 0.3.

   The values stand on both sides of where the constraints give way — the rank's changes early, the card's move, the lag — so that if no candidate meets them, it is not for want of a grid that reaches them.
2. **The sweep's best** is the main candidate of the highest score among those that meet every constraint as the sweep reads them, or, if none does, among those that miss the fewest.
3. **The refinement** (`-rules refine`, the sweep's children) runs the points one number away from the sweep's best on finer ladders — v_T from 0.05 to 1.2, s from 0.25 to 1, q from 0 to 0.04 and to 0.06, L from 0.22 to none — that the sweep did not run, and the backup at the best's numbers and with each of its q one rung either way.
4. **The nomination** for the decision run, read off the sweep and the refinement together:
   - the five main candidates of the highest score that meet every constraint, and for each count of numbers added, 0 to 3, the best that does, so that the choice has the simpler ones to choose;
   - if fewer than five meet every constraint, the best of those one constraint short, up to five in all;
   - the best backup and the best two speeds;
   - the full filter and the estimate over the whole history at the best main candidate's numbers;
   - the best main candidate with each of its parts taken away: the limit, q for the overall level, q for a topic, all three at once, and the service's gains in place of the model's;
   - the three floors.

   The rule did not foresee a sweep in which no main candidate meets every constraint and none misses one alone, which is what the sweep and the refinement came to. In that case the order the sweep's best is found by — the fewest constraints missed, then the score — stands in for meeting them, for the main candidates, the best one, the backup and the two speeds alike: the five first, and the first for each count of numbers added. This was settled after the sweep and the refinement were read and before the decision run.
5. **The decision run** (`-rules decision -children 4000`, the working children) reads the choice of "Choosing the step" itself and writes it in `criterion.md` under "The choice". Three readings of it are fixed here:
   - the one nearest the service's rule is the one whose steps differ least from the service's: the mean difference of the overall level's steps over answers 6–200, plus that of a topic's over its first 40 answers;
   - the backup's margin of 0.05 is read against the main step chosen; when no main step meets the constraints, a backup that meets them is chosen;
   - when no floor meets the constraints of not worse and of the screen, the service's step stays.

   For a chosen candidate it writes, before the held-out children are drawn, the chance each constraint holds on as many new children; the exit, taken without a confirmation, has none written.
6. **The parts** of the chosen step, if it is not the one whose parts the decision run took away, are run on the working children as well (`-rules parts`). A floor taken as the exit has no parts to take away.
7. **The confirmation** (`-held-out -children 4000`) runs the chosen rule alone beside the service, the slow constant step and the oracle. The exit is taken without one, so a run of the held-out children refuses once the choice comes to the exit: they stay held out for a candidate.

A candidate exactly as good as the service fails at least one of the nineteen checks of the error about one time in six, each passing 99 times in 100, and the confirmation runs that risk again. The criterion was set so in T72.3; the bench only names it.

### What the runs came to

**The sweep and the refinement**, 300 children a cell on the sweep's children, `results/sweep/` and `results/refine/`: no main candidate met every constraint, and none missed one alone; the fewest any missed was three. What gives way shows across the grid. To follow the child who learns at half speed within the goal, the overall level needs some 0.003 of uncertainty added for every answer, a lasting step of 0.13; with it the overall rank changes about four times in a hundred answers late in a run, where the service's changes three times at the start and less than once late. Without it the rank stays still, and the goals of the learner and of the jump are missed. Averaged over the other numbers of the grid, 36 points each:

| q for the overall level | Limit | Lag on G2-half | Corridor on G2-half | Not caught up on G3 | Rank changes, answers 6–20 | Rank changes, answers 150–200 |
|---|---|---:|---:|---:|---:|---:|
| 0 | none | −0.538 | 37.8 % | 61.7 % | 8.2 | 0.9 |
| 0 | 0.42 | −0.526 | 38.2 % | 60.3 % | 7.3 | 0.9 |
| 0 | 0.3 | −0.459 | 39.1 % | 53.2 % | 6.6 | 0.9 |
| 0.003 | none | −0.342 | 41.0 % | 23.2 % | 8.8 | 4.1 |
| 0.003 | 0.42 | −0.328 | 41.3 % | 22.4 % | 7.8 | 4.1 |
| 0.003 | 0.3 | −0.211 | 42.2 % | 15.9 % | 7.1 | 4.2 |
| 0.01 | none | −0.251 | 41.1 % | 11.5 % | 9.8 | 8.1 |
| 0.01 | 0.42 | −0.189 | 41.7 % | 9.2 % | 8.8 | 8.0 |
| 0.01 | 0.3 | +0.034 | 41.7 % | 4.8 % | 8.0 | 8.0 |

On the sweep's children the goals were a lag of at most 0.268, a corridor of at least 42.7 %, at most 25 % not caught up, and at most 3.1 changes of the rank in a hundred answers. The limit helps the learner as much as it calms the card: a wrong answer is a larger surprise than a right one, so holding back the large moves holds back mostly the moves down.

**The decision run**, 4,000 children a cell on the working children, `results/decision/`: **no candidate meets every constraint, and the exit is taken — the floor of 0.05 under the overall level's step**, with a score of 0.056 [0.055, 0.058]. Of the other floors, 0.1 scores 0.161 but is worse than the service on the error after 200 answers on six generators (G4, G5, G6, G0-miss0.25, G0-start0.5, G0-topics0.3), and 0.02 meets every check but closes nothing of the way, 0.0001.

The best main candidate — v_T 0.13, s 0.7, q 0.003 and 0.006, the limit 0.3 — scores 0.255 [0.251, 0.259] and meets the goals of the lag, 0.212 against at most 0.276, and of the jump, 10 % not caught up against at most 25 %. It misses:

- the corridor of the child who learns at half speed, by a hair: 41.8 % against at least 42.2 %;
- the overall rank's changes late in a run: 4.6 in a hundred answers on G0 against the service's 3.7 at the start;
- not worse on the error after 200 answers, on nine generators: on G0 0.512 against the service's 0.493, past a tolerance of 0.010;
- not worse on the corridor, on three generators.

The backup scores 0.245 and meets the learner's goals — a lag of 0.245, 42.6 % in the corridor, 18 % not caught up — but its overall rank changes 12.6 times in a hundred answers early, against the service's 3.7: the uncertainty the five answers of the series leave is near 1, so its first steps are large.

So the choice does not fail for want of a step that follows a learner. The step that does is held back by two constraints a lasting step meets head on: a child who stays put may not lose more than about a hundredth of a logit of accuracy, and the overall rank may not change more often late in a run than the service's does at the start.

**What each part of the best main step gives**, its score less the score of the step without it: q for the overall level 0.192; the limit 0.065; the model's gains in place of the service's 0.048; q for a topic 0.016; all three of q, q and the limit 0.233. The paired intervals are in `results/decision/criterion.md`, under "Every rule against the main step of the highest score".

**The comparisons.** The full filter on the best main step's numbers scores 0.197, less than the step it generalises: its exact gain is smaller than κ·v, so it follows the learner less, a lag of 0.328; the step's simplicity costs nothing here. The estimate over the whole history errs 0.464 on G0, against the service's 0.493: on a child who stays put, no rule of answers does much better than the service now. The two speeds, at a threshold of 2, score 0.106: 37 % not caught up after a jump against the service's 72 %, at an error of 0.514 on G0.

**The confirmation** was not run: the exit is taken without one, so the held-out children have not been drawn and stay held out for a candidate.

#### Every generator, in the decision run

The rules the choice turned on, on every generator; every rule of the run is in `results/decision/scenarios.md`, and every number with its interval in `cells.csv`. "Best main step" is v_T 0.13, s 0.7, q 0.003 and 0.006, the limit 0.3; the backup, the full filter and the whole history take its numbers.

##### The error after 200 answers, in logits

| Generator | Service | Floor 0.05, the exit | Floor 0.1 | Best main step | Backup | Full filter | Whole history | Two speeds | Oracle |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| G0 | 0.493 | 0.488 | 0.496 | 0.512 | 0.495 | 0.497 | 0.464 | 0.514 | 0.000 |
| G1 | 0.513 | 0.500 | 0.496 | 0.514 | 0.494 | 0.494 | 0.463 | 0.509 | 0.000 |
| G2 | 1.428 | 1.187 | 0.860 | 0.664 | 0.762 | 0.774 | 1.248 | 0.880 | 0.000 |
| G3 | 0.839 | 0.717 | 0.595 | 0.535 | 0.547 | 0.557 | 0.779 | 0.661 | 0.000 |
| G4 | 0.654 | 0.655 | 0.668 | 0.625 | 0.614 | 0.637 | 0.630 | 0.679 | 0.000 |
| G5 | 0.550 | 0.545 | 0.563 | 0.585 | 0.579 | 0.567 | 0.538 | 0.572 | 0.000 |
| G6 | 0.743 | 0.753 | 0.780 | 0.740 | 0.707 | 0.746 | 0.732 | 0.785 | 0.000 |
| G7 | 0.568 | 0.559 | 0.570 | 0.579 | 0.562 | 0.565 | 0.540 | 0.583 | 0.000 |
| G8 | 0.495 | 0.488 | 0.497 | 0.517 | 0.493 | 0.498 | 0.467 | 0.519 | 0.000 |
| G0-exact | 0.477 | 0.472 | 0.482 | 0.512 | 0.485 | 0.490 | 0.454 | 0.502 | 0.000 |
| G0-miss0.25 | 0.480 | 0.476 | 0.488 | 0.510 | 0.489 | 0.490 | 0.455 | 0.503 | 0.000 |
| G0-miss1 | 0.544 | 0.539 | 0.542 | 0.542 | 0.522 | 0.530 | 0.507 | 0.564 | 0.000 |
| G2-half | 0.835 | 0.724 | 0.609 | 0.542 | 0.553 | 0.573 | 0.757 | 0.679 | 0.000 |
| G2-fading | 0.658 | 0.589 | 0.532 | 0.511 | 0.499 | 0.513 | 0.590 | 0.582 | 0.000 |
| G3-drop | 0.692 | 0.601 | 0.541 | 0.564 | 0.575 | 0.543 | 0.712 | 0.630 | 0.000 |
| G0-start0.5 | 0.490 | 0.483 | 0.496 | 0.509 | 0.491 | 0.495 | 0.463 | 0.512 | 0.000 |
| G0-start2 | 0.515 | 0.503 | 0.505 | 0.523 | 0.497 | 0.505 | 0.469 | 0.524 | 0.000 |
| G0-topics0.3 | 0.437 | 0.430 | 0.443 | 0.473 | 0.449 | 0.453 | 0.414 | 0.459 | 0.000 |
| G0-topics1 | 0.740 | 0.728 | 0.733 | 0.718 | 0.708 | 0.714 | 0.704 | 0.760 | 0.000 |

##### The tasks in the corridor, %

| Generator | Service | Floor 0.05, the exit | Floor 0.1 | Best main step | Backup | Full filter | Whole history | Two speeds | Oracle |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| G0 | 42.6 | 42.7 | 43.0 | 42.5 | 42.8 | 42.7 | 45.2 | 42.8 | 61.4 |
| G1 | 35.1 | 35.3 | 35.9 | 36.0 | 36.5 | 35.9 | 38.0 | 35.7 | 53.4 |
| G2 | 26.0 | 28.1 | 33.2 | 38.6 | 39.2 | 35.5 | 30.9 | 31.2 | 60.4 |
| G3 | 33.6 | 35.0 | 37.6 | 40.0 | 39.7 | 38.5 | 36.4 | 35.9 | 61.3 |
| G4 | 38.6 | 38.7 | 38.8 | 38.4 | 38.5 | 38.4 | 40.3 | 38.6 | 47.0 |
| G5 | 40.8 | 40.9 | 40.8 | 40.3 | 40.3 | 40.5 | 42.4 | 40.8 | 60.6 |
| G6 | 42.0 | 42.3 | 43.1 | 41.2 | 40.5 | 41.8 | 45.0 | 43.0 | 35.8 |
| G7 | 42.4 | 42.6 | 42.8 | 42.5 | 42.8 | 42.6 | 44.9 | 42.5 | 56.3 |
| G8 | 42.8 | 43.0 | 43.1 | 42.6 | 42.8 | 42.9 | 45.0 | 42.8 | 61.4 |
| G0-exact | 54.3 | 54.4 | 54.3 | 52.9 | 53.1 | 53.6 | 58.1 | 54.3 | 97.1 |
| G0-miss0.25 | 50.0 | 50.0 | 50.4 | 49.1 | 49.6 | 49.7 | 53.4 | 50.4 | 84.2 |
| G0-miss1 | 29.6 | 29.7 | 29.8 | 30.0 | 30.3 | 29.9 | 30.7 | 29.6 | 35.7 |
| G2-half | 35.6 | 37.0 | 39.4 | 41.8 | 42.6 | 40.4 | 39.2 | 37.7 | 60.2 |
| G2-fading | 37.4 | 38.3 | 40.3 | 42.0 | 43.3 | 41.1 | 40.9 | 38.9 | 59.9 |
| G3-drop | 38.8 | 39.4 | 40.3 | 38.4 | 37.9 | 38.9 | 39.7 | 39.6 | 60.4 |
| G0-start0.5 | 43.3 | 43.4 | 43.5 | 43.2 | 43.2 | 43.4 | 45.5 | 43.4 | 61.7 |
| G0-start2 | 39.5 | 39.7 | 40.1 | 39.9 | 40.4 | 40.0 | 42.3 | 40.0 | 58.2 |
| G0-topics0.3 | 44.6 | 44.8 | 45.1 | 44.6 | 44.8 | 44.8 | 47.3 | 44.9 | 61.4 |
| G0-topics1 | 37.3 | 37.0 | 37.0 | 36.5 | 36.6 | 36.7 | 38.6 | 37.0 | 61.7 |

##### The estimate less the level from the 101st answer on, in logits

| Generator | Service | Floor 0.05, the exit | Floor 0.1 | Best main step | Backup | Full filter | Whole history | Two speeds | Oracle |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| G0 | -0.129 | -0.127 | -0.118 | 0.020 | 0.025 | -0.041 | -0.075 | -0.127 | 0.000 |
| G1 | -0.153 | -0.144 | -0.120 | 0.006 | 0.018 | -0.049 | -0.062 | -0.117 | 0.000 |
| G2 | -1.105 | -0.983 | -0.734 | -0.467 | -0.532 | -0.621 | -0.932 | -0.774 | 0.000 |
| G3 | -0.714 | -0.644 | -0.510 | -0.299 | -0.350 | -0.412 | -0.633 | -0.548 | 0.000 |
| G4 | -0.491 | -0.495 | -0.504 | -0.371 | -0.372 | -0.424 | -0.466 | -0.513 | 0.000 |
| G5 | -0.121 | -0.119 | -0.119 | 0.019 | 0.012 | -0.040 | -0.082 | -0.131 | 0.000 |
| G6 | -0.234 | -0.239 | -0.266 | -0.118 | -0.096 | -0.174 | -0.223 | -0.273 | 0.000 |
| G7 | -0.184 | -0.178 | -0.171 | -0.032 | -0.034 | -0.098 | -0.130 | -0.177 | 0.000 |
| G8 | -0.117 | -0.114 | -0.107 | 0.034 | 0.032 | -0.028 | -0.065 | -0.117 | 0.000 |
| G0-exact | -0.082 | -0.077 | -0.068 | 0.068 | 0.065 | 0.009 | -0.024 | -0.076 | 0.000 |
| G0-miss0.25 | -0.085 | -0.084 | -0.078 | 0.055 | 0.057 | 0.002 | -0.037 | -0.080 | 0.000 |
| G0-miss1 | -0.234 | -0.232 | -0.231 | -0.074 | -0.070 | -0.143 | -0.186 | -0.242 | 0.000 |
| G2-half | -0.614 | -0.547 | -0.418 | -0.212 | -0.245 | -0.328 | -0.514 | -0.472 | 0.000 |
| G2-fading | -0.481 | -0.428 | -0.319 | -0.132 | -0.152 | -0.233 | -0.381 | -0.355 | 0.000 |
| G3-drop | 0.385 | 0.326 | 0.212 | 0.300 | 0.348 | 0.282 | 0.439 | 0.218 | 0.000 |
| G0-start0.5 | -0.129 | -0.126 | -0.120 | 0.019 | 0.019 | -0.041 | -0.075 | -0.125 | 0.000 |
| G0-start2 | -0.142 | -0.134 | -0.117 | 0.016 | 0.028 | -0.039 | -0.062 | -0.125 | 0.000 |
| G0-topics0.3 | -0.117 | -0.112 | -0.101 | 0.036 | 0.036 | -0.027 | -0.066 | -0.114 | 0.000 |
| G0-topics1 | -0.141 | -0.148 | -0.149 | -0.019 | -0.024 | -0.069 | -0.085 | -0.151 | 0.000 |

##### The masteries declared falsely, %, for the choice of mastery

| Generator | Service | Floor 0.05, the exit | Floor 0.1 | Best main step | Backup | Full filter | Whole history | Two speeds | Oracle |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| G0 | 63.5 | 64.6 | 66.1 | 69.1 | 70.9 | 67.5 | 66.1 | 64.0 | 75.0 |
| G1 | 60.8 | 61.5 | 62.7 | 64.7 | 65.5 | 62.7 | 63.2 | 61.5 | 66.4 |
| G2 | 41.2 | 44.2 | 47.8 | 50.8 | 50.7 | 48.6 | 44.8 | 44.6 | 67.0 |
| G3 | 51.2 | 53.4 | 55.7 | 58.5 | 58.1 | 57.0 | 52.7 | 53.3 | 71.7 |
| G4 | 58.8 | 59.7 | 61.1 | 63.8 | 64.8 | 62.4 | 60.0 | 58.3 | 75.0 |
| G5 | 62.6 | 63.3 | 64.8 | 67.8 | 68.3 | 65.7 | 64.2 | 62.8 | 74.8 |
| G6 | 66.9 | 67.1 | 68.3 | 71.5 | 72.4 | 69.8 | 68.0 | 66.8 | 68.9 |
| G7 | 64.0 | 64.6 | 66.2 | 69.0 | 70.1 | 67.4 | 65.6 | 64.4 | 68.2 |
| G8 | 64.5 | 65.2 | 67.4 | 69.7 | 70.8 | 68.8 | 66.6 | 65.3 | 75.2 |
| G0-exact | 65.2 | 66.2 | 67.5 | 70.1 | 71.0 | 68.7 | 66.7 | 65.4 | 75.2 |
| G0-miss0.25 | 63.9 | 64.6 | 66.2 | 69.1 | 70.2 | 67.9 | 65.8 | 64.5 | 74.7 |
| G0-miss1 | 62.3 | 63.2 | 64.3 | 67.4 | 68.9 | 66.1 | 63.5 | 62.6 | 74.6 |
| G2-half | 52.4 | 54.3 | 56.8 | 59.4 | 60.4 | 57.6 | 55.1 | 54.2 | 73.1 |
| G2-fading | 53.4 | 55.2 | 57.5 | 60.9 | 62.1 | 58.7 | 56.5 | 55.0 | 71.5 |
| G3-drop | 72.0 | 72.3 | 72.7 | 75.6 | 78.4 | 73.9 | 74.6 | 71.0 | 78.8 |
| G0-start0.5 | 67.7 | 68.3 | 69.6 | 71.4 | 72.3 | 70.7 | 67.8 | 67.7 | 75.2 |
| G0-start2 | 61.6 | 62.4 | 64.4 | 67.5 | 69.8 | 65.9 | 64.8 | 62.3 | 73.3 |
| G0-topics0.3 | 65.6 | 66.2 | 67.7 | 70.5 | 71.9 | 68.8 | 66.5 | 66.0 | 76.0 |
| G0-topics1 | 59.0 | 59.8 | 61.5 | 63.0 | 64.4 | 62.5 | 60.4 | 59.1 | 73.1 |

#### Remarks

- The screen counts every change of the overall rank, a rating that crosses a rank's floor and back again among them. A step that keeps following a child moves the rating a little on every answer, and near a floor every such move can change the rank: on a child who stays put, every change of the best main step late in a run is that, since the child does not move. They are counted, as the criterion says (decided 2026-10-04, by the author's leave): a rank shown with a margin, changing only once the rating is clearly past a floor, would calm the screen under any step, the service's own included, and is an improvement of the screen of its own, not a reason to choose the step again — the best main step misses the error and the corridor besides.
- The tolerance on the error of a child who stays put, about a hundredth of a logit at the decision size, leaves little room for a lasting step: the floor of 0.1 already goes past it on six generators. It stays (decided 2026-10-04): it is the goal of no worse than the service's 0.495, and widening it once the candidates' numbers are known would fit the criterion to a candidate.
- The nomination rule did not foresee a sweep in which no candidate missed one constraint alone; how it was read is written under "How they are searched and chosen".

## The held-out seeds

A run given `-held-out` draws from seeds kept for confirmation: seed 20261003, experiment `held-out`. It runs the confirmation alone — the chosen step beside the service, the slow constant step and the oracle — refuses to run before a step is chosen, or when the step chosen is the exit, which is taken without a confirmation, refuses `-seed`, `-experiment` and any other `-rules` beside it, and writes under `held-out` in the directory it is given, never over the working run. Nobody runs them before the step's confirmation: a set that has been looked at is held out no longer.

The sweep and its refinement draw children of their own too, the paper's seed under the experiment `sweep`, and refuse `-seed` and `-experiment` likewise. A test proves that no child of any generator is drawn alike in any two of the three sets — the working children, the held-out ones and the sweep's.

## Running it

- `just learners` runs the bench's own set, a thousand children a cell, writes `tools/learners/results/` over the run kept there, and prints the summary. It takes four to five minutes on the development machine.
- Arguments go to the bench: `-children` and `-answers` (at least 200, the answer the comparisons read the error at), `-rules` and the set to run, `-seed` and `-experiment` or `-held-out`, and `-out`. A quick look goes elsewhere, so that the run kept is a whole one: `just learners -children 100 -out /tmp/learners`.
- The sets other than the bench's write under a directory of their own within `-out`, `results/<set>`:
  - `just learners -rules sweep -children 300` — about an hour on the development machine;
  - `just learners -rules refine -children 300` — minutes, once the sweep's best is set;
  - `just learners -rules decision -children 4000` — about an hour, once the candidates are nominated;
  - `just learners -rules parts -children 4000`, if the chosen step is a candidate whose parts the decision run did not take away;
  - `just learners -rules mastery-pilot -children 1000` — the pilot of mastery, on the sweep's children, about three minutes;
  - `just learners -rules mastery -children 4000` — the decision run of mastery, about a quarter of an hour;
  - `just learners -held-out -children 4000` — the confirmation of the last candidate chosen, once, about ten minutes: the model of mastery chosen over the step, once one is; the step's exit is taken without one.
- `just learners-test` runs the bench's tests with the race detector, and `just learners-lint` holds the module to what the service is held to. CI runs both on every pull request. A change to the product's `go.mod` is followed by `just learners-tidy`.

## The results

`tools/learners/results/` keeps the whole run of the bench's own set, as the line every change to the student model is measured from, and beside it, each in its own directory, the runs the step was chosen by: `sweep/`, `refine/` and `decision/` — and `parts/` when a chosen candidate needs them, which the exit did not — and the runs the rule of mastery was chosen by: `mastery-pilot/`, `mastery/` and the confirmation of the model chosen, `held-out/`. Every run writes:

- `summary.md`, the numbers a change is judged by at a glance, for every rule: the error after 200 answers on G0, the share in the corridor on G0 and G2, the share of false masteries on G0, the lag on G2, and the share of G3's children not caught up after the jump;
- `scenarios.md`, every rule on every generator, measure by measure: the error after 200 answers, the corridor, false masteries, the lag, the children not caught up after a jump or a drop, and the error of the overall level after ten answers;
- `criterion.md` and `criterion.csv`, the criterion evaluated for every rule: the choice it comes to, or on the held-out children the confirmation; each constraint with its value, interval, bound and verdict, whether its goal is reached, on the edge or not reached, and the score with its interval; beside them, the bench's resolution of each check of not worse;
- `cells.csv`, every measure of every cell, with its interval — but for the sweep and the refinement, where it would run to tens of megabytes and any run gives it again to the last digit, so it is not kept;
- `comparisons.csv`, the four groups of comparisons;
- `run.txt`, what the run was given and what computed it: the seed, the name, the children and answers, the set of rules, the service's version, Go's, and the processor's architecture.

A change to the rule, the rating, the profile or the catalog's topics moves the numbers. `TestASmallRun` runs every cell with ten children and holds the run's `summary.md` and `criterion.md` to their snapshots in `tools/learners/testdata/`, so such a change fails `just learners-test` until they are rewritten with `go test -run TestASmallRun -update` in `tools/learners`; the rewritten snapshots, and a new whole run when the change is to the student model, are part of the change's review. The same run must give every row of `testdata/carried-over/`, the bench's numbers as it was carried over, which nothing rewrites: what is added to the bench leaves them as they were, and only a change to the service's own rule moves them. `TestStepRulesAreTheServicesUpdate` holds the bench's copy of the step to `rating.Update` at the service's constants, so a new step in the service fails it until the bench's copy follows.

## The paper's numbers

A run given neither `-seed` nor `-experiment` draws the children of the paper's run: seed 20261001, experiment E-A3. On amd64 the bench's cells and the comparisons it shares with the research run repeat `research/experiments/learnersim/results` to the last digit, so the summary's first row is the paper's account of the service's rule. On arm64 the last digits may differ, since Go fuses a multiplication and an addition there into one step that rounds once.
