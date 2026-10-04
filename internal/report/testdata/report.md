# What the log adds up to

135 lines of the service's, from 2026-09-28 10:00:00 UTC to 2026-09-30 08:16:30 UTC. 2 more lines, not the service's, were left out.

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
| user_rate | 2 |

## Tool calls

Milliseconds are the service's own time for a call, its calls to Drive included; without Drive, the same less the time its calls to Drive took, tied to the call by the request they were made in.

| Host | Tool | Calls | Answered | Refused | Failed | Invalid | Milliseconds, median | Milliseconds, 95th percentile | Without Drive, median | Without Drive, 95th percentile |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| chatgpt | get_profile | 1 | 0 | 0 | 1 | 0 | 5000 | 5000 | 5000 | 5000 |
| chatgpt | next_task | 2 | 2 | 0 | 0 | 0 | 90 | 140 | 90 | 140 |
| chatgpt | other | 1 | 0 | 0 | 0 | 1 | 1 | 1 | 1 | 1 |
| chatgpt | submit_task | 3 | 0 | 3 | 0 | 0 | 310 | 320 | 310 | 320 |
| claude | get_profile | 1 | 1 | 0 | 0 | 0 | 5 | 5 | 5 | 5 |
| claude | get_progress | 1 | 1 | 0 | 0 | 0 | 400 | 400 | 70 | 70 |
| claude | next_task | 3 | 2 | 1 | 0 | 0 | 100 | 120 | 100 | 120 |
| claude | read_task | 2 | 2 | 0 | 0 | 0 | 70 | 90 | 20 | 30 |
| claude | submit_answer | 39 | 39 | 0 | 0 | 0 | 80 | 80 | 80 | 80 |
| claude | submit_task | 3 | 2 | 1 | 0 | 0 | 250 | 300 | 250 | 300 |

## Traces

A request's trace is kept or dropped as the request arrives. The spans of a kept trace are delivered before the request ends, and the measurements with whichever request finds them due, so a delivery that failed lost the spans of a kept trace, or the measurements. A delivery that ran out of time is past the deadline it is given. 0.250 of the 4 requests that carried a trace kept it. The telemetry was built to keep 0.1. Deliveries that failed for a request whose own line does not tell whether it kept its trace — a line not read, or an id requests of both kinds came with: 1, past the deadline 1. Lines that say the telemetry failed on its own, beside a delivery: 1.

| Trace | Requests | Deliveries failed | Of them past the deadline |
|---|---:|---:|---:|
| kept | 1 | 1 | 1 |
| dropped | 3 | 1 | 0 |

## The busiest minute

The most a minute by the clock held, the minute a pace is counted over. An account's pace counts every message it sends the MCP endpoint, and a tool call is the one a line names the account on; an instance's counts every request it is sent, at the MCP endpoint and at the sign-in apart. The lines of requests name 2 instances. The MCP endpoint was sent 0.07 requests for each tool call.

| What | Most in a minute |
|---|---:|
| Tool calls of one account | 3 |
| Requests to the MCP endpoint on one instance | 2 |
| Other requests on one instance | 1 |
| Requests to the MCP endpoint on all instances | 3 |

## Children, a day at a time

The children the counts for grant applications are taken from, by the rules those counts follow: a child is counted under the name it has that month, on each day a task was handed to it; a child the load tool or MCP Inspector handed a task to is left out, and so is every line of its; a line the log repeated is counted once; a day is a day in UTC. The same days of `just impact private` hold the same numbers.

| Day | Children | Tasks | Answers | Topics won |
|---|---:|---:|---:|---:|
| 2026-09-29 | 2 | 2 | 1 | 1 |

## The rules of the log

Every line of the service's is held to an event the service is decided to write, to the fields decided for that event, to the form each field of the lines the children are counted from may hold, and to carrying nothing shaped like an email address. A line that breaks one is named by its event and its field, never by what it held: an event the table does not name is not named, since its words may be anything.

| Event | Field | Rule | Lines |
|---|---|---|---:|
| (an event not in the table) | (none) | an event the service is not decided to write | 1 |
| cimd_fetch | host | text shaped like an email address | 1 |
| limit_hit | (withheld) | a field its event is not decided to carry | 1 |
| limit_hit | (withheld) | text shaped like an email address | 1 |
| task_accepted | country | a value its field may not hold | 1 |
| tool_call | pseudonym | a field its event is not decided to carry | 1 |
