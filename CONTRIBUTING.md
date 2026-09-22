# Contributing

Thank you for looking. This is a small project with one maintainer, and it is built for children, so it is worth saying plainly what helps and what will be turned down — before you spend an evening on something.

## Before you write anything

**Open an issue first** for anything beyond a typo. The order of the work is fixed: [`RUN.md`](RUN.md) lists the tasks in the order they are done, and what the product is meant to be is settled in [`PRODUCT-V1.md`](PRODUCT-V1.md). A change that does not fit either is not rejected because it is bad — it is rejected because the plan is what keeps this finishable by one person.

If you are reporting a bug, an issue with what you did, what happened and what you expected is already a real contribution. Reports from actual use with a child are the most valuable thing this project can receive.

**A security fault is not an issue.** It goes through the private channel in [`SECURITY.md`](SECURITY.md).

## What is especially welcome

- **Bugs found in real use**, in any chat host, in any language.
- **Translations.** The interface is looked up by locale, and a language read by a native speaker is worth more than any amount of machine translation.
- **Reference tasks** for grades 5–6, the part of the content still being written. The rules for these are strict and are below.
- **A wording that is clearer to a child** than what is there now, in any of the texts the child or the parent reads.

## What will probably be declined

- A new topic, trap or skill outside the catalogs. They are closed on purpose: a child's history is only comparable with itself while the names stay the same.
- A new dependency. Every one of them has to be read, licensed compatibly with MIT, and kept current by one person.
- A feature that is not in `PRODUCT-V1.md`. The document has a section for what is deliberately out.
- A change that makes the service store something about a child. It stores nothing, and that is a design decision rather than an unfinished piece.

## Setting up

Everything happens inside the development container. The host needs Docker and VS Code and nothing else:

```
git clone https://github.com/MathTrail/mathtrail-standalone
code mathtrail-standalone      # then: Reopen in Container
```

Inside it, `just --list` shows every command. The ones you need:

```
just fmt        # format
just test       # the tests
just ci-lint    # what the checks run: formatting and the linter
just ci-test    # what the checks run: the tests, with the race detector and coverage
```

## What a change has to satisfy

The conventions are written down in [`CLAUDE.md`](CLAUDE.md) — it is addressed to an AI assistant, but it is the house style for everybody, and a reviewer will hold a change to it.

The short version:

- **`just ci-lint` and `just ci-test` are green.** They are what the pull request checks run, so a failure here is a failure there.
- **Tests come with the change.** A pull request that drops below 90% coverage of the lines it touches cannot be merged. A new check is not considered covered until its test has been seen to fail with the check switched off.
- **English everywhere in the repository**: code, every comment, configuration and documents.
- **Nothing personal reaches a log.** Log lines carry aggregates — never a task, an answer, a pseudonym or a profile.
- **The answer stays hidden** until the child has answered: not in a widget payload, not in an open field of the profile, not in a log.
- **No secrets in the repository.** The whole history is scanned on every pull request.

One commit or several, as suits the change; the title of the pull request is what a reader sees in the history, so make it a sentence about what changed.

## If you are contributing content

A reference task is few-shot material for the model that writes the child's task, so a weak one is copied thousands of times:

- **The wording is your own.** Ideas from open sources are fine — a wording lifted from a book or an olympiad is not, whatever its age.
- **Five options, all different, exactly one correct**, and every wrong one names a trap from the catalog and explains, in the child's words, what mistake leads there.
- **A solver.** Every task carries a short Starlark program that brute-forces it and confirms that exactly one option is right. A task nobody can check by enumeration does not belong in the catalog.
- **Nothing a child of that grade has not met.** The excluded skills exist for that, and they apply to the traps as much as to the wording.

The format, field by field, is in `SPEC.md`, section 1.5, and every file in `content/examples/` is an example of it.

## Licensing, and the name

The code and the content are MIT, and a contribution arrives under the same licence — by opening a pull request you are offering it on those terms.

**The name is not part of the licence.** MathTrail is the name of this service; a fork is free to take the code and the content, and is asked to run under a name of its own. That is about telling two things apart for the families using them, not about ownership.
