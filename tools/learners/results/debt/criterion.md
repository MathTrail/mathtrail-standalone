# The criterion, read on this run

Seed 20261001, experiment E-A3: 4000 children a generator, 200 answers each. The score is the share of the way from the baseline, floor_0.05+cautious_z1/both — the model the service runs — to a card no wrong answer moves, a rule closes, with its 95 % interval.

## The choice

Candidates that meet every constraint and are better than the baseline: 0.

No candidate meets every constraint and is better than the baseline. **The exit: floor_0.05+cautious_z1/both**, the baseline itself, kept as it is.

### Every rule against the main debt of the highest score

The main debt of the highest score, whatever it meets: floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both, 0.761 [0.758, 0.764]. Its score less every other rule's, on the same children:

| Rule | Score | The highest main debt's less this |
|---|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 0.761 [0.758, 0.764] |
| constant_slow/both | 0.138 [0.136, 0.140] | 0.623 [0.620, 0.625] |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 0.695 [0.692, 0.698] | 0.066 [0.065, 0.067] |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 0.766 [0.762, 0.769] | -0.005 [-0.006, -0.004] |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 0.748 [0.745, 0.751] | 0.013 [0.012, 0.014] |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 0.759 [0.755, 0.762] | 0.002 [0.002, 0.003] |

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 79 of 83 | 4 | 0 |
| constant_slow/both | 0.138 [0.136, 0.140] | 74 of 83 | 9 | 0 |
| floor_0.05+cautious_z1/both | 0.000 [0.000, 0.000] | 79 of 83 | 4 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 0.761 [0.758, 0.764] | 35 of 83 | 48 | 0 |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 0.695 [0.692, 0.698] | 45 of 83 | 38 | 0 |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 0.766 [0.762, 0.769] | 42 of 83 | 41 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 0.748 [0.745, 0.751] | 32 of 83 | 51 | 0 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 0.759 [0.755, 0.762] | 51 of 83 | 32 | 0 |
| oracle/both (the ceiling) | — | 65 of 83 | 2 | 16 |

## Goals

Every rule's value against each goal's bound, and whether its interval reaches the bound, stands on its edge or does not reach it. The goals:

1. the card's fall after a wrong answer in answers 6–20 at most 50 % of the baseline's — G0, r8_wrong_fall_6_20, bound 34.131
2. right answers that leave the card still in answers 6–20 at most 10 % — G0, r8_right_still_6_20, bound 0.100
3. the card's fall after a wrong answer in answers 150–200 at most 50 % of the baseline's — G0, r8_wrong_fall_150_200, bound 17.165
4. right answers that leave the card still in answers 150–200 at most 10 % — G0, r8_right_still_150_200, bound 0.100
5. the card's fall after a wrong answer in answers 6–20 at most 50 % of the baseline's — G2-half, r8_wrong_fall_6_20, bound 34.174
6. right answers that leave the card still in answers 6–20 at most 10 % — G2-half, r8_right_still_6_20, bound 0.100
7. the card's fall after a wrong answer in answers 150–200 at most 50 % of the baseline's — G2-half, r8_wrong_fall_150_200, bound 18.012
8. right answers that leave the card still in answers 150–200 at most 10 % — G2-half, r8_right_still_150_200, bound 0.100
9. the bias after 200 answers at most 0.15 of a rank — G0, r1_mean_200, bound 0.143
10. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_move_p95_6_20, bound 73.000
11. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G0, r8_rank_6_20, bound 3.677
12. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_move_p95_150_200, bound 73.000
13. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G0, r8_rank_150_200, bound 3.677
14. the card's move in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_6_20, bound 73.000
15. the rank's changes in answers 6–20 no more than the baseline's in answers 6–20 — G2-half, r8_rank_6_20, bound 3.587
16. the card's move in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_move_p95_150_200, bound 73.000
17. the rank's changes in answers 150–200 no more than the baseline's in answers 6–20 — G2-half, r8_rank_150_200, bound 3.587

