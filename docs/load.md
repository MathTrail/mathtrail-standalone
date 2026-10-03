# Load

What one instance of the service takes, measured in a container of its size. The instance is sized by these numbers (R125); T64 measures the same things against the deployed service and compares them with this page.

## How it was measured, and how to repeat it

The load tool in `tools/load` (R124) starts the service from its image in a container of one vCPU and 1 GiB with no swap, since the platform gives none. The container is told what a deployment of that size is told: one solver slot (`MATHTRAIL_SOLVER_CONCURRENCY=1`) and nine tenths of its memory as the Go runtime's soft limit (`GOMEMLIMIT=921MiB`). The tool reads what the instance spends through its cgroup and writes what each run came to in Markdown.

- `just ci-load` runs every scenario against the image of the working tree, each held to the memory in the table at the end of this page. It is what the weekly workflow runs.
- `just load <scenario> [flags]` runs one scenario. For example, `just load adversarial -memory 512m` or `just load saturation -env MATHTRAIL_SOLVER_CONCURRENCY=4`.

The numbers below were taken on 2026-09-29 on the development machine: Docker inside the devcontainer, cgroups of the second version, kernel 7.1.5. A shared runner or a vCPU of the platform is slower than this machine, so times will differ there; what a run keeps in memory and whether an instance survives do not depend on the machine.

## The instance

| Setting | Value | Why |
|---|---|---|
| `cpu` | 1 vCPU | The free tier is counted in vCPU-seconds, and those run out first (below) |
| `memory` | 1 GiB | One solver's worth of memory (381 MiB at most), the collector's as much again, and the service beside them. Up to 2 GiB for each vCPU costs nothing inside the free tier, since vCPU-seconds run out before GiB-seconds do |
| `GOMEMLIMIT` | 921 MiB, nine tenths of the memory | The collector works harder as the heap nears this limit, so garbage a run has let go of does not outgrow the instance |
| Solver slots | 1, one for each vCPU | The clock is wall time. Two runs on one processor each spend it on the other's work |
| Steps | 25,000,000 | About ten times the costliest reference solver (2,413,617), and a limit of 381 MiB on what the built-ins of a run keep |
| Clock | 2 s | Most runs at the step ceiling take 65–400 ms on a processor of their own, and for them the clock is insurance. Sums and products of large numbers, which make a number at every step, are the runs the clock stops first |
| Wait for a slot | 3 s | Keeps the slowest `submit_task` within the platform's minute (R122) |
| `concurrency` | 80 requests | Most requests wait on Drive. The sandbox's queue is bounded in time, not in length |
| `max_instances` | 3 | Unchanged: a cost ceiling (О-24) |

## The price, and what it bounds

Everything the sandbox declares pays a step for every element it walks. It also pays a step for every sixteen bytes it builds and hands back, estimated from the sizes it was handed, with a quarter on top for the allocator's rounding (SPEC 6.5). At the step ceiling, then, the built-ins of one run can keep at most 25,000,000 × 16 bytes, which is 381 MiB. A test measures every built-in against the price: what a call keeps is at most sixteen bytes for every step it paid, and what it allocates on the way at most twice that.

What the language builds by itself is not priced: literals, operators, and the methods of values (see "The gap the price does not reach" below).

Under the price, the costliest reference solver is `pig-34-d4-2` at 2,413,617 steps. It was 2,081,362 under the old price of one step a value. No reference solver or template comes to a tenth of the ceiling.

## A lesson, and what it costs

One child's lesson, run five times: a task asked for, handed in and answered.

| | |
|---|---|
| Calls | 18, all answered. Median 4 ms, longest 10 ms. The development sign-in keeps profiles in memory, so no Drive is involved |
| Billed | 0.36 vCPU-seconds and 0.36 GiB-seconds per task |
| Spent, as the cgroup counted | 0.019 CPU-seconds per task |
| Peak memory | 14 MiB |
| The free tier a month | 500,000 tasks. vCPU-seconds run out first; GiB-seconds would last for 1,000,000 |

The tool counts only its own requests. A chat host sends more for every call it makes (SPEC remark 62), and T64 counts those.

These numbers were measured before T62.8. Since then a lesson asks for each task's package with a call of its own, and the card a task comes to asks how it stands, once on either side of the hand-in in the tool's lesson: six calls a task where there were three, each of the new ones a read (R152). In a host the card asks every four seconds while the task is written, about fifteen times for a minute's wait; T64 measures that against the deployed service.

## Slots and the clock

