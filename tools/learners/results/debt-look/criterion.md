# The criterion, read on this run

Seed 20261001, experiment sweep: 300 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the baseline, floor_0.05+cautious_z1/both — the model the service runs — to a card no wrong answer moves, a rule closes, with its 95 % interval.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the baseline: 0.

No candidate meets every constraint and is better than the baseline. **The exit: floor_0.05+cautious_z1/both**, the baseline itself, kept as it is.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 79 of 83 | 4 | 0 |
| constant_slow/both | 0.135 [0.127, 0.142] | 76 of 83 | 7 | 0 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 79 of 83 | 4 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 0.757 [0.745, 0.769] | 57 of 83 | 26 | 0 |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 0.693 [0.682, 0.704] | 71 of 83 | 12 | 0 |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 0.763 [0.751, 0.775] | 67 of 83 | 16 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 0.745 [0.733, 0.757] | 36 of 83 | 47 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 0.756 [0.743, 0.768] | 65 of 83 | 18 | 0 |
| oracle/both (the ceiling) | — | 65 of 83 | 2 | 16 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the card's fall after a wrong answer in answers 6–20 at most 50 % of the baseline's — G0, r8_wrong_fall_6_20, bound 34.272
2. right answers that leave the card still in answers 6–20 at most 10 % — G0, r8_right_still_6_20, bound 0.100
3. the card's fall after a wrong answer in answers 150–200 at most 50 % of the baseline's — G0, r8_wrong_fall_150_200, bound 17.131
4. right answers that leave the card still in answers 150–200 at most 10 % — G0, r8_right_still_150_200, bound 0.100
5. the card's fall after a wrong answer in answers 6–20 at most 50 % of the baseline's — G2-half, r8_wrong_fall_6_20, bound 34.011
6. right answers that leave the card still in answers 6–20 at most 10 % — G2-half, r8_right_still_6_20, bound 0.100
7. the card's fall after a wrong answer in answers 150–200 at most 50 % of the baseline's — G2-half, r8_wrong_fall_150_200, bound 17.825
8. right answers that leave the card still in answers 150–200 at most 10 % — G2-half, r8_right_still_150_200, bound 0.100
9. the bias after 200 answers at most 0.15 of a rank — G0, r1_mean_200, bound 0.143
10. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_move_p95_6_20, bound 74.000
11. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_rank_6_20, bound 3.200
12. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_move_p95_150_200, bound 74.000
13. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_rank_150_200, bound 3.200
14. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
15. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.200
16. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
17. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.200

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 68.543, not reached | 0.000, reached | 34.263, not reached | 0.000, reached | 68.021, not reached | 0.000, reached | 35.650, not reached | 0.000, reached | -0.100, reached | 74.000, reached | 3.200, on the edge | 40.000, reached | 1.863, reached | 73.000, on the edge | 3.200, on the edge | 40.000, reached | 2.072, reached |
| constant_slow/both | 39.998, not reached | 0.000, reached | 40.005, not reached | 0.000, reached | 39.812, not reached | 0.000, reached | 40.083, not reached | 0.000, reached | -0.091, reached | 42.000, reached | 2.644, on the edge | 42.000, reached | 2.961, on the edge | 42.000, reached | 2.644, on the edge | 42.000, reached | 2.980, on the edge |
| floor_0.05+cautious_z1/both | 68.543, not reached | 0.000, reached | 34.263, not reached | 0.000, reached | 68.021, not reached | 0.000, reached | 35.650, not reached | 0.000, reached | -0.100, reached | 74.000, reached | 3.200, on the edge | 40.000, reached | 1.863, reached | 73.000, reached | 3.200, on the edge | 40.000, reached | 2.072, reached |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 19.212, reached | 0.000, reached | 8.197, reached | 0.003, reached | 18.325, reached | 0.000, reached | 6.558, reached | 0.004, reached | 0.140, on the edge | 30.000, reached | 1.222, reached | 14.000, reached | 0.340, reached | 29.000, reached | 1.289, reached | 13.000, reached | 0.627, reached |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 23.431, reached | 0.000, reached | 10.372, reached | 0.001, reached | 22.465, reached | 0.000, reached | 9.011, reached | 0.002, reached | 0.037, reached | 30.000, reached | 1.111, reached | 13.000, reached | 0.451, reached | 29.000, reached | 1.333, reached | 13.000, reached | 0.810, reached |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 19.205, reached | 0.224, not reached | 7.628, reached | 0.690, not reached | 18.330, reached | 0.212, not reached | 6.287, reached | 0.450, not reached | -0.028, reached | 30.000, reached | 1.067, reached | 13.000, reached | 0.307, reached | 29.000, reached | 1.111, reached | 13.000, reached | 0.673, reached |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 19.317, reached | 0.000, reached | 8.987, reached | 0.001, reached | 18.326, reached | 0.000, reached | 7.312, reached | 0.000, reached | 0.397, not reached | 30.000, reached | 1.333, reached | 16.000, reached | 0.425, reached | 29.000, reached | 1.644, reached | 15.000, reached | 0.556, reached |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 19.212, reached | 0.000, reached | 8.417, reached | 0.003, reached | 18.325, reached | 0.000, reached | 6.425, reached | 0.002, reached | 0.132, on the edge | 30.000, reached | 1.222, reached | 15.000, reached | 0.386, reached | 29.000, reached | 1.289, reached | 13.000, reached | 0.634, reached |
| oracle/both (the ceiling) | — | — | — | — | — | — | — | — | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the baseline

