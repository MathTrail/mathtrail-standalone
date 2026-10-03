# The criterion, read on this run

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 65 of 68 | 3 | 0 |
| constant/both | 0.210 [0.200, 0.219] | 29 of 68 | 39 | 0 |
| constant_slow/both | 0.153 [0.146, 0.159] | 61 of 68 | 7 | 0 |
| floor_0.05/both | 0.055 [0.052, 0.059] | 65 of 68 | 3 | 0 |
| no_trial/both | -0.121 [-0.133, -0.109] | 59 of 68 | 9 | 0 |
| glicko2_floor/general | 0.194 [0.184, 0.204] | 44 of 68 | 24 | 0 |
| glicko2_floor/topics | — | 23 of 68 | 41 | 4 |
| oracle/both (the ceiling) | 1.000 [0.987, 1.012] | 41 of 68 | 19 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it.

| Goal | Generator | Bound | shrinking/both (the service) | constant/both | constant_slow/both | floor_0.05/both | no_trial/both | glicko2_floor/general | glicko2_floor/topics | oracle/both (the ceiling) |
|---|---|---:|---|---|---|---|---|---|---|---|
| the lag at most 45 % of the service's | G2-half | 0.274 | 0.609, not reached | 0.286, on the edge | 0.423, not reached | 0.544, not reached | 0.617, not reached | 0.354, not reached | 0.600, not reached | 0.000, reached |
| the corridor a share of the way to the ceiling | G2-half | 0.424 | 0.357, not reached | 0.404, not reached | 0.394, not reached | 0.372, not reached | 0.362, not reached | 0.411, not reached | 0.332, not reached | 0.605, reached |
| children not caught up after the jump at most 25 % | G3 | 0.250 | 0.711, not reached | 0.091, reached | 0.319, not reached | 0.556, not reached | 0.727, not reached | 0.328, not reached | 0.736, not reached | 0.000, reached |
| the card's move in answers 6–20 no more than the service's in answers 6–20 | G0 | 73.000 | 73.000, on the edge | 84.000, not reached | 42.000, reached | 73.000, on the edge | 73.000, on the edge | 82.000, not reached | 324.000, not reached | — |
| the rank's changes in answers 6–20 no more than the service's in answers 6–20 | G0 | 3.813 | 3.813, on the edge | 6.940, not reached | 2.960, reached | 3.813, on the edge | 4.313, not reached | 16.680, not reached | — | — |
| the card's move in answers 150–200 no more than the service's in answers 6–20 | G0 | 73.000 | 37.000, reached | 84.000, not reached | 42.000, reached | 41.000, reached | 36.000, reached | 20.000, reached | 59.000, reached | — |
| the rank's changes in answers 150–200 no more than the service's in answers 6–20 | G0 | 3.813 | 0.629, reached | 6.818, not reached | 3.480, on the edge | 1.635, reached | 0.973, reached | 5.176, not reached | — | — |
| the card's move in answers 6–20 no more than the service's in answers 6–20 | G2-half | 73.000 | 73.000, reached | 84.000, not reached | 42.000, reached | 73.000, reached | 73.000, on the edge | 81.000, not reached | 324.000, not reached | — |
| the rank's changes in answers 6–20 no more than the service's in answers 6–20 | G2-half | 3.600 | 3.600, on the edge | 6.267, not reached | 2.747, reached | 3.600, on the edge | 4.927, not reached | 16.147, not reached | — | — |
| the card's move in answers 150–200 no more than the service's in answers 6–20 | G2-half | 73.000 | 36.000, reached | 83.000, not reached | 41.000, reached | 40.000, reached | 35.000, reached | 20.000, reached | 51.000, reached | — |
| the rank's changes in answers 150–200 no more than the service's in answers 6–20 | G2-half | 3.600 | 0.502, reached | 6.496, not reached | 3.065, reached | 1.504, reached | 0.629, reached | 5.067, not reached | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 57 of 57 |  |
| constant/both | 28 of 57 | G0 r1_rms_200 0.098 [0.085, 0.111] against 0.020; G0 r4_false 0.048 [0.034, 0.060] against 0.026; G1 r1_rms_200 0.062 [0.048, 0.076] against 0.021; G2 r4_false 0.106 [0.093, 0.120] against 0.026; G3 r4_false 0.072 [0.060, 0.085] against 0.026; G4 r1_rms_200 0.091 [0.074, 0.109] against 0.026; G4 r4_false 0.039 [0.027, 0.052] against 0.026; G5 r1_rms_200 0.100 [0.086, 0.113] against 0.021; G5 r4_false 0.040 [0.028, 0.054] against 0.026; G6 r1_rms_200 0.111 [0.088, 0.135] against 0.029; G7 r1_rms_200 0.084 [0.069, 0.099] against 0.023; G8 r1_rms_200 0.099 [0.086, 0.114] against 0.020; G8 r4_false 0.048 [0.034, 0.062] against 0.030; G0-exact r1_rms_200 0.103 [0.091, 0.115] against 0.018; G0-exact r3_inside -0.034 [-0.042, -0.026] against 0.015; G0-exact r4_false 0.040 [0.027, 0.052] against 0.026; G0-miss0.25 r1_rms_200 0.104 [0.092, 0.116] against 0.018; G0-miss0.25 r3_inside -0.023 [-0.029, -0.017] against 0.012; G0-miss0.25 r4_false 0.048 [0.036, 0.060] against 0.025; G0-miss1 r1_rms_200 0.093 [0.078, 0.107] against 0.022; G0-miss1 r4_false 0.041 [0.028, 0.053] against 0.027; G2-half r4_false 0.076 [0.062, 0.089] against 0.026; G2-fading r4_false 0.074 [0.061, 0.087] against 0.025; G0-start0.5 r1_rms_200 0.094 [0.083, 0.107] against 0.018; G0-start0.5 r3_inside -0.015 [-0.019, -0.011] against 0.010; G0-start2 r1_rms_200 0.084 [0.069, 0.099] against 0.022; G0-start2 r4_false 0.061 [0.049, 0.075] against 0.027; G0-topics0.3 r1_rms_200 0.122 [0.110, 0.135] against 0.019; G0-topics0.3 r4_false 0.041 [0.029, 0.054] against 0.026 |
| constant_slow/both | 53 of 57 | G2 r4_false 0.051 [0.039, 0.063] against 0.026; G2-half r4_false 0.041 [0.030, 0.053] against 0.026; G0-topics1 r1_rms_200 0.055 [0.045, 0.064] against 0.022; G0-topics1 r3_inside -0.019 [-0.023, -0.015] against 0.010 |
| floor_0.05/both | 57 of 57 |  |
| no_trial/both | 53 of 57 | G1 r1_rms_200 0.153 [0.141, 0.166] against 0.021; G1 r3_inside -0.113 [-0.121, -0.105] against 0.010; G0-start2 r1_rms_200 0.075 [0.060, 0.091] against 0.022; G0-start2 r3_inside -0.040 [-0.048, -0.032] against 0.010 |
| glicko2_floor/general | 42 of 57 | G0 r1_rms_200 0.099 [0.085, 0.112] against 0.020; G1 r1_rms_200 0.091 [0.076, 0.107] against 0.021; G1 r4_false 0.055 [0.036, 0.073] against 0.026; G4 r1_rms_200 0.169 [0.152, 0.187] against 0.026; G5 r1_rms_200 0.137 [0.120, 0.152] against 0.021; G6 r1_rms_200 0.093 [0.073, 0.112] against 0.029; G7 r1_rms_200 0.105 [0.090, 0.121] against 0.023; G8 r1_rms_200 0.097 [0.084, 0.110] against 0.020; G0-exact r1_rms_200 0.121 [0.108, 0.133] against 0.018; G0-miss0.25 r1_rms_200 0.120 [0.108, 0.132] against 0.018; G0-miss1 r1_rms_200 0.095 [0.080, 0.110] against 0.022; G0-start0.5 r1_rms_200 0.109 [0.097, 0.122] against 0.018; G0-start2 r1_rms_200 0.082 [0.068, 0.097] against 0.022; G0-topics1 r1_rms_200 0.465 [0.443, 0.488] against 0.022; G0-topics1 r3_inside -0.059 [-0.064, -0.053] against 0.010 |
| glicko2_floor/topics | 21 of 57 | G0 r1_rms_200 0.339 [0.320, 0.357] against 0.020; G0 r3_inside -0.043 [-0.049, -0.037] against 0.010; G1 r1_rms_200 0.297 [0.279, 0.315] against 0.021; G1 r3_inside -0.108 [-0.116, -0.100] against 0.010; G2 r1_rms_200 0.510 [0.486, 0.535] against 0.032; G3 r1_rms_200 0.472 [0.452, 0.492] against 0.030; G3 r3_inside -0.020 [-0.026, -0.014] against 0.010; G4 r1_rms_200 0.327 [0.309, 0.345] against 0.026; G4 r3_inside -0.035 [-0.040, -0.030] against 0.010; G5 r1_rms_200 0.268 [0.249, 0.289] against 0.021; G5 r3_inside -0.034 [-0.040, -0.028] against 0.010; G6 r1_rms_200 0.220 [0.199, 0.239] against 0.029; G6 r3_inside -0.055 [-0.064, -0.046] against 0.010; G7 r1_rms_200 0.306 [0.288, 0.324] against 0.023; G7 r3_inside -0.045 [-0.052, -0.039] against 0.010; G8 r1_rms_200 0.343 [0.326, 0.360] against 0.020; G8 r3_inside -0.044 [-0.050, -0.037] against 0.010; G0-exact r1_rms_200 0.340 [0.322, 0.357] against 0.018; G0-exact r3_inside -0.085 [-0.096, -0.073] against 0.015; G0-miss0.25 r1_rms_200 0.315 [0.299, 0.332] against 0.018; G0-miss0.25 r3_inside -0.063 [-0.071, -0.054] against 0.012; G0-miss1 r1_rms_200 0.353 [0.337, 0.370] against 0.022; G0-miss1 r3_inside -0.018 [-0.021, -0.014] against 0.010; G2-half r1_rms_200 0.443 [0.422, 0.464] against 0.027; G2-half r3_inside -0.025 [-0.032, -0.019] against 0.010; G2-fading r1_rms_200 0.469 [0.448, 0.490] against 0.025; G2-fading r3_inside -0.028 [-0.034, -0.021] against 0.010; G3-drop r1_rms_200 0.149 [0.133, 0.166] against 0.029; G3-drop r3_inside -0.038 [-0.044, -0.033] against 0.010; G0-start0.5 r1_rms_200 0.372 [0.356, 0.389] against 0.018; G0-start0.5 r3_inside -0.036 [-0.041, -0.030] against 0.010; G0-start2 r1_rms_200 0.371 [0.351, 0.394] against 0.022; G0-start2 r3_inside -0.075 [-0.082, -0.068] against 0.010; G0-topics0.3 r1_rms_200 0.389 [0.371, 0.407] against 0.019; G0-topics0.3 r3_inside -0.062 [-0.069, -0.055] against 0.010; G0-topics1 r1_rms_200 0.088 [0.069, 0.107] against 0.022 |
| oracle/both (the ceiling) | 38 of 57 | G0 r4_false 0.118 [0.102, 0.136] against 0.026; G1 r4_false 0.050 [0.035, 0.066] against 0.026; G2 r4_false 0.257 [0.243, 0.273] against 0.026; G3 r4_false 0.198 [0.182, 0.214] against 0.026; G4 r4_false 0.161 [0.142, 0.181] against 0.026; G5 r4_false 0.131 [0.113, 0.151] against 0.026; G6 r3_inside -0.066 [-0.075, -0.057] against 0.010; G7 r4_false 0.047 [0.029, 0.064] against 0.025; G8 r4_false 0.109 [0.091, 0.127] against 0.030; G0-exact r4_false 0.112 [0.095, 0.128] against 0.026; G0-miss0.25 r4_false 0.111 [0.095, 0.127] against 0.025; G0-miss1 r4_false 0.114 [0.097, 0.132] against 0.027; G2-half r4_false 0.214 [0.198, 0.230] against 0.026; G2-fading r4_false 0.182 [0.166, 0.196] against 0.025; G3-drop r4_false 0.057 [0.042, 0.073] against 0.025; G0-start0.5 r4_false 0.075 [0.059, 0.091] against 0.025; G0-start2 r4_false 0.118 [0.101, 0.134] against 0.027; G0-topics0.3 r4_false 0.105 [0.086, 0.125] against 0.026; G0-topics1 r4_false 0.147 [0.130, 0.163] against 0.026 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 1000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0046 | 0.020 | 100.0 % | 99.1 % |
| G0 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0061 | 0.026 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0050 | 0.021 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0075 | 0.032 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0061 | 0.026 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0069 | 0.030 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0061 | 0.026 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0060 | 0.026 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0016 | 0.010 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0049 | 0.021 | 100.0 % | 99.0 % |
| G5 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0059 | 0.026 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0068 | 0.029 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0055 | 0.023 | 100.0 % | 99.0 % |
| G7 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0046 | 0.020 | 100.0 % | 99.2 % |
| G8 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0069 | 0.030 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0041 | 0.018 | 100.0 % | 99.8 % |
| G0-exact | r3_inside | 0.010 | 0.0035 | 0.015 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0042 | 0.018 | 100.0 % | 99.8 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0027 | 0.012 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0051 | 0.022 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0011 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0062 | 0.027 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0062 | 0.027 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0059 | 0.026 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0058 | 0.025 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0058 | 0.025 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0068 | 0.029 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0043 | 0.018 | 100.0 % | 99.7 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0018 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0059 | 0.025 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0050 | 0.022 | 100.0 % | 99.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0063 | 0.027 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0045 | 0.019 | 100.0 % | 99.4 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0051 | 0.022 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0060 | 0.026 | 100.0 % | 100.0 % |
