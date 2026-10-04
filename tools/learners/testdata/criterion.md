# The criterion, read on this run

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service. **The exit: floor_0.05/both**, the floor of the highest score, 0.052 [0.020, 0.088], among those that meet the constraints of not worse and of the screen.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| constant/both | 0.215 [0.129, 0.296] | 41 of 49 | 8 | 0 |
| constant_slow/both | 0.175 [0.125, 0.228] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.052 [0.020, 0.088] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.258 [-0.416, -0.098] | 44 of 49 | 5 | 0 |
| glicko2_floor/general | 0.226 [0.144, 0.306] | 42 of 49 | 7 | 0 |
| glicko2_floor/topics | — | 29 of 49 | 16 | 4 |
| floor_0.05+cautious_z1/both | 0.090 [0.031, 0.145] | 47 of 49 | 2 | 0 |
| oracle/both (the ceiling) | 1.000 [0.893, 1.113] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.311
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.405
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 74.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 6.667
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 74.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 6.667
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 72.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 2.000
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 72.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 2.000

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.691, not reached | 0.332, not reached | 0.700, not reached | 74.000, on the edge | 6.667, on the edge | 36.000, reached | 0.392, reached | 72.000, on the edge | 2.000, on the edge | 36.000, reached | 1.176, on the edge |
| constant/both | 0.418, on the edge | 0.409, on the edge | 0.100, on the edge | 83.000, not reached | 7.333, on the edge | 84.000, not reached | 5.490, on the edge | 84.000, not reached | 6.000, on the edge | 82.000, not reached | 9.020, not reached |
| constant_slow/both | 0.505, not reached | 0.382, on the edge | 0.100, on the edge | 42.000, reached | 4.667, on the edge | 42.000, reached | 2.157, reached | 42.000, reached | 1.333, on the edge | 41.000, reached | 3.333, on the edge |
| floor_0.05/both | 0.633, not reached | 0.344, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 5.098, on the edge | 72.000, on the edge | 2.000, on the edge | 39.000, reached | 0.784, on the edge |
| no_trial/both | 0.655, not reached | 0.366, not reached | 0.800, not reached | 73.000, on the edge | 4.000, on the edge | 36.000, reached | 0.392, reached | 74.000, on the edge | 6.000, on the edge | 36.000, reached | 1.373, on the edge |
| glicko2_floor/general | 0.366, on the edge | 0.408, on the edge | 0.200, on the edge | 86.000, on the edge | 18.000, not reached | 20.000, reached | 3.922, on the edge | 82.000, on the edge | 12.667, not reached | 20.000, reached | 6.863, not reached |
| glicko2_floor/topics | 0.514, not reached | 0.378, on the edge | 0.600, not reached | 324.000, not reached | — | 67.000, on the edge | — | 323.000, not reached | — | 54.000, reached | — |
| floor_0.05+cautious_z1/both | 0.553, not reached | 0.378, on the edge | 0.100, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 3.725, on the edge | 72.000, on the edge | 2.000, on the edge | 38.000, reached | 1.961, on the edge |
| oracle/both (the ceiling) | 0.000, reached | 0.602, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| constant/both | 38 of 38 |  |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 38 of 38 |  |
| glicko2_floor/general | 37 of 38 | G0-topics1 r1_rms_200 0.527 [0.308, 0.801] against 0.258 |
| glicko2_floor/topics | 27 of 38 | G0 r1_rms_200 0.334 [0.176, 0.494] against 0.107; G1 r1_rms_200 0.528 [0.290, 0.765] against 0.200; G3 r1_rms_200 0.635 [0.406, 0.867] against 0.248; G6 r3_inside -0.138 [-0.256, -0.048] against 0.037; G8 r1_rms_200 0.435 [0.298, 0.583] against 0.205; G0-exact r1_rms_200 0.350 [0.190, 0.539] against 0.170; G2-half r1_rms_200 0.546 [0.258, 0.863] against 0.216; G0-start0.5 r1_rms_200 0.336 [0.177, 0.517] against 0.167; G0-start2 r1_rms_200 0.353 [0.171, 0.555] against 0.114; G0-start2 r3_inside -0.110 [-0.168, -0.053] against 0.053; G0-topics0.3 r1_rms_200 0.465 [0.354, 0.576] against 0.267 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.152 [-0.225, -0.085] against 0.037 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0248 | 0.107 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0200 | 0.086 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0466 | 0.200 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0195 | 0.084 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0784 | 0.337 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0124 | 0.053 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0577 | 0.248 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0163 | 0.070 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0791 | 0.340 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0177 | 0.076 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0329 | 0.142 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0967 | 0.416 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0087 | 0.037 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0453 | 0.195 | 100.0 % | 99.3 % |
| G7 | r3_inside | 0.010 | 0.0165 | 0.071 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0477 | 0.205 | 100.0 % | 99.0 % |
| G8 | r3_inside | 0.010 | 0.0259 | 0.111 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0396 | 0.170 | 100.0 % | 99.9 % |
| G0-exact | r3_inside | 0.010 | 0.0441 | 0.190 | 100.0 % | 99.5 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0468 | 0.201 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0182 | 0.078 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.1014 | 0.436 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0120 | 0.052 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0503 | 0.216 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0136 | 0.059 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0678 | 0.291 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0151 | 0.065 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0670 | 0.288 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0207 | 0.089 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0389 | 0.167 | 100.0 % | 99.9 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0203 | 0.087 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0265 | 0.114 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0122 | 0.053 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0622 | 0.267 | 100.0 % | 99.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0126 | 0.054 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0599 | 0.258 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0156 | 0.067 | 100.0 % | 100.0 % |