What of not worse each rule does not meet, or the run cannot read: its difference from the baseline, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 66 of 66 |  |
| constant_slow/both | 63 of 66 | G0 r7_error_10 0.019 [0.013, 0.024] against 0.013; G1 r7_error_10 0.019 [0.013, 0.025] against 0.013; G0-topics1 r1_rms_200 0.053 [0.038, 0.069] against 0.034 |
| floor_0.05+cautious_z1/both | 66 of 66 |  |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 40 of 66 | G0 r3_inside -0.025 [-0.035, -0.015] against 0.013; G0 r4_false 0.045 [0.027, 0.067] against 0.020; G0 r7_error_10 0.030 [0.020, 0.040] against 0.013; G1 r7_error_10 0.045 [0.034, 0.055] against 0.013; G1 r7_longest_wrong 0.167 [0.110, 0.233] against 0.100; G5 r3_inside -0.040 [-0.051, -0.029] against 0.013; G6 r3_inside -0.051 [-0.065, -0.037] against 0.012; G6 r4_false 0.055 [0.035, 0.077] against 0.020; G7 r3_inside -0.035 [-0.045, -0.026] against 0.013; G7 r4_false 0.040 [0.023, 0.059] against 0.020; G8 r3_inside -0.025 [-0.034, -0.015] against 0.013; G8 r4_false 0.037 [0.023, 0.054] against 0.020; G0-exact r3_inside -0.068 [-0.088, -0.048] against 0.022; G0-miss0.25 r1_rms_200 0.055 [0.034, 0.077] against 0.031; G0-miss0.25 r3_inside -0.051 [-0.065, -0.036] against 0.019; G0-miss0.25 r4_false 0.036 [0.021, 0.054] against 0.020; G3-drop r1_rms_200 0.310 [0.284, 0.334] against 0.035; G3-drop r3_inside -0.077 [-0.085, -0.068] against 0.011; G3-drop r4_false 0.076 [0.060, 0.097] against 0.025; G3-drop r6_jump_answers 39.260 [35.113, 43.283] against 6.209; G0-start0.5 r3_inside -0.027 [-0.036, -0.016] against 0.013; G0-start0.5 r4_false 0.035 [0.020, 0.049] against 0.020; G0-start2 r3_inside -0.027 [-0.037, -0.017] against 0.013; G0-topics1 r1_rms_200 0.064 [0.048, 0.082] against 0.034; G0-topics1 r3_inside -0.027 [-0.035, -0.019] against 0.013; G0-topics1 r4_false 0.044 [0.027, 0.060] against 0.020 |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 54 of 66 | G0 r7_error_10 0.027 [0.017, 0.036] against 0.013; G1 r7_error_10 0.037 [0.028, 0.047] against 0.013; G5 r3_inside -0.027 [-0.035, -0.018] against 0.013; G6 r3_inside -0.037 [-0.048, -0.025] against 0.012; G6 r4_false 0.041 [0.025, 0.060] against 0.020; G7 r3_inside -0.025 [-0.033, -0.017] against 0.013; G0-exact r3_inside -0.050 [-0.066, -0.033] against 0.022; G3-drop r1_rms_200 0.206 [0.186, 0.227] against 0.035; G3-drop r3_inside -0.057 [-0.064, -0.049] against 0.011; G3-drop r4_false 0.056 [0.040, 0.073] against 0.025; G3-drop r6_jump_answers 32.847 [29.260, 36.547] against 6.209; G0-topics1 r3_inside -0.021 [-0.028, -0.014] against 0.013 |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 54 of 66 | G0 r7_error_10 0.028 [0.018, 0.037] against 0.013; G1 r7_error_10 0.040 [0.030, 0.050] against 0.013; G1 r7_longest_wrong 0.167 [0.107, 0.230] against 0.100; G5 r3_inside -0.026 [-0.034, -0.017] against 0.013; G6 r3_inside -0.035 [-0.046, -0.024] against 0.012; G6 r4_false 0.039 [0.022, 0.057] against 0.020; G7 r3_inside -0.025 [-0.033, -0.018] against 0.013; G0-exact r3_inside -0.038 [-0.054, -0.023] against 0.022; G3-drop r1_rms_200 0.159 [0.140, 0.180] against 0.035; G3-drop r3_inside -0.053 [-0.060, -0.046] against 0.011; G3-drop r4_false 0.050 [0.034, 0.067] against 0.025; G3-drop r6_jump_answers 28.283 [24.950, 31.940] against 6.209 |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 20 of 66 | G0 r1_rms_200 0.117 [0.091, 0.143] against 0.033; G0 r3_inside -0.045 [-0.057, -0.034] against 0.013; G0 r4_false 0.089 [0.069, 0.109] against 0.020; G0 r7_error_10 0.029 [0.017, 0.040] against 0.013; G1 r1_rms_200 0.102 [0.078, 0.126] against 0.027; G1 r3_inside -0.025 [-0.036, -0.014] against 0.012; G1 r7_error_10 0.047 [0.036, 0.057] against 0.013; G1 r7_longest_wrong 0.167 [0.107, 0.237] against 0.100; G3 r4_false 0.042 [0.030, 0.055] against 0.020; G4 r3_inside -0.032 [-0.040, -0.023] against 0.011; G4 r4_false 0.031 [0.021, 0.044] against 0.020; G5 r1_rms_200 0.117 [0.089, 0.142] against 0.033; G5 r3_inside -0.055 [-0.067, -0.044] against 0.013; G5 r4_false 0.065 [0.050, 0.081] against 0.026; G6 r3_inside -0.070 [-0.087, -0.053] against 0.012; G6 r4_false 0.079 [0.057, 0.102] against 0.020; G7 r3_inside -0.047 [-0.059, -0.037] against 0.013; G7 r4_false 0.060 [0.043, 0.077] against 0.020; G8 r1_rms_200 0.122 [0.100, 0.145] against 0.030; G8 r3_inside -0.047 [-0.058, -0.036] against 0.013; G8 r4_false 0.066 [0.051, 0.083] against 0.020; G0-exact r1_rms_200 0.145 [0.121, 0.169] against 0.030; G0-exact r3_inside -0.107 [-0.130, -0.084] against 0.022; G0-exact r4_false 0.070 [0.052, 0.088] against 0.020; G0-miss0.25 r1_rms_200 0.151 [0.125, 0.177] against 0.031; G0-miss0.25 r3_inside -0.078 [-0.095, -0.061] against 0.019; G0-miss0.25 r4_false 0.060 [0.044, 0.078] against 0.020; G0-miss1 r4_false 0.052 [0.038, 0.066] against 0.020; G2-half r4_false 0.036 [0.026, 0.047] against 0.020; G2-fading r4_false 0.041 [0.030, 0.054] against 0.020; G3-drop r1_rms_200 0.423 [0.401, 0.446] against 0.035; G3-drop r3_inside -0.105 [-0.114, -0.097] against 0.011; G3-drop r4_false 0.104 [0.085, 0.124] against 0.025; G3-drop r6_jump_answers 47.457 [43.143, 51.530] against 6.209; G0-start0.5 r1_rms_200 0.127 [0.102, 0.151] against 0.029; G0-start0.5 r3_inside -0.044 [-0.055, -0.033] against 0.013; G0-start0.5 r4_false 0.065 [0.048, 0.083] against 0.020; G0-start2 r1_rms_200 0.115 [0.091, 0.139] against 0.034; G0-start2 r3_inside -0.040 [-0.052, -0.028] against 0.013; G0-start2 r4_false 0.063 [0.048, 0.082] against 0.020; G0-topics0.3 r1_rms_200 0.118 [0.093, 0.144] against 0.029; G0-topics0.3 r3_inside -0.038 [-0.051, -0.026] against 0.012; G0-topics0.3 r4_false 0.069 [0.053, 0.086] against 0.020; G0-topics1 r1_rms_200 0.079 [0.058, 0.100] against 0.034; G0-topics1 r3_inside -0.039 [-0.048, -0.031] against 0.013; G0-topics1 r4_false 0.071 [0.055, 0.088] against 0.020 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 48 of 66 | G0 r3_inside -0.025 [-0.035, -0.015] against 0.013; G0 r7_error_10 0.030 [0.019, 0.041] against 0.013; G1 r7_error_10 0.045 [0.034, 0.055] against 0.013; G1 r7_longest_wrong 0.167 [0.107, 0.233] against 0.100; G4 r3_inside -0.022 [-0.030, -0.014] against 0.011; G5 r3_inside -0.038 [-0.049, -0.028] against 0.013; G6 r3_inside -0.050 [-0.064, -0.036] against 0.012; G7 r3_inside -0.036 [-0.047, -0.026] against 0.013; G8 r3_inside -0.029 [-0.039, -0.019] against 0.013; G0-exact r3_inside -0.071 [-0.090, -0.051] against 0.022; G0-miss0.25 r3_inside -0.052 [-0.067, -0.037] against 0.019; G3-drop r1_rms_200 0.298 [0.274, 0.321] against 0.035; G3-drop r3_inside -0.079 [-0.088, -0.070] against 0.011; G3-drop r6_jump_answers 38.790 [34.773, 42.727] against 6.209; G0-start0.5 r3_inside -0.025 [-0.035, -0.016] against 0.013; G0-start2 r3_inside -0.029 [-0.039, -0.019] against 0.013; G0-topics1 r1_rms_200 0.054 [0.040, 0.069] against 0.034; G0-topics1 r3_inside -0.031 [-0.040, -0.023] against 0.013 |
| oracle/both (the ceiling) | 64 of 66 | G6 r3_inside -0.079 [-0.097, -0.062] against 0.012; G6 r4_false 0.077 [0.050, 0.106] against 0.020 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the baseline, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the baseline passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 300 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0076 | 0.033 | 100.0 % | 99.8 % |
| G0 | r3_inside | 0.010 | 0.0031 | 0.013 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0046 | 0.020 | 100.0 % | 100.0 % |
| G0 | r7_error_10 | 0.010 | 0.0029 | 0.013 | 100.0 % | 100.0 % |
| G0 | r7_longest_wrong | 0.100 | 0.0238 | 0.102 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0063 | 0.027 | 100.0 % | 100.0 % |
| G1 | r3_inside | 0.010 | 0.0027 | 0.012 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0018 | 0.020 | 100.0 % | 100.0 % |
| G1 | r7_error_10 | 0.010 | 0.0029 | 0.013 | 100.0 % | 100.0 % |
| G1 | r7_longest_wrong | 0.100 | 0.0187 | 0.100 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0113 | 0.049 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0028 | 0.012 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0016 | 0.020 | 100.0 % | 100.0 % |
| G2 | abs(r6_lag) | 0.010 | 0.0101 | 0.043 | 100.0 % | 99.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0087 | 0.037 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0028 | 0.012 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0034 | 0.020 | 100.0 % | 100.0 % |
| G3 | r6_jump_answers | 1.000 | 1.7177 | 7.386 | 100.0 % | 99.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0092 | 0.040 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0026 | 0.011 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0027 | 0.020 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0078 | 0.033 | 100.0 % | 99.7 % |
| G5 | r3_inside | 0.010 | 0.0031 | 0.013 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0099 | 0.042 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0028 | 0.012 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0036 | 0.020 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0079 | 0.034 | 100.0 % | 99.6 % |
| G7 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0043 | 0.020 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0071 | 0.030 | 100.0 % | 99.9 % |
| G8 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0042 | 0.020 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0070 | 0.030 | 100.0 % | 99.9 % |
| G0-exact | r3_inside | 0.010 | 0.0050 | 0.022 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0038 | 0.020 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0072 | 0.031 | 100.0 % | 99.9 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0044 | 0.019 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0045 | 0.020 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0073 | 0.031 | 100.0 % | 99.9 % |
| G0-miss1 | r3_inside | 0.010 | 0.0020 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0040 | 0.020 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0090 | 0.039 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0031 | 0.014 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0025 | 0.020 | 100.0 % | 100.0 % |
| G2-half | abs(r6_lag) | 0.010 | 0.0102 | 0.044 | 100.0 % | 99.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0085 | 0.036 | 100.0 % | 99.1 % |
| G2-fading | r3_inside | 0.010 | 0.0028 | 0.012 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0035 | 0.020 | 100.0 % | 100.0 % |
| G2-fading | abs(r6_lag) | 0.010 | 0.0090 | 0.039 | 100.0 % | 99.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0082 | 0.035 | 100.0 % | 99.4 % |
| G3-drop | r3_inside | 0.010 | 0.0026 | 0.011 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G3-drop | r6_jump_answers | 1.000 | 1.4439 | 6.209 | 100.0 % | 99.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0068 | 0.029 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0026 | 0.020 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0078 | 0.034 | 100.0 % | 99.7 % |
| G0-start2 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0038 | 0.020 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0068 | 0.029 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0028 | 0.012 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0036 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0080 | 0.034 | 100.0 % | 99.6 % |
| G0-topics1 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0045 | 0.020 | 100.0 % | 100.0 % |
