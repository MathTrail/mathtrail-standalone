# The criterion, read on this run

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service. **The exit: floor_0.05/both**, the floor of the highest score, 0.000 [0.000, 0.000], among those that meet the constraints of not worse and of the screen.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| earlier/both (the service's earlier rule) | -0.092 [-0.145, -0.038] | 45 of 49 | 4 | 0 |
| constant/both | 0.147 [0.070, 0.227] | 38 of 49 | 11 | 0 |
| constant_slow/both | 0.045 [-0.006, 0.092] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.369 [-0.533, -0.206] | 44 of 49 | 5 | 0 |
| glicko2_floor/general | 0.107 [0.021, 0.190] | 41 of 49 | 8 | 0 |
| glicko2_floor/topics | — | 30 of 49 | 15 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.895, 1.111] | 41 of 49 | 0 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.249
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.445
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 74.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 6.667
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 74.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 6.667
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 2.000
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 2.000

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.554, not reached | 0.385, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 4.510, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| earlier/both (the service's earlier rule) | 0.694, not reached | 0.318, not reached | 0.600, not reached | 74.000, on the edge | 6.667, on the edge | 37.000, reached | 0.784, reached | 73.000, on the edge | 2.000, on the edge | 37.000, reached | 0.784, on the edge |
| constant/both | 0.416, not reached | 0.392, not reached | 0.100, on the edge | 84.000, not reached | 7.333, on the edge | 84.000, not reached | 9.608, on the edge | 85.000, not reached | 3.333, on the edge | 83.000, not reached | 8.431, not reached |
| constant_slow/both | 0.472, not reached | 0.374, not reached | 0.100, on the edge | 42.000, reached | 4.000, on the edge | 42.000, reached | 5.098, on the edge | 43.000, reached | 2.000, on the edge | 41.000, reached | 6.078, not reached |
| floor_0.05/both | 0.554, not reached | 0.385, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 4.510, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| no_trial/both | 0.648, not reached | 0.382, not reached | 0.500, on the edge | 73.000, on the edge | 4.000, on the edge | 37.000, reached | 0.000, reached | 74.000, on the edge | 6.000, on the edge | 30.000, reached | 0.588, reached |
| glicko2_floor/general | 0.480, not reached | 0.398, not reached | 0.100, on the edge | 86.000, on the edge | 16.667, not reached | 21.000, reached | 4.314, on the edge | 82.000, on the edge | 12.667, not reached | 20.000, reached | 1.961, on the edge |
| glicko2_floor/topics | 0.790, not reached | 0.310, not reached | 0.600, not reached | 324.000, not reached | — | 54.000, reached | — | 323.000, not reached | — | 36.000, reached | — |
| floor_0.05+cautious_z1/both | 0.554, not reached | 0.385, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 4.510, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.609, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 37 of 38 | G2 r1_rms_200 0.239 [0.140, 0.346] against 0.126 |
| constant/both | 37 of 38 | G0-topics0.3 r1_rms_200 0.221 [0.150, 0.298] against 0.101 |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 38 of 38 |  |
| glicko2_floor/general | 36 of 38 | G0-topics1 r1_rms_200 0.303 [0.199, 0.417] against 0.139; G0-topics1 r3_inside -0.069 [-0.110, -0.031] against 0.027 |
| glicko2_floor/topics | 28 of 38 | G1 r1_rms_200 0.296 [0.187, 0.409] against 0.155; G2 r1_rms_200 0.517 [0.362, 0.674] against 0.126; G3 r1_rms_200 0.285 [0.161, 0.411] against 0.143; G8 r1_rms_200 0.257 [0.109, 0.416] against 0.103; G0-exact r1_rms_200 0.359 [0.221, 0.492] against 0.089; G2-fading r1_rms_200 0.315 [0.169, 0.455] against 0.148; G0-start0.5 r1_rms_200 0.440 [0.332, 0.555] against 0.150; G0-start2 r1_rms_200 0.293 [0.146, 0.484] against 0.104; G0-topics0.3 r1_rms_200 0.290 [0.194, 0.383] against 0.101; G0-topics0.3 r3_inside -0.130 [-0.201, -0.064] against 0.054 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 38 of 38 |  |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0257 | 0.111 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0131 | 0.056 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0360 | 0.155 | 100.0 % | 100.0 % |
| G1 | r3_inside | 0.010 | 0.0263 | 0.113 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0294 | 0.126 | 100.0 % | 100.0 % |
| G2 | r3_inside | 0.010 | 0.0159 | 0.069 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0333 | 0.143 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0193 | 0.083 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0542 | 0.233 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0180 | 0.077 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0185 | 0.080 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0409 | 0.176 | 100.0 % | 99.8 % |
| G6 | r3_inside | 0.010 | 0.0149 | 0.064 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0374 | 0.161 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0165 | 0.071 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0239 | 0.103 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0167 | 0.072 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0207 | 0.089 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0295 | 0.127 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0810 | 0.348 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0249 | 0.107 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0887 | 0.381 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0110 | 0.047 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0483 | 0.208 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0108 | 0.047 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0344 | 0.148 | 100.0 % | 100.0 % |
| G2-fading | r3_inside | 0.010 | 0.0139 | 0.060 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0551 | 0.237 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0111 | 0.048 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0349 | 0.150 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0145 | 0.063 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0241 | 0.104 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0219 | 0.094 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0235 | 0.101 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0125 | 0.054 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0324 | 0.139 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0063 | 0.027 | 100.0 % | 100.0 % |