| Rule | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| shrinking/both (the service) | 68.262, not reached | 0.000, reached | 34.330, not reached | 0.000, reached | 68.348, not reached | 0.000, reached | 36.024, not reached | 0.000, reached | -0.098, reached | 73.000, reached | 3.677, on the edge | 40.000, reached | 1.689, reached | 73.000, reached | 3.587, on the edge | 40.000, reached | 1.508, reached |
| constant_slow/both | 39.975, not reached | 0.000, reached | 39.894, not reached | 0.000, reached | 39.960, not reached | 0.000, reached | 40.171, not reached | 0.000, reached | -0.094, reached | 42.000, reached | 2.732, reached | 42.000, reached | 3.447, reached | 42.000, reached | 2.603, reached | 42.000, reached | 3.295, reached |
| floor_0.05+cautious_z1/both | 68.262, not reached | 0.000, reached | 34.330, not reached | 0.000, reached | 68.348, not reached | 0.000, reached | 36.024, not reached | 0.000, reached | -0.098, reached | 73.000, reached | 3.677, on the edge | 40.000, reached | 1.689, reached | 73.000, reached | 3.587, on the edge | 40.000, reached | 1.508, reached |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 18.766, reached | 0.000, reached | 8.073, reached | 0.003, reached | 18.313, reached | 0.000, reached | 6.445, reached | 0.002, reached | 0.128, reached | 29.000, reached | 1.382, reached | 14.000, reached | 0.468, reached | 29.000, reached | 1.317, reached | 13.000, reached | 0.418, reached |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 22.894, reached | 0.000, reached | 10.297, reached | 0.002, reached | 22.575, reached | 0.000, reached | 9.210, reached | 0.001, reached | 0.031, reached | 29.000, reached | 1.460, reached | 14.000, reached | 0.467, reached | 29.000, reached | 1.448, reached | 13.000, reached | 0.588, reached |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 18.737, reached | 0.227, not reached | 7.658, reached | 0.679, not reached | 18.286, reached | 0.218, not reached | 6.208, reached | 0.450, not reached | -0.030, reached | 29.000, reached | 1.277, reached | 14.000, reached | 0.305, reached | 29.000, reached | 1.232, reached | 13.000, reached | 0.449, reached |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 18.833, reached | 0.000, reached | 8.870, reached | 0.002, reached | 18.343, reached | 0.000, reached | 7.401, reached | 0.000, reached | 0.382, not reached | 29.000, reached | 1.637, reached | 16.000, reached | 0.631, reached | 29.000, reached | 1.618, reached | 15.000, reached | 0.676, reached |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 18.766, reached | 0.000, reached | 8.239, reached | 0.002, reached | 18.313, reached | 0.000, reached | 6.594, reached | 0.001, reached | 0.124, reached | 29.000, reached | 1.382, reached | 14.000, reached | 0.450, reached | 29.000, reached | 1.317, reached | 13.000, reached | 0.444, reached |
| oracle/both (the ceiling) | — | — | — | — | — | — | — | — | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the baseline

