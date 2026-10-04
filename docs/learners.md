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

The step rules start from the trial series' estimate, as the service does; Glicko-2 and the oracle start from the first answer. Every rule runs on all nineteen generators: 152 cells.

The oracle is the ceiling of the corridor at each miss of the model, and of mastery under the mastery rule the service has now, given a perfect estimate. It knows a child's level but not their slope or their floor (G6, G7). It is not a ceiling for a new mastery rule: to set the oracle beside one, the bench will need a way to put that rule in the service's place. It is compared with no rule as one, and the measures of what the child is shown leave it out: its rating moves only when the child does.

## The measures

A topic is **truly mastered** at a level when the child's true chance on a task of that level and topic at difficulty 3, written as asked, is at least 0.775, the corridor's middle. The names are those of `cells.csv`.

| Measure | Names | What it reads |
|---|---|---|
| R1 error of the level | `r1_rms_N`, `r1_mean_N` | Over the topics the child has answered, the root mean square and the mean of the estimate less the truth, after 5, 10, 20, 50, 100 and 200 answers |
| R2 calibration | `r2_brier`, `r2_bias`, `r2_ece` | The Brier score of the rule's own chance against the answers, the chance it predicted less the share right, and the calibration error over ten bins |
| R3 corridor | `r3_inside`, `r3_below_0.5`, `r3_above_0.95`, `r3_reachable` | The share of tasks whose true chance lies in 0.70–0.85, below 0.50 and above 0.95; and the share of tasks whose topic's ladder held a point at a true chance in 0.70–0.85 at all — the ceiling of the ladder itself, which no rule passes when the model misses nothing |
| R4 false "mastered" | `r4_false`, `r4_false_0.70`, `r4_declared` | Of the masteries declared, the share whose topic was not truly mastered at its level, or not at 0.70; how many were declared |
| R4b false "mastered" within m attempts | `r4b_3` … `r4b_50`, `r4b_chance_*` | The chance that a topic not truly mastered is declared mastered within its first m tasks set at a chance of at most 0.775, the ones mastery counts, and the true chances on those tasks |
| R5 late mastery | `r5_late_answers`, `r5_never` | For a topic and level the child truly masters, the answers in the topic until the rule declares it, and the share never declared |
| R6 lag | `r6_lag`, `r6_jump_answers`, `r6_jump_unsettled` | The estimate less the truth from the 101st answer on, below zero when the estimate lags behind; after a jump or a drop, the answers until the error stays under 0.5 for ten answers, and the share of children it never does for |
| R7 placement | `r7_error_5`, `r7_error_10`, `r7_longest_wrong`, `r7_hard_first` | The error of the overall level after 5 and 10 answers, the longest run of wrong answers among the first 15, and the share of the first 10 tasks whose true chance is below 0.50 |
| R8 the screen | `r8_move_p95_W`, `r8_rank_W`, `r8_topic_rank_W` | What the child is shown, in two windows W of answers, 6–20 (`6_20`), the first after the trial series, which shows no rating, and 150–200 (`150_200`): how far the topic's rating on the card moves after an answer, `|before − after|` in rating points as the card shows them, at the 95th percentile of every answer in the window of every child of the cell; and how many times in a hundred answers the overall rank, on the progress screen, and the rank of the topic answered change |

R8 is read off the rule's own estimate, through the product's own `rating.Elo`, `rating.Shown` and `rating.Rank`, as the card and the progress screen would show it.

Every number is read off all the children of its cell, with a 95 % interval from 2,000 resamples of the children. A comparison is the difference between two cells over the same children, with an interval from the same resamples taken in both. Four groups of comparisons are made on every run, in `comparisons.csv`:

1. on G0, the service against every other rule but the oracle: R1 after 200 answers, R3, R4;
2. on G2, the service against every other rule but the oracle: R6, R3;
3. on G1, the service against no trial series: R7;
4. the floor under the step against the service: R6 on G3, R1 after 200 answers on G0.

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
  - On every generator, the error after 200 answers, `r1_rms_200`, the corridor, `r3_inside`, and false masteries, `r4_false`. On G0 the first two are the goals of no worse than now.
  - The tolerance of a check is the larger of the author's — 0.01 logit for the error, 1 percentage point for the corridor, 2 for false masteries — and the bench's resolution on that check: 4.3 standard errors of the paired difference between the slow constant step and the service there, in the same run. A candidate exactly as good as the service passes such a check 99 times in 100: a tolerance finer than the bench resolves would fail good candidates by chance across some fifty checks.
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

