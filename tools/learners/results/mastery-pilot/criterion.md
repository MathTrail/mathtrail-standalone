# The criterion, read on this run

Seed 20261001, experiment sweep: 1000 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the baseline, floor_0.05/both — the step chosen under the service's rule of mastery — to perfect, no false mastery and no wait, a rule closes, with its 95 % interval.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the baseline: 2.

| Rule | Kind | Score | Numbers added | Fields in the profile | Distance from the service's rule of mastery |
|---|---|---:|---:|---|---:|
| floor_0.05+cautious_z1/both | main | 0.687 [0.678, 0.697] | 1 | no | 1.000 |
| floor_0.05+cautious_z1.28/both | main | 0.681 [0.671, 0.691] | 1 | no | 1.280 |

A*, the main rule of mastery of the highest score: floor_0.05+cautious_z1/both, 0.687 [0.678, 0.697].

No main rule of mastery's score less A*'s holds nothing: A* has no equals.

The simplest of them, the main rule of mastery: floor_0.05+cautious_z1/both — 1 numbers added to the service's rule, no field added to the profile, a distance of 1.000 from the service's rule of mastery.

No backup meets every constraint and is better than the baseline.

**The choice: floor_0.05+cautious_z1/both.**

### The chosen rule's constraints on new children

The chance each constraint holds on as many new children, the likeliest to fail first: the value there falls around the value here with the standard error its interval shows, and a check of not worse holds where the worse end of its interval does. The chance all 71 hold, taken as independent: 1.1 %.

