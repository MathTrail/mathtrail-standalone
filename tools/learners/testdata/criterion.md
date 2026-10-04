# The criterion, read on this run

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 65 of 68 | 3 | 0 |
| constant/both | 0.215 [0.129, 0.296] | 60 of 68 | 8 | 0 |
| constant_slow/both | 0.175 [0.125, 0.228] | 65 of 68 | 3 | 0 |
| floor_0.05/both | 0.052 [0.020, 0.088] | 65 of 68 | 3 | 0 |
| no_trial/both | -0.258 [-0.416, -0.098] | 63 of 68 | 5 | 0 |
| glicko2_floor/general | 0.226 [0.144, 0.306] | 61 of 68 | 7 | 0 |
| glicko2_floor/topics | — | 48 of 68 | 16 | 4 |
| oracle/both (the ceiling) | 1.000 [0.893, 1.113] | 59 of 68 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it.

| Goal | Generator | Bound | shrinking/both (the service) | constant/both | constant_slow/both | floor_0.05/both | no_trial/both | glicko2_floor/general | glicko2_floor/topics | oracle/both (the ceiling) |
|---|---|---:|---|---|---|---|---|---|---|---|
| the lag at most 45 % of the service's | G2-half | 0.311 | 0.691, not reached | 0.418, on the edge | 0.505, not reached | 0.633, not reached | 0.655, not reached | 0.366, on the edge | 0.514, not reached | 0.000, reached |
| the corridor a share of the way to the ceiling | G2-half | 0.405 | 0.332, not reached | 0.409, on the edge | 0.382, on the edge | 0.344, on the edge | 0.366, not reached | 0.408, on the edge | 0.378, on the edge | 0.602, reached |
| children not caught up after the jump at most 25 % | G3 | 0.250 | 0.700, not reached | 0.100, on the edge | 0.100, on the edge | 0.400, on the edge | 0.800, not reached | 0.200, on the edge | 0.600, not reached | 0.000, reached |
| the card's move in answers 6–20 no more than the service's in answers 6–20 | G0 | 74.000 | 74.000, on the edge | 83.000, not reached | 42.000, reached | 74.000, on the edge | 73.000, on the edge | 86.000, on the edge | 324.000, not reached | — |
| the rank's changes in answers 6–20 no more than the service's in answers 6–20 | G0 | 6.667 | 6.667, on the edge | 7.333, on the edge | 4.667, on the edge | 6.667, on the edge | 4.000, on the edge | 18.000, not reached | — | — |
| the card's move in answers 150–200 no more than the service's in answers 6–20 | G0 | 74.000 | 36.000, reached | 84.000, not reached | 42.000, reached | 41.000, reached | 36.000, reached | 20.000, reached | 67.000, on the edge | — |
| the rank's changes in answers 150–200 no more than the service's in answers 6–20 | G0 | 6.667 | 0.392, reached | 5.490, on the edge | 2.157, reached | 5.098, on the edge | 0.392, reached | 3.922, on the edge | — | — |
| the card's move in answers 6–20 no more than the service's in answers 6–20 | G2-half | 72.000 | 72.000, on the edge | 84.000, not reached | 42.000, reached | 72.000, on the edge | 74.000, on the edge | 82.000, on the edge | 323.000, not reached | — |
| the rank's changes in answers 6–20 no more than the service's in answers 6–20 | G2-half | 2.000 | 2.000, on the edge | 6.000, on the edge | 1.333, on the edge | 2.000, on the edge | 6.000, on the edge | 12.667, not reached | — | — |
| the card's move in answers 150–200 no more than the service's in answers 6–20 | G2-half | 72.000 | 36.000, reached | 82.000, not reached | 41.000, reached | 39.000, reached | 36.000, reached | 20.000, reached | 54.000, reached | — |
| the rank's changes in answers 150–200 no more than the service's in answers 6–20 | G2-half | 2.000 | 1.176, on the edge | 9.020, not reached | 3.333, on the edge | 0.784, on the edge | 1.373, on the edge | 6.863, not reached | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 57 of 57 |  |
| constant/both | 57 of 57 |  |
| constant_slow/both | 57 of 57 |  |
| floor_0.05/both | 57 of 57 |  |
| no_trial/both | 57 of 57 |  |
| glicko2_floor/general | 56 of 57 | G0-topics1 r1_rms_200 0.527 [0.308, 0.801] against 0.258 |
| glicko2_floor/topics | 46 of 57 | G0 r1_rms_200 0.334 [0.176, 0.494] against 0.107; G1 r1_rms_200 0.528 [0.290, 0.765] against 0.200; G3 r1_rms_200 0.635 [0.406, 0.867] against 0.248; G6 r3_inside -0.138 [-0.256, -0.048] against 0.037; G8 r1_rms_200 0.435 [0.298, 0.583] against 0.205; G0-exact r1_rms_200 0.350 [0.190, 0.539] against 0.170; G2-half r1_rms_200 0.546 [0.258, 0.863] against 0.216; G0-start0.5 r1_rms_200 0.336 [0.177, 0.517] against 0.167; G0-start2 r1_rms_200 0.353 [0.171, 0.555] against 0.114; G0-start2 r3_inside -0.110 [-0.168, -0.053] against 0.053; G0-topics0.3 r1_rms_200 0.465 [0.354, 0.576] against 0.267 |
| oracle/both (the ceiling) | 56 of 57 | G6 r3_inside -0.152 [-0.225, -0.085] against 0.037 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 10 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0248 | 0.107 | 100.0 % | 100.0 % |
| G0 | r3_inside | 0.010 | 0.0200 | 0.086 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0555 | 0.239 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0466 | 0.200 | 100.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0195 | 0.084 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0494 | 0.212 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0784 | 0.337 | 100.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0124 | 0.053 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0627 | 0.270 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0577 | 0.248 | 100.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0163 | 0.070 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0640 | 0.275 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0791 | 0.340 | 100.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0177 | 0.076 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0368 | 0.158 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0329 | 0.142 | 100.0 % | 100.0 % |
| G5 | r3_inside | 0.010 | 0.0148 | 0.064 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0574 | 0.247 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0967 | 0.416 | 100.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0087 | 0.037 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0440 | 0.189 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0453 | 0.195 | 100.0 % | 99.3 % |
| G7 | r3_inside | 0.010 | 0.0165 | 0.071 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0514 | 0.221 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0477 | 0.205 | 100.0 % | 99.0 % |
| G8 | r3_inside | 0.010 | 0.0259 | 0.111 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0576 | 0.248 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0396 | 0.170 | 100.0 % | 99.9 % |
| G0-exact | r3_inside | 0.010 | 0.0441 | 0.190 | 100.0 % | 99.5 % |
| G0-exact | r4_false | 0.020 | 0.0540 | 0.232 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0468 | 0.201 | 100.0 % | 99.0 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0182 | 0.078 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0583 | 0.251 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.1014 | 0.436 | 100.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0120 | 0.052 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0491 | 0.211 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0503 | 0.216 | 100.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0136 | 0.059 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0585 | 0.251 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0678 | 0.291 | 100.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0151 | 0.065 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0216 | 0.093 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0670 | 0.288 | 100.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0207 | 0.089 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0603 | 0.259 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0389 | 0.167 | 100.0 % | 99.9 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0203 | 0.087 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0597 | 0.257 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0265 | 0.114 | 100.0 % | 100.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0122 | 0.053 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0538 | 0.231 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0622 | 0.267 | 100.0 % | 99.0 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0126 | 0.054 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0829 | 0.356 | 100.0 % | 99.8 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0599 | 0.258 | 100.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0156 | 0.067 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0708 | 0.304 | 100.0 % | 100.0 % |
