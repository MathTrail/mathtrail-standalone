# Learners

How the service's student model does with children whose true level is known: simulated children answer tasks through the service's own code, and the bench measures how each way of estimating a child places them, keeps their tasks in the corridor, declares topics mastered and follows a child who changes. A change to the rating, the rule or mastery is measured here before it ships.

## What it is, and where it comes from

The bench in `tools/learners` is a Go module of its own that imports the product, as the load tool does (R124, R155). It was carried over from the research program's offline experiment with simulated learners, `research/experiments/learnersim`, whose protocol, [`research/experiments/PROTOCOL-A-offline.md`](../research/experiments/PROTOCOL-A-offline.md) (section 2), defines every measure in full. The research copy stays as it is: it describes the rule the paper was written about. The bench follows the product as it changes.

Every step of a lesson goes through the product's domain code, the same as in the service:

1. a new profile at the start the child's grade gives (`profile.New`);
2. the rule's brief for the next task (`tutor.Next`);
3. the request opened and a task issued at the brief's point (`Profile.Ask`, `Profile.Issue`), sealed with a key made for the run;
4. the child's answer recorded (`Profile.Record`): the trial series, the step, the runs of mastery and the failures, as the service does them.

A task is answered two minutes after it is issued, the next is issued at once, and every tenth answer starts a new day. A child gives 200 answers.

## The children

Every child is synthetic and drawn from a seed of its own, and the same children are run under every rule, meeting the same random draws answer by answer. The tasks themselves differ between rules, since each rule chooses from its own estimate.

The **base population**: a grade from 1 to 6, and the start the service gives that grade; a true level of the start plus N(0, 1); a true offset per topic of N(0, 0.5²); and a task that is truly harder or easier than the difficulty asked for, by N(0, 0.5²), since a chat's model does not hit the difficulty it is asked for. A child answers correctly with chance 0.2 + 0.8·σ(level + offset − difficulty), and never asks for the hint.

The **generators** each change the base population in one respect:

| Generator | The children |
|---|---|
| G0 | stay put: the model's own assumptions, but for the model's misses |
| G1 | were placed a level off: their level is 2.5 below the start for half of them, 2.5 above it for the rest |
| G2 | learn: their level grows by 0.01 an answer, the topic answered by 0.02 more, and the other topics slip back by 1 % of the way to where they began |
| G3 | jump: their level grows by 1.0 at once, at an answer from the 50th to the 150th |
| G4 | meet two hosts in turn, the second writing its tasks harder by 0.75 |
| G5 | are strong in one half of the topics and weak in the other, by 0.75 each way |
| G6 | answer with another slope than the model's: 0.5 for half of them, 2 for the rest |
| G7 | guess otherwise than the model: never for half of them, three times in ten for the rest |
| G8 | ask for the hint on one answer in five, which keeps a right answer from counting towards mastery |

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

The step rules start from the trial series' estimate, as the service does; Glicko-2 starts from the first answer. Every rule runs on all nine generators: 63 cells.

## The measures

A topic is **truly mastered** at a level when the child's true chance on a task of that level and topic at difficulty 3, written as asked, is at least 0.775, the corridor's middle. The names are those of `cells.csv`.