| Constraint | Generator | Measure | Value | Bound | Chance it holds |
|---|---|---|---:|---:|---:|
| not worse than the baseline | G3-drop | r3_inside | -0.009 | 0.010 | 21.1 % |
| the card's move in answers 6–20 no more than the baseline's in answers 6–20 | G0 | r8_move_p95_6_20 | 73.000 | 73.000 | 50.0 % |
| the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 | G0 | r8_rank_6_20 | 3.153 | 3.153 | 50.0 % |
| the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 | G2-half | r8_rank_6_20 | 3.720 | 3.720 | 50.0 % |
| not worse than the baseline | G0-topics1 | r5_never | 0.045 | 0.080 | 79.5 % |
| not worse than the baseline | G2 | r1_rms_200 | 0.018 | 0.030 | 81.2 % |
| not worse than the baseline | G4 | r1_rms_200 | 0.014 | 0.030 | 92.6 % |
| not worse than the baseline | G2-half | r1_rms_200 | 0.014 | 0.030 | 93.9 % |

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.013 [0.009, 0.017] | 61 of 71 | 10 | 0 |
| constant_slow/both | 0.008 [0.002, 0.014] | 66 of 71 | 5 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 69 of 71 | 2 | 0 |
| floor_0.05+run5/both | -0.133 [-0.145, -0.121] | 66 of 71 | 5 | 0 |
| floor_0.05+cautious_z1/both | 0.687 [0.678, 0.697] | 71 of 71 | 0 | 0 |
| floor_0.05+cautious_z1.28/both | 0.681 [0.671, 0.691] | 71 of 71 | 0 | 0 |
| floor_0.05+cautious_z1.64/both | 0.679 [0.669, 0.689] | 69 of 71 | 2 | 0 |
| floor_0.05+wald/both | 0.383 [0.374, 0.392] | 69 of 71 | 2 | 0 |
| floor_0.05+cautious_z0/both | 0.651 [0.641, 0.661] | 71 of 71 | 0 | 0 |
| oracle/both (the ceiling) | -0.038 [-0.046, -0.029] | 41 of 71 | 22 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. masteries declared falsely at most 20 % — G0, r4_false, bound 0.200
2. the answers until mastery at most 1.5 times the baseline's — G0, r5_late_answers, bound 14.836
3. masteries declared falsely at most 20 % — G0-topics1, r4_false, bound 0.200
4. the answers until mastery at most 1.5 times the baseline's — G0-topics1, r5_late_answers, bound 13.870
5. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
6. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_rank_6_20, bound 3.153
7. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
8. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_rank_150_200, bound 3.153
9. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
10. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.720
11. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
12. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.720

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.643, not reached | 9.783, reached | 0.581, not reached | 9.219, reached | 73.000, on the edge | 3.153, on the edge | 37.000, reached | 0.535, reached | 73.000, on the edge | 3.720, on the edge | 35.000, reached | 0.696, reached |
| constant_slow/both | 0.672, not reached | 9.348, reached | 0.596, not reached | 8.766, reached | 42.000, reached | 2.267, reached | 42.000, reached | 3.324, on the edge | 42.000, reached | 2.580, reached | 42.000, reached | 3.163, reached |
| floor_0.05/both | 0.653, not reached | 9.891, reached | 0.591, not reached | 9.247, reached | 73.000, on the edge | 3.153, on the edge | 41.000, reached | 1.473, reached | 73.000, on the edge | 3.720, on the edge | 40.000, reached | 1.859, reached |
| floor_0.05+run5/both | 0.697, not reached | 11.929, reached | 0.572, not reached | 11.075, reached | 73.000, on the edge | 3.153, on the edge | 42.000, reached | 1.537, reached | 73.000, on the edge | 3.720, on the edge | 41.000, reached | 1.796, reached |
| floor_0.05+cautious_z1/both | 0.040, reached | 4.699, reached | 0.061, reached | 4.876, reached | 73.000, on the edge | 3.153, on the edge | 40.000, reached | 1.557, reached | 73.000, reached | 3.720, on the edge | 40.000, reached | 1.676, reached |
| floor_0.05+cautious_z1.28/both | 0.025, reached | 4.938, reached | 0.044, reached | 4.931, reached | 73.000, on the edge | 3.153, on the edge | 41.000, reached | 1.598, reached | 73.000, on the edge | 3.720, on the edge | 40.000, reached | 1.716, reached |
| floor_0.05+cautious_z1.64/both | 0.010, reached | 5.041, reached | 0.029, reached | 5.075, reached | 73.000, on the edge | 3.153, on the edge | 41.000, reached | 1.465, reached | 73.000, on the edge | 3.720, on the edge | 40.000, reached | 1.761, reached |
| floor_0.05+wald/both | 0.005, reached | 11.025, reached | 0.005, reached | 10.769, reached | 73.000, on the edge | 3.153, on the edge | 42.000, reached | 1.522, reached | 73.000, on the edge | 3.720, on the edge | 41.000, reached | 1.800, reached |
| floor_0.05+cautious_z0/both | 0.160, reached | 4.208, reached | 0.162, reached | 4.263, reached | 73.000, on the edge | 3.153, on the edge | 40.000, reached | 1.439, reached | 73.000, on the edge | 3.720, on the edge | 40.000, reached | 1.682, reached |
| oracle/both (the ceiling) | 0.749, not reached | 6.403, reached | 0.727, not reached | 6.529, reached | — | — | — | — | — | — | — | — |

## Not worse than the baseline

