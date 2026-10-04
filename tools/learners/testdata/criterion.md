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
| constant/both | 0.184 [0.109, 0.258] | 40 of 49 | 9 | 0 |
| constant_slow/both | 0.136 [0.091, 0.183] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.318 [-0.476, -0.162] | 44 of 49 | 5 | 0 |
| glicko2_floor/general | 0.187 [0.105, 0.265] | 41 of 49 | 8 | 0 |
| glicko2_floor/topics | — | 29 of 49 | 16 | 4 |
| floor_0.05+cautious_z1/both | 0.037 [-0.017, 0.088] | 47 of 49 | 2 | 0 |
| oracle/both (the ceiling) | 1.000 [0.896, 1.110] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.285
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.414
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
| shrinking/both (the service) | 0.633, not reached | 0.344, not reached | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 5.098, on the edge | 72.000, on the edge | 2.000, on the edge | 39.000, reached | 0.784, on the edge |
| constant/both | 0.418, on the edge | 0.409, on the edge | 0.100, on the edge | 83.000, not reached | 7.333, on the edge | 84.000, not reached | 5.490, on the edge | 84.000, not reached | 6.000, on the edge | 82.000, not reached | 9.020, not reached |
| constant_slow/both | 0.505, not reached | 0.382, on the edge | 0.100, on the edge | 42.000, reached | 4.667, on the edge | 42.000, reached | 2.157, reached | 42.000, reached | 1.333, on the edge | 41.000, reached | 3.333, on the edge |
| floor_0.05/both | 0.633, not reached | 0.344, not reached | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 5.098, on the edge | 72.000, on the edge | 2.000, on the edge | 39.000, reached | 0.784, on the edge |
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
| glicko2_floor/general | 37 of 38 | G0-topics1 r1_rms_200 0.526 [0.323, 0.775] against 0.223 |
| glicko2_floor/topics | 27 of 38 | G0 r1_rms_200 0.355 [0.215, 0.504] against 0.075; G1 r1_rms_200 0.584 [0.330, 0.839] against 0.169; G2 r1_rms_200 0.612 [0.367, 0.851] against 0.309; G3 r1_rms_200 0.724 [0.473, 0.984] against 0.225; G6 r3_inside -0.143 [-0.256, -0.064] against 0.057; G8 r1_rms_200 0.479 [0.313, 0.656] against 0.129; G0-exact r1_rms_200 0.382 [0.210, 0.574] against 0.160; G2-half r1_rms_200 0.660 [0.376, 0.982] against 0.227; G2-fading r1_rms_200 0.500 [0.299, 0.706] against 0.268; G0-start0.5 r1_rms_200 0.357 [0.221, 0.520] against 0.094; G0-topics0.3 r1_rms_200 0.463 [0.363, 0.564] against 0.146 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.157 [-0.227, -0.091] against 0.057 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0174 | 0.075 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0163 | 0.070 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0392 | 0.169 | 100.0 % | 99.9 % |
| G1 | r3_inside | 0.010 | 0.0233 | 0.100 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0720 | 0.309 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0154 | 0.066 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0524 | 0.225 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0205 | 0.088 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0549 | 0.236 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0184 | 0.079 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0293 | 0.126 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0103 | 0.044 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0514 | 0.221 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0133 | 0.057 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0355 | 0.153 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0130 | 0.056 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0300 | 0.129 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0185 | 0.080 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0372 | 0.160 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0434 | 0.186 | 100.0 % | 99.6 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0407 | 0.175 | 100.0 % | 99.8 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0130 | 0.056 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0543 | 0.233 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0088 | 0.038 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0528 | 0.227 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0129 | 0.055 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0622 | 0.268 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0166 | 0.071 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0485 | 0.208 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0219 | 0.094 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0163 | 0.070 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0460 | 0.198 | 100.0 % | 99.1 % |
| G0-start2 | r3_inside | 0.010 | 0.0116 | 0.050 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0340 | 0.146 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0108 | 0.047 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0519 | 0.223 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0162 | 0.070 | 100.0 % | 100.0 % |
