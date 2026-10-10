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
| earlier/both (the service's earlier rule) | -0.111 [-0.165, -0.059] | 46 of 49 | 3 | 0 |
| constant/both | 0.166 [0.083, 0.250] | 39 of 49 | 10 | 0 |
| constant_slow/both | 0.073 [0.026, 0.121] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.402 [-0.573, -0.239] | 44 of 49 | 5 | 0 |
| glicko2_floor/general | 0.068 [-0.028, 0.163] | 40 of 49 | 9 | 0 |
| glicko2_floor/topics | — | 27 of 49 | 18 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.895, 1.110] | 41 of 49 | 0 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.267
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.441
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
| shrinking/both (the service) | 0.593, not reached | 0.377, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 1.176, reached | 73.000, on the edge | 2.000, on the edge | 37.000, reached | 0.784, on the edge |
| earlier/both (the service's earlier rule) | 0.683, not reached | 0.321, not reached | 0.700, not reached | 74.000, on the edge | 6.667, on the edge | 37.000, reached | 2.549, reached | 73.000, on the edge | 2.000, on the edge | 35.000, reached | 1.176, on the edge |
| constant/both | 0.369, on the edge | 0.436, on the edge | 0.100, on the edge | 83.000, not reached | 7.333, on the edge | 84.000, not reached | 8.039, on the edge | 85.000, not reached | 4.667, on the edge | 83.000, not reached | 7.451, not reached |
| constant_slow/both | 0.484, not reached | 0.399, on the edge | 0.200, on the edge | 42.000, reached | 4.667, on the edge | 42.000, reached | 4.706, on the edge | 42.000, reached | 1.333, on the edge | 41.000, reached | 4.510, on the edge |
| floor_0.05/both | 0.593, not reached | 0.377, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 1.176, reached | 73.000, on the edge | 2.000, on the edge | 37.000, reached | 0.784, on the edge |
| no_trial/both | 0.713, not reached | 0.367, not reached | 0.600, not reached | 73.000, on the edge | 4.000, on the edge | 36.000, reached | 3.137, on the edge | 74.000, on the edge | 6.000, on the edge | 30.000, reached | 1.961, on the edge |
| glicko2_floor/general | 0.511, not reached | 0.391, not reached | 0.100, on the edge | 86.000, on the edge | 16.667, not reached | 21.000, reached | 4.706, on the edge | 82.000, on the edge | 12.667, not reached | 20.000, reached | 3.333, on the edge |
| glicko2_floor/topics | 0.694, not reached | 0.326, not reached | 0.400, on the edge | 324.000, not reached | — | 45.000, reached | — | 323.000, not reached | — | 32.000, reached | — |
| floor_0.05+cautious_z1/both | 0.593, not reached | 0.377, on the edge | 0.400, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 1.176, reached | 73.000, on the edge | 2.000, on the edge | 37.000, reached | 0.784, on the edge |
| oracle/both (the ceiling) | 0.000, reached | 0.614, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 38 of 38 |  |
| constant/both | 38 of 38 |  |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 38 of 38 |  |
| glicko2_floor/general | 36 of 38 | G0-topics1 r1_rms_200 0.350 [0.206, 0.518] against 0.155; G0-topics1 r3_inside -0.051 [-0.082, -0.024] against 0.021 |
| glicko2_floor/topics | 25 of 38 | G0 r1_rms_200 0.541 [0.410, 0.676] against 0.100; G1 r1_rms_200 0.524 [0.297, 0.786] against 0.140; G2 r1_rms_200 0.646 [0.409, 0.877] against 0.247; G3 r1_rms_200 0.756 [0.508, 1.013] against 0.155; G8 r1_rms_200 0.573 [0.383, 0.795] against 0.105; G0-exact r1_rms_200 0.500 [0.305, 0.676] against 0.131; G2-half r1_rms_200 0.709 [0.463, 0.934] against 0.189; G2-fading r1_rms_200 0.587 [0.368, 0.814] against 0.093; G3-drop r1_rms_200 0.412 [0.201, 0.630] against 0.167; G0-start0.5 r1_rms_200 0.339 [0.235, 0.454] against 0.180; G0-start2 r1_rms_200 0.445 [0.230, 0.685] against 0.166; G0-topics0.3 r1_rms_200 0.583 [0.466, 0.701] against 0.107; G0-topics1 r1_rms_200 0.341 [0.205, 0.504] against 0.155 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 38 of 38 |  |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0232 | 0.100 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0124 | 0.053 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0325 | 0.140 | 100.0 % | 100.0 % |
| G1 | r3_inside | 0.010 | 0.0245 | 0.105 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0575 | 0.247 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0188 | 0.081 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0361 | 0.155 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0181 | 0.078 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0505 | 0.217 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0050 | 0.021 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0132 | 0.057 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0114 | 0.049 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0540 | 0.232 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0121 | 0.052 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0467 | 0.201 | 100.0 % | 99.0 % |
| G7 | r3_inside | 0.010 | 0.0133 | 0.057 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0244 | 0.105 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0306 | 0.131 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0302 | 0.130 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0511 | 0.220 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0208 | 0.089 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0779 | 0.335 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0131 | 0.056 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0439 | 0.189 | 100.0 % | 99.5 % |
| G2-half | r3_inside | 0.010 | 0.0135 | 0.058 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0216 | 0.093 | 100.0 % | 100.0 % |
| G2-fading | r3_inside | 0.010 | 0.0126 | 0.054 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0388 | 0.167 | 100.0 % | 99.9 % |
| G3-drop | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0419 | 0.180 | 100.0 % | 99.8 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0151 | 0.065 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0386 | 0.166 | 100.0 % | 99.9 % |
| G0-start2 | r3_inside | 0.010 | 0.0182 | 0.078 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0248 | 0.107 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0073 | 0.031 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0361 | 0.155 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0050 | 0.021 | 100.0 % | 100.0 % |
