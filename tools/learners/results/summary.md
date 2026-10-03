# The student model on simulated children

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.495 [0.486, 0.503] | 42.4 % [41.9, 43.0] | 25.7 % [25.1, 26.3] | 63.2 % [61.5, 65.0] | -1.122 [-1.143, -1.100] | 71.1 % [68.3, 73.8] |
| constant/both | 0.593 [0.581, 0.606] | 41.4 % [41.0, 41.8] | 37.7 % [37.3, 38.1] | 67.9 % [66.5, 69.3] | -0.479 [-0.494, -0.464] | 9.1 % [7.4, 10.9] |
| constant_slow/both | 0.481 [0.472, 0.489] | 42.7 % [42.2, 43.3] | 32.5 % [31.9, 33.1] | 65.6 % [64.0, 67.3] | -0.764 [-0.781, -0.748] | 31.9 % [29.0, 35.0] |
| floor_0.05/both | 0.491 [0.483, 0.499] | 42.4 % [41.9, 43.0] | 27.8 % [27.1, 28.4] | 64.7 % [63.0, 66.3] | -0.997 [-1.018, -0.978] | 55.6 % [52.5, 58.6] |
| no_trial/both | 0.470 [0.462, 0.478] | 43.5 % [42.9, 44.0] | 26.4 % [25.7, 27.0] | 61.5 % [59.9, 63.2] | -1.115 [-1.134, -1.096] | 72.7 % [69.8, 75.5] |
| glicko2_floor/general | 0.594 [0.582, 0.606] | 43.0 % [42.5, 43.5] | 35.9 % [35.4, 36.3] | 63.6 % [61.6, 65.7] | -0.625 [-0.641, -0.610] | 32.8 % [29.8, 35.8] |
| glicko2_floor/topics | 0.834 [0.816, 0.851] | 38.2 % [37.7, 38.6] | 24.7 % [24.2, 25.1] | 53.9 % [52.0, 55.7] | -1.129 [-1.149, -1.107] | 73.6 % [70.8, 76.4] |
