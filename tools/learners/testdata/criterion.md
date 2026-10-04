# The criterion, read on this run

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service. **The exit: floor_0.05/both**, the floor of the highest score, 0.000 [0.000, 0.000], among those that meet the constraints of not worse and of the screen.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 47 of 49 | 2 | 0 |
| earlier/both (the service's earlier rule) | -0.108 [-0.168, -0.044] | 45 of 49 | 4 | 0 |
| constant/both | 0.170 [0.090, 0.255] | 40 of 49 | 9 | 0 |
| constant_slow/both | 0.080 [0.026, 0.136] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 47 of 49 | 2 | 0 |
| no_trial/both | -0.353 [-0.510, -0.192] | 43 of 49 | 6 | 0 |
| glicko2_floor/general | 0.084 [-0.013, 0.178] | 39 of 49 | 10 | 0 |
| glicko2_floor/topics | — | 26 of 49 | 19 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 47 of 49 | 2 | 0 |
| oracle/both (the ceiling) | 1.000 [0.904, 1.098] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.249
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.440
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
| shrinking/both (the service) | 0.553, not reached | 0.378, not reached | 0.100, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 3.725, on the edge | 72.000, on the edge | 2.000, on the edge | 38.000, reached | 1.961, on the edge |
| earlier/both (the service's earlier rule) | 0.691, not reached | 0.332, not reached | 0.700, not reached | 74.000, on the edge | 6.667, on the edge | 36.000, reached | 0.392, reached | 72.000, on the edge | 2.000, on the edge | 36.000, reached | 1.176, on the edge |
| constant/both | 0.312, on the edge | 0.404, not reached | 0.100, on the edge | 83.000, not reached | 7.333, on the edge | 83.000, not reached | 5.882, on the edge | 84.000, not reached | 6.000, on the edge | 84.000, not reached | 6.667, not reached |
| constant_slow/both | 0.534, not reached | 0.385, on the edge | 0.100, on the edge | 42.000, reached | 4.667, on the edge | 42.000, reached | 4.118, on the edge | 42.000, reached | 1.333, on the edge | 41.000, reached | 4.510, on the edge |
| floor_0.05/both | 0.553, not reached | 0.378, not reached | 0.100, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 3.725, on the edge | 72.000, on the edge | 2.000, on the edge | 38.000, reached | 1.961, on the edge |
| no_trial/both | 0.653, not reached | 0.381, not reached | 0.900, not reached | 73.000, on the edge | 4.000, on the edge | 36.000, reached | 2.941, on the edge | 74.000, on the edge | 6.000, on the edge | 32.000, reached | 0.392, reached |
| glicko2_floor/general | 0.461, not reached | 0.395, not reached | 0.300, on the edge | 86.000, on the edge | 18.000, not reached | 20.000, reached | 7.451, on the edge | 82.000, on the edge | 12.667, not reached | 20.000, reached | 5.294, on the edge |
| glicko2_floor/topics | 0.592, not reached | 0.398, not reached | 0.400, on the edge | 324.000, not reached | — | 52.000, reached | — | 323.000, not reached | — | 38.000, reached | — |
| floor_0.05+cautious_z1/both | 0.553, not reached | 0.378, not reached | 0.100, on the edge | 74.000, on the edge | 6.667, on the edge | 41.000, reached | 3.725, on the edge | 72.000, on the edge | 2.000, on the edge | 38.000, reached | 1.961, on the edge |
| oracle/both (the ceiling) | 0.000, reached | 0.606, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 37 of 38 | G2-half r1_rms_200 0.156 [0.057, 0.244] against 0.050 |
| constant/both | 38 of 38 |  |
| constant_slow/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 37 of 38 | G3 r1_rms_200 0.175 [0.102, 0.259] against 0.083 |
| glicko2_floor/general | 37 of 38 | G0-topics1 r1_rms_200 0.375 [0.196, 0.613] against 0.135 |
| glicko2_floor/topics | 24 of 38 | G0 r1_rms_200 0.422 [0.235, 0.622] against 0.106; G1 r1_rms_200 0.524 [0.309, 0.759] against 0.213; G2 r1_rms_200 0.651 [0.419, 0.880] against 0.222; G3 r1_rms_200 0.742 [0.493, 0.984] against 0.083; G4 r1_rms_200 0.340 [0.201, 0.505] against 0.177; G7 r1_rms_200 0.293 [0.199, 0.397] against 0.124; G8 r1_rms_200 0.570 [0.446, 0.712] against 0.103; G0-exact r1_rms_200 0.534 [0.352, 0.730] against 0.170; G2-half r1_rms_200 0.734 [0.449, 1.042] against 0.050; G2-fading r1_rms_200 0.531 [0.313, 0.753] against 0.205; G0-start0.5 r1_rms_200 0.440 [0.297, 0.586] against 0.181; G0-start2 r1_rms_200 0.460 [0.264, 0.684] against 0.083; G0-topics0.3 r1_rms_200 0.509 [0.402, 0.621] against 0.092; G0-topics1 r1_rms_200 0.313 [0.149, 0.506] against 0.135 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.146 [-0.248, -0.048] against 0.040 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0246 | 0.106 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0103 | 0.044 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0494 | 0.213 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0194 | 0.083 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0516 | 0.222 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0214 | 0.092 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0193 | 0.083 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0131 | 0.056 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0412 | 0.177 | 100.0 % | 99.8 % |
| G4 | r3_inside | 0.010 | 0.0129 | 0.055 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0186 | 0.080 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0133 | 0.057 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0530 | 0.228 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0093 | 0.040 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0288 | 0.124 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0121 | 0.052 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0239 | 0.103 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0130 | 0.056 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0394 | 0.170 | 100.0 % | 99.9 % |
| G0-exact | r3_inside | 0.010 | 0.0224 | 0.097 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0498 | 0.214 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0124 | 0.053 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0859 | 0.369 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0117 | 0.050 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0115 | 0.050 | 100.0 % | 100.0 % |
| G2-half | r3_inside | 0.010 | 0.0121 | 0.052 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0476 | 0.205 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0091 | 0.039 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0421 | 0.181 | 100.0 % | 99.7 % |
| G3-drop | r3_inside | 0.010 | 0.0121 | 0.052 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0421 | 0.181 | 100.0 % | 99.7 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0157 | 0.067 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0194 | 0.083 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0163 | 0.070 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0215 | 0.092 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0115 | 0.049 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0313 | 0.135 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0139 | 0.060 | 100.0 % | 100.0 % |