What of not worse each rule does not meet, or the run cannot read: its difference from the baseline, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 66 of 66 |  |
| constant_slow/both | 61 of 66 | G0 r7_error_10 0.020 [0.019, 0.022] against 0.010; G1 r7_error_10 0.019 [0.017, 0.020] against 0.010; G5 r1_rms_200 0.020 [0.016, 0.025] against 0.010; G0-topics1 r1_rms_200 0.051 [0.047, 0.055] against 0.010; G0-topics1 r3_inside -0.017 [-0.019, -0.016] against 0.010 |
| floor_0.05+cautious_z1/both | 66 of 66 |  |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 18 of 66 | G0 r1_rms_200 0.031 [0.026, 0.037] against 0.010; G0 r3_inside -0.025 [-0.028, -0.023] against 0.010; G0 r4_false 0.037 [0.033, 0.042] against 0.020; G0 r7_error_10 0.027 [0.024, 0.030] against 0.010; G0 r7_longest_wrong 0.139 [0.125, 0.154] against 0.100; G1 r1_rms_200 0.028 [0.023, 0.033] against 0.010; G1 r3_inside -0.015 [-0.018, -0.013] against 0.010; G1 r7_error_10 0.040 [0.037, 0.043] against 0.010; G1 r7_longest_wrong 0.137 [0.122, 0.153] against 0.100; G2 r1_rms_200 0.038 [0.034, 0.043] against 0.013; G3 r4_false 0.029 [0.025, 0.033] against 0.020; G4 r3_inside -0.023 [-0.025, -0.021] against 0.010; G4 r4_false 0.024 [0.021, 0.028] against 0.020; G5 r1_rms_200 0.037 [0.032, 0.043] against 0.010; G5 r3_inside -0.027 [-0.029, -0.024] against 0.010; G5 r4_false 0.035 [0.031, 0.039] against 0.020; G6 r3_inside -0.048 [-0.052, -0.045] against 0.010; G6 r4_false 0.061 [0.054, 0.067] against 0.020; G7 r1_rms_200 0.008 [0.002, 0.015] against 0.010; G7 r3_inside -0.021 [-0.023, -0.018] against 0.010; G7 r4_false 0.034 [0.030, 0.039] against 0.020; G8 r1_rms_200 0.031 [0.026, 0.036] against 0.010; G8 r3_inside -0.026 [-0.029, -0.024] against 0.010; G8 r4_false 0.038 [0.034, 0.043] against 0.020; G0-exact r1_rms_200 0.038 [0.033, 0.043] against 0.010; G0-exact r3_inside -0.049 [-0.055, -0.044] against 0.010; G0-exact r4_false 0.031 [0.027, 0.036] against 0.020; G0-miss0.25 r1_rms_200 0.039 [0.034, 0.045] against 0.010; G0-miss0.25 r3_inside -0.045 [-0.049, -0.041] against 0.010; G0-miss0.25 r4_false 0.036 [0.032, 0.041] against 0.020; G0-miss1 r4_false 0.038 [0.034, 0.043] against 0.020; G2-half r4_false 0.021 [0.017, 0.024] against 0.020; G3-drop r1_rms_200 0.299 [0.292, 0.306] against 0.010; G3-drop r3_inside -0.078 [-0.080, -0.075] against 0.010; G3-drop r4_false 0.072 [0.067, 0.077] against 0.020; G3-drop r6_jump_answers 36.702 [35.623, 37.805] against 1.588; G0-start0.5 r1_rms_200 0.022 [0.017, 0.027] against 0.010; G0-start0.5 r3_inside -0.023 [-0.025, -0.020] against 0.010; G0-start0.5 r4_false 0.031 [0.026, 0.036] against 0.020; G0-start2 r1_rms_200 0.030 [0.025, 0.036] against 0.010; G0-start2 r3_inside -0.023 [-0.026, -0.021] against 0.010; G0-start2 r4_false 0.035 [0.031, 0.039] against 0.020; G0-topics0.3 r1_rms_200 0.013 [0.008, 0.019] against 0.010; G0-topics0.3 r3_inside -0.024 [-0.026, -0.021] against 0.010; G0-topics0.3 r4_false 0.034 [0.029, 0.038] against 0.020; G0-topics1 r1_rms_200 0.067 [0.063, 0.072] against 0.010; G0-topics1 r3_inside -0.029 [-0.031, -0.026] against 0.010; G0-topics1 r4_false 0.045 [0.041, 0.049] against 0.020 |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 28 of 66 | G0 r3_inside -0.015 [-0.018, -0.013] against 0.010; G0 r4_false 0.027 [0.023, 0.031] against 0.020; G0 r7_error_10 0.024 [0.021, 0.027] against 0.010; G0 r7_longest_wrong 0.118 [0.105, 0.132] against 0.100; G1 r7_error_10 0.034 [0.032, 0.037] against 0.010; G1 r7_longest_wrong 0.113 [0.100, 0.128] against 0.100; G2 r1_rms_200 0.051 [0.047, 0.056] against 0.013; G3 r1_rms_200 0.009 [0.004, 0.013] against 0.011; G3 r4_false 0.022 [0.019, 0.025] against 0.020; G4 r3_inside -0.015 [-0.017, -0.013] against 0.010; G4 r4_false 0.017 [0.015, 0.020] against 0.020; G5 r1_rms_200 0.012 [0.007, 0.016] against 0.010; G5 r3_inside -0.017 [-0.019, -0.015] against 0.010; G5 r4_false 0.022 [0.018, 0.026] against 0.020; G6 r3_inside -0.036 [-0.039, -0.033] against 0.010; G6 r4_false 0.045 [0.040, 0.051] against 0.020; G7 r3_inside -0.011 [-0.014, -0.009] against 0.010; G7 r4_false 0.022 [0.018, 0.026] against 0.020; G8 r3_inside -0.017 [-0.019, -0.015] against 0.010; G8 r4_false 0.025 [0.021, 0.029] against 0.020; G0-exact r3_inside -0.028 [-0.032, -0.023] against 0.010; G0-exact r4_false 0.019 [0.015, 0.022] against 0.020; G0-miss0.25 r3_inside -0.027 [-0.031, -0.024] against 0.010; G0-miss0.25 r4_false 0.023 [0.019, 0.027] against 0.020; G0-miss1 r4_false 0.026 [0.022, 0.030] against 0.020; G3-drop r1_rms_200 0.198 [0.192, 0.205] against 0.010; G3-drop r3_inside -0.060 [-0.062, -0.057] against 0.010; G3-drop r4_false 0.052 [0.048, 0.057] against 0.020; G3-drop r6_jump_answers 30.356 [29.336, 31.387] against 1.588; G0-start0.5 r3_inside -0.012 [-0.015, -0.010] against 0.010; G0-start0.5 r4_false 0.018 [0.014, 0.022] against 0.020; G0-start2 r3_inside -0.013 [-0.016, -0.011] against 0.010; G0-start2 r4_false 0.023 [0.019, 0.026] against 0.020; G0-topics0.3 r3_inside -0.012 [-0.015, -0.010] against 0.010; G0-topics0.3 r4_false 0.023 [0.018, 0.027] against 0.020; G0-topics1 r1_rms_200 0.043 [0.039, 0.047] against 0.010; G0-topics1 r3_inside -0.021 [-0.023, -0.019] against 0.010; G0-topics1 r4_false 0.031 [0.027, 0.035] against 0.020 |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 29 of 66 | G0 r3_inside -0.014 [-0.016, -0.012] against 0.010; G0 r4_false 0.024 [0.019, 0.028] against 0.020; G0 r7_error_10 0.025 [0.022, 0.028] against 0.010; G0 r7_longest_wrong 0.136 [0.123, 0.151] against 0.100; G1 r7_error_10 0.036 [0.034, 0.039] against 0.010; G1 r7_longest_wrong 0.133 [0.118, 0.148] against 0.100; G2 r1_rms_200 0.061 [0.056, 0.065] against 0.013; G3 r1_rms_200 0.032 [0.027, 0.036] against 0.011; G3 r4_false 0.021 [0.018, 0.025] against 0.020; G4 r3_inside -0.013 [-0.015, -0.011] against 0.010; G5 r1_rms_200 0.010 [0.006, 0.015] against 0.010; G5 r3_inside -0.016 [-0.018, -0.014] against 0.010; G5 r4_false 0.019 [0.016, 0.023] against 0.020; G6 r3_inside -0.033 [-0.036, -0.030] against 0.010; G6 r4_false 0.042 [0.036, 0.047] against 0.020; G7 r3_inside -0.010 [-0.012, -0.008] against 0.010; G7 r4_false 0.019 [0.015, 0.023] against 0.020; G8 r3_inside -0.015 [-0.017, -0.012] against 0.010; G8 r4_false 0.021 [0.017, 0.025] against 0.020; G0-exact r3_inside -0.024 [-0.028, -0.020] against 0.010; G0-miss0.25 r3_inside -0.023 [-0.026, -0.020] against 0.010; G0-miss0.25 r4_false 0.020 [0.016, 0.024] against 0.020; G0-miss1 r4_false 0.025 [0.020, 0.029] against 0.020; G2-half r1_rms_200 0.019 [0.015, 0.024] against 0.010; G3-drop r1_rms_200 0.156 [0.150, 0.162] against 0.010; G3-drop r3_inside -0.054 [-0.056, -0.051] against 0.010; G3-drop r4_false 0.045 [0.041, 0.050] against 0.020; G3-drop r6_jump_answers 26.120 [25.133, 27.041] against 1.588; G0-start0.5 r3_inside -0.011 [-0.013, -0.009] against 0.010; G0-start0.5 r4_false 0.017 [0.013, 0.021] against 0.020; G0-start2 r3_inside -0.012 [-0.014, -0.010] against 0.010; G0-start2 r4_false 0.019 [0.016, 0.023] against 0.020; G0-topics0.3 r3_inside -0.010 [-0.012, -0.008] against 0.010; G0-topics0.3 r4_false 0.020 [0.015, 0.024] against 0.020; G0-topics1 r1_rms_200 0.044 [0.040, 0.048] against 0.010; G0-topics1 r3_inside -0.021 [-0.023, -0.019] against 0.010; G0-topics1 r4_false 0.030 [0.026, 0.034] against 0.020 |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 16 of 66 | G0 r1_rms_200 0.108 [0.102, 0.115] against 0.010; G0 r3_inside -0.041 [-0.044, -0.038] against 0.010; G0 r4_false 0.068 [0.064, 0.073] against 0.020; G0 r7_error_10 0.026 [0.023, 0.029] against 0.010; G0 r7_longest_wrong 0.143 [0.128, 0.157] against 0.100; G1 r1_rms_200 0.104 [0.097, 0.110] against 0.010; G1 r3_inside -0.025 [-0.028, -0.022] against 0.010; G1 r4_false 0.021 [0.018, 0.024] against 0.020; G1 r7_error_10 0.044 [0.041, 0.047] against 0.010; G1 r7_longest_wrong 0.139 [0.124, 0.154] against 0.100; G3 r4_false 0.041 [0.038, 0.045] against 0.020; G4 r3_inside -0.035 [-0.037, -0.033] against 0.010; G4 r4_false 0.034 [0.031, 0.038] against 0.020; G5 r1_rms_200 0.116 [0.109, 0.123] against 0.010; G5 r3_inside -0.043 [-0.046, -0.040] against 0.010; G5 r4_false 0.067 [0.062, 0.071] against 0.020; G6 r3_inside -0.068 [-0.072, -0.064] against 0.010; G6 r4_false 0.088 [0.082, 0.094] against 0.020; G7 r1_rms_200 0.049 [0.040, 0.059] against 0.010; G7 r3_inside -0.034 [-0.038, -0.031] against 0.010; G7 r4_false 0.068 [0.064, 0.073] against 0.020; G8 r1_rms_200 0.113 [0.106, 0.120] against 0.010; G8 r3_inside -0.042 [-0.045, -0.038] against 0.010; G8 r4_false 0.069 [0.065, 0.074] against 0.020; G0-exact r1_rms_200 0.129 [0.123, 0.135] against 0.010; G0-exact r3_inside -0.089 [-0.094, -0.083] against 0.010; G0-exact r4_false 0.065 [0.061, 0.069] against 0.020; G0-miss0.25 r1_rms_200 0.130 [0.123, 0.136] against 0.010; G0-miss0.25 r3_inside -0.076 [-0.080, -0.071] against 0.010; G0-miss0.25 r4_false 0.070 [0.065, 0.075] against 0.020; G0-miss1 r1_rms_200 0.036 [0.028, 0.044] against 0.010; G0-miss1 r4_false 0.065 [0.061, 0.070] against 0.020; G2-half r4_false 0.038 [0.035, 0.041] against 0.020; G2-fading r4_false 0.041 [0.037, 0.044] against 0.020; G3-drop r1_rms_200 0.413 [0.406, 0.420] against 0.010; G3-drop r3_inside -0.103 [-0.105, -0.101] against 0.010; G3-drop r4_false 0.102 [0.096, 0.107] against 0.020; G3-drop r6_jump_answers 45.265 [44.041, 46.484] against 1.588; G0-start0.5 r1_rms_200 0.105 [0.099, 0.112] against 0.010; G0-start0.5 r3_inside -0.039 [-0.042, -0.036] against 0.010; G0-start0.5 r4_false 0.058 [0.053, 0.063] against 0.020; G0-start2 r1_rms_200 0.106 [0.099, 0.113] against 0.010; G0-start2 r3_inside -0.036 [-0.039, -0.033] against 0.010; G0-start2 r4_false 0.067 [0.063, 0.072] against 0.020; G0-topics0.3 r1_rms_200 0.118 [0.110, 0.124] against 0.010; G0-topics0.3 r3_inside -0.044 [-0.047, -0.040] against 0.010; G0-topics0.3 r4_false 0.068 [0.063, 0.073] against 0.020; G0-topics1 r1_rms_200 0.082 [0.076, 0.088] against 0.010; G0-topics1 r3_inside -0.038 [-0.041, -0.036] against 0.010; G0-topics1 r4_false 0.070 [0.066, 0.075] against 0.020 |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 34 of 66 | G0 r1_rms_200 0.026 [0.021, 0.032] against 0.010; G0 r3_inside -0.027 [-0.030, -0.024] against 0.010; G0 r7_error_10 0.027 [0.024, 0.030] against 0.010; G0 r7_longest_wrong 0.139 [0.125, 0.154] against 0.100; G1 r1_rms_200 0.030 [0.026, 0.035] against 0.010; G1 r3_inside -0.012 [-0.015, -0.010] against 0.010; G1 r7_error_10 0.040 [0.038, 0.043] against 0.010; G1 r7_longest_wrong 0.137 [0.123, 0.152] against 0.100; G2 r1_rms_200 0.039 [0.034, 0.044] against 0.013; G4 r3_inside -0.025 [-0.027, -0.023] against 0.010; G5 r1_rms_200 0.035 [0.030, 0.041] against 0.010; G5 r3_inside -0.026 [-0.029, -0.024] against 0.010; G6 r3_inside -0.050 [-0.053, -0.046] against 0.010; G7 r1_rms_200 0.006 [-0.000, 0.013] against 0.010; G7 r3_inside -0.021 [-0.024, -0.019] against 0.010; G8 r1_rms_200 0.028 [0.023, 0.033] against 0.010; G8 r3_inside -0.028 [-0.031, -0.025] against 0.010; G0-exact r1_rms_200 0.034 [0.029, 0.039] against 0.010; G0-exact r3_inside -0.052 [-0.057, -0.047] against 0.010; G0-miss0.25 r1_rms_200 0.036 [0.031, 0.041] against 0.010; G0-miss0.25 r3_inside -0.047 [-0.051, -0.043] against 0.010; G3-drop r1_rms_200 0.289 [0.283, 0.296] against 0.010; G3-drop r3_inside -0.080 [-0.082, -0.077] against 0.010; G3-drop r6_jump_answers 36.589 [35.488, 37.703] against 1.588; G0-start0.5 r1_rms_200 0.021 [0.016, 0.026] against 0.010; G0-start0.5 r3_inside -0.023 [-0.026, -0.020] against 0.010; G0-start2 r1_rms_200 0.028 [0.023, 0.033] against 0.010; G0-start2 r3_inside -0.025 [-0.027, -0.022] against 0.010; G0-topics0.3 r1_rms_200 0.011 [0.005, 0.016] against 0.010; G0-topics0.3 r3_inside -0.025 [-0.028, -0.022] against 0.010; G0-topics1 r1_rms_200 0.061 [0.056, 0.065] against 0.010; G0-topics1 r3_inside -0.032 [-0.034, -0.029] against 0.010 |
| oracle/both (the ceiling) | 64 of 66 | G6 r3_inside -0.076 [-0.080, -0.070] against 0.010; G6 r4_false 0.078 [0.071, 0.086] against 0.020 |

