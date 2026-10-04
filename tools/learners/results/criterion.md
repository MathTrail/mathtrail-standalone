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
| earlier/both (the service's earlier rule) | -0.069 [-0.074, -0.063] | 37 of 49 | 12 | 0 |
| constant/both | 0.172 [0.163, 0.181] | 22 of 49 | 27 | 0 |
| constant_slow/both | 0.100 [0.094, 0.106] | 43 of 49 | 6 | 0 |
| floor_0.05/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| no_trial/both | -0.177 [-0.189, -0.165] | 35 of 49 | 14 | 0 |
| glicko2_floor/general | 0.128 [0.118, 0.138] | 25 of 49 | 24 | 0 |
| glicko2_floor/topics | — | 3 of 49 | 42 | 4 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| oracle/both (the ceiling) | 1.000 [0.987, 1.012] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.242
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.439
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
| shrinking/both (the service) | 0.538, not reached | 0.375, not reached | 0.568, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.633, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.673, reached |
| earlier/both (the service's earlier rule) | 0.609, not reached | 0.357, not reached | 0.711, not reached | 73.000, on the edge | 3.813, on the edge | 37.000, reached | 0.629, reached | 73.000, reached | 3.600, on the edge | 36.000, reached | 0.502, reached |
| constant/both | 0.279, not reached | 0.407, not reached | 0.108, reached | 84.000, not reached | 6.940, not reached | 84.000, not reached | 7.059, not reached | 84.000, not reached | 6.267, not reached | 84.000, not reached | 6.947, not reached |
| constant_slow/both | 0.417, not reached | 0.395, not reached | 0.327, not reached | 42.000, reached | 2.960, reached | 42.000, reached | 3.631, on the edge | 42.000, reached | 2.747, reached | 42.000, reached | 3.316, on the edge |
| floor_0.05/both | 0.538, not reached | 0.375, not reached | 0.568, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.633, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.673, reached |
| no_trial/both | 0.602, not reached | 0.368, not reached | 0.730, not reached | 73.000, on the edge | 4.313, not reached | 36.000, reached | 0.810, reached | 73.000, on the edge | 4.927, not reached | 34.000, reached | 0.690, reached |
| glicko2_floor/general | 0.374, not reached | 0.399, not reached | 0.301, not reached | 82.000, not reached | 16.680, not reached | 21.000, reached | 4.884, not reached | 81.000, not reached | 16.147, not reached | 20.000, reached | 4.431, not reached |
| glicko2_floor/topics | 0.684, not reached | 0.325, not reached | 0.665, not reached | 324.000, not reached | — | 41.000, reached | — | 324.000, not reached | — | 37.000, reached | — |
| floor_0.05+cautious_z1/both | 0.538, not reached | 0.375, not reached | 0.568, not reached | 73.000, on the edge | 3.813, on the edge | 41.000, reached | 1.633, reached | 73.000, reached | 3.600, on the edge | 40.000, reached | 1.673, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.611, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| earlier/both (the service's earlier rule) | 29 of 38 | G2 r1_rms_200 0.232 [0.223, 0.241] against 0.024; G2 r3_inside -0.018 [-0.021, -0.015] against 0.010; G3 r1_rms_200 0.109 [0.100, 0.118] against 0.023; G3 r3_inside -0.015 [-0.018, -0.012] against 0.010; G2-half r1_rms_200 0.097 [0.088, 0.105] against 0.021; G2-half r3_inside -0.017 [-0.020, -0.014] against 0.010; G2-fading r1_rms_200 0.049 [0.041, 0.058] against 0.019; G2-fading r3_inside -0.015 [-0.018, -0.012] against 0.010; G3-drop r1_rms_200 0.132 [0.121, 0.143] against 0.021 |
| constant/both | 21 of 38 | G0 r1_rms_200 0.117 [0.105, 0.129] against 0.017; G1 r1_rms_200 0.080 [0.068, 0.092] against 0.017; G4 r1_rms_200 0.109 [0.095, 0.125] against 0.021; G5 r1_rms_200 0.113 [0.099, 0.126] against 0.018; G6 r1_rms_200 0.130 [0.110, 0.150] against 0.023; G7 r1_rms_200 0.110 [0.097, 0.124] against 0.018; G8 r1_rms_200 0.112 [0.100, 0.125] against 0.015; G0-exact r1_rms_200 0.115 [0.104, 0.126] against 0.015; G0-exact r3_inside -0.030 [-0.037, -0.022] against 0.013; G0-miss0.25 r1_rms_200 0.127 [0.116, 0.139] against 0.016; G0-miss0.25 r3_inside -0.028 [-0.033, -0.023] against 0.010; G0-miss1 r1_rms_200 0.103 [0.088, 0.117] against 0.019; G3-drop r1_rms_200 0.049 [0.036, 0.063] against 0.021; G0-start0.5 r1_rms_200 0.108 [0.097, 0.120] against 0.015; G0-start2 r1_rms_200 0.095 [0.082, 0.108] against 0.017; G0-topics0.3 r1_rms_200 0.137 [0.125, 0.149] against 0.017; G0-topics0.3 r3_inside -0.016 [-0.020, -0.012] against 0.010 |
| constant_slow/both | 35 of 38 | G5 r1_rms_200 0.029 [0.021, 0.037] against 0.018; G0-topics1 r1_rms_200 0.044 [0.036, 0.052] against 0.017; G0-topics1 r3_inside -0.020 [-0.023, -0.016] against 0.010 |
| floor_0.05/both | 38 of 38 |  |
| no_trial/both | 29 of 38 | G1 r1_rms_200 0.162 [0.150, 0.174] against 0.017; G1 r3_inside -0.121 [-0.129, -0.113] against 0.010; G2 r1_rms_200 0.243 [0.227, 0.259] against 0.024; G3 r1_rms_200 0.128 [0.114, 0.142] against 0.023; G2-half r1_rms_200 0.103 [0.091, 0.116] against 0.021; G2-fading r1_rms_200 0.061 [0.048, 0.073] against 0.019; G3-drop r1_rms_200 0.065 [0.054, 0.076] against 0.021; G0-start2 r1_rms_200 0.095 [0.078, 0.111] against 0.017; G0-start2 r3_inside -0.045 [-0.052, -0.037] against 0.010 |
| glicko2_floor/general | 23 of 38 | G0 r1_rms_200 0.085 [0.074, 0.095] against 0.017; G1 r1_rms_200 0.073 [0.062, 0.084] against 0.017; G4 r1_rms_200 0.090 [0.077, 0.103] against 0.021; G5 r1_rms_200 0.124 [0.110, 0.138] against 0.018; G6 r1_rms_200 0.087 [0.070, 0.105] against 0.023; G7 r1_rms_200 0.081 [0.070, 0.092] against 0.018; G8 r1_rms_200 0.085 [0.075, 0.096] against 0.015; G0-exact r1_rms_200 0.091 [0.081, 0.100] against 0.015; G0-miss0.25 r1_rms_200 0.092 [0.082, 0.102] against 0.016; G0-miss0.25 r3_inside -0.017 [-0.024, -0.010] against 0.010; G0-miss1 r1_rms_200 0.060 [0.049, 0.072] against 0.019; G0-start0.5 r1_rms_200 0.081 [0.072, 0.091] against 0.015; G0-start2 r1_rms_200 0.071 [0.060, 0.083] against 0.017; G0-topics1 r1_rms_200 0.355 [0.339, 0.370] against 0.017; G0-topics1 r3_inside -0.059 [-0.064, -0.054] against 0.010 |
| glicko2_floor/topics | 1 of 38 | G0 r1_rms_200 0.411 [0.393, 0.429] against 0.017; G0 r3_inside -0.028 [-0.034, -0.021] against 0.010; G1 r1_rms_200 0.351 [0.332, 0.370] against 0.017; G1 r3_inside -0.105 [-0.113, -0.097] against 0.010; G2 r1_rms_200 0.785 [0.763, 0.808] against 0.024; G2 r3_inside -0.051 [-0.058, -0.044] against 0.010; G3 r1_rms_200 0.639 [0.619, 0.662] against 0.023; G3 r3_inside -0.025 [-0.032, -0.019] against 0.010; G4 r1_rms_200 0.337 [0.317, 0.355] against 0.021; G4 r3_inside -0.029 [-0.033, -0.024] against 0.010; G5 r1_rms_200 0.336 [0.317, 0.357] against 0.018; G5 r3_inside -0.017 [-0.023, -0.011] against 0.010; G6 r1_rms_200 0.248 [0.227, 0.267] against 0.023; G6 r3_inside -0.035 [-0.043, -0.026] against 0.010; G7 r1_rms_200 0.365 [0.347, 0.383] against 0.018; G7 r3_inside -0.030 [-0.036, -0.024] against 0.010; G8 r1_rms_200 0.400 [0.382, 0.418] against 0.015; G8 r3_inside -0.034 [-0.039, -0.027] against 0.010; G0-exact r1_rms_200 0.404 [0.386, 0.422] against 0.015; G0-exact r3_inside -0.045 [-0.057, -0.033] against 0.013; G0-miss0.25 r1_rms_200 0.392 [0.375, 0.410] against 0.016; G0-miss0.25 r3_inside -0.039 [-0.048, -0.030] against 0.010; G0-miss1 r1_rms_200 0.407 [0.389, 0.425] against 0.019; G0-miss1 r3_inside -0.014 [-0.018, -0.011] against 0.010; G2-half r1_rms_200 0.585 [0.564, 0.606] against 0.021; G2-half r3_inside -0.050 [-0.057, -0.043] against 0.010; G2-fading r1_rms_200 0.559 [0.538, 0.581] against 0.019; G2-fading r3_inside -0.036 [-0.044, -0.029] against 0.010; G3-drop r1_rms_200 0.423 [0.404, 0.442] against 0.021; G3-drop r3_inside -0.018 [-0.024, -0.013] against 0.010; G0-start0.5 r1_rms_200 0.431 [0.415, 0.447] against 0.015; G0-start0.5 r3_inside -0.017 [-0.023, -0.011] against 0.010; G0-start2 r1_rms_200 0.438 [0.416, 0.463] against 0.017; G0-start2 r3_inside -0.063 [-0.070, -0.055] against 0.010; G0-topics0.3 r1_rms_200 0.453 [0.435, 0.470] against 0.017; G0-topics0.3 r3_inside -0.046 [-0.053, -0.040] against 0.010; G0-topics1 r1_rms_200 0.189 [0.169, 0.208] against 0.017 |
| floor_0.05+cautious_z1/both | 38 of 38 |  |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.079 [-0.088, -0.069] against 0.010 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 1000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0040 | 0.017 | 100.0 % | 99.9 % |
| G0 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0038 | 0.017 | 100.0 % | 99.9 % |
| G1 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0057 | 0.024 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0053 | 0.023 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0048 | 0.021 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0014 | 0.010 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0042 | 0.018 | 100.0 % | 99.7 % |
| G5 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0054 | 0.023 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0041 | 0.018 | 100.0 % | 99.8 % |
| G7 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0036 | 0.015 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0036 | 0.015 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0029 | 0.013 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0037 | 0.016 | 100.0 % | 100.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0022 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0045 | 0.019 | 100.0 % | 99.4 % |
| G0-miss1 | r3_inside | 0.010 | 0.0011 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0049 | 0.021 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0044 | 0.019 | 100.0 % | 99.5 % |
| G2-fading | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0048 | 0.021 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0034 | 0.015 | 100.0 % | 100.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0014 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0040 | 0.017 | 100.0 % | 99.9 % |
| G0-start2 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0040 | 0.017 | 100.0 % | 99.9 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0015 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0040 | 0.017 | 100.0 % | 99.9 % |
| G0-topics1 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
