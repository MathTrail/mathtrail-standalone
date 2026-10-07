# The criterion, read on this run

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. A rough look, not a decision: a decision runs 4000 children a cell, and a rough look reads not worse as no clear harm. The score is the share of the way from the service to the ceiling a rule closes, with its 95 % interval; the ceiling is no candidate, and is read to show the way.

## The choice

As a rough look reads it, which decides nothing.

Candidates that meet every constraint and are better than the service: 0.

No candidate meets every constraint and is better than the service, and no floor of the exit meets the constraints of not worse and of the screen: **the service's step stays.**

## Scores

| Rule | Score | Constraints met | Not met | Unread |
|---|---:|---:|---:|---:|
| shrinking/both (the service) | 0.000 [0.000, 0.000] | 46 of 49 | 3 | 0 |
| constant_slow/both | 0.100 [0.094, 0.106] | 43 of 49 | 6 | 0 |
| shrinking_ahead/both (the service, the next task written ahead) | -0.003 [-0.012, 0.006] | 44 of 49 | 5 | 0 |
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
| constant_slow/both | 0.417, not reached | 0.395, not reached | 0.327, not reached | 42.000, reached | 2.960, reached | 42.000, reached | 3.631, on the edge | 42.000, reached | 2.747, reached | 42.000, reached | 3.316, on the edge |
| shrinking_ahead/both (the service, the next task written ahead) | 0.531, not reached | 0.374, not reached | 0.563, not reached | 73.000, on the edge | 4.467, not reached | 41.000, reached | 1.563, reached | 73.000, reached | 4.160, not reached | 41.000, reached | 1.410, reached |
| oracle/both (the ceiling) | 0.000, reached | 0.611, reached | 0.000, reached | — | — | — | — | — | — | — | — |

## Not worse than the service

What of not worse each rule does not meet, or the run cannot read: its difference from the service, with its interval, and the tolerance it is read with.

| Rule | Checks met | Not met or unread |
|---|---:|---|
| shrinking/both (the service) | 38 of 38 |  |
| constant_slow/both | 35 of 38 | G5 r1_rms_200 0.029 [0.021, 0.037] against 0.018; G0-topics1 r1_rms_200 0.044 [0.036, 0.052] against 0.017; G0-topics1 r3_inside -0.020 [-0.023, -0.016] against 0.010 |
| shrinking_ahead/both (the service, the next task written ahead) | 38 of 38 |  |
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
