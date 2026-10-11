# The student model on simulated children

Seed 20261001, experiment E-A3: 4000 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.484 [0.480, 0.488] | 42.8 % [42.6, 43.1] | 28.0 % [27.7, 28.3] | 3.7 % [3.4, 4.0] | -0.986 [-0.996, -0.976] | 56.8 % [55.2, 58.3] |
| constant_slow/both | 0.478 [0.474, 0.482] | 42.7 % [42.5, 43.0] | 32.6 % [32.3, 32.9] | 3.8 % [3.5, 4.1] | -0.749 [-0.758, -0.741] | 31.7 % [30.3, 33.2] |
| floor_0.05+cautious_z1/both | 0.484 [0.480, 0.488] | 42.8 % [42.6, 43.1] | 28.0 % [27.7, 28.4] | 3.7 % [3.4, 4.0] | -0.986 [-0.996, -0.975] | 56.8 % [55.2, 58.3] |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 0.515 [0.510, 0.520] | 40.3 % [40.0, 40.6] | 29.2 % [28.9, 29.5] | 7.4 % [6.8, 8.1] | -0.923 [-0.934, -0.912] | 52.2 % [50.7, 53.8] |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 0.485 [0.481, 0.489] | 41.3 % [41.0, 41.6] | 28.5 % [28.2, 28.8] | 6.4 % [5.9, 6.9] | -0.963 [-0.974, -0.953] | 58.5 % [56.9, 60.0] |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 0.481 [0.477, 0.485] | 41.4 % [41.1, 41.7] | 28.2 % [27.9, 28.5] | 6.1 % [5.6, 6.6] | -0.981 [-0.991, -0.971] | 61.7 % [60.1, 63.2] |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 0.592 [0.586, 0.597] | 38.7 % [38.4, 39.0] | 35.9 % [35.6, 36.3] | 10.5 % [9.9, 11.2] | -0.596 [-0.607, -0.585] | 12.2 % [11.2, 13.3] |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 0.510 [0.505, 0.515] | 40.1 % [39.8, 40.5] | 29.4 % [29.1, 29.7] | 1.0 % [0.9, 1.2] | -0.918 [-0.930, -0.906] | 52.5 % [51.0, 54.1] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 60.7 % [60.5, 60.8] | 60.5 % [60.3, 60.6] | 0.0 % [0.0, 0.0] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
