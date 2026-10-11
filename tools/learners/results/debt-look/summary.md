# The student model on simulated children

Seed 20261001, experiment sweep: 300 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.487 [0.472, 0.502] | 43.1 % [42.0, 44.1] | 28.2 % [27.0, 29.4] | 4.5 % [3.2, 6.1] | -0.980 [-1.021, -0.942] | 53.0 % [47.3, 58.3] |
| constant_slow/both | 0.480 [0.467, 0.495] | 43.0 % [42.1, 43.9] | 32.3 % [31.1, 33.4] | 5.0 % [3.5, 6.5] | -0.755 [-0.790, -0.721] | 34.0 % [29.0, 39.3] |
| floor_0.05+cautious_z1/both | 0.487 [0.471, 0.503] | 43.1 % [42.1, 44.3] | 28.2 % [26.9, 29.4] | 4.5 % [3.2, 6.0] | -0.980 [-1.019, -0.942] | 53.0 % [47.7, 58.7] |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1/both | 0.522 [0.503, 0.543] | 40.6 % [39.4, 41.9] | 28.9 % [27.8, 30.1] | 9.1 % [6.7, 11.7] | -0.939 [-0.984, -0.895] | 48.7 % [43.0, 54.3] |
| floor_0.05+debt_0.2_0.4_p0.8+cautious_z1/both | 0.485 [0.469, 0.502] | 41.8 % [40.6, 43.0] | 28.3 % [27.2, 29.5] | 7.0 % [4.8, 9.4] | -0.969 [-1.011, -0.929] | 52.3 % [46.7, 58.0] |
| floor_0.05+debt_0.1_0.4_p1+cautious_z1/both | 0.482 [0.467, 0.498] | 41.9 % [40.7, 43.1] | 27.9 % [26.8, 29.0] | 6.5 % [4.5, 9.0] | -0.999 [-1.038, -0.960] | 58.0 % [52.3, 63.3] |
| floor_0.05+debt_0.1_0.4_p0.8_x2+cautious_z1/both | 0.604 [0.584, 0.625] | 38.7 % [37.5, 39.8] | 35.5 % [34.3, 36.7] | 13.4 % [10.7, 16.1] | -0.600 [-0.643, -0.557] | 10.7 % [7.3, 14.3] |
| floor_0.05+debt_0.1_0.4_p0.8+cautious_z1_settled/both | 0.514 [0.498, 0.533] | 40.7 % [39.4, 42.0] | 28.7 % [27.5, 29.9] | 1.0 % [0.5, 1.5] | -0.937 [-0.983, -0.893] | 48.7 % [43.0, 54.3] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 60.6 % [59.9, 61.2] | 60.7 % [60.3, 61.1] | 0.0 % [0.0, 0.0] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
