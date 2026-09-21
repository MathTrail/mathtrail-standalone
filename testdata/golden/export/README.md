# Golden vectors from the prototype

The numbers in `testdata/golden/` are the prototype's, produced by the prototype's own code (RUN.md T16). The Go implementations of the rating formulas, the readability filter and the rule are checked against them, so **no Go test ever rewrites these files** (CLAUDE.md, "Tests"). They change only by running the export again.

## What is in each file

| File | Holds | Checked against it |
|---|---|---|
| `readability.json` | The Flesch–Kincaid grade, the sentence count and the longest sentence in words for all 450 reference questions, plus the eight anchors of the prototype's own `tests/test_filters.py` with their pass or fail at each grade in English and in Russian | T33, the readability check of SPEC 5.5 |
| `trgm_similarity.json` | PostgreSQL `similarity()` for fifteen hand-picked pairs, seven of them Cyrillic; the nearest neighbour of every one of the 450 questions; and the fifty most similar pairs in the corpus | T32, the near-duplicate check of SPEC 5.6 |
| `ratings.json` | The worked example of the prototype's `docs/architecture/05-ratings.md`, the same sequence with the separate `k0` of `config.yaml`, a corridor for every level from −2.0 to +3.5 in steps of 0.25, and the replayed ratings of the five seed profiles | T25, the ratings of SPEC 2 |
| `rule_briefs.json` | The rule's brief for each of the five seed profiles, with the profile it was computed from | T27, the rule of SPEC 3 |

Every file carries a `source` block naming what produced the numbers and what v1 changed. No file contains a timestamp, a path or anything in random order: a second run produces byte-identical files, and that is part of the acceptance of T16.

## How to run it

Everything runs in containers; nothing is installed into the devcontainer (CLAUDE.md, "Devcontainer only"). Python comes from the `uv` image pinned by tag **and** digest, PostgreSQL from the prototype's own `docker-compose.yml`, which pins its image the same way.

```sh
just golden
```

That recipe does four things, and they can be run by hand just as well:

```sh
# 1. PostgreSQL, from the prototype's compose file
cd prototype && docker compose up -d --wait

# 2-4. dependencies from uv.lock, the schema, the seed profiles, then the export
docker run --rm --network host \
  -v "$PWD:/repo" -w /repo/prototype --user "$(id -u):$(id -g)" \
  -e HOME=/tmp -e UV_PROJECT_ENVIRONMENT=/tmp/venv -e UV_CACHE_DIR=/tmp/uvcache \
  -e DATABASE_URL=postgresql://taskgen:taskgen@127.0.0.1:5432/taskgen \
  ghcr.io/astral-sh/uv:0.12.13-python3.12-trixie-slim@sha256:87bc72093c0aa93cc962bd7c0498ddf416dad3ce9e1434724e936e02b72afe5d \
  bash -lc "uv sync --frozen \
    && uv run python -m taskgen.apply_schema --force \
    && uv run python -m taskgen.seed \
    && uv run python /repo/testdata/golden/export/export_golden.py --out /repo/testdata/golden"
```

Three details that are not decoration:

- `UV_PROJECT_ENVIRONMENT=/tmp/venv` keeps the virtual environment out of `prototype/`, which is a read-only reference copy (CLAUDE.md, "Reference copies").
- `--user "$(id -u):$(id -g)"` keeps the written files owned by you rather than by root.
- `uv sync --frozen` forbids re-locking: the versions are `prototype/uv.lock` exactly.

`--force` on the schema step drops and recreates the prototype's tables. It touches only the container's database, never the product.

## How it was checked

- **Reproducible.** Two runs into different directories, compared byte for byte.
- **Against the prototype's own documents.** Every number of the worked example in `prototype/docs/architecture/05-ratings.md` appears in `ratings.json`: P of 0.7848, 0.6343 and 0.5383; θ of 0.0861, −0.1556 and −0.3513; the corridors [−1.294; −0.339], [−1.778; −0.822] and [−2.169; −1.213]; and the table of P by difficulty at θ = 0 — 0.905, 0.785, 0.600, 0.415, 0.295.
- **Against the prototype's own tests.** The Flesch–Kincaid values commented in `tests/test_filters.py` — −0.1, 5.4, 3.5, 4.1 and 11.5 — are what `readability.json` reports for the same anchors.

## What v1 does not take from these vectors

The prototype is not v1, and three of these numbers are history rather than a specification:

- **β updates.** `ratings.json` carries `beta_after` because the prototype updated task difficulty. v1 does not: every task is written for one child and never reused (SPEC 2.2).
- **The goal `motivate`.** It appears nowhere in `rule_briefs.json` as it happens, but the prototype could produce it; v1 has two goals (О-34, SPEC 3.1).
- **The setting rotation.** The prototype rotates the interests by the number of history rows; v1 rotates by the total answer count, because a bounded history window would make the row count go backwards (SPEC 3.2).
- **The third answer.** The prototype's `"?"` is an answer worth S = 0; in v1 "I don't understand" is a flag beside an answer, not an answer (SPEC 2.2, 04-profile).

Everything else — the P formula, the K decay, the corridor, the chess scale, the readability measurements, the similarity measure, the choice of topic, difficulty and traps — is the same in v1, which is why it is worth being checked against.

## What the export turned up

Two findings that belong to T32 rather than to this export:

1. **`similarity()` reaches 1.0 on genuinely different tasks.** `kl-34-d3-3` and `kl-34-d3-5` are different knights-and-liars problems with different answers, and their trigram sets are identical: the measure is a set, not a multiset, so two texts built from the same vocabulary in a different arrangement are indistinguishable to it. The same happens to `tri-12-d2-3` and `tri-12-d4-2`, which are different arithmetic questions over the same digits.
2. **207 of the 450 reference questions have a nearest neighbour at or above 0.6**, the prototype's threshold, with a median nearest neighbour of 0.57. Formulaic topics cluster hardest: the weighing-and-pouring questions sit at 0.96 to each other.

Neither breaks anything — a near-duplicate refusal costs one attempt, not a wrong answer — but both say that 0.6 is an aggressive threshold for a corpus this formulaic, and the histogram in `trgm_similarity.json` is the data to calibrate it against.
