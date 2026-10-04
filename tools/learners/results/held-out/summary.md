# The student model on simulated children

Seed 20261003, experiment held-out: 4000 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.494 [0.490, 0.498] | 42.6 % [42.3, 42.9] | 25.9 % [25.6, 26.2] | 64.1 % [63.3, 64.9] | -1.109 [-1.120, -1.098] | 72.8 % [71.4, 74.1] |
| constant_slow/both | 0.483 [0.479, 0.487] | 42.8 % [42.6, 43.1] | 32.7 % [32.4, 33.0] | 65.4 % [64.6, 66.2] | -0.751 [-0.760, -0.743] | 30.9 % [29.4, 32.3] |
| floor_0.05/both | 0.486 [0.483, 0.490] | 42.7 % [42.5, 43.0] | 28.1 % [27.8, 28.4] | 65.0 % [64.2, 65.8] | -0.983 [-0.993, -0.973] | 56.6 % [55.1, 58.2] |
| floor_0.05+cautious_z1/both | 0.491 [0.487, 0.496] | 42.8 % [42.6, 43.1] | 27.8 % [27.5, 28.1] | 3.8 % [3.5, 4.1] | -0.993 [-1.002, -0.982] | 58.7 % [57.1, 60.2] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 61.4 % [61.2, 61.6] | 60.4 % [60.2, 60.5] | 74.3 % [73.5, 75.1] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
