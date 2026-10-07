# The student model on simulated children

Seed 20261001, experiment E-A3: 1000 children a generator, 200 answers each. G0 is children who stay put, G2 children who learn, G3 children who jump once. Each number is read off all the children of its cell, with a 95 % interval from 2000 samples of them drawn again.

| Rule | Error after 200 answers, G0 | In the corridor, G0 | In the corridor, G2 | Mastered falsely, G0 | Estimate less level, G2 | Not caught up after the jump, G3 |
|---|---:|---:|---:|---:|---:|---:|
| shrinking/both (the service) | 0.484 [0.476, 0.492] | 42.8 % [42.3, 43.3] | 27.5 % [26.9, 28.1] | 3.6 % [3.1, 4.2] | -1.008 [-1.027, -0.988] | 56.8 % [53.8, 59.8] |
| constant_slow/both | 0.481 [0.473, 0.490] | 42.5 % [42.0, 43.0] | 32.4 % [31.8, 33.0] | 3.9 % [3.3, 4.5] | -0.757 [-0.774, -0.741] | 32.7 % [29.8, 35.7] |
| shrinking_ahead/both (the service, the next task written ahead) | 0.489 [0.481, 0.497] | 42.9 % [42.3, 43.4] | 27.7 % [27.1, 28.3] | 3.8 % [3.2, 4.4] | -1.002 [-1.021, -0.982] | 56.3 % [53.2, 59.2] |
| oracle/both (the ceiling) | 0.000 [0.000, 0.000] | 60.7 % [60.3, 61.1] | 60.6 % [60.3, 60.8] | 0.0 % [0.0, 0.0] | 0.000 [0.000, 0.000] | 0.0 % [0.0, 0.0] |
