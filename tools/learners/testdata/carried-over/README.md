# The numbers the bench was carried over with

These tables are a small run of the bench as it was carried over from the research, before any scenario, rule or measure was added to it: the paper's seed and experiment, ten children a cell, two hundred answers each, computed on amd64 at commit `474da49` with `go run . -children 10 -answers 200`.

A run of the bench today must still give every row of them, in the order written here, with new rows allowed between them: what was added since must leave the carried-over cells, comparisons and summary as they were. Nothing rewrites these files; they change only when the service's own rule does, and then by hand, with the reason in the change.

**The service's step gained a floor** (2026-10-04, R187): its overall level's step stays at 0.05 from the 62nd answer. The rows of `shrinking/both` were taken from a run with the floor, by the same command — 233 cells, 32 comparisons and the service's row of the summary, whose numbers are now those of `floor_0.05/both`. The rule's measures read before the 62nd answer came out as they were, and every row of the other rules stayed as it was.

**The service masters a topic by a cautious estimate** (2026-10-04, R187): a topic is mastered once the level in it, less how far that level may be off, clears the middle task of one of its grade levels, rather than after a run of three right answers. The service's rule reads mastery when it chooses the next topic, and every rule of these tables runs under the service's mastery, so the rows of every rule were taken from a run with it, by the same command — 1,775 cells, 25 comparisons and the seven rows of the summary. The measures read within the first twenty answers, and those of the start, came out as they were.

## The service as it was

`earlier.csv` keeps the values of the service's cells as the bench was carried over with them, before the floor under its step and the cautious estimate of mastery: the 344 rows of `shrinking/both` in `cells.csv` as they stood at commit `55dd676`, renamed to `earlier/both`, without the ends of their intervals. The bench's earlier rule, the service's step with no floor and mastery by a run of three in the service's rule's place, must give every one of those values, in that order; that is what shows it is the service as it was. The intervals are left out because each is drawn from a stream named after its cell, and the earlier rule's cells have names of their own. Nothing rewrites this file. It was made by:

```
git show 55dd676:tools/learners/testdata/carried-over/cells.csv | awk -F, -v OFS=, 'NR==1{print $1,$2,$3,$4,$5; next} $1=="shrinking"{$1="earlier"; print $1,$2,$3,$4,$5}'
```
