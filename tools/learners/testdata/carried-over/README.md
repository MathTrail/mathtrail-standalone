# The numbers the bench was carried over with

These tables are a small run of the bench as it was carried over from the research, before any scenario, rule or measure was added to it: the paper's seed and experiment, ten children a cell, two hundred answers each, computed on amd64 at commit `474da49` with `go run . -children 10 -answers 200`.

A run of the bench today must still give every row of them, in the order written here, with new rows allowed between them: what was added since must leave the carried-over cells, comparisons and summary as they were. Nothing rewrites these files; they change only when the service's own rule does, and then by hand, with the reason in the change.

**The service's step gained a floor** (2026-10-04, R187): its overall level's step stays at 0.05 from the 62nd answer. The rows of `shrinking/both` were taken from a run with the floor, by the same command — 233 cells, 32 comparisons and the service's row of the summary, whose numbers are now those of `floor_0.05/both`. The rule's measures read before the 62nd answer came out as they were, and every row of the other rules stayed as it was.