What of not worse each rule does not meet, or the run cannot read: its difference from the baseline, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 51 of 59 | G2 r1_rms_200 0.244 [0.236, 0.252] against 0.030; G2 r3_inside -0.022 [-0.024, -0.020] against 0.010; G3 r1_rms_200 0.117 [0.110, 0.125] against 0.030; G3 r3_inside -0.014 [-0.016, -0.011] against 0.010; G2-half r1_rms_200 0.112 [0.105, 0.119] against 0.030; G2-half r3_inside -0.014 [-0.016, -0.012] against 0.010; G2-fading r1_rms_200 0.064 [0.057, 0.070] against 0.030; G3-drop r1_rms_200 0.088 [0.080, 0.095] against 0.030 |
| constant_slow/both | 57 of 59 | G0-topics1 r1_rms_200 0.069 [0.061, 0.077] against 0.030; G0-topics1 r3_inside -0.016 [-0.020, -0.012] against 0.010 |
| floor_0.05/both | 59 of 59 |  |
| floor_0.05+run5/both | 56 of 59 | G0 r5_never 0.525 [0.507, 0.543] against 0.036; G3-drop r4_false 0.049 [0.030, 0.070] against 0.027; G0-topics1 r5_never 0.498 [0.480, 0.516] against 0.080 |
| floor_0.05+cautious_z1/both | 59 of 59 |  |
| floor_0.05+cautious_z1.28/both | 59 of 59 |  |
| floor_0.05+cautious_z1.64/both | 57 of 59 | G0 r5_never 0.062 [0.037, 0.087] against 0.036; G0-topics1 r5_never 0.142 [0.121, 0.165] against 0.080 |
| floor_0.05+wald/both | 57 of 59 | G0 r5_never 0.372 [0.352, 0.392] against 0.036; G0-topics1 r5_never 0.408 [0.392, 0.424] against 0.080 |
| floor_0.05+cautious_z0/both | 59 of 59 |  |
| oracle/both (the ceiling) | 39 of 59 | G0 r4_false 0.096 [0.080, 0.112] against 0.024; G0 r5_never 0.405 [0.380, 0.430] against 0.036; G1 r4_false 0.063 [0.047, 0.078] against 0.024; G2 r4_false 0.216 [0.202, 0.229] against 0.023; G3 r4_false 0.180 [0.164, 0.196] against 0.024; G4 r4_false 0.153 [0.135, 0.173] against 0.026; G5 r4_false 0.108 [0.090, 0.125] against 0.024; G6 r3_inside -0.068 [-0.076, -0.059] against 0.010; G8 r4_false 0.093 [0.075, 0.111] against 0.030; G0-exact r4_false 0.081 [0.064, 0.098] against 0.024; G0-miss0.25 r4_false 0.093 [0.075, 0.110] against 0.025; G0-miss1 r4_false 0.105 [0.087, 0.121] against 0.025; G2-half r4_false 0.188 [0.175, 0.202] against 0.023; G2-fading r4_false 0.165 [0.151, 0.180] against 0.023; G3-drop r4_false 0.056 [0.039, 0.073] against 0.027; G0-start0.5 r4_false 0.068 [0.054, 0.083] against 0.026; G0-start2 r4_false 0.098 [0.081, 0.115] against 0.024; G0-topics0.3 r4_false 0.106 [0.088, 0.124] against 0.024; G0-topics1 r4_false 0.136 [0.120, 0.151] against 0.026; G0-topics1 r5_never 0.445 [0.422, 0.468] against 0.080 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the baseline, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the baseline passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 1000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.030 | 0.0042 | 0.030 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0056 | 0.024 | 100.0 % | 100.0 % |
| G0 | r5_never | 0.020 | 0.0084 | 0.036 | 100.0 % | 99.7 % |
| G1 | r1_rms_200 | 0.030 | 0.0038 | 0.030 | 100.0 % | 100.0 % |
| G1 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0056 | 0.024 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.030 | 0.0060 | 0.030 | 100.0 % | 100.0 % |
| G2 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0054 | 0.023 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.030 | 0.0055 | 0.030 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0056 | 0.024 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.030 | 0.0049 | 0.030 | 100.0 % | 100.0 % |
| G4 | r3_inside | 0.010 | 0.0014 | 0.010 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0059 | 0.026 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.030 | 0.0047 | 0.030 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0055 | 0.024 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.030 | 0.0057 | 0.030 | 100.0 % | 100.0 % |
| G6 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.030 | 0.0044 | 0.030 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0058 | 0.025 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.030 | 0.0038 | 0.030 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0069 | 0.030 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.030 | 0.0042 | 0.030 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0035 | 0.015 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0056 | 0.024 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.030 | 0.0040 | 0.030 | 100.0 % | 100.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0024 | 0.010 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0058 | 0.025 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.030 | 0.0041 | 0.030 | 100.0 % | 100.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0011 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0057 | 0.025 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.030 | 0.0050 | 0.030 | 100.0 % | 100.0 % |
| G2-half | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0054 | 0.023 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.030 | 0.0043 | 0.030 | 100.0 % | 100.0 % |
| G2-fading | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0054 | 0.023 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.030 | 0.0052 | 0.030 | 100.0 % | 100.0 % |
| G3-drop | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0063 | 0.027 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.030 | 0.0040 | 0.030 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.030 | 0.0041 | 0.030 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0056 | 0.024 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.030 | 0.0040 | 0.030 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0057 | 0.024 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.030 | 0.0042 | 0.030 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0059 | 0.026 | 100.0 % | 100.0 % |
| G0-topics1 | r5_never | 0.080 | 0.0080 | 0.080 | 100.0 % | 100.0 % |
