# The package for grant applications

Everything an application for MathTrail draws on, in one place: what the product is, the rationale behind it, the numbers it may show, and the documents about data a school asks for. The calendar of applications, a year ahead, is section 8 of [usage-evidence-for-grants.md](../usage-evidence-for-grants.md), in Russian.

## MathTrail in brief

As of 5 October 2026.

- **What it is.** A free, open-source (MIT) app of olympiad-style maths for children in grades 1–6, used inside Claude and ChatGPT as an MCP app. An adult, a parent or a tutor, runs the lesson in their own chat, and the child answers on a card in it.
- **How a lesson works.**
  - The service picks the topic and the difficulty by a deterministic rule from the child's ratings: a learner model of the Elo and IRT family keeps practice in a band of likely success.
  - The chat's own model writes each task from the brief and reference examples the service hands it.
  - The service checks every task before the child sees it: its structure, a solver program that tries every option, the model's own check, readability, near-duplicates of earlier tasks, and the text drawing.
  - The card shows the task without the answer, which stays sealed until the child answers, and the hint comes only when asked for.
  - A wrong answer is explained from the trap behind the option chosen, the mistake of reasoning it stands for. A topic counts as mastered cautiously, and the mastery can be taken back.
- **The child's data.** The child is known by a pseudonym alone. The profile is one file in the adult's own Google Drive, and the service keeps no database. The counts kept for grant applications hold no identifier ([privacy policy](https://mathtrail.app/en/privacy/), [data inventory](data-inventory.md)).
- **Schools.** A pilot in a school means families using MathTrail at home on the school's recommendation: the school hands over no student data and receives none, and the free app has no school or class accounts.
- **Content.** 603 reference tasks in 17 topics, each with a worked solution and every wrong option labelled with one of 20 catalogued traps; 33 solver programs in Starlark; 25 skills.
- **Reach.** The card speaks 22 languages, the site English and Russian.
- **Status.** Live since the first release, v0.1.0, on 22 September 2026; the current release is v0.2.7, of 5 October 2026. The service answers at `mcp.mathtrail.app` and the site is [mathtrail.app](https://mathtrail.app). Use has been counted since 4 October 2026, and the first month with public numbers can be November 2026.
- **Public goods.** The code, the reference tasks with their solvers and trap labels, and the SQL of the views that publish the counts, all under MIT; and the public report of the counts, once it is published.
- **Contact.** [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).

## What the package holds

| Document | What it answers | State |
|---|---|---|
| This page | What MathTrail is, where each question of an application is answered | Written |
| Calendar, [usage-evidence-for-grants.md](../usage-evidence-for-grants.md) §8 | Which applications fall when, and who may make them | Written, checked 5 October 2026 |
| [logic-model.md](logic-model.md) | The rationale at ESSA Tier 4: the problem, resources, activities, outputs, outcomes and impacts, and the plan of a study of the effects | Written; one of its seven components waits for a work to be read |
| [sources.md](sources.md) | The empirical works the logic model rests on, each read, with the passage relied on | Five works read, and a sixth beside the first; a seventh waits for a copy |
| `metrics-<year>-<month>.md` | The public numbers of each closed month | None yet: the first can be made on 1 December 2026, if November had ten children or more ([how](#the-monthly-numbers)) |
| [data-inventory.md](data-inventory.md) | What the service keeps, where, for how long and who can read it, what the chat's model sees, and what is done when a key leaks | Written |
| [sdpc-ndpa-v2.md](sdpc-ndpa-v2.md) | Answers to the SDPC National Data Privacy Agreement v2.2, for a school whose families use MathTrail at home | Written; the rows marked for the maintainer are open |
| [ferpa-letter.md](ferpa-letter.md) | A letter on FERPA for such a school | Written, not signed |

## The monthly numbers

On the first day of each month, after 02:00 UTC, once the nightly count has taken in the month just closed, an owner of the project with `bq` signed in writes the numbers to a file outside the repository:

```bash
just impact public 1 > ~/metrics.md
```

The command prints a title, a line on what the public views show, a table with the latest month they show, and a line saying whether every day so far was counted. The public views show a month only once it is over, every day of it counted, and only if it had ten children or more. If the table's row names the month just closed, the file is moved to `docs/grants/metrics-<year>-<month>.md` as it is. If the row names an earlier month, if there is no row, or if the command failed, the file is deleted: the month has no public numbers. The commit that adds a file dates it.

## Where an application's questions are answered

### Tools Competition 2027, first phase

The abstract is due on 13 October 2026. Its form asks for the prize level and a description of the entrant, in about a thousand words, and the organizers assess fit with the competition, alignment with the track's objectives, novelty and quality ([how proposals are evaluated](https://tools-competition.org/how-are-proposals-evaluated/)).

| What the abstract needs | Where in the package |
|---|---|
| The problem, and who it serves | Logic model: the problem |
| The tool, and what is new in it | MathTrail in brief |
| Why it belongs in Reimagining Assessment | MathTrail in brief: every wrong option a labelled trap, and the traps a child keeps falling for counted as they happen; logic model |
| Evidence behind the design | Logic model, sources |
| Users and reach | MathTrail in brief: status, reach; the monthly numbers from December 2026 |
| The public goods it creates | MathTrail in brief: public goods |
| How the child's data is protected | Privacy policy; [data inventory](data-inventory.md) |
| The entrant and the team | Not in the package |

The second phase, due on 21 January 2027, is scored by a rubric the organizers send to those invited on 24 November 2026; its map is added here then. The map of an ED/IES SBIR proposal is added once the solicitation of the 2027 fiscal year is published.

## What the package leaves out, on purpose

- The texts of the applications themselves.
- Budgets, the team's CVs, letters of support, registrations such as SAM.gov, and a plan of commercialization.
- Any number the public views do not show: an application quotes the public numbers alone.
