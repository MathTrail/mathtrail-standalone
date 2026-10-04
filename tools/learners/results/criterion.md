# The criterion, read on this run

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service. **The exit: floor_0.05/both**, the floor of the highest score, 0.000 [0.000, 0.000], among those that meet the constraints of not worse and of the screen.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| constant/both | 0.175 [0.166, 0.184] | 22 of 49 | 27 | 0 |
| constant_slow/both | 0.108 [0.102, 0.114] | 44 of 49 | 5 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.183 [-0.195, -0.171] | 35 of 49 | 14 | 0 |
| glicko2_floor/general | 0.155 [0.146, 0.165] | 25 of 49 | 24 | 0 |
| glicko2_floor/topics | — | 3 of 49 | 42 | 4 |
| floor_0.05+cautious_z1/both | 0.007 [0.002, 0.012] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.987, 1.012] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.245
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.435
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 3.813
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 3.813
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.600
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.600

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.544, not reached | 0.372, not reached | 0.556, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.635, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.504, reached |
| constant/both | 0.286, not reached | 0.404, not reached | 0.091, reached | 84.000, not reached | 6.940, not reached | 84.000, not reached | 6.818, not reached | 84.000, not reached | 6.267, not reached | 83.000, not reached | 6.496, not reached |
| constant_slow/both | 0.423, not reached | 0.394, not reached | 0.319, not reached | 42.000, reached | 2.960, reached | 42.000, reached | 3.480, on the edge | 42.000, reached | 2.747, reached | 41.000, reached | 3.065, reached |
| floor_0.05/both | 0.544, not reached | 0.372, not reached | 0.556, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.635, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.504, reached |
| no_trial/both | 0.617, not reached | 0.362, not reached | 0.727, not reached | 73.000, on the edge | 4.313, not reached | 36.000, reached | 0.973, reached | 73.000, on the edge | 4.927, not reached | 35.000, reached | 0.629, reached |
| glicko2_floor/general | 0.354, not reached | 0.411, not reached | 0.328, not reached | 82.000, not reached | 16.680, not reached | 20.000, reached | 5.176, not reached | 81.000, not reached | 16.147, not reached | 20.000, reached | 5.067, not reached |
| glicko2_floor/topics | 0.600, not reached | 0.332, not reached | 0.736, not reached | 324.000, not reached | — | 59.000, reached | — | 324.000, not reached | — | 51.000, reached | — |
| floor_0.05+cautious_z1/both | 0.538, not reached | 0.375, not reached | 0.568, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.633, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.673, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.605, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| constant/both | 21 of 38 | G0 r1_rms_200 0.102 [0.090, 0.115] against 0.018; G1 r1_rms_200 0.079 [0.067, 0.091] against 0.017; G4 r1_rms_200 0.091 [0.076, 0.106] against 0.021; G5 r1_rms_200 0.109 [0.096, 0.122] against 0.019; G6 r1_rms_200 0.101 [0.080, 0.120] against 0.023; G7 r1_rms_200 0.093 [0.080, 0.106] against 0.019; G8 r1_rms_200 0.109 [0.096, 0.122] against 0.018; G0-exact r1_rms_200 0.108 [0.097, 0.120] against 0.016; G0-exact r3_inside -0.035 [-0.043, -0.027] against 0.014; G0-miss0.25 r1_rms_200 0.106 [0.096, 0.117] against 0.016; G0-miss0.25 r3_inside -0.023 [-0.028, -0.017] against 0.010; G0-miss1 r1_rms_200 0.098 [0.085, 0.112] against 0.019; G0-start0.5 r1_rms_200 0.099 [0.088, 0.110] against 0.017; G0-start0.5 r3_inside -0.016 [-0.020, -0.013] against 0.010; G0-start2 r1_rms_200 0.099 [0.086, 0.112] against 0.018; G0-topics0.3 r1_rms_200 0.127 [0.116, 0.139] against 0.016; G0-topics0.3 r3_inside -0.016 [-0.020, -0.012] against 0.010 |
| constant_slow/both | 36 of 38 | G0-topics1 r1_rms_200 0.063 [0.054, 0.073] against 0.020; G0-topics1 r3_inside -0.016 [-0.019, -0.012] against 0.010 |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 29 of 38 | G1 r1_rms_200 0.171 [0.159, 0.182] against 0.017; G1 r3_inside -0.116 [-0.123, -0.107] against 0.010; G2 r1_rms_200 0.235 [0.219, 0.251] against 0.025; G3 r1_rms_200 0.117 [0.103, 0.132] against 0.022; G2-half r1_rms_200 0.098 [0.086, 0.110] against 0.020; G2-fading r1_rms_200 0.066 [0.054, 0.078] against 0.019; G3-drop r1_rms_200 0.060 [0.048, 0.073] against 0.023; G0-start2 r1_rms_200 0.090 [0.074, 0.106] against 0.018; G0-start2 r3_inside -0.042 [-0.050, -0.035] against 0.010 |
| glicko2_floor/general | 23 of 38 | G0 r1_rms_200 0.103 [0.090, 0.115] against 0.018; G1 r1_rms_200 0.109 [0.095, 0.123] against 0.017; G4 r1_rms_200 0.169 [0.154, 0.185] against 0.021; G5 r1_rms_200 0.146 [0.130, 0.161] against 0.019; G6 r1_rms_200 0.083 [0.065, 0.100] against 0.023; G7 r1_rms_200 0.115 [0.101, 0.128] against 0.019; G8 r1_rms_200 0.106 [0.094, 0.118] against 0.018; G0-exact r1_rms_200 0.126 [0.114, 0.138] against 0.016; G0-miss0.25 r1_rms_200 0.122 [0.111, 0.134] against 0.016; G0-miss1 r1_rms_200 0.100 [0.087, 0.114] against 0.019; G2-fading r1_rms_200 0.047 [0.033, 0.061] against 0.019; G0-start0.5 r1_rms_200 0.114 [0.102, 0.125] against 0.017; G0-start2 r1_rms_200 0.097 [0.084, 0.110] against 0.018; G0-topics1 r1_rms_200 0.473 [0.452, 0.497] against 0.020; G0-topics1 r3_inside -0.056 [-0.061, -0.050] against 0.010 |
| glicko2_floor/topics | 1 of 38 | G0 r1_rms_200 0.343 [0.325, 0.361] against 0.018; G0 r3_inside -0.042 [-0.048, -0.037] against 0.010; G1 r1_rms_200 0.315 [0.296, 0.332] against 0.017; G1 r3_inside -0.111 [-0.118, -0.103] against 0.010; G2 r1_rms_200 0.757 [0.734, 0.782] against 0.025; G2 r3_inside -0.031 [-0.038, -0.025] against 0.010; G3 r1_rms_200 0.585 [0.566, 0.605] against 0.022; G3 r3_inside -0.033 [-0.038, -0.026] against 0.010; G4 r1_rms_200 0.327 [0.308, 0.345] against 0.021; G4 r3_inside -0.036 [-0.041, -0.031] against 0.010; G5 r1_rms_200 0.277 [0.258, 0.298] against 0.019; G5 r3_inside -0.034 [-0.040, -0.028] against 0.010; G6 r1_rms_200 0.209 [0.189, 0.229] against 0.023; G6 r3_inside -0.056 [-0.064, -0.047] against 0.010; G7 r1_rms_200 0.315 [0.298, 0.332] against 0.019; G7 r3_inside -0.046 [-0.053, -0.040] against 0.010; G8 r1_rms_200 0.352 [0.335, 0.370] against 0.018; G8 r3_inside -0.046 [-0.052, -0.040] against 0.010; G0-exact r1_rms_200 0.345 [0.327, 0.362] against 0.016; G0-exact r3_inside -0.086 [-0.097, -0.075] against 0.014; G0-miss0.25 r1_rms_200 0.318 [0.301, 0.334] against 0.016; G0-miss0.25 r3_inside -0.063 [-0.071, -0.054] against 0.010; G0-miss1 r1_rms_200 0.359 [0.342, 0.376] against 0.019; G0-miss1 r3_inside -0.018 [-0.022, -0.015] against 0.010; G2-half r1_rms_200 0.553 [0.532, 0.574] against 0.020; G2-half r3_inside -0.040 [-0.046, -0.034] against 0.010; G2-fading r1_rms_200 0.538 [0.517, 0.558] against 0.019; G2-fading r3_inside -0.037 [-0.043, -0.030] against 0.010; G3-drop r1_rms_200 0.234 [0.218, 0.250] against 0.023; G3-drop r3_inside -0.044 [-0.049, -0.039] against 0.010; G0-start0.5 r1_rms_200 0.376 [0.360, 0.392] against 0.017; G0-start0.5 r3_inside -0.036 [-0.042, -0.031] against 0.010; G0-start2 r1_rms_200 0.386 [0.365, 0.411] against 0.018; G0-start2 r3_inside -0.077 [-0.084, -0.071] against 0.010; G0-topics0.3 r1_rms_200 0.395 [0.375, 0.413] against 0.016; G0-topics0.3 r3_inside -0.065 [-0.071, -0.058] against 0.010; G0-topics1 r1_rms_200 0.097 [0.078, 0.116] against 0.020 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.067 [-0.075, -0.058] against 0.010 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 1000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0041 | 0.018 | 100.0 % | 99.8 % |
| G0 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0040 | 0.017 | 100.0 % | 99.9 % |
| G1 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0057 | 0.025 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0052 | 0.022 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0048 | 0.021 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0044 | 0.019 | 100.0 % | 99.5 % |
| G5 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0055 | 0.023 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0045 | 0.019 | 100.0 % | 99.3 % |
| G7 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0042 | 0.018 | 100.0 % | 99.8 % |
| G8 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0038 | 0.016 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0032 | 0.014 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0038 | 0.016 | 100.0 % | 100.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0024 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0043 | 0.019 | 100.0 % | 99.6 % |
| G0-miss1 | r3_inside | 0.010 | 0.0011 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0046 | 0.020 | 100.0 % | 99.1 % |
| G2-half | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0044 | 0.019 | 100.0 % | 99.5 % |
| G2-fading | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0054 | 0.023 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0038 | 0.017 | 100.0 % | 99.9 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0041 | 0.018 | 100.0 % | 99.8 % |
| G0-start2 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0038 | 0.016 | 100.0 % | 100.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0047 | 0.020 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
