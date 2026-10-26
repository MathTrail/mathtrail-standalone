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
| earlier/both (the service's earlier rule) | -0.094 [-0.158, -0.035] | 45 of 49 | 4 | 0 |
| constant/both | 0.125 [0.033, 0.221] | 39 of 49 | 10 | 0 |
| constant_slow/both | 0.046 [-0.005, 0.096] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.384 [-0.550, -0.221] | 44 of 49 | 5 | 0 |
| glicko2_floor/general | 0.067 [-0.030, 0.160] | 40 of 49 | 9 | 0 |
| glicko2_floor/topics | — | 30 of 49 | 15 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.895, 1.115] | 41 of 49 | 0 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.256
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
| shrinking/both (the service) | 0.569, not reached | 0.378, on the edge | 0.500, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 3.725, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| earlier/both (the service's earlier rule) | 0.686, not reached | 0.318, not reached | 0.600, not reached | 74.000, on the edge | 6.667, on the edge | 37.000, reached | 0.784, reached | 73.000, on the edge | 2.000, on the edge | 37.000, reached | 0.784, on the edge |
| constant/both | 0.447, not reached | 0.396, not reached | 0.100, on the edge | 83.000, not reached | 7.333, on the edge | 84.000, not reached | 9.412, on the edge | 85.000, not reached | 3.333, on the edge | 83.000, not reached | 7.843, not reached |
| constant_slow/both | 0.492, not reached | 0.375, not reached | 0.100, on the edge | 42.000, reached | 4.000, on the edge | 42.000, reached | 5.098, on the edge | 42.000, reached | 1.333, on the edge | 41.000, reached | 5.294, on the edge |
| floor_0.05/both | 0.569, not reached | 0.378, on the edge | 0.500, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 3.725, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| no_trial/both | 0.648, not reached | 0.382, not reached | 0.500, on the edge | 73.000, on the edge | 4.000, on the edge | 37.000, reached | 0.000, reached | 74.000, on the edge | 6.000, on the edge | 30.000, reached | 0.588, reached |
| glicko2_floor/general | 0.511, not reached | 0.391, not reached | 0.100, on the edge | 86.000, on the edge | 16.667, not reached | 21.000, reached | 4.706, on the edge | 82.000, on the edge | 12.667, not reached | 20.000, reached | 3.333, on the edge |
| glicko2_floor/topics | 0.790, not reached | 0.310, not reached | 0.600, not reached | 324.000, not reached | — | 54.000, reached | — | 323.000, not reached | — | 36.000, reached | — |
| floor_0.05+cautious_z1/both | 0.569, not reached | 0.378, on the edge | 0.500, on the edge | 74.000, on the edge | 6.667, on the edge | 42.000, reached | 3.725, on the edge | 73.000, on the edge | 2.000, on the edge | 40.000, reached | 0.588, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.609, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 37 of 38 | G2 r1_rms_200 0.239 [0.140, 0.346] against 0.131 |
| constant/both | 38 of 38 |  |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 38 of 38 |  |
| glicko2_floor/general | 36 of 38 | G0-topics1 r1_rms_200 0.374 [0.252, 0.500] against 0.091; G0-topics1 r3_inside -0.075 [-0.115, -0.036] against 0.031 |
| glicko2_floor/topics | 28 of 38 | G0 r3_inside -0.092 [-0.108, -0.075] against 0.058; G1 r1_rms_200 0.311 [0.208, 0.416] against 0.177; G2 r1_rms_200 0.517 [0.362, 0.674] against 0.131; G3 r1_rms_200 0.309 [0.171, 0.443] against 0.151; G8 r1_rms_200 0.275 [0.121, 0.435] against 0.104; G0-exact r1_rms_200 0.359 [0.221, 0.492] against 0.088; G0-start0.5 r1_rms_200 0.461 [0.354, 0.568] against 0.192; G0-start2 r1_rms_200 0.292 [0.141, 0.485] against 0.104; G0-topics0.3 r1_rms_200 0.288 [0.191, 0.382] against 0.103; G0-topics0.3 r3_inside -0.133 [-0.204, -0.067] against 0.052 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 38 of 38 |  |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0268 | 0.115 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0135 | 0.058 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0411 | 0.177 | 100.0 % | 99.8 % |
| G1 | r3_inside | 0.010 | 0.0268 | 0.115 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0306 | 0.131 | 100.0 % | 100.0 % |
| G2 | r3_inside | 0.010 | 0.0136 | 0.059 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0351 | 0.151 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0145 | 0.063 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0447 | 0.192 | 100.0 % | 99.4 % |
| G4 | r3_inside | 0.010 | 0.0179 | 0.077 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0177 | 0.076 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0131 | 0.056 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0405 | 0.174 | 100.0 % | 99.9 % |
| G6 | r3_inside | 0.010 | 0.0149 | 0.064 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0352 | 0.152 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0153 | 0.066 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0241 | 0.104 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0162 | 0.070 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0205 | 0.088 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0291 | 0.125 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0486 | 0.209 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0235 | 0.101 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0857 | 0.369 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0120 | 0.052 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0579 | 0.249 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0125 | 0.054 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0446 | 0.192 | 100.0 % | 99.4 % |
| G2-fading | r3_inside | 0.010 | 0.0139 | 0.060 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0425 | 0.183 | 100.0 % | 99.7 % |
| G3-drop | r3_inside | 0.010 | 0.0114 | 0.049 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0446 | 0.192 | 100.0 % | 99.4 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0115 | 0.049 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0242 | 0.104 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0172 | 0.074 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0239 | 0.103 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0120 | 0.052 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0212 | 0.091 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0073 | 0.031 | 100.0 % | 100.0 % |
