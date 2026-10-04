# The criterion, read on this run

Seed 20261001, experiment sweep: 300 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service, and no floor of the exit meets the constraints of not worse and of the screen: **the service's step stays.**

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| constant_slow/both | 0.156 [0.143, 0.168] | 45 of 49 | 4 | 0 |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 0.227 [0.211, 0.242] | 46 of 49 | 3 | 0 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.252 [0.236, 0.267] | 46 of 49 | 3 | 0 |
| uncertain_v0.1_s0.6_qt0.003_qd0.006_L0.3/both | 0.206 [0.191, 0.220] | 45 of 49 | 4 | 0 |
| uncertain_v0.1_s0.85_qt0.003_qd0.006_L0.3/both | 0.257 [0.240, 0.275] | 32 of 49 | 17 | 0 |
| uncertain_v0.1_s0.7_qt0.0015_qd0.006_L0.3/both | 0.177 [0.163, 0.191] | 46 of 49 | 3 | 0 |
| uncertain_v0.1_s0.7_qt0.006_qd0.006_L0.3/both | 0.302 [0.283, 0.320] | 44 of 49 | 5 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 0.234 [0.218, 0.250] | 46 of 49 | 3 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.015_L0.3/both | 0.267 [0.251, 0.283] | 41 of 49 | 8 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.22/both | 0.207 [0.181, 0.232] | 24 of 49 | 25 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.36/both | 0.196 [0.180, 0.211] | 45 of 49 | 4 | 0 |
| stored_s0.7_qt0.003_qd0.006_L0.3/both | 0.268 [0.248, 0.287] | 44 of 49 | 5 | 0 |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 0.235 [0.216, 0.253] | 46 of 49 | 3 | 0 |
| stored_s0.7_qt0.006_qd0.006_L0.3/both | 0.299 [0.278, 0.320] | 40 of 49 | 9 | 0 |
| stored_s0.7_qt0.003_qd0.003_L0.3/both | 0.265 [0.246, 0.285] | 45 of 49 | 4 | 0 |
| stored_s0.7_qt0.003_qd0.015_L0.3/both | 0.279 [0.259, 0.299] | 40 of 49 | 9 | 0 |
| oracle/both (the ceiling) | 1.000 [0.978, 1.023] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.268
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.427
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 74.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 3.089
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 74.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 3.089
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.067
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.067

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.597, not reached | 0.363, not reached | 0.713, not reached | 74.000, reached | 3.089, on the edge | 37.000, reached | 0.562, reached | 73.000, reached | 3.067, on the edge | 35.000, reached | 0.732, reached |
| constant_slow/both | 0.408, not reached | 0.401, not reached | 0.317, not reached | 42.000, reached | 2.489, on the edge | 42.000, reached | 2.908, on the edge | 42.000, reached | 2.578, on the edge | 42.000, reached | 2.627, on the edge |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 0.223, reached | 0.413, not reached | 0.103, reached | 52.000, reached | 1.022, reached | 52.000, reached | 3.791, not reached | 52.000, reached | 0.667, reached | 52.000, reached | 4.033, not reached |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.200, reached | 0.423, on the edge | 0.113, reached | 52.000, reached | 2.489, on the edge | 52.000, reached | 3.595, on the edge | 52.000, reached | 2.222, reached | 52.000, reached | 4.033, not reached |
| uncertain_v0.1_s0.6_qt0.003_qd0.006_L0.3/both | 0.314, not reached | 0.411, not reached | 0.230, on the edge | 52.000, reached | 2.444, on the edge | 52.000, reached | 3.235, on the edge | 52.000, reached | 2.044, reached | 52.000, reached | 4.392, not reached |
| uncertain_v0.1_s0.85_qt0.003_qd0.006_L0.3/both | 0.075, reached | 0.422, on the edge | 0.033, reached | 52.000, reached | 1.644, reached | 52.000, reached | 4.033, not reached | 52.000, reached | 1.311, reached | 52.000, reached | 3.686, not reached |
| uncertain_v0.1_s0.7_qt0.0015_qd0.006_L0.3/both | 0.330, not reached | 0.405, not reached | 0.253, on the edge | 52.000, reached | 1.689, reached | 52.000, reached | 2.647, on the edge | 52.000, reached | 1.444, reached | 52.000, reached | 2.294, reached |
| uncertain_v0.1_s0.7_qt0.006_qd0.006_L0.3/both | 0.033, reached | 0.426, on the edge | 0.023, reached | 52.000, reached | 2.178, reached | 52.000, reached | 6.176, not reached | 52.000, reached | 2.044, reached | 52.000, reached | 6.418, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 0.228, reached | 0.416, not reached | 0.140, reached | 52.000, reached | 1.867, reached | 52.000, reached | 3.758, not reached | 52.000, reached | 1.667, reached | 52.000, reached | 3.882, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.015_L0.3/both | 0.105, reached | 0.427, on the edge | 0.050, reached | 52.000, reached | 2.022, reached | 52.000, reached | 4.092, not reached | 52.000, reached | 1.622, reached | 52.000, reached | 4.085, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.22/both | 0.091, reached | 0.425, on the edge | 0.013, reached | 39.000, reached | 1.778, reached | 39.000, reached | 4.621, not reached | 39.000, reached | 1.378, reached | 39.000, reached | 4.307, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.36/both | 0.319, not reached | 0.409, not reached | 0.203, reached | 63.000, reached | 2.267, reached | 57.000, reached | 3.850, not reached | 63.000, reached | 1.867, reached | 57.000, reached | 3.686, not reached |
| stored_s0.7_qt0.003_qd0.006_L0.3/both | 0.172, reached | 0.425, on the edge | 0.123, reached | 52.000, reached | 13.133, not reached | 52.000, reached | 3.948, not reached | 52.000, reached | 12.044, not reached | 52.000, reached | 4.131, not reached |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 0.234, reached | 0.427, on the edge | 0.200, reached | 52.000, reached | 13.000, not reached | 51.000, reached | 2.810, on the edge | 52.000, reached | 12.156, not reached | 51.000, reached | 2.745, on the edge |
| stored_s0.7_qt0.006_qd0.006_L0.3/both | 0.046, reached | 0.425, on the edge | 0.057, reached | 52.000, reached | 13.667, not reached | 52.000, reached | 5.529, not reached | 52.000, reached | 12.222, not reached | 52.000, reached | 6.359, not reached |
| stored_s0.7_qt0.003_qd0.003_L0.3/both | 0.184, reached | 0.429, on the edge | 0.133, reached | 52.000, reached | 13.156, not reached | 52.000, reached | 3.863, not reached | 52.000, reached | 12.089, not reached | 52.000, reached | 4.020, not reached |
| stored_s0.7_qt0.003_qd0.015_L0.3/both | 0.089, reached | 0.431, on the edge | 0.070, reached | 52.000, reached | 13.267, not reached | 52.000, reached | 4.549, not reached | 52.000, reached | 12.022, not reached | 52.000, reached | 3.582, on the edge |
| oracle/both (the ceiling) | 0.000, reached | 0.602, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| constant_slow/both | 37 of 38 | G0-topics1 r1_rms_200 0.062 [0.045, 0.079] against 0.037 |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 38 of 38 |  |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 38 of 38 |  |
| uncertain_v0.1_s0.6_qt0.003_qd0.006_L0.3/both | 38 of 38 |  |
| uncertain_v0.1_s0.85_qt0.003_qd0.006_L0.3/both | 24 of 38 | G0 r1_rms_200 0.076 [0.053, 0.098] against 0.040; G0 r3_inside -0.028 [-0.038, -0.018] against 0.015; G5 r3_inside -0.026 [-0.036, -0.017] against 0.015; G6 r3_inside -0.039 [-0.050, -0.026] against 0.016; G8 r1_rms_200 0.067 [0.045, 0.088] against 0.033; G8 r3_inside -0.030 [-0.039, -0.020] against 0.015; G0-exact r3_inside -0.058 [-0.077, -0.041] against 0.027; G0-miss0.25 r1_rms_200 0.077 [0.055, 0.099] against 0.037; G0-miss0.25 r3_inside -0.039 [-0.053, -0.024] against 0.021; G3-drop r3_inside -0.033 [-0.040, -0.026] against 0.012; G0-start0.5 r1_rms_200 0.068 [0.048, 0.089] against 0.037; G0-start0.5 r3_inside -0.027 [-0.036, -0.017] against 0.015; G0-topics0.3 r1_rms_200 0.097 [0.074, 0.118] against 0.037; G0-topics0.3 r3_inside -0.025 [-0.034, -0.015] against 0.013 |
| uncertain_v0.1_s0.7_qt0.0015_qd0.006_L0.3/both | 38 of 38 |  |
| uncertain_v0.1_s0.7_qt0.006_qd0.006_L0.3/both | 36 of 38 | G8 r1_rms_200 0.070 [0.049, 0.092] against 0.033; G0-miss0.25 r1_rms_200 0.072 [0.049, 0.096] against 0.037 |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 38 of 38 |  |
| uncertain_v0.1_s0.7_qt0.003_qd0.015_L0.3/both | 33 of 38 | G8 r1_rms_200 0.061 [0.041, 0.083] against 0.033; G0-miss0.25 r1_rms_200 0.076 [0.053, 0.100] against 0.037; G3-drop r3_inside -0.019 [-0.025, -0.012] against 0.012; G0-start0.5 r1_rms_200 0.060 [0.040, 0.080] against 0.037; G0-topics0.3 r1_rms_200 0.069 [0.048, 0.091] against 0.037 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.22/both | 16 of 38 | G0 r1_rms_200 0.111 [0.083, 0.139] against 0.040; G0 r3_inside -0.038 [-0.049, -0.028] against 0.015; G1 r1_rms_200 0.090 [0.062, 0.119] against 0.037; G4 r3_inside -0.024 [-0.032, -0.016] against 0.013; G5 r1_rms_200 0.101 [0.074, 0.129] against 0.039; G5 r3_inside -0.044 [-0.054, -0.034] against 0.015; G6 r3_inside -0.059 [-0.074, -0.044] against 0.016; G7 r3_inside -0.035 [-0.046, -0.024] against 0.013; G8 r1_rms_200 0.110 [0.085, 0.134] against 0.033; G8 r3_inside -0.043 [-0.053, -0.032] against 0.015; G0-exact r1_rms_200 0.096 [0.069, 0.121] against 0.039; G0-exact r3_inside -0.090 [-0.109, -0.070] against 0.027; G0-miss0.25 r1_rms_200 0.131 [0.105, 0.158] against 0.037; G0-miss0.25 r3_inside -0.065 [-0.084, -0.048] against 0.021; G3-drop r3_inside -0.059 [-0.067, -0.052] against 0.012; G0-start0.5 r1_rms_200 0.119 [0.093, 0.145] against 0.037; G0-start0.5 r3_inside -0.039 [-0.049, -0.028] against 0.015; G0-start2 r1_rms_200 0.087 [0.060, 0.114] against 0.039; G0-start2 r3_inside -0.028 [-0.039, -0.016] against 0.015; G0-topics0.3 r1_rms_200 0.123 [0.097, 0.151] against 0.037; G0-topics0.3 r3_inside -0.031 [-0.042, -0.019] against 0.013; G0-topics1 r3_inside -0.032 [-0.040, -0.023] against 0.016 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.36/both | 38 of 38 |  |
| stored_s0.7_qt0.003_qd0.006_L0.3/both | 38 of 38 |  |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 38 of 38 |  |
| stored_s0.7_qt0.006_qd0.006_L0.3/both | 34 of 38 | G5 r1_rms_200 0.075 [0.049, 0.100] against 0.039; G6 r3_inside -0.037 [-0.053, -0.021] against 0.016; G0-exact r3_inside -0.049 [-0.068, -0.029] against 0.027; G0-miss0.25 r1_rms_200 0.063 [0.040, 0.086] against 0.037 |
| stored_s0.7_qt0.003_qd0.003_L0.3/both | 38 of 38 |  |
| stored_s0.7_qt0.003_qd0.015_L0.3/both | 33 of 38 | G6 r3_inside -0.035 [-0.052, -0.019] against 0.016; G0-exact r3_inside -0.049 [-0.069, -0.030] against 0.027; G0-miss0.25 r1_rms_200 0.070 [0.048, 0.092] against 0.037; G3-drop r3_inside -0.022 [-0.030, -0.015] against 0.012; G0-topics0.3 r1_rms_200 0.068 [0.049, 0.090] against 0.037 |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.066 [-0.084, -0.049] against 0.016 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 300 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0092 | 0.040 | 100.0 % | 99.0 % |
| G0 | r3_inside | 0.010 | 0.0035 | 0.015 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0087 | 0.037 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0032 | 0.014 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0140 | 0.060 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0031 | 0.013 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0124 | 0.053 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0035 | 0.015 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0102 | 0.044 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0091 | 0.039 | 100.0 % | 99.0 % |
| G5 | r3_inside | 0.010 | 0.0034 | 0.015 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0137 | 0.059 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0038 | 0.016 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0089 | 0.038 | 100.0 % | 99.0 % |
| G7 | r3_inside | 0.010 | 0.0030 | 0.013 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0077 | 0.033 | 100.0 % | 99.7 % |
| G8 | r3_inside | 0.010 | 0.0034 | 0.015 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0091 | 0.039 | 100.0 % | 99.0 % |
| G0-exact | r3_inside | 0.010 | 0.0062 | 0.027 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0085 | 0.037 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0050 | 0.021 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0090 | 0.039 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0020 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0114 | 0.049 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0034 | 0.014 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0109 | 0.047 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0034 | 0.014 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0116 | 0.050 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0027 | 0.012 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0087 | 0.037 | 100.0 % | 99.0 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0034 | 0.015 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0091 | 0.039 | 100.0 % | 99.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0034 | 0.015 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0087 | 0.037 | 100.0 % | 99.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0031 | 0.013 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0085 | 0.037 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0037 | 0.016 | 100.0 % | 100.0 % |