A run of `loop` that spends the whole budget takes 65 ms on this machine. A run that builds takes longer:

| Variant | Time per run |
|---|---|
| `tuples` | 214 ms |
| `pairs` | 236 ms |
| `sets` | 251 ms |
| `appends` | 393 ms |

`helper` and `product` are stopped by the clock.

To see what the clock does when runs share a processor, `helper` was sized to 5,000,000 steps. Alone it takes 620 ms. It was then handed in twice a second for twenty seconds:

| Slots | What the hand-ins got | What the runs came to |
|---|---|---|
| 1 | 13 examined, 26 told the sandbox was busy | 28 runs, all within their clock: 621 ms median, 710 ms longest |
| 4 | 39 examined, none told busy | 44 runs, 38 of them stopped by the clock at 2.03 s |

With four slots, a solver that needs 0.62 seconds is rejected as too slow, and in a real hand-in that costs the child an attempt. What stopped it was the work of its neighbours. With one slot the extra hand-ins are told the sandbox is busy instead: the task is not checked and no attempt is spent (SPEC 10.3).

## Eighty at once

`saturation` hands in the costly `loop` for requests nobody opened, so every hand-in costs both of its runs and nothing else.

**Through the pace of the instance.** Three hand-ins a second is about as many as the instance's limit of 200 messages a minute lets in. In thirty seconds, 90 hand-ins were all examined, with a median of 134 ms, and the one slot kept up.

**Past the pace**, with the instance's limit lifted: 80 children, 27 hand-ins a second, so that about eighty are in flight at once.

| Slots | Told busy | Examined | Longest answer | Median run | Peak |
|---|---|---|---|---|---|
| 1 | 685 of 810 (85 %), median 3.0 s | 125, median 6.1 s | 6.2 s | 68 ms | 36 MiB |
| 4 | 682 of 810 (84 %) | 128 | 6.9 s | 294 ms | 36 MiB |

Every call was answered and none was left without an answer. A hand-in that finds no slot is told so within its wait: 3 s for its first run, and about 6 s if its second run finds none either. That is what becomes of the seventy-sixth of eighty requests: whatever the platform lets in hears within seven seconds whether its task was checked. The chat hosts wait about a minute for a tool call before they give up. That comes from developers' reports, not from documentation: [Claude](https://github.com/anthropics/claude-code/issues/49910), [ChatGPT](https://community.openai.com/t/how-to-configure-long-mcp-tool-call-times-for-chatgpt-app/1379834), [Codex](https://developers.openai.com/codex/mcp).

## Memory, by costly solver

Each costly solver is handed in three times a second for twenty seconds, on an instance of its own, and spends 90–99 % of the step ceiling.

| Solver | 1 GiB, 1 slot, four runs | 512 MiB, 1 slot | 1 GiB, 4 slots | Before: 512 MiB, 4 slots, old price, no `GOMEMLIMIT` (T52a.4) |
|---|---|---|---|---|
| `appends` | 177–183 MiB | 188 MiB | 428 MiB | 436 MiB |
| `helper` | 83–92 MiB | 83 MiB | 270 MiB | survived; its peak was not recorded |
| `tuples` | 773–782 MiB | 443 MiB | killed for memory | killed for memory |
| `pairs` | 668–705 MiB | 437 MiB | killed for memory | killed for memory |
| `sets` | 411–580 MiB | 460 MiB | 942 MiB, and 4 runs stopped by the clock | not measured |
| `product` | 179–208 MiB | 189 MiB | 838 MiB | killed for memory |

All runs at 1 GiB with one slot were clean. The costliest solvers were told the sandbox was busy for a fifth to four fifths of their hand-ins, and every answer came within 6.7 s.

- **512 MiB with one slot survives too, but only just.** Its peaks sit at its soft limit of 460 MiB, 50 MiB short of the kill. The collector working that close to the limit makes the heaviest runs a fifth to a third slower: `tuples` 282 ms against 214, `pairs` 275 against 236. 1 GiB leaves room for the service's own requests beside a run, at no cost inside the free tier.
- **1 GiB with four slots does not survive.** Four runs keep more than an instance holds, so `tuples` and `pairs` end it. Four runs on one processor also stop `sets` and `pairs` at the clock, although each needs a quarter of a second alone.
- **Peaks near the soft limit are the collector's, not the run's.** `tuples` keeps at most 381 MiB, and the collector lets the heap grow towards 921 MiB before it works hard.
- **The soft limit is what keeps garbage inside the instance.** `product` keeps little and lets go of a great deal. Four of it at once on 512 MiB peak at 469 MiB with `GOMEMLIMIT`; with `-env GOMEMLIMIT=off`, the same run ends the instance for want of memory.

