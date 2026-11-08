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
| earlier/both (the service's earlier rule) | -0.080 [-0.125, -0.037] | 46 of 49 | 3 | 0 |
| constant/both | 0.178 [0.106, 0.257] | 39 of 49 | 10 | 0 |
| constant_slow/both | 0.093 [0.044, 0.146] | 45 of 49 | 4 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.373 [-0.546, -0.194] | 43 of 49 | 6 | 0 |
| glicko2_floor/general | 0.116 [0.020, 0.208] | 40 of 49 | 9 | 0 |
| glicko2_floor/topics | — | 26 of 49 | 19 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.896, 1.113] | 41 of 49 | 0 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.284
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.403
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 74.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 6.667
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 74.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 6.667
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 72.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 2.667
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 72.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 2.667

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.630, not reached | 0.322, not reached | 0.300, on the edge | 74.000, on the edge | 6.667, on the edge | 44.000, reached | 0.392, reached | 72.000, on the edge | 2.667, on the edge | 42.000, reached | 0.980, reached |
| earlier/both (the service's earlier rule) | 0.685, not reached | 0.323, not reached | 0.700, not reached | 74.000, on the edge | 6.667, on the edge | 39.000, reached | 0.784, reached | 72.000, on the edge | 2.667, on the edge | 39.000, reached | 0.000, reached |
| constant/both | 0.390, on the edge | 0.402, on the edge | 0.200, on the edge | 84.000, not reached | 7.333, on the edge | 84.000, not reached | 8.039, on the edge | 85.000, not reached | 4.667, on the edge | 84.000, not reached | 6.078, on the edge |
| constant_slow/both | 0.483, on the edge | 0.363, on the edge | 0.400, on the edge | 42.000, reached | 4.000, on the edge | 41.000, reached | 6.471, on the edge | 42.000, reached | 0.667, reached | 41.000, reached | 5.294, on the edge |
| floor_0.05/both | 0.630, not reached | 0.322, not reached | 0.300, on the edge | 74.000, on the edge | 6.667, on the edge | 44.000, reached | 0.392, reached | 72.000, on the edge | 2.667, on the edge | 42.000, reached | 0.980, reached |
| no_trial/both | 0.637, not reached | 0.374, on the edge | 0.800, not reached | 73.000, on the edge | 4.000, on the edge | 37.000, reached | 2.549, on the edge | 74.000, on the edge | 10.667, not reached | 34.000, reached | 2.353, on the edge |
| glicko2_floor/general | 0.443, not reached | 0.398, on the edge | 0.500, on the edge | 86.000, on the edge | 20.000, not reached | 20.000, reached | 4.706, reached | 82.000, on the edge | 12.000, not reached | 21.000, reached | 6.275, on the edge |
| glicko2_floor/topics | 0.898, not reached | 0.279, not reached | 0.900, not reached | 324.000, not reached | — | 61.000, reached | — | 323.000, not reached | — | 43.000, reached | — |
| floor_0.05+cautious_z1/both | 0.630, not reached | 0.322, on the edge | 0.300, on the edge | 74.000, on the edge | 6.667, on the edge | 44.000, reached | 0.392, reached | 72.000, on the edge | 2.667, on the edge | 42.000, reached | 0.980, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.621, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 38 of 38 |  |
| constant/both | 38 of 38 |  |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 37 of 38 | G3 r1_rms_200 0.246 [0.138, 0.368] against 0.138 |
| glicko2_floor/general | 37 of 38 | G0-topics1 r1_rms_200 0.265 [0.159, 0.382] against 0.126 |
| glicko2_floor/topics | 24 of 38 | G0 r1_rms_200 0.307 [0.210, 0.426] against 0.083; G0 r3_inside -0.129 [-0.166, -0.092] against 0.055; G1 r1_rms_200 0.318 [0.227, 0.419] against 0.213; G1 r3_inside -0.135 [-0.181, -0.082] against 0.082; G2 r1_rms_200 0.553 [0.359, 0.770] against 0.184; G3 r1_rms_200 0.440 [0.252, 0.650] against 0.138; G3 r3_inside -0.096 [-0.136, -0.056] against 0.054; G8 r1_rms_200 0.323 [0.212, 0.449] against 0.111; G0-exact r1_rms_200 0.370 [0.282, 0.461] against 0.212; G2-fading r1_rms_200 0.297 [0.202, 0.396] against 0.153; G0-start0.5 r1_rms_200 0.333 [0.280, 0.383] against 0.211; G0-start2 r1_rms_200 0.410 [0.238, 0.591] against 0.123; G0-start2 r3_inside -0.151 [-0.239, -0.056] against 0.054; G0-topics0.3 r1_rms_200 0.307 [0.194, 0.429] against 0.144 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 38 of 38 |  |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0193 | 0.083 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0128 | 0.055 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0495 | 0.213 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0190 | 0.082 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0428 | 0.184 | 100.0 % | 99.7 % |
| G2 | r3_inside | 0.010 | 0.0171 | 0.073 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0320 | 0.138 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0125 | 0.054 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0524 | 0.225 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0131 | 0.056 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0520 | 0.224 | 100.0 % | 99.0 % |
| G5 | r3_inside | 0.010 | 0.0088 | 0.038 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0495 | 0.213 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0119 | 0.051 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0380 | 0.164 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0191 | 0.082 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0258 | 0.111 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0157 | 0.067 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0493 | 0.212 | 100.0 % | 99.0 % |
| G0-exact | r3_inside | 0.010 | 0.0323 | 0.139 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0191 | 0.082 | 100.0 % | 100.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0162 | 0.070 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0846 | 0.364 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0083 | 0.036 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0338 | 0.145 | 100.0 % | 100.0 % |
| G2-half | r3_inside | 0.010 | 0.0140 | 0.060 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0357 | 0.153 | 100.0 % | 100.0 % |
| G2-fading | r3_inside | 0.010 | 0.0133 | 0.057 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0293 | 0.126 | 100.0 % | 100.0 % |
| G3-drop | r3_inside | 0.010 | 0.0120 | 0.052 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0491 | 0.211 | 100.0 % | 99.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0156 | 0.067 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0287 | 0.123 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0126 | 0.054 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0335 | 0.144 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0135 | 0.058 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0293 | 0.126 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0154 | 0.066 | 100.0 % | 100.0 % |
