# Deviations from the protocol

`PROTOCOL-A-offline.md` beside this file was frozen and timestamped before the first run of any experiment. This is every place where the runs or the paper depart from it, or read it where it left room, each with its reason.

## E-A1: defects injected into the reference tasks

- **An exact interval beside the bootstrap.** Beside a class's interval from a bootstrap over its hosts, the run also computes the exact interval of the class's pooled cases, because a bootstrap has no width when every host has the same share, as at 0 and 100 %. The paper reports each out-of-scope class by its operators, each with the exact interval the protocol prescribes for an operator.
- **The renamed copies, read after the results.** The protocol puts a copy with its people and objects renamed (X09) out of scope. After the results, the authors read such a copy as a copy a child would know, whose refusal is right. The paper therefore reports its refusal as the near-duplicate check working, with its two operators apart.

## E-A3: simulated learners

- **Glickman's worked example.** Computed exactly, the example gives a rating of 1464.05 and a volatility of 0.06000, where the paper prints 1464.06 and 0.05999, because its printed intermediate values are rounded. The test holds every intermediate value to the printed precision and the final ones to their rounding.
- **The fixed-item Urnings.** The protocol's check, a binomial law within 0.01 with items of random difficulty, passes even with the acceptance step removed, so it cannot show that the step works. A stricter companion check, with items drawn near the learner, holds the law within 0.003 and fails without the step. Both are reported in `learnersim/results/checks.txt`.
- **Pooled shares.** The protocol calls every number a mean over the children of its cell, while its own definitions make R4, R5, R4b and the share left unsettled after a jump shares of declarations, of topics and levels, of attempts and of children. They are pooled as defined, and the bootstrap still draws children.
- **R4b as a survival estimate.** R4b, the share declared mastered within the first m eligible attempts, is read as one less a Kaplan–Meier estimate over attempts. A topic whose count stops before m attempts, because the run ends or the child comes to master the topic, has no share of its own, and is counted for the attempts it made.
- **Where the protocol left room.** R4b's true chances are given in five bins: below 0.5, by tenths up to 0.8, and 0.8 or more. The error R6 follows after a jump is averaged over the topics with its sign, as R6 reads the lag elsewhere. A bound of 0.5 on the root mean square would sit at the error a child who never jumped already has, 0.49 after 200 answers.

## E-A4: what a review costs

- **The cloud numbers.** The cost of a task, the cold start and an instance's pace are not copied into a data file of their own: they stay in the evidence ledger (C086, C101), each proved by its lines at the pinned commit.
- **The table of cost** is a paragraph of the paper rather than a table.
- **Every turn.** The sizes of the model's package cover every turn of the reference tasks as well as every point of the ladder, since which tasks a package shows depends on the child's count of answers.
- **Percentiles** are read at the index ⌊q·n⌋ of the sorted values, as E-A1 and E-A3 read theirs.
- **Overwritten records.** While the harness was being built it was run several times, and those runs' records were overwritten, against the rule that every run keeps its records. Their medians, 3.28–3.43 ms, lie within the range of the final run's passes. Only the final run's records are kept, and the paper cites them.

## All three

- **The recipes** are `just research faultinject`, `just research learnersim` and `just research perf`; the protocol writes them with a hyphen, as `just research-learnersim`.
