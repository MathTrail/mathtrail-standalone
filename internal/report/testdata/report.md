# What the log adds up to

114 lines of the service's, from 2026-09-28 10:00:00 UTC to 2026-09-29 10:40:00 UTC. 2 more lines, not the service's, were left out.

## Tasks

Asked for counts the requests opened; handed in, the attempts judged, of which the checks refused some; out of attempts, the requests whose last attempt was refused.

| Instructions | Host | Asked for | Handed in | Refused | Out of attempts |
|---|---|---:|---:|---:|---:|
| 0a1b2c3d4e5f | chatgpt | 1 | 3 | 3 | 1 |
| 0a1b2c3d4e5f | claude | 1 | 2 | 1 | 0 |
| 88b63e22129c | (call not read) | 0 | 1 | 0 | 0 |
| 88b63e22129c | claude | 1 | 1 | 0 | 0 |

## Accepted tasks

The attempts an accepted task took, and the seconds from its request to its acceptance: the time the chat's model took to write it.

| Instructions | Host | Accepted | At the first attempt | Attempts, mean | Seconds, median | Seconds, 90th percentile | Seconds, longest |
|---|---|---:|---:|---:|---:|---:|---:|
| 0a1b2c3d4e5f | claude | 1 | 0 | 2.0 | 95 | 95 | 95 |
| 88b63e22129c | (call not read) | 1 | 1 | 1.0 | 70 | 70 | 70 |
| 88b63e22129c | claude | 1 | 1 | 1.0 | 40 | 40 | 40 |

## Why attempts were refused

An attempt is counted by one check, the first its refusal names, and fails that check and any others beside it.

| Instructions | Check | Counted by it | Failed it |
|---|---|---:|---:|
| 0a1b2c3d4e5f | solver_disagrees | 2 | 2 |
| 0a1b2c3d4e5f | bad_structure | 1 | 1 |
| 0a1b2c3d4e5f | readability | 1 | 2 |
| 0a1b2c3d4e5f | near_duplicate | 0 | 1 |

## Chances and what came of them

Answers to tasks the rule chose, after the trial series and without the hint, by the chance of a right answer each task was handed out at: how many there were, from how many children, the chance promised on average, and the share that came out right. A cell of fewer than 30 answers says too few. Left out of this table and the next are the answers to tasks the model chose (1), in the trial series (1), given with the hint (1) and on lines that carry no chance or no range of answers (1). Answers given with the hint, and tasks left unanswered — 2 in these lines, of every kind — are mostly ones that looked too hard, so the share right reads a little high by what they take away.

| Instructions | Host | Chance | Answers | Children | Promised | Came true |
|---|---|---|---:|---:|---:|---:|
| 88b63e22129c | claude | 0.70-0.77 | 3 | 3 | too few | too few |
| 88b63e22129c | claude | 0.78-0.85 | 32 | 4 | 0.80 | 0.81 |

## How the estimate keeps up

The same answers by which of the child's answers each was: a right answer as 1 and a wrong one as 0, less the chance promised, on average, with the standard error of that mean. Above zero the child did better than the estimate promised, which is how an estimate that falls behind a learning child shows; below zero, worse.

| Instructions | Host | Answer number | Answers | Children | Came true less promised | Standard error |
|---|---|---|---:|---:|---:|---:|
| 88b63e22129c | claude | 6-20 | 32 | 4 | +0.01 | 0.04 |
| 88b63e22129c | claude | 21-50 | 3 | 3 | too few | too few |

## Limits reached

A pace writes one line for a flood of refusals, and a day's ceiling one for every call it refused.

| Limit | Lines |
|---|---:|
| daily_tasks | 1 |
| ip_rate | 1 |
| user_rate | 1 |

## Tool calls

Milliseconds are the service's own time for a call, its calls to Drive included.

| Host | Tool | Calls | Answered | Refused | Failed | Invalid | Milliseconds, median | Milliseconds, 95th percentile |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| chatgpt | get_profile | 1 | 0 | 0 | 1 | 0 | 5000 | 5000 |
| chatgpt | next_task | 2 | 2 | 0 | 0 | 0 | 90 | 140 |
| chatgpt | other | 1 | 0 | 0 | 0 | 1 | 1 | 1 |
| chatgpt | submit_task | 3 | 0 | 3 | 0 | 0 | 310 | 320 |
| claude | next_task | 3 | 2 | 1 | 0 | 0 | 100 | 120 |
| claude | submit_answer | 39 | 39 | 0 | 0 | 0 | 80 | 80 |
| claude | submit_task | 3 | 2 | 1 | 0 | 0 | 250 | 300 |
