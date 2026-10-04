# The criterion, read on this run

Seed 20261001, experiment E-A3: 4000 children a generator, 200 answers each. The score is the share of the way from the baseline, floor_0.05/both — the step chosen under the service's rule of mastery — to perfect, no false mastery and no wait, a rule closes, with its 95 % interval.

## The choice

Candidates that meet every constraint and are better than the baseline: 0.

No candidate meets every constraint and is better than the baseline. **The exit: floor_0.05/both**, the baseline itself, kept as it is.

### Every rule against the main rule of mastery of the highest score

The main rule of mastery of the highest score, whatever it meets: floor_0.05+cautious_z1/both, 0.688 [0.684, 0.693]. Its score less every other rule's, on the same children:

| Rule | Score | The highest main rule of mastery's less this |
|---|---:|---:|
| shrinking/both (the service) | 0.010 [0.007, 0.012] | 0.679 [0.674, 0.684] |
| constant_slow/both | 0.008 [0.005, 0.011] | 0.681 [0.676, 0.685] |
| floor_0.05+run5/both | -0.119 [-0.125, -0.114] | 0.807 [0.801, 0.813] |
| floor_0.05+cautious_z1.28/both | 0.684 [0.680, 0.689] | 0.004 [0.003, 0.005] |
| floor_0.05+cautious_z1.64/both | 0.681 [0.676, 0.685] | 0.007 [0.005, 0.009] |
| floor_0.05+wald/both | 0.388 [0.384, 0.392] | 0.300 [0.297, 0.303] |
| floor_0.05+cautious_z0/both | 0.654 [0.649, 0.658] | 0.035 [0.032, 0.037] |

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.010 [0.007, 0.012] | 56 of 71 | 15 | 0 |
| constant_slow/both | 0.008 [0.005, 0.011] | 63 of 71 | 8 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 69 of 71 | 2 | 0 |
| floor_0.05+run5/both | -0.119 [-0.125, -0.114] | 53 of 71 | 18 | 0 |
| floor_0.05+cautious_z1/both | 0.688 [0.684, 0.693] | 64 of 71 | 7 | 0 |
| floor_0.05+cautious_z1.28/both | 0.684 [0.680, 0.689] | 63 of 71 | 8 | 0 |
| floor_0.05+cautious_z1.64/both | 0.681 [0.676, 0.685] | 65 of 71 | 6 | 0 |
| floor_0.05+wald/both | 0.388 [0.384, 0.392] | 67 of 71 | 4 | 0 |
| floor_0.05+cautious_z0/both | 0.654 [0.649, 0.658] | 57 of 71 | 14 | 0 |
| oracle/both (the ceiling) | -0.037 [-0.040, -0.033] | 39 of 71 | 24 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. masteries declared falsely at most 20 % — G0, r4_false, bound 0.200
2. the answers until mastery at most 1.5 times the baseline's — G0, r5_late_answers, bound 14.727
3. masteries declared falsely at most 20 % — G0-topics1, r4_false, bound 0.200
4. the answers until mastery at most 1.5 times the baseline's — G0-topics1, r5_late_answers, bound 13.956
5. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
6. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_rank_6_20, bound 3.667
7. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
8. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_rank_150_200, bound 3.667
9. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
10. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.633
11. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
12. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.633

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.635, not reached | 9.922, reached | 0.590, not reached | 9.329, reached | 73.000, reached | 3.667, on the edge | 37.000, reached | 0.678, reached | 73.000, reached | 3.633, on the edge | 36.000, reached | 0.604, reached |
| constant_slow/both | 0.656, not reached | 9.505, reached | 0.606, not reached | 8.845, reached | 42.000, reached | 2.737, reached | 42.000, reached | 3.442, reached | 42.000, reached | 2.598, reached | 41.000, reached | 3.224, reached |
| floor_0.05/both | 0.646, not reached | 9.818, reached | 0.598, not reached | 9.304, reached | 73.000, reached | 3.667, on the edge | 41.000, reached | 1.639, reached | 73.000, reached | 3.633, on the edge | 40.000, reached | 1.534, reached |
| floor_0.05+run5/both | 0.676, not reached | 12.058, reached | 0.576, not reached | 11.147, reached | 73.000, reached | 3.667, on the edge | 42.000, reached | 1.633, reached | 73.000, reached | 3.633, on the edge | 41.000, reached | 1.592, reached |
| floor_0.05+cautious_z1/both | 0.037, reached | 4.637, reached | 0.059, reached | 4.907, reached | 73.000, reached | 3.667, on the edge | 41.000, reached | 1.681, reached | 73.000, reached | 3.633, on the edge | 40.000, reached | 1.641, reached |
| floor_0.05+cautious_z1.28/both | 0.023, reached | 4.766, reached | 0.042, reached | 5.065, reached | 73.000, reached | 3.667, on the edge | 41.000, reached | 1.755, reached | 73.000, reached | 3.633, on the edge | 40.000, reached | 1.616, reached |
| floor_0.05+cautious_z1.64/both | 0.011, reached | 4.856, reached | 0.027, reached | 5.211, reached | 73.000, reached | 3.667, on the edge | 41.000, reached | 1.640, reached | 73.000, reached | 3.633, on the edge | 40.000, reached | 1.662, reached |
| floor_0.05+wald/both | 0.007, reached | 10.816, reached | 0.005, reached | 10.713, reached | 73.000, reached | 3.667, on the edge | 42.000, reached | 1.723, reached | 73.000, reached | 3.633, on the edge | 41.000, reached | 1.651, reached |
| floor_0.05+cautious_z0/both | 0.154, reached | 4.193, reached | 0.168, reached | 4.318, reached | 73.000, reached | 3.667, on the edge | 40.000, reached | 1.665, reached | 73.000, reached | 3.633, on the edge | 41.000, reached | 1.601, reached |
| oracle/both (the ceiling) | 0.750, not reached | 6.402, reached | 0.731, not reached | 6.547, reached | — | — | — | — | — | — | — | — |

