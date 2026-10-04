# The student model on simulated children

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.484 [0.476, 0.492] | 42.8 % [42.3, 43.3] | 27.5 % [26.9, 28.1] | 3.6 % [3.1, 4.2] | -1.008 [-1.027, -0.988] | 56.8 % [53.8, 59.8] |
| constant/both | 0.601 [0.590, 0.613] | 41.6 % [41.2, 42.0] | 37.8 % [37.4, 38.2] | 6.8 % [6.1, 7.5] | -0.464 [-0.480, -0.448] | 10.8 % [9.0, 12.7] |
| constant_slow/both | 0.481 [0.473, 0.490] | 42.5 % [42.0, 43.0] | 32.4 % [31.8, 33.0] | 3.9 % [3.3, 4.5] | -0.757 [-0.774, -0.741] | 32.7 % [29.8, 35.7] |
| floor_0.05/both | 0.484 [0.476, 0.492] | 42.8 % [42.3, 43.4] | 27.5 % [26.9, 28.1] | 3.6 % [3.1, 4.2] | -1.008 [-1.028, -0.989] | 56.8 % [53.7, 59.8] |
| no_trial/both | 0.469 [0.461, 0.477] | 44.2 % [43.6, 44.7] | 26.2 % [25.6, 26.8] | 2.9 % [2.4, 3.4] | -1.126 [-1.146, -1.107] | 73.0 % [70.3, 75.7] |
| glicko2_floor/general | 0.568 [0.557, 0.579] | 42.6 % [42.1, 43.0] | 34.5 % [34.0, 35.0] | 4.3 % [3.6, 5.0] | -0.652 [-0.667, -0.638] | 30.1 % [27.2, 33.0] |
| glicko2_floor/topics | 0.894 [0.878, 0.911] | 40.1 % [39.5, 40.6] | 22.4 % [21.9, 22.9] | 10.5 % [9.8, 11.3] | -1.242 [-1.265, -1.220] | 66.5 % [63.5, 69.5] |
| floor_0.05+cautious_z1/both | 0.484 [0.475, 0.492] | 42.8 % [42.3, 43.4] | 27.5 % [26.9, 28.1] | 3.6 % [3.1, 4.2] | -1.008 [-1.028, -0.989] | 56.8 % [53.8, 60.0] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 60.7 % [60.3, 61.1] | 60.6 % [60.3, 60.8] | 0.0 % [0.0, 0.0] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