The rule of mastery is chosen in the same way, over the step already chosen, with these constraints and this score.

- **Constraints,** on point estimates:
  - false masteries are at most 20 %, on G0 and on G0-topics1, the widest spread of topics, on which a margin of mastery is chosen; they are read at the level a topic is held mastered at, the service's included, and never against the share the old reading gave;
  - the answers until a mastery the child has is declared, `r5_late_answers`, are at most 1.5 times the service's, on the same generators.
- **Not worse than the service:** the share of masteries never declared, `r5_never`, with the tolerance of false masteries — or a rule would wait less only by declaring less — and every constraint of not worse and of the screen of the step.
- **The score:** the share of the way to perfect — no false mastery and no wait — closed on false masteries and on the answers until mastery, with the weights of the step.
- **Better, the choice, the confirmation and the exit** as for the step; the exit keeps the service's rule of mastery.

The bench evaluates this part once the candidates of mastery are on it; the service's numbers for it are in `cells.csv` already.

## The held-out seeds

A run given `-held-out` draws from seeds kept for confirmation: seed 20261003, experiment `held-out`. It refuses `-seed` and `-experiment` beside it, and writes under `held-out` in the directory it is given, never over the working run. Nobody runs them before the step's confirmation: a set that has been looked at is held out no longer. A test proves that no child of any generator is drawn alike in the two sets.

## Running it

- `just learners` runs every cell, a thousand children each, writes `tools/learners/results/` over the run kept there, and prints the summary. It takes four to five minutes on the development machine.
- Arguments go to the bench: `-children` and `-answers` (at least 200, the answer the comparisons read the error at), `-seed` and `-experiment` or `-held-out`, and `-out`. A quick look goes elsewhere, so that the run kept is a whole one: `just learners -children 100 -out /tmp/learners`. A decision run is `-children 4000`.
- `just learners-test` runs the bench's tests with the race detector, and `just learners-lint` holds the module to what the service is held to. CI runs both on every pull request. A change to the product's `go.mod` is followed by `just learners-tidy`.

## The results

`tools/learners/results/` keeps the whole run of the rule the service has now, as the line every change to the student model is measured from:

- `summary.md`, the numbers a change is judged by at a glance, for every rule: the error after 200 answers on G0, the share in the corridor on G0 and G2, the share of false masteries on G0, the lag on G2, and the share of G3's children not caught up after the jump;
- `criterion.md` and `criterion.csv`, the criterion evaluated for every rule: each constraint with its value, interval, bound and verdict, whether its goal is reached, on the edge or not reached, and the score with its interval; beside them, the bench's resolution of each check of not worse;
- `cells.csv`, every measure of every cell, with its interval;
- `comparisons.csv`, the four groups of comparisons;
- `run.txt`, what the run was given and what computed it: the seed, the name, the children and answers, the service's version, Go's, and the processor's architecture.

A change to the rule, the rating, the profile or the catalog's topics moves the numbers. `TestASmallRun` runs every cell with ten children and holds the run's `summary.md` and `criterion.md` to their snapshots in `tools/learners/testdata/`, so such a change fails `just learners-test` until they are rewritten with `go test -run TestASmallRun -update` in `tools/learners`; the rewritten snapshots, and a new whole run when the change is to the student model, are part of the change's review. The same run must give every row of `testdata/carried-over/`, the bench's numbers as it was carried over, which nothing rewrites: what is added to the bench leaves them as they were, and only a change to the service's own rule moves them. `TestStepRulesAreTheServicesUpdate` holds the bench's copy of the step to `rating.Update` at the service's constants, so a new step in the service fails it until the bench's copy follows.

## The paper's numbers

A run given neither `-seed` nor `-experiment` draws the children of the paper's run: seed 20261001, experiment E-A3. On amd64 the bench's cells and the comparisons it shares with the research run repeat `research/experiments/learnersim/results` to the last digit, so the summary's first row is the paper's account of the service's rule. On arm64 the last digits may differ, since Go fuses a multiplication and an addition there into one step that rounds once.
