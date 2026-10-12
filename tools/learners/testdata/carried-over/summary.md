# The student model on simulated children

Seed 20261001, experiment E-A3: 10 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.474 [0.420, 0.524] | 45.9 % [43.0, 48.6] | 31.9 % [25.8, 37.1] | 3.3 % [0.0, 7.1] | -0.901 [-1.082, -0.742] | 40.0 % [10.0, 70.0] |
| constant/both | 0.533 [0.461, 0.623] | 43.9 % [42.0, 46.2] | 41.2 % [38.6, 44.0] | 6.2 % [2.6, 10.3] | -0.368 [-0.479, -0.264] | 10.0 % [0.0, 30.0] |
| constant_slow/both | 0.502 [0.441, 0.571] | 44.1 % [40.6, 47.0] | 36.8 % [31.4, 41.8] | 3.1 % [0.0, 7.2] | -0.626 [-0.734, -0.526] | 10.0 % [0.0, 30.0] |
| floor_0.05/both | 0.474 [0.419, 0.524] | 45.9 % [43.1, 48.7] | 31.9 % [25.7, 37.2] | 3.3 % [0.0, 7.3] | -0.901 [-1.093, -0.747] | 40.0 % [10.0, 70.0] |
| no_trial/both | 0.431 [0.373, 0.481] | 47.5 % [43.5, 51.6] | 28.3 % [21.6, 34.3] | 2.2 % [0.0, 4.8] | -1.077 [-1.287, -0.886] | 50.0 % [20.0, 80.0] |
| glicko2_floor/general | 0.507 [0.443, 0.568] | 47.8 % [44.7, 50.4] | 39.5 % [35.4, 43.6] | 7.5 % [1.9, 15.2] | -0.475 [-0.554, -0.392] | 10.0 % [0.0, 30.0] |
| glicko2_floor/topics | 0.650 [0.546, 0.744] | 38.6 % [36.2, 40.5] | 21.7 % [16.8, 26.5] | 9.1 % [3.8, 15.5] | -1.276 [-1.533, -1.032] | 60.0 % [30.0, 90.0] |