## Not worse than the baseline

What of not worse each rule does not meet, or the run cannot read: its difference from the baseline, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 46 of 59 | G1 r1_rms_200 0.013 [0.010, 0.016] against 0.010; G2 r1_rms_200 0.241 [0.238, 0.245] against 0.013; G2 r3_inside -0.021 [-0.022, -0.021] against 0.010; G3 r1_rms_200 0.122 [0.118, 0.125] against 0.011; G3 r3_inside -0.014 [-0.015, -0.013] against 0.010; G7 r1_rms_200 0.008 [0.005, 0.011] against 0.010; G2-half r1_rms_200 0.111 [0.107, 0.114] against 0.011; G2-half r3_inside -0.014 [-0.015, -0.013] against 0.010; G2-fading r1_rms_200 0.069 [0.065, 0.072] against 0.010; G2-fading r3_inside -0.009 [-0.010, -0.008] against 0.010; G3-drop r1_rms_200 0.091 [0.087, 0.095] against 0.012; G0-start2 r1_rms_200 0.012 [0.009, 0.014] against 0.010; G0-topics1 r1_rms_200 0.011 [0.008, 0.014] against 0.010 |
| constant_slow/both | 53 of 59 | G2 r4_false 0.016 [0.011, 0.021] against 0.020; G5 r1_rms_200 0.024 [0.020, 0.028] against 0.010; G8 r4_false 0.014 [0.008, 0.021] against 0.020; G2-fading r4_false 0.018 [0.013, 0.024] against 0.020; G0-topics1 r1_rms_200 0.066 [0.062, 0.071] against 0.010; G0-topics1 r3_inside -0.016 [-0.018, -0.014] against 0.010 |
| floor_0.05/both | 59 of 59 |  |
| floor_0.05+run5/both | 43 of 59 | G0 r4_false 0.030 [0.020, 0.040] against 0.020; G0 r5_never 0.542 [0.532, 0.550] against 0.020; G1 r4_false 0.027 [0.015, 0.038] against 0.020; G4 r4_false 0.018 [0.008, 0.028] against 0.020; G5 r4_false 0.022 [0.012, 0.032] against 0.020; G6 r4_false 0.033 [0.023, 0.044] against 0.020; G7 r4_false 0.015 [0.005, 0.025] against 0.020; G8 r4_false 0.038 [0.025, 0.051] against 0.020; G0-exact r4_false 0.018 [0.007, 0.028] against 0.020; G0-miss0.25 r4_false 0.031 [0.021, 0.041] against 0.020; G0-miss1 r4_false 0.023 [0.013, 0.033] against 0.020; G3-drop r4_false 0.068 [0.058, 0.078] against 0.020; G0-start0.5 r4_false 0.020 [0.010, 0.030] against 0.020; G0-start2 r4_false 0.029 [0.018, 0.040] against 0.020; G0-topics0.3 r4_false 0.025 [0.014, 0.035] against 0.020; G0-topics1 r5_never 0.499 [0.490, 0.507] against 0.020 |
| floor_0.05+cautious_z1/both | 52 of 59 | G2 r1_rms_200 0.010 [0.007, 0.014] against 0.013; G3 r1_rms_200 0.010 [0.005, 0.014] against 0.011; G4 r1_rms_200 0.013 [0.009, 0.017] against 0.011; G2-half r1_rms_200 0.014 [0.010, 0.018] against 0.011; G2-fading r1_rms_200 0.018 [0.014, 0.022] against 0.010; G0-topics0.3 r1_rms_200 0.007 [0.003, 0.010] against 0.010; G0-topics1 r5_never 0.053 [0.042, 0.063] against 0.020 |
| floor_0.05+cautious_z1.28/both | 51 of 59 | G0 r5_never 0.013 [0.001, 0.026] against 0.020; G2 r1_rms_200 0.010 [0.006, 0.014] against 0.013; G3 r1_rms_200 0.008 [0.004, 0.013] against 0.011; G4 r1_rms_200 0.013 [0.009, 0.017] against 0.011; G2-half r1_rms_200 0.014 [0.010, 0.017] against 0.011; G2-fading r1_rms_200 0.017 [0.013, 0.020] against 0.010; G0-topics0.3 r1_rms_200 0.008 [0.004, 0.011] against 0.010; G0-topics1 r5_never 0.091 [0.082, 0.102] against 0.020 |
| floor_0.05+cautious_z1.64/both | 53 of 59 | G0 r5_never 0.077 [0.064, 0.089] against 0.020; G4 r1_rms_200 0.013 [0.009, 0.017] against 0.011; G2-half r1_rms_200 0.010 [0.006, 0.014] against 0.011; G2-fading r1_rms_200 0.013 [0.009, 0.016] against 0.010; G0-topics0.3 r1_rms_200 0.007 [0.003, 0.010] against 0.010; G0-topics1 r5_never 0.143 [0.133, 0.154] against 0.020 |
| floor_0.05+wald/both | 55 of 59 | G0 r5_never 0.388 [0.378, 0.397] against 0.020; G3-drop r3_inside -0.009 [-0.010, -0.007] against 0.010; G0-topics1 r3_inside -0.009 [-0.011, -0.008] against 0.010; G0-topics1 r5_never 0.406 [0.398, 0.415] against 0.020 |
| floor_0.05+cautious_z0/both | 45 of 59 | G0 r1_rms_200 0.006 [0.002, 0.010] against 0.010; G1 r1_rms_200 0.012 [0.008, 0.016] against 0.010; G3 r1_rms_200 0.010 [0.006, 0.015] against 0.011; G4 r1_rms_200 0.011 [0.007, 0.015] against 0.011; G7 r1_rms_200 0.010 [0.006, 0.014] against 0.010; G0-exact r1_rms_200 0.009 [0.006, 0.013] against 0.010; G0-miss0.25 r1_rms_200 0.008 [0.005, 0.012] against 0.010; G0-miss1 r1_rms_200 0.008 [0.004, 0.012] against 0.010; G2-half r1_rms_200 0.007 [0.004, 0.011] against 0.011; G2-fading r1_rms_200 0.014 [0.010, 0.018] against 0.010; G0-start0.5 r1_rms_200 0.010 [0.006, 0.013] against 0.010; G0-start2 r1_rms_200 0.008 [0.004, 0.012] against 0.010; G0-topics0.3 r1_rms_200 0.008 [0.005, 0.012] against 0.010; G0-topics1 r1_rms_200 0.007 [0.003, 0.011] against 0.010 |
| oracle/both (the ceiling) | 37 of 59 | G0 r4_false 0.104 [0.095, 0.112] against 0.020; G0 r5_never 0.422 [0.409, 0.435] against 0.020; G1 r4_false 0.049 [0.041, 0.056] against 0.020; G2 r4_false 0.228 [0.221, 0.234] against 0.020; G3 r4_false 0.182 [0.174, 0.190] against 0.020; G4 r4_false 0.154 [0.144, 0.163] against 0.020; G5 r4_false 0.115 [0.106, 0.124] against 0.020; G6 r3_inside -0.065 [-0.070, -0.060] against 0.010; G6 r4_false 0.018 [0.009, 0.027] against 0.020; G7 r4_false 0.036 [0.028, 0.045] against 0.020; G8 r4_false 0.100 [0.091, 0.108] against 0.020; G0-exact r4_false 0.091 [0.083, 0.099] against 0.020; G0-miss0.25 r4_false 0.100 [0.092, 0.108] against 0.020; G0-miss1 r4_false 0.114 [0.106, 0.123] against 0.020; G2-half r4_false 0.188 [0.181, 0.194] against 0.020; G2-fading r4_false 0.163 [0.156, 0.171] against 0.020; G3-drop r4_false 0.065 [0.057, 0.073] against 0.020; G0-start0.5 r4_false 0.070 [0.062, 0.077] against 0.020; G0-start2 r4_false 0.108 [0.100, 0.117] against 0.020; G0-topics0.3 r4_false 0.098 [0.089, 0.107] against 0.020; G0-topics1 r4_false 0.132 [0.125, 0.140] against 0.020; G0-topics1 r5_never 0.446 [0.434, 0.457] against 0.020 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the baseline, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the baseline passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 4000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G0 | r5_never | 0.020 | 0.0040 | 0.020 | 99.9 % | 99.9 % |
| G1 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G1 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0029 | 0.013 | 99.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0026 | 0.020 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0007 | 0.010 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0022 | 0.010 | 99.4 % | 99.4 % |
| G5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0028 | 0.012 | 99.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0023 | 0.010 | 99.1 % | 99.1 % |
| G7 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.7 % | 99.7 % |
| G8 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0033 | 0.020 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0-exact | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 99.9 % | 99.9 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0012 | 0.010 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.8 % | 99.8 % |
| G0-miss1 | r3_inside | 0.010 | 0.0006 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.7 % | 99.7 % |
| G2-fading | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0028 | 0.020 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0028 | 0.012 | 99.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0030 | 0.020 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.7 % | 99.7 % |
| G0-start2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 99.9 % | 99.9 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0023 | 0.010 | 99.1 % | 99.1 % |
| G0-topics1 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0029 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r5_never | 0.020 | 0.0040 | 0.020 | 99.9 % | 99.9 % |
