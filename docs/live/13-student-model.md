# The student model against live children (T72.14)

**T72.14.1, the first look: too few to read.** 2026-10-05. Read nineteen hours after the model of R187 rolled out, over the eighteen hours before, the service's log holds one child, one task accepted and one answer weighed, the child's answer in the range 6–20. Neither table the rule reads has a row. The floor under the step wants 20 children with answers in both ranges and an error of 0.010; the cautious mastery wants 20 children shown masteries and 100 settled by the tenth answer. At this pace no date can be given: the reading waits for families to use the service (T72.14.2).

> The repository is public, so this record carries counts and versions rather than identifiers: no account, child or instance is named.

## What this is

- **The question.** Remarks 68 and 69 of SPEC. The floor of 0.05 under the overall step and the cautious estimate of mastery (R187) were chosen on simulated children. Two tables of the service's report hold them to live ones (R199), by a rule written before any live number was read (`docs/learners.md`, "Reading the live numbers").
- **The model.** R187 came out in v0.2.6. No release since has changed the model. From v0.2.6 to v0.3.0, `internal/domain/rating` only renamed the corridor's bounds, while `internal/domain/tutor`, the mastery of `internal/domain/profile` and `tools/learners/testdata/guard.csv` stayed as they were.
- **The versions.** The instructions version changed with nearly every release. Each version below was computed from its release's content, as the service computes it:

  | Release | Rolled out (UTC) | Instructions version |
  |---|---|---|
  | v0.2.6 | 2026-10-04 23:37:52 | `7ed3c64f063d` |
  | v0.2.7 | 2026-10-05 03:08:32 | `7ed3c64f063d` |
  | v0.2.8 | 2026-10-05 14:17:38 | `38d142f805be` |
  | v0.3.0 | 2026-10-05 17:36:27 | `ce93a7cf05b5` |

- **The row read.** The rule reads the last row of both tables, `(every version)`, which R212 added on 2026-10-05, before any live number was read. A child's answers 6–50 and 101–200 fall under different versions, and until then the report paired a child within one version alone.
- **Roles.** The author ran `just report` over the production log, from a clean copy of T72.14.1's change, as the rule has it. Claude read the tables and wrote this record.
- **Cost.** $0: reading the log costs nothing.

## The rule, fixed before the data

As `docs/learners.md` gives it, which wins if the two ever differ:

| Check | Enough to read | Holds when |
|---|---|---|
| The floor under the step | at least 20 children and 300 answers in each range, and a standard error of the later less the earlier of at most 0.010 | the difference, two errors either way, meets the band from 0.006 to 0.077 |
| The cautious mastery | at least 20 children shown masteries, 100 masteries settled by the 10th answer, and a standard error of the share of at most 0.03 | the share taken back by the 10th answer, less two errors, is at most 0.394 |

The window starts after the rollout: `just report <N>d`, with N the whole days since 23:37:52 UTC on 2026-10-04 and at most 30, or `<H>h` in whole hours within the first day. It ends before the first release that changes the model. A look before enough has gathered records the counts and the errors alone; the difference and the share are read once, when there is enough.

## Readings

| Read on | Task | Lines (UTC) | Versions in the lines | Floor: children, answers earlier and later | Floor | Mastery: children, settled | Mastery |
|---|---|---|---|---|---|---|---|
| 2026-10-05 | T72.14.1 | 00:31:08 to 17:54:35 on 2026-10-05 | `38d142f805be` | 0 children | too few to read | 0 children, 0 settled | too few to read |

### The first look, 2026-10-05

- **The lines.** 538 lines of the service's, from 00:31:08 to 17:54:35 UTC on 2026-10-05, read at about 18:30 UTC with `just report 18h`.
- **What they hold.** One child, through Claude:
  - one task asked for, after two attempts: one refused for its structure, and one accepted under `38d142f805be`;
  - one answer weighed, in the range 6–20 of the child's answers, at a chance of 0.78–0.85.
- **The two tables.** "Later answers against earlier" and "Masteries taken back" both say "None in these lines". "How the estimate keeps up" holds the one answer, too few to read.
- **The other versions.** No line falls under `7ed3c64f063d` or `ce93a7cf05b5`. v0.3.0 had been out eighteen minutes when the lines end.
- **The verdict.** Too few to read, for both checks: no child with answers in both ranges, against the 20 the floor wants, and no mastery settled, against the 100 the mastery wants.

### When to look again

At this pace the rule gives no date, and the bench gives the scale. The bench reads the service's difference on a child who stays put to ±0.006 at a thousand children, a standard error near 0.003. An error of 0.010 then wants some ninety children past their 101st answer. Its share taken back reads to ±3.45 points, an error near 0.0176, so an error of 0.03 wants some three hundred and forty children's worth of masteries.

Twenty children past their 101st answer within the thirty days the log keeps would each have to answer three or four tasks a day for a month. The next look is worth taking once the report's table of children a day counts tens of children. Until then T72.14.2 waits, and a reading past thirty days (R199, *Open*) is likely to be needed.

## How to repeat it

```sh
# the window: whole days since the rollout, at most 30, or whole hours within the first day
H=$(( ($(date -u +%s) - $(date -u -d 2026-10-04T23:37:52Z +%s)) / 3600 ))
if [ "$H" -lt 24 ]; then window="${H}h"; else window="$(( H / 24 < 30 ? H / 24 : 30 ))d"; fi
just report "$window" > reading.md

# whether a release changed the model since v0.2.6: read the diff, renames aside
git diff v0.2.6 <release> -- internal/domain/rating internal/domain/tutor internal/domain/profile tools/learners/testdata/guard.csv
```

The rows to read are the last, `(every version)`, of "Later answers against earlier" and "Masteries taken back". The versions' rows of "How the estimate keeps up", laid side by side range by range, show how far the versions taken together lean the difference.