| Measure | Names | What it reads |
|---|---|---|
| R1 error of the level | `r1_rms_N`, `r1_mean_N` | Over the topics the child has answered, the root mean square and the mean of the estimate less the truth, after 5, 10, 20, 50, 100 and 200 answers |
| R2 calibration | `r2_brier`, `r2_bias`, `r2_ece` | The Brier score of the rule's own chance against the answers, the chance it predicted less the share right, and the calibration error over ten bins |
| R3 corridor | `r3_inside`, `r3_below_0.5`, `r3_above_0.95` | The share of tasks whose true chance lies in 0.70–0.85, below 0.50 and above 0.95 |
| R4 false "mastered" | `r4_false`, `r4_false_0.70`, `r4_declared` | Of the masteries declared, the share whose topic was not truly mastered at its level, or not at 0.70; how many were declared |
| R4b false "mastered" within m attempts | `r4b_3` … `r4b_50`, `r4b_chance_*` | The chance that a topic not truly mastered is declared mastered within its first m tasks set at a chance of at most 0.775, the ones mastery counts, and the true chances on those tasks |
| R5 late mastery | `r5_late_answers`, `r5_never` | For a topic and level the child truly masters, the answers in the topic until the rule declares it, and the share never declared |
| R6 lag | `r6_lag`, `r6_jump_answers`, `r6_jump_unsettled` | The estimate less the truth from the 101st answer on, below zero when the estimate lags behind; after a jump, the answers until the error stays under 0.5 for ten answers, and the share of children it never does for |
| R7 placement | `r7_error_5`, `r7_error_10`, `r7_longest_wrong`, `r7_hard_first` | The error of the overall level after 5 and 10 answers, the longest run of wrong answers among the first 15, and the share of the first 10 tasks whose true chance is below 0.50 |

Every number is read off all the children of its cell, with a 95 % interval from 2,000 resamples of the children. A comparison is the difference between two cells over the same children, with an interval from the same resamples taken in both. Four groups of comparisons are made on every run, in `comparisons.csv`:

1. on G0, the service against every other rule: R1 after 200 answers, R3, R4;
2. on G2, the service against every other rule: R6, R3;
3. on G1, the service against no trial series: R7;
4. the floor under the step against the service: R6 on G3, R1 after 200 answers on G0.

A comparison is read as finding a difference when its interval leaves zero out. None is corrected for the others.

## Running it

- `just learners` runs every cell, a thousand children each, writes `tools/learners/results/` over the run kept there, and prints the summary. It takes about two minutes on the development machine.
- Arguments go to the bench: `-children` and `-answers` (at least 200, the answer the comparisons read the error at), `-seed` and `-experiment`, and `-out`. A quick look goes elsewhere, so that the run kept is a whole one: `just learners -children 100 -out /tmp/learners`.
- `just learners-test` runs the bench's tests with the race detector, and `just learners-lint` holds the module to what the service is held to. CI runs both on every pull request. A change to the product's `go.mod` is followed by `just learners-tidy`.

## The results

`tools/learners/results/` keeps the whole run of the rule the service has now, as the line every change to the student model is measured from:

- `summary.md`, the numbers a change is judged by, for every rule: the error after 200 answers on G0, the share in the corridor on G0 and G2, the share of false masteries on G0, the lag on G2, and the share of G3's children not caught up after the jump;
- `cells.csv`, every measure of every cell, with its interval;
- `comparisons.csv`, the four groups of comparisons;
- `run.txt`, what the run was given and what computed it: the seed, the name, the children and answers, the service's version, Go's, and the processor's architecture.

A change to the rule, the rating, the profile or the catalog's topics moves the numbers. The test `TestASmallRunIsSummedUpAsItsSnapshot` holds the summary of a small run — ten children a cell — to `tools/learners/testdata/summary.md`, so such a change fails `just learners-test` until the snapshot is rewritten with `go test -run TestASmallRunIsSummedUpAsItsSnapshot -update` in `tools/learners`; the rewritten snapshot, and a new whole run when the change is to the student model, are part of the change's review. `TestStepRulesAreTheServicesUpdate` holds the bench's copy of the step to `rating.Update` at the service's constants, so a new step in the service fails it until the bench's copy follows.

## The paper's numbers

A run given neither `-seed` nor `-experiment` draws the children of the paper's run: seed 20261001, experiment E-A3. On amd64 the bench's cells and the comparisons it shares with the research run repeat `research/experiments/learnersim/results` to the last digit, so the summary's first row is the paper's account of the service's rule. On arm64 the last digits may differ, since Go fuses a multiplication and an addition there into one step that rounds once.
