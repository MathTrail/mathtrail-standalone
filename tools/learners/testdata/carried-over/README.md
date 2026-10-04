# The numbers the bench was carried over with

These tables are a small run of the bench as it was carried over from the research, before any scenario, rule or measure was added to it: the paper's seed and experiment, ten children a cell, two hundred answers each, computed on amd64 at commit `474da49` with `go run . -children 10 -answers 200`.

A run of the bench today must still give every row of them, in the order written here, with new rows allowed between them: what was added since must leave the carried-over cells, comparisons and summary as they were. Nothing rewrites these files; they change only when the service's own rule does, and then by hand, with the reason in the change.
