# The student model on simulated children

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.458 [0.406, 0.500] | 47.0 % [44.9, 48.9] | 31.7 % [24.7, 37.6] | 2.8 % [0.0, 6.9] | -0.927 [-1.078, -0.798] | 30.0 % [0.0, 60.0] |
| constant/both | 0.549 [0.466, 0.648] | 45.5 % [42.8, 48.0] | 42.1 % [38.6, 46.0] | 6.4 % [1.1, 15.3] | -0.404 [-0.478, -0.322] | 20.0 % [0.0, 50.0] |
| constant_slow/both | 0.473 [0.431, 0.516] | 48.0 % [45.2, 50.9] | 36.9 % [30.4, 43.1] | 4.8 % [1.2, 8.9] | -0.597 [-0.702, -0.490] | 40.0 % [10.0, 70.0] |
| floor_0.05/both | 0.458 [0.406, 0.501] | 47.0 % [45.0, 48.8] | 31.7 % [24.6, 37.6] | 2.8 % [0.0, 7.1] | -0.927 [-1.084, -0.795] | 30.0 % [0.0, 60.0] |
| no_trial/both | 0.457 [0.388, 0.526] | 46.6 % [42.1, 51.3] | 25.8 % [19.7, 30.9] | 3.9 % [0.0, 8.9] | -1.131 [-1.357, -0.942] | 80.0 % [60.0, 100.0] |
| glicko2_floor/general | 0.507 [0.454, 0.568] | 45.0 % [42.8, 47.1] | 40.0 % [36.7, 43.2] | 9.0 % [2.2, 18.5] | -0.543 [-0.639, -0.443] | 50.0 % [20.0, 80.0] |
| glicko2_floor/topics | 0.765 [0.684, 0.871] | 34.1 % [30.9, 36.9] | 20.8 % [15.0, 26.2] | 8.1 % [3.4, 15.4] | -1.333 [-1.622, -1.051] | 90.0 % [70.0, 100.0] |
