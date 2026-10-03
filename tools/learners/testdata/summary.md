# The student model on simulated children

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.518 [0.468, 0.563] | 46.3 % [42.7, 49.9] | 30.9 % [24.2, 36.5] | 74.1 % [65.7, 82.4] | -0.957 [-1.161, -0.749] | 70.0 % [40.0, 100.0] |
| constant/both | 0.578 [0.474, 0.699] | 43.3 % [40.4, 45.9] | 42.5 % [39.2, 45.7] | 78.4 % [68.9, 87.6] | -0.419 [-0.490, -0.346] | 10.0 % [0.0, 30.0] |
| constant_slow/both | 0.481 [0.431, 0.536] | 45.1 % [42.5, 47.4] | 37.1 % [31.2, 42.2] | 80.6 % [70.6, 89.4] | -0.667 [-0.780, -0.554] | 10.0 % [0.0, 30.0] |
| floor_0.05/both | 0.497 [0.456, 0.538] | 45.5 % [43.3, 48.2] | 32.5 % [25.8, 38.4] | 75.7 % [69.3, 83.3] | -0.837 [-1.008, -0.679] | 40.0 % [10.0, 70.0] |
| no_trial/both | 0.448 [0.396, 0.508] | 47.1 % [42.8, 51.5] | 28.9 % [22.7, 33.9] | 64.6 % [53.9, 76.8] | -1.050 [-1.273, -0.884] | 80.0 % [50.0, 100.0] |
| glicko2_floor/general | 0.466 [0.402, 0.530] | 47.1 % [45.0, 49.6] | 42.4 % [38.6, 46.2] | 81.7 % [68.0, 92.5] | -0.483 [-0.564, -0.399] | 20.0 % [0.0, 50.0] |
| glicko2_floor/topics | 0.852 [0.705, 1.007] | 37.2 % [32.0, 42.3] | 28.0 % [21.4, 34.1] | 51.1 % [36.5, 68.3] | -1.083 [-1.347, -0.867] | 60.0 % [30.0, 90.0] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 63.4 % [61.8, 64.9] | 60.1 % [57.2, 63.2] | 82.4 % [75.0, 90.6] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
