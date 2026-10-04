# The criterion, read on this run

Seed 20261001, experiment E-A3: 4000 children a generator, 200 answers each. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service. **The exit: floor_0.05/both**, the floor of the highest score, 0.056 [0.055, 0.058], among those that meet the constraints of not worse and of the screen.

### Every rule against the main step of the highest score

The main step of the highest score, whatever it meets: uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both, 0.255 [0.251, 0.259]. Its score less every other rule's, on the same children:

| Rule | Score | The highest main step's less this |
|---|---:|---:|
| constant_slow/both | 0.154 [0.151, 0.158] | 0.101 [0.098, 0.104] |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.3/both | 0.247 [0.243, 0.252] | 0.008 [0.006, 0.010] |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 0.239 [0.235, 0.243] | 0.016 [0.014, 0.018] |
| uncertain_v0.1_s0.7_qt0.003_qd0_L0.3/both | 0.232 [0.228, 0.236] | 0.023 [0.021, 0.025] |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 0.230 [0.226, 0.234] | 0.025 [0.023, 0.028] |
| uncertain_v0.17_s0.7_qt0_qd0_L0.3/both | 0.076 [0.073, 0.079] | 0.179 [0.176, 0.183] |
| uncertain_v0.1_s0.7_qt0_qd0_Lnone/both | -0.001 [-0.003, 0.002] | 0.256 [0.251, 0.260] |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_Lnone/both | 0.190 [0.186, 0.194] | 0.065 [0.062, 0.067] |
| uncertain_v0.13_s0.7_qt0_qd0.006_L0.3/both | 0.063 [0.060, 0.066] | 0.192 [0.188, 0.195] |
| uncertain_v0.13_s0.7_qt0.003_qd0_L0.3/both | 0.239 [0.235, 0.243] | 0.016 [0.013, 0.018] |
| uncertain_v0.13_s0.7_qt0_qd0_Lnone/both | 0.022 [0.019, 0.024] | 0.233 [0.229, 0.238] |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3_kservice/both | 0.207 [0.203, 0.211] | 0.048 [0.045, 0.051] |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 0.245 [0.240, 0.250] | 0.010 [0.007, 0.013] |
| filter_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.197 [0.193, 0.200] | 0.058 [0.056, 0.061] |
| history_s0.7/both | 0.103 [0.099, 0.108] | 0.152 [0.148, 0.156] |
| switch_h2/both | 0.106 [0.102, 0.109] | 0.149 [0.146, 0.153] |
| floor_0.02/both | 0.000 [0.000, 0.000] | 0.255 [0.251, 0.259] |
| floor_0.05/both | 0.056 [0.055, 0.058] | 0.199 [0.195, 0.202] |
| floor_0.1/both | 0.161 [0.157, 0.164] | 0.094 [0.091, 0.097] |

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| constant_slow/both | 0.154 [0.151, 0.158] | 42 of 49 | 7 | 0 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.255 [0.251, 0.259] | 34 of 49 | 15 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.3/both | 0.247 [0.243, 0.252] | 33 of 49 | 16 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 0.239 [0.235, 0.243] | 36 of 49 | 13 | 0 |
| uncertain_v0.1_s0.7_qt0.003_qd0_L0.3/both | 0.232 [0.228, 0.236] | 35 of 49 | 14 | 0 |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 0.230 [0.226, 0.234] | 33 of 49 | 16 | 0 |
| uncertain_v0.17_s0.7_qt0_qd0_L0.3/both | 0.076 [0.073, 0.079] | 46 of 49 | 3 | 0 |
| uncertain_v0.1_s0.7_qt0_qd0_Lnone/both | -0.001 [-0.003, 0.002] | 36 of 49 | 13 | 0 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_Lnone/both | 0.190 [0.186, 0.194] | 30 of 49 | 19 | 0 |
| uncertain_v0.13_s0.7_qt0_qd0.006_L0.3/both | 0.063 [0.060, 0.066] | 40 of 49 | 9 | 0 |
| uncertain_v0.13_s0.7_qt0.003_qd0_L0.3/both | 0.239 [0.235, 0.243] | 37 of 49 | 12 | 0 |
| uncertain_v0.13_s0.7_qt0_qd0_Lnone/both | 0.022 [0.019, 0.024] | 44 of 49 | 5 | 0 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3_kservice/both | 0.207 [0.203, 0.211] | 24 of 49 | 25 | 0 |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 0.245 [0.240, 0.250] | 40 of 49 | 9 | 0 |
| filter_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.197 [0.193, 0.200] | 39 of 49 | 10 | 0 |
| history_s0.7/both | 0.103 [0.099, 0.108] | 41 of 49 | 8 | 0 |
| switch_h2/both | 0.106 [0.102, 0.109] | 33 of 49 | 16 | 0 |
| floor_0.02/both | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| floor_0.05/both | 0.056 [0.055, 0.058] | 46 of 49 | 3 | 0 |
| floor_0.1/both | 0.161 [0.157, 0.164] | 40 of 49 | 9 | 0 |
| oracle/both (the ceiling) | 1.000 [0.994, 1.007] | 40 of 49 | 1 | 8 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the lag at most 45 % of the service's — G2-half, r6_lag, bound 0.276
2. the corridor a share of the way to the ceiling — G2-half, r3_inside, bound 0.422
3. children not caught up after the jump at most 25 % — G3, r6_jump_unsettled, bound 0.250
4. the card's move in answers 6–20 no more than the service's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
5. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G0, r8_rank_6_20, bound 3.667
6. the card's move in answers 150–200 no more than the service's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
7. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G0, r8_rank_150_200, bound 3.667
8. the card's move in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
9. the rank's changes in answers 6–20 no more than the service's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.633
10. the card's move in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
11. the rank's changes in answers 150–200 no more than the service's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.633

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 0.614, not reached | 0.356, not reached | 0.724, not reached | 73.000, reached | 3.667, on the edge | 37.000, reached | 0.678, reached | 73.000, reached | 3.633, on the edge | 36.000, reached | 0.604, reached |
| constant_slow/both | 0.425, not reached | 0.394, not reached | 0.321, not reached | 42.000, reached | 2.737, reached | 42.000, reached | 3.442, reached | 42.000, reached | 2.598, reached | 41.000, reached | 3.224, reached |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.212, reached | 0.418, not reached | 0.104, reached | 52.000, reached | 2.615, reached | 52.000, reached | 4.589, not reached | 52.000, reached | 2.672, reached | 52.000, reached | 4.456, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.3/both | 0.222, reached | 0.415, not reached | 0.108, reached | 52.000, reached | 1.997, reached | 52.000, reached | 4.514, not reached | 52.000, reached | 2.043, reached | 52.000, reached | 4.352, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 0.248, reached | 0.414, not reached | 0.132, reached | 52.000, reached | 2.003, reached | 52.000, reached | 4.597, not reached | 52.000, reached | 2.065, reached | 52.000, reached | 4.450, not reached |
| uncertain_v0.1_s0.7_qt0.003_qd0_L0.3/both | 0.269, on the edge | 0.412, not reached | 0.151, reached | 52.000, reached | 2.033, reached | 52.000, reached | 4.568, not reached | 52.000, reached | 2.068, reached | 52.000, reached | 4.366, not reached |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 0.238, reached | 0.409, not reached | 0.126, reached | 52.000, reached | 0.895, reached | 52.000, reached | 4.361, not reached | 52.000, reached | 0.900, reached | 52.000, reached | 4.387, not reached |
| uncertain_v0.17_s0.7_qt0_qd0_L0.3/both | 0.497, not reached | 0.379, not reached | 0.575, not reached | 52.000, reached | 3.048, reached | 37.000, reached | 1.058, reached | 52.000, reached | 2.963, reached | 36.000, reached | 0.941, reached |
| uncertain_v0.1_s0.7_qt0_qd0_Lnone/both | 0.597, not reached | 0.353, not reached | 0.689, not reached | 74.000, not reached | 2.023, reached | 37.000, reached | 0.891, reached | 74.000, on the edge | 1.888, reached | 35.000, reached | 0.816, reached |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_Lnone/both | 0.359, not reached | 0.400, not reached | 0.191, reached | 79.000, not reached | 3.462, on the edge | 57.000, reached | 4.543, not reached | 79.000, not reached | 3.365, reached | 57.000, reached | 4.410, not reached |
| uncertain_v0.13_s0.7_qt0_qd0.006_L0.3/both | 0.499, not reached | 0.376, not reached | 0.553, not reached | 52.000, reached | 2.183, reached | 43.000, reached | 0.964, reached | 52.000, reached | 2.233, reached | 42.000, reached | 0.918, reached |
| uncertain_v0.13_s0.7_qt0.003_qd0_L0.3/both | 0.263, reached | 0.415, not reached | 0.154, reached | 52.000, reached | 2.630, reached | 52.000, reached | 4.703, not reached | 52.000, reached | 2.667, reached | 52.000, reached | 4.292, not reached |
| uncertain_v0.13_s0.7_qt0_qd0_Lnone/both | 0.584, not reached | 0.360, not reached | 0.680, not reached | 77.000, not reached | 2.917, reached | 37.000, reached | 0.900, reached | 77.000, not reached | 2.798, reached | 36.000, reached | 0.835, reached |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3_kservice/both | 0.204, reached | 0.409, not reached | 0.122, reached | 52.000, reached | 1.270, reached | 52.000, reached | 2.945, reached | 52.000, reached | 1.292, reached | 52.000, reached | 2.765, reached |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 0.245, reached | 0.426, reached | 0.181, reached | 52.000, reached | 12.635, not reached | 51.000, reached | 3.306, reached | 52.000, reached | 12.917, not reached | 51.000, reached | 3.078, reached |
| filter_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 0.328, not reached | 0.404, not reached | 0.196, reached | 52.000, reached | 2.448, reached | 52.000, reached | 4.151, not reached | 52.000, reached | 2.412, reached | 52.000, reached | 4.169, not reached |
| history_s0.7/both | 0.514, not reached | 0.392, not reached | 0.665, not reached | 128.000, not reached | 18.378, not reached | 34.000, reached | 1.270, reached | 128.000, not reached | 17.913, not reached | 35.000, reached | 1.227, reached |
| switch_h2/both | 0.472, not reached | 0.377, not reached | 0.373, not reached | 73.000, reached | 3.667, on the edge | 37.000, reached | 0.799, reached | 73.000, reached | 3.633, on the edge | 38.000, reached | 1.056, reached |
| floor_0.02/both | 0.613, not reached | 0.356, not reached | 0.724, not reached | 73.000, reached | 3.667, on the edge | 37.000, reached | 0.682, reached | 73.000, reached | 3.633, on the edge | 36.000, reached | 0.623, reached |
| floor_0.05/both | 0.547, not reached | 0.370, not reached | 0.573, not reached | 73.000, reached | 3.667, on the edge | 41.000, reached | 1.639, reached | 73.000, reached | 3.633, on the edge | 40.000, reached | 1.534, reached |
| floor_0.1/both | 0.418, not reached | 0.394, not reached | 0.295, not reached | 73.000, reached | 3.667, on the edge | 48.000, reached | 3.605, on the edge | 73.000, reached | 3.633, on the edge | 48.000, reached | 3.234, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.602, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| constant_slow/both | 34 of 38 | G5 r1_rms_200 0.019 [0.015, 0.024] against 0.011; G6 r1_rms_200 0.010 [0.004, 0.017] against 0.015; G0-topics1 r1_rms_200 0.055 [0.050, 0.060] against 0.011; G0-topics1 r3_inside -0.018 [-0.020, -0.016] against 0.010 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 26 of 38 | G0 r1_rms_200 0.019 [0.014, 0.024] against 0.010; G5 r1_rms_200 0.035 [0.030, 0.040] against 0.011; G6 r3_inside -0.008 [-0.011, -0.006] against 0.010; G7 r1_rms_200 0.011 [0.005, 0.017] against 0.011; G8 r1_rms_200 0.022 [0.017, 0.028] against 0.010; G0-exact r1_rms_200 0.036 [0.031, 0.041] against 0.010; G0-exact r3_inside -0.015 [-0.019, -0.010] against 0.010; G0-miss0.25 r1_rms_200 0.030 [0.024, 0.035] against 0.010; G0-miss0.25 r3_inside -0.009 [-0.012, -0.006] against 0.010; G0-start0.5 r1_rms_200 0.020 [0.014, 0.025] against 0.010; G0-start2 r1_rms_200 0.009 [0.003, 0.015] against 0.011; G0-topics0.3 r1_rms_200 0.036 [0.031, 0.041] against 0.010 |
| uncertain_v0.1_s0.7_qt0.003_qd0.006_L0.3/both | 25 of 38 | G0 r1_rms_200 0.020 [0.015, 0.026] against 0.010; G5 r1_rms_200 0.033 [0.028, 0.039] against 0.011; G6 r3_inside -0.009 [-0.011, -0.006] against 0.010; G7 r1_rms_200 0.013 [0.007, 0.019] against 0.011; G8 r1_rms_200 0.023 [0.018, 0.029] against 0.010; G0-exact r1_rms_200 0.033 [0.028, 0.038] against 0.010; G0-exact r3_inside -0.016 [-0.020, -0.012] against 0.010; G0-miss0.25 r1_rms_200 0.033 [0.028, 0.038] against 0.010; G0-miss0.25 r3_inside -0.011 [-0.014, -0.007] against 0.010; G0-start0.5 r1_rms_200 0.018 [0.013, 0.023] against 0.010; G0-start2 r1_rms_200 0.009 [0.002, 0.014] against 0.011; G0-topics0.3 r1_rms_200 0.035 [0.030, 0.040] against 0.010; G0-topics1 r3_inside -0.009 [-0.011, -0.006] against 0.010 |
| uncertain_v0.1_s0.7_qt0.003_qd0.003_L0.3/both | 28 of 38 | G0 r1_rms_200 0.016 [0.010, 0.021] against 0.010; G5 r1_rms_200 0.029 [0.023, 0.034] against 0.011; G7 r1_rms_200 0.010 [0.004, 0.016] against 0.011; G8 r1_rms_200 0.015 [0.010, 0.020] against 0.010; G0-exact r1_rms_200 0.027 [0.021, 0.031] against 0.010; G0-exact r3_inside -0.013 [-0.017, -0.009] against 0.010; G0-miss0.25 r1_rms_200 0.021 [0.016, 0.026] against 0.010; G0-miss0.25 r3_inside -0.008 [-0.011, -0.005] against 0.010; G0-start0.5 r1_rms_200 0.016 [0.011, 0.021] against 0.010; G0-topics0.3 r1_rms_200 0.028 [0.023, 0.033] against 0.010 |
| uncertain_v0.1_s0.7_qt0.003_qd0_L0.3/both | 27 of 38 | G0 r1_rms_200 0.011 [0.006, 0.016] against 0.010; G5 r1_rms_200 0.027 [0.022, 0.033] against 0.011; G6 r1_rms_200 0.009 [0.000, 0.018] against 0.015; G7 r1_rms_200 0.006 [-0.000, 0.013] against 0.011; G8 r1_rms_200 0.012 [0.007, 0.017] against 0.010; G0-exact r1_rms_200 0.022 [0.018, 0.027] against 0.010; G0-exact r3_inside -0.011 [-0.016, -0.007] against 0.010; G0-miss0.25 r1_rms_200 0.018 [0.013, 0.023] against 0.010; G0-start0.5 r1_rms_200 0.013 [0.008, 0.018] against 0.010; G0-topics0.3 r1_rms_200 0.019 [0.014, 0.024] against 0.010; G0-topics1 r3_inside -0.008 [-0.010, -0.006] against 0.010 |
| uncertain_v0.05_s0.7_qt0.003_qd0.006_L0.3/both | 25 of 38 | G0 r1_rms_200 0.023 [0.018, 0.029] against 0.010; G5 r1_rms_200 0.033 [0.028, 0.039] against 0.011; G6 r3_inside -0.010 [-0.012, -0.007] against 0.010; G7 r1_rms_200 0.016 [0.010, 0.022] against 0.011; G8 r1_rms_200 0.022 [0.016, 0.027] against 0.010; G0-exact r1_rms_200 0.033 [0.028, 0.039] against 0.010; G0-exact r3_inside -0.021 [-0.025, -0.017] against 0.010; G0-miss0.25 r1_rms_200 0.031 [0.026, 0.036] against 0.010; G0-miss0.25 r3_inside -0.016 [-0.019, -0.013] against 0.010; G0-start0.5 r1_rms_200 0.021 [0.016, 0.026] against 0.010; G0-start2 r1_rms_200 0.011 [0.004, 0.017] against 0.011; G0-topics0.3 r1_rms_200 0.037 [0.031, 0.042] against 0.010; G0-topics1 r3_inside -0.011 [-0.013, -0.009] against 0.010 |
| uncertain_v0.17_s0.7_qt0_qd0_L0.3/both | 38 of 38 |  |
| uncertain_v0.1_s0.7_qt0_qd0_Lnone/both | 30 of 38 | G0 r1_rms_200 0.009 [0.006, 0.012] against 0.010; G0-exact r1_rms_200 0.007 [0.004, 0.011] against 0.010; G0-exact r3_inside -0.009 [-0.012, -0.005] against 0.010; G0-miss0.25 r3_inside -0.008 [-0.010, -0.005] against 0.010; G0-miss1 r1_rms_200 0.012 [0.008, 0.015] against 0.011; G0-start0.5 r1_rms_200 0.007 [0.004, 0.010] against 0.010; G0-start2 r1_rms_200 0.009 [0.006, 0.013] against 0.011; G0-topics0.3 r1_rms_200 0.011 [0.007, 0.014] against 0.010 |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_Lnone/both | 25 of 38 | G0 r1_rms_200 0.033 [0.028, 0.038] against 0.010; G4 r1_rms_200 0.040 [0.033, 0.048] against 0.013; G5 r1_rms_200 0.039 [0.034, 0.044] against 0.011; G6 r1_rms_200 0.061 [0.052, 0.070] against 0.015; G7 r1_rms_200 0.026 [0.021, 0.033] against 0.011; G8 r1_rms_200 0.031 [0.026, 0.037] against 0.010; G0-exact r1_rms_200 0.036 [0.031, 0.041] against 0.010; G0-exact r3_inside -0.011 [-0.015, -0.007] against 0.010; G0-miss0.25 r1_rms_200 0.035 [0.030, 0.040] against 0.010; G0-miss1 r1_rms_200 0.029 [0.023, 0.035] against 0.011; G0-start0.5 r1_rms_200 0.031 [0.026, 0.035] against 0.010; G0-start2 r1_rms_200 0.019 [0.013, 0.025] against 0.011; G0-topics0.3 r1_rms_200 0.045 [0.040, 0.050] against 0.010 |
| uncertain_v0.13_s0.7_qt0_qd0.006_L0.3/both | 32 of 38 | G6 r3_inside -0.011 [-0.013, -0.009] against 0.010; G0-exact r3_inside -0.009 [-0.013, -0.005] against 0.010; G0-miss0.25 r1_rms_200 0.007 [0.003, 0.011] against 0.010; G0-miss0.25 r3_inside -0.010 [-0.012, -0.007] against 0.010; G3-drop r3_inside -0.012 [-0.013, -0.010] against 0.010; G0-topics0.3 r1_rms_200 0.010 [0.007, 0.014] against 0.010 |
| uncertain_v0.13_s0.7_qt0.003_qd0_L0.3/both | 29 of 38 | G0 r1_rms_200 0.010 [0.005, 0.015] against 0.010; G5 r1_rms_200 0.027 [0.021, 0.032] against 0.011; G6 r1_rms_200 0.008 [-0.000, 0.017] against 0.015; G8 r1_rms_200 0.011 [0.006, 0.016] against 0.010; G0-exact r1_rms_200 0.021 [0.016, 0.026] against 0.010; G0-exact r3_inside -0.009 [-0.013, -0.005] against 0.010; G0-miss0.25 r1_rms_200 0.017 [0.012, 0.023] against 0.010; G0-start0.5 r1_rms_200 0.012 [0.007, 0.018] against 0.010; G0-topics0.3 r1_rms_200 0.015 [0.010, 0.021] against 0.010 |
| uncertain_v0.13_s0.7_qt0_qd0_Lnone/both | 38 of 38 |  |
| uncertain_v0.13_s0.7_qt0.003_qd0.006_L0.3_kservice/both | 14 of 38 | G0 r1_rms_200 0.046 [0.041, 0.052] against 0.010; G0 r3_inside -0.013 [-0.016, -0.011] against 0.010; G1 r1_rms_200 0.025 [0.020, 0.031] against 0.010; G4 r3_inside -0.014 [-0.016, -0.012] against 0.010; G5 r1_rms_200 0.043 [0.038, 0.049] against 0.011; G5 r3_inside -0.015 [-0.018, -0.013] against 0.010; G6 r3_inside -0.030 [-0.033, -0.027] against 0.010; G7 r1_rms_200 0.030 [0.024, 0.036] against 0.011; G7 r3_inside -0.011 [-0.014, -0.009] against 0.010; G8 r1_rms_200 0.047 [0.042, 0.052] against 0.010; G8 r3_inside -0.014 [-0.016, -0.011] against 0.010; G0-exact r1_rms_200 0.064 [0.059, 0.069] against 0.010; G0-exact r3_inside -0.035 [-0.040, -0.030] against 0.010; G0-miss0.25 r1_rms_200 0.059 [0.053, 0.064] against 0.010; G0-miss0.25 r3_inside -0.029 [-0.033, -0.026] against 0.010; G0-miss1 r1_rms_200 0.014 [0.008, 0.020] against 0.011; G3-drop r3_inside -0.030 [-0.032, -0.028] against 0.010; G0-start0.5 r1_rms_200 0.047 [0.041, 0.052] against 0.010; G0-start0.5 r3_inside -0.013 [-0.015, -0.011] against 0.010; G0-start2 r1_rms_200 0.034 [0.028, 0.040] against 0.011; G0-start2 r3_inside -0.008 [-0.010, -0.006] against 0.010; G0-topics0.3 r1_rms_200 0.079 [0.073, 0.084] against 0.010; G0-topics0.3 r3_inside -0.019 [-0.022, -0.017] against 0.010; G0-topics1 r3_inside -0.010 [-0.012, -0.008] against 0.010 |
| stored_s0.7_qt0.0015_qd0.006_L0.3/both | 31 of 38 | G5 r1_rms_200 0.030 [0.024, 0.036] against 0.011; G6 r3_inside -0.016 [-0.019, -0.012] against 0.010; G0-exact r1_rms_200 0.009 [0.004, 0.014] against 0.010; G0-exact r3_inside -0.013 [-0.018, -0.007] against 0.010; G0-miss0.25 r1_rms_200 0.009 [0.004, 0.014] against 0.010; G3-drop r3_inside -0.010 [-0.012, -0.007] against 0.010; G0-topics0.3 r1_rms_200 0.012 [0.007, 0.017] against 0.010 |
| filter_v0.13_s0.7_qt0.003_qd0.006_L0.3/both | 32 of 38 | G5 r1_rms_200 0.017 [0.012, 0.022] against 0.011; G0-exact r1_rms_200 0.014 [0.009, 0.018] against 0.010; G0-exact r3_inside -0.007 [-0.011, -0.003] against 0.010; G0-miss0.25 r1_rms_200 0.010 [0.005, 0.015] against 0.010; G0-start0.5 r1_rms_200 0.006 [0.001, 0.010] against 0.010; G0-topics0.3 r1_rms_200 0.016 [0.011, 0.021] against 0.010 |
| history_s0.7/both | 37 of 38 | G3-drop r1_rms_200 0.020 [0.014, 0.026] against 0.014 |
| switch_h2/both | 25 of 38 | G0 r1_rms_200 0.020 [0.016, 0.025] against 0.010; G4 r1_rms_200 0.025 [0.020, 0.031] against 0.013; G5 r1_rms_200 0.022 [0.018, 0.027] against 0.011; G6 r1_rms_200 0.042 [0.035, 0.049] against 0.015; G7 r1_rms_200 0.015 [0.011, 0.020] against 0.011; G8 r1_rms_200 0.025 [0.020, 0.029] against 0.010; G0-exact r1_rms_200 0.025 [0.021, 0.030] against 0.010; G0-miss0.25 r1_rms_200 0.023 [0.019, 0.027] against 0.010; G0-miss1 r1_rms_200 0.020 [0.015, 0.025] against 0.011; G0-start0.5 r1_rms_200 0.022 [0.018, 0.026] against 0.010; G0-start2 r1_rms_200 0.009 [0.004, 0.015] against 0.011; G0-topics0.3 r1_rms_200 0.022 [0.018, 0.026] against 0.010; G0-topics1 r1_rms_200 0.020 [0.016, 0.025] against 0.011 |
| floor_0.02/both | 38 of 38 |  |
| floor_0.05/both | 38 of 38 |  |
| floor_0.1/both | 32 of 38 | G4 r1_rms_200 0.014 [0.008, 0.020] against 0.013; G5 r1_rms_200 0.013 [0.009, 0.018] against 0.011; G6 r1_rms_200 0.037 [0.030, 0.044] against 0.015; G0-miss0.25 r1_rms_200 0.008 [0.003, 0.013] against 0.010; G0-start0.5 r1_rms_200 0.007 [0.002, 0.011] against 0.010; G0-topics0.3 r1_rms_200 0.006 [0.001, 0.010] against 0.010 |
| oracle/both (the ceiling) | 37 of 38 | G6 r3_inside -0.062 [-0.067, -0.058] against 0.010 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the service, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the service passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 4000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0023 | 0.010 | 99.0 % | 99.0 % |
| G0 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0024 | 0.010 | 99.0 % | 99.0 % |
| G1 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0038 | 0.016 | 99.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0033 | 0.014 | 99.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0030 | 0.013 | 99.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0007 | 0.010 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G5 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0034 | 0.015 | 99.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0010 | 0.010 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0026 | 0.011 | 99.0 % | 99.0 % |
| G7 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0024 | 0.010 | 99.0 % | 99.0 % |
| G8 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0022 | 0.010 | 99.5 % | 99.5 % |
| G0-exact | r3_inside | 0.010 | 0.0017 | 0.010 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0022 | 0.010 | 99.5 % | 99.5 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0013 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G0-miss1 | r3_inside | 0.010 | 0.0006 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0032 | 0.014 | 99.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0029 | 0.012 | 99.0 % | 99.0 % |
| G2-fading | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0034 | 0.014 | 99.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0022 | 0.010 | 99.5 % | 99.5 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0026 | 0.011 | 99.0 % | 99.0 % |
| G0-start2 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0023 | 0.010 | 99.2 % | 99.2 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0026 | 0.011 | 99.0 % | 99.0 % |
| G0-topics1 | r3_inside | 0.010 | 0.0010 | 0.010 | 100.0 % | 100.0 % |