## The gap the price does not reach

`dicts` writes empty dictionaries in a loop, `[{} for i in range(n)]`. Each `{}` is one instruction and 512 bytes, so at the ceiling that is 2.6 million dictionaries, about 1.4 GB. On the instance of 1 GiB with one slot it was killed for want of memory at 1,008 MiB, and every call in flight went unanswered. A list literal of many dictionaries is worse. The operators of R45 and R53 (`[0] * n`, `xs = xs + xs`) and the methods of values (`x.extend(range(n))`) are the same kind of gap: memory the language takes by itself, one step at a time, where no price of the sandbox stands in front of it. `GOMEMLIMIT` does not close it, since it tells the collector when to work harder and refuses nothing.

The recorded fix is a solver in a process of its own with a memory limit on it (R45). Until it is made, a solver written to exhaust memory takes its instance down. The platform then starts another one, in about a quarter of a second (below), and the requests that were in flight on the old instance fail. `dicts` is therefore not part of the weekly check: it is handed in only when it is named (`just load adversarial -variants dicts`).

## The planted regression

The check has to fail when the numbers stop seeing each other. `just ci-load -env MATHTRAIL_SOLVER_CONCURRENCY=32` runs every scenario with thirty-two slots on one processor, and it fails:

- `tuples`, `pairs` and `sets` end the instance for want of memory, and the calls in flight on it go unanswered: 42 to 52 refused and 7 to 15 reset in each;
- `appends`, `helper` and `product` hold more than their ceilings;
- the lesson, both saturations, the limits and the cold starts stay clean, since they build nothing a slot more could multiply.

## Cold starts

Five starts of the service, each an instance of its own:

| | Median |
|---|---|
| From the container asked for to the probe's first answer | 254 ms |
| From the process started to the probe's first answer | 180 ms |
| From the container asked for to the first `get_profile` | 258 ms |
| Memory once it has answered | 12 MiB |

## The weekly check

`.github/workflows/load.yml` runs `just ci-load` on Mondays at 05:23 UTC against the image of what is on `main`. It also runs when started by hand, and on a pull request that changes the workflow itself. Every report goes into the run's summary.

The workflow fails on what the service should never do: a status of an error, a call nobody answered, a panic, an instance killed or ended by itself, a service that does not come back, or a peak past the ceiling its run is held to. Latencies and throughput fail nothing.

Each ceiling is the peak measured here, plus the larger of a quarter and 32 MiB, rounded up to 32 MiB. The solvers that keep hundreds of MiB are the exception. Their peak is wherever the collector chose to work, anywhere up to the soft limit: `sets` peaked at 411 MiB in one run and at 580 in another. So they are held just under the instance, at 960 MiB:

| Run | Peak measured | Ceiling |
|---|---|---|
| `lesson` | 14 MiB | 64m |
| `saturation` | 16 MiB | 64m |
| `saturation`, eighty in flight | 36 MiB | 96m |
| `limits` | 17–18 MiB | 64m |
| `cold` | 12–13 MiB | 64m |
| `adversarial` `appends` | 177–183 MiB | 256m |
| `adversarial` `helper` | 83–92 MiB | 128m |
| `adversarial` `tuples` | 773–782 MiB | 960m |
| `adversarial` `pairs` | 668–705 MiB | 960m |
| `adversarial` `sets` | 411–580 MiB | 960m |
| `adversarial` `product` | 179–208 MiB | 288m |

The ceilings of the small runs catch a regression in the service's own footprint, such as a leak or a cache that grows. For `tuples`, `pairs` and `sets`, it is the instance being killed for its memory that catches a regression.

## What is open

- **A hand-in for no open request is free to the account.** It spends no attempt, and the sandbox runs both of its runs before the profile is read. One account handing in solvers that spend their whole clock can keep the single slot of an instance busy, and the others on that instance then wait or are told it is busy; a freed slot goes to whoever waited longest. A limit on the sandbox time of an account, or a check of the request before the solver runs, would close this. It is not decided.
- **Time inside a single call.** A set of keys that hash alike is built in quadratic time inside one call, which the clock cannot stop. It is the same class as the long sorts R53 records.
- **The gap of the language**, above.
- **The runner.** The ceilings were measured on the development machine. The first runs of the workflow on a GitHub runner confirm them or move them.