## What the bench resolves

Each check of not worse: the author's tolerance, the standard error of the paired difference between constant_slow and the baseline, the tolerance the check is read with — never under 4.3 standard errors — and the chance a candidate exactly as good as the baseline passes it, read as this run reads it and as a decision run does.

| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, 4000 children | Passes, 4000 children |
|---|---|---:|---:|---:|---:|---:|
| G0 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0 | r4_false | 0.020 | 0.0011 | 0.020 | 100.0 % | 100.0 % |
| G0 | r7_error_10 | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0 | r7_longest_wrong | 0.100 | 0.0055 | 0.100 | 100.0 % | 100.0 % |
| G1 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G1 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G1 | r4_false | 0.020 | 0.0005 | 0.020 | 100.0 % | 100.0 % |
| G1 | r7_error_10 | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G1 | r7_longest_wrong | 0.100 | 0.0049 | 0.100 | 100.0 % | 100.0 % |
| G2 | r1_rms_200 | 0.010 | 0.0030 | 0.013 | 99.0 % | 99.0 % |
| G2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2 | r4_false | 0.020 | 0.0005 | 0.020 | 100.0 % | 100.0 % |
| G2 | abs(r6_lag) | 0.010 | 0.0027 | 0.012 | 99.0 % | 99.0 % |
| G3 | r1_rms_200 | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3 | r4_false | 0.020 | 0.0008 | 0.020 | 100.0 % | 100.0 % |
| G3 | r6_jump_answers | 1.000 | 0.4065 | 1.748 | 99.0 % | 99.0 % |
| G4 | r1_rms_200 | 0.010 | 0.0024 | 0.010 | 99.0 % | 99.0 % |
| G4 | r3_inside | 0.010 | 0.0007 | 0.010 | 100.0 % | 100.0 % |
| G4 | r4_false | 0.020 | 0.0007 | 0.020 | 100.0 % | 100.0 % |
| G5 | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.7 % | 99.7 % |
| G5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G5 | r4_false | 0.020 | 0.0012 | 0.020 | 100.0 % | 100.0 % |
| G6 | r1_rms_200 | 0.010 | 0.0027 | 0.012 | 99.0 % | 99.0 % |
| G6 | r3_inside | 0.010 | 0.0009 | 0.010 | 100.0 % | 100.0 % |
| G6 | r4_false | 0.020 | 0.0014 | 0.020 | 100.0 % | 100.0 % |
| G7 | r1_rms_200 | 0.010 | 0.0022 | 0.010 | 99.4 % | 99.4 % |
| G7 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G7 | r4_false | 0.020 | 0.0012 | 0.020 | 100.0 % | 100.0 % |
| G8 | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G8 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G8 | r4_false | 0.020 | 0.0011 | 0.020 | 100.0 % | 100.0 % |
| G0-exact | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r3_inside | 0.010 | 0.0014 | 0.010 | 100.0 % | 100.0 % |
| G0-exact | r4_false | 0.020 | 0.0012 | 0.020 | 100.0 % | 100.0 % |
| G0-miss0.25 | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 99.9 % | 99.9 % |
| G0-miss0.25 | r3_inside | 0.010 | 0.0011 | 0.010 | 100.0 % | 100.0 % |
| G0-miss0.25 | r4_false | 0.020 | 0.0013 | 0.020 | 100.0 % | 100.0 % |
| G0-miss1 | r1_rms_200 | 0.010 | 0.0021 | 0.010 | 99.7 % | 99.7 % |
| G0-miss1 | r3_inside | 0.010 | 0.0006 | 0.010 | 100.0 % | 100.0 % |
| G0-miss1 | r4_false | 0.020 | 0.0011 | 0.020 | 100.0 % | 100.0 % |
| G2-half | r1_rms_200 | 0.010 | 0.0024 | 0.010 | 99.0 % | 99.0 % |
| G2-half | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2-half | r4_false | 0.020 | 0.0007 | 0.020 | 100.0 % | 100.0 % |
| G2-half | abs(r6_lag) | 0.010 | 0.0025 | 0.011 | 99.0 % | 99.0 % |
| G2-fading | r1_rms_200 | 0.010 | 0.0023 | 0.010 | 99.2 % | 99.2 % |
| G2-fading | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G2-fading | r4_false | 0.020 | 0.0009 | 0.020 | 100.0 % | 100.0 % |
| G2-fading | abs(r6_lag) | 0.010 | 0.0026 | 0.011 | 99.0 % | 99.0 % |
| G3-drop | r1_rms_200 | 0.010 | 0.0024 | 0.010 | 99.0 % | 99.0 % |
| G3-drop | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G3-drop | r4_false | 0.020 | 0.0015 | 0.020 | 100.0 % | 100.0 % |
| G3-drop | r6_jump_answers | 1.000 | 0.3692 | 1.588 | 99.0 % | 99.0 % |
| G0-start0.5 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0-start0.5 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start0.5 | r4_false | 0.020 | 0.0010 | 0.020 | 100.0 % | 100.0 % |
| G0-start2 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.9 % | 99.9 % |
| G0-start2 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-start2 | r4_false | 0.020 | 0.0011 | 0.020 | 100.0 % | 100.0 % |
| G0-topics0.3 | r1_rms_200 | 0.010 | 0.0019 | 0.010 | 99.9 % | 99.9 % |
| G0-topics0.3 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-topics0.3 | r4_false | 0.020 | 0.0011 | 0.020 | 100.0 % | 100.0 % |
| G0-topics1 | r1_rms_200 | 0.010 | 0.0020 | 0.010 | 99.8 % | 99.8 % |
| G0-topics1 | r3_inside | 0.010 | 0.0008 | 0.010 | 100.0 % | 100.0 % |
| G0-topics1 | r4_false | 0.020 | 0.0013 | 0.020 | 100.0 % | 100.0 % |
