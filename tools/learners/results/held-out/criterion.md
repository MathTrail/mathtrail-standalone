# The criterion, read on this run

Seed 20261003, experiment held-out: 4000 children a generator, 200 answers each. The score is the share of the way from the baseline, floor_0.05/both — the step chosen under the service's rule of mastery — to perfect, no false mastery and no wait, a rule closes, with its 95 % interval.

## The confirmation

floor_0.05+cautious_z1/both: 71 of 71 constraints met, the score 0.690 [0.686, 0.695]. **Confirmed:** it meets every constraint and is better than the baseline on the held-out children.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.010 [0.008, 0.012] | 60 of 71 | 11 | 0 |
| constant_slow/both | 0.010 [0.007, 0.013] | 64 of 71 | 7 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 69 of 71 | 2 | 0 |
| floor_0.05+cautious_z1/both | 0.690 [0.686, 0.695] | 71 of 71 | 0 | 0 |
| oracle/both (the ceiling) | -0.032 [-0.036, -0.028] | 39 of 71 | 24 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. masteries declared falsely at most 20 % — G0, r4_false, bound 0.200
2. the answers until mastery at most 1.5 times the baseline's — G0, r5_late_answers, bound 14.885
3. masteries declared falsely at most 20 % — G0-topics1, r4_false, bound 0.200
4. the answers until mastery at most 1.5 times the baseline's — G0-topics1, r5_late_answers, bound 13.798
5. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
6. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_rank_6_20, bound 3.630
7. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
8. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_rank_150_200, bound 3.630
9. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
10. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.382
11. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
12. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.382

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.641, not reached | 9.940, reached | 0.592, not reached | 9.206, reached | 73.000, reached | 3.630, on the edge | 37.000, reached | 0.688, reached | 73.000, reached | 3.382, on the edge | 36.000, reached | 0.673, reached |
| constant_slow/both | 0.654, not reached | 9.433, reached | 0.606, not reached | 8.894, reached | 42.000, reached | 2.675, reached | 42.000, reached | 3.518, on the edge | 42.000, reached | 2.545, reached | 42.000, reached | 3.386, on the edge |
| floor_0.05/both | 0.650, not reached | 9.923, reached | 0.600, not reached | 9.199, reached | 73.000, reached | 3.630, on the edge | 41.000, reached | 1.743, reached | 73.000, reached | 3.382, on the edge | 40.000, reached | 1.536, reached |
| floor_0.05+cautious_z1/both | 0.038, reached | 4.685, reached | 0.060, reached | 4.898, reached | 73.000, reached | 3.630, on the edge | 41.000, reached | 1.675, reached | 73.000, reached | 3.382, on the edge | 40.000, reached | 1.450, reached |
| oracle/both (the ceiling) | 0.743, not reached | 6.502, reached | 0.727, not reached | 6.542, reached | — | — | — | — | — | — | — | — |

## Not worse than the baseline

What of not worse each rule does not meet, or the run cannot read: its difference from the baseline, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 50 of 59 | G2 r1_rms_200 0.240 [0.236, 0.244] against 0.030; G2 r3_inside -0.022 [-0.023, -0.021] against 0.010; G3 r1_rms_200 0.123 [0.119, 0.127] against 0.030; G3 r3_inside -0.014 [-0.015, -0.013] against 0.010; G2-half r1_rms_200 0.108 [0.104, 0.111] against 0.030; G2-half r3_inside -0.013 [-0.014, -0.012] against 0.010; G2-fading r1_rms_200 0.072 [0.069, 0.076] against 0.030; G2-fading r3_inside -0.010 [-0.011, -0.009] against 0.010; G3-drop r1_rms_200 0.090 [0.086, 0.094] against 0.030 |
| constant_slow/both | 55 of 59 | G1 r4_false 0.021 [0.015, 0.027] against 0.020; G2 r4_false 0.017 [0.012, 0.022] against 0.020; G0-topics1 r1_rms_200 0.066 [0.062, 0.070] against 0.030; G0-topics1 r3_inside -0.017 [-0.019, -0.015] against 0.010 |
| floor_0.05/both | 59 of 59 |  |
| floor_0.05+cautious_z1/both | 59 of 59 |  |
| oracle/both (the ceiling) | 37 of 59 | G0 r4_false 0.093 [0.085, 0.102] against 0.020; G0 r5_never 0.418 [0.406, 0.431] against 0.020; G1 r4_false 0.052 [0.045, 0.059] against 0.020; G2 r4_false 0.228 [0.221, 0.235] against 0.020; G3 r4_false 0.186 [0.178, 0.194] against 0.020; G4 r4_false 0.154 [0.145, 0.164] against 0.020; G5 r4_false 0.110 [0.101, 0.120] against 0.020; G6 r3_inside -0.067 [-0.071, -0.063] against 0.010; G6 r4_false 0.011 [0.002, 0.020] against 0.020; G7 r4_false 0.029 [0.021, 0.038] against 0.020; G8 r4_false 0.096 [0.087, 0.104] against 0.020; G0-exact r4_false 0.097 [0.089, 0.105] against 0.020; G0-miss0.25 r4_false 0.098 [0.090, 0.106] against 0.020; G0-miss1 r4_false 0.105 [0.097, 0.114] against 0.020; G2-half r4_false 0.181 [0.174, 0.188] against 0.020; G2-fading r4_false 0.155 [0.148, 0.162] against 0.020; G3-drop r4_false 0.060 [0.052, 0.068] against 0.020; G0-start0.5 r4_false 0.071 [0.063, 0.079] against 0.020; G0-start2 r4_false 0.113 [0.105, 0.121] against 0.020; G0-topics0.3 r4_false 0.100 [0.091, 0.109] against 0.020; G0-topics1 r4_false 0.127 [0.119, 0.135] against 0.020; G0-topics1 r5_never 0.433 [0.422, 0.444] against 0.080 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the baseline, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the baseline passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 4000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0 | r5_never | 0.020 | 0.0042 | 0.020 | 99.8 % | 99.8 % |
| G1 | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G1 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.030 | 0.0028 | 0.030 | 100.0 % | 100.0 % |
| G2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0026 | 0.020 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.030 | 0.0025 | 0.030 | 100.0 % | 100.0 % |
| G3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.030 | 0.0026 | 0.030 | 100.0 % | 100.0 % |
| G4 | r3_inside | 0.010 | 0.0007 | 0.010 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.030 | 0.0021 | 0.030 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.030 | 0.0030 | 0.030 | 100.0 % | 100.0 % |
| G6 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.030 | 0.0023 | 0.030 | 100.0 % | 100.0 % |
| G7 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0033 | 0.020 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0012 | 0.010 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.030 | 0.0022 | 0.030 | 100.0 % | 100.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0005 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.030 | 0.0024 | 0.030 | 100.0 % | 100.0 % |
| G2-half | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.030 | 0.0021 | 0.030 | 100.0 % | 100.0 % |
| G2-fading | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.030 | 0.0027 | 0.030 | 100.0 % | 100.0 % |
| G3-drop | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.030 | 0.0020 | 0.030 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.030 | 0.0022 | 0.030 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.030 | 0.0019 | 0.030 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.030 | 0.0023 | 0.030 | 100.0 % | 100.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r5_never | 0.080 | 0.0040 | 0.080 | 100.0 % | 100.0 % |
