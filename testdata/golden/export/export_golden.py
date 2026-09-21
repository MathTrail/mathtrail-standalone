"""Export golden vectors from the prototype (RUN.md T16).

The Go implementations of the rating formulas, the readability filter, the near-duplicate measure and the rule are
checked against these files. They are the prototype's own numbers, produced by the prototype's own code: this script
imports `taskgen` and asks PostgreSQL for `similarity()` rather than reimplementing anything.

Run it from the repository root; see README.md next to this file. It writes four files into testdata/golden/ and
contains no timestamps, no paths and no random order, so a second run produces byte-identical files.
"""

import argparse
import importlib.util
import json
import os
import statistics
import sys
from pathlib import Path

import psycopg

from taskgen import ROOT as PROTOTYPE_ROOT
from taskgen import db, filters, rating, tutor_rule
from taskgen.validate_examples import load_examples

CORRIDOR_GRID_START, CORRIDOR_GRID_STOP, CORRIDOR_GRID_STEP = -2.0, 3.5, 0.25
SEED_STUDENTS = ("dima", "masha", "olya", "petya", "sasha")
TOP_PAIRS = 50

# The prototype's docs/architecture/05-ratings.md worked example predates D37 and uses one k0 for all three
# parameters. It is exported as it stands so that the numbers printed in that document can be checked against code.
DOC_EXAMPLE_K0 = 0.4
DOC_EXAMPLE_STEPS = (
    {"difficulty": 2, "correct": True},
    {"difficulty": 3, "correct": False},
    {"difficulty": 3, "correct": None},  # "?", the prototype's third outcome; v1 has no such answer
)

# Pairs for pg_trgm similarity(): the shapes a near-duplicate check meets, in Latin and in Cyrillic.
TRGM_PAIRS = (
    ("identical", "Four spaceships dock in pairs. How many different pairs can they make?",
     "Four spaceships dock in pairs. How many different pairs can they make?"),
    ("one_word_changed", "Four spaceships dock in pairs. How many different pairs can they make?",
     "Four spaceships dock in pairs. How many different pairs can the ships make?"),
    ("numbers_changed", "Four spaceships dock in pairs. How many different pairs can they make?",
     "Seven spaceships dock in pairs. How many different pairs can they make?"),
    ("same_idea_new_setting", "Four spaceships dock in pairs. How many different pairs can they make?",
     "Four kittens play in pairs. How many different pairs can they make?"),
    ("sentences_reordered", "Masha has 3 red balls and 2 blue balls. How many balls does she have?",
     "How many balls does Masha have? She has 3 red balls and 2 blue balls."),
    ("case_and_spacing", "Four spaceships dock in pairs.", "FOUR   SPACESHIPS   DOCK   IN   PAIRS."),
    ("unrelated", "Four spaceships dock in pairs. How many different pairs can they make?",
     "A clock strikes 3 times at three o'clock. How long does it strike at six o'clock?"),
    ("cyrillic_identical", "Четыре корабля стыкуются парами. Сколько разных пар получится?",
     "Четыре корабля стыкуются парами. Сколько разных пар получится?"),
    ("cyrillic_one_word_changed", "Четыре корабля стыкуются парами. Сколько разных пар получится?",
     "Четыре корабля стыкуются парами. Сколько разных пар выйдет?"),
    ("cyrillic_numbers_changed", "Четыре корабля стыкуются парами. Сколько разных пар получится?",
     "Семь кораблей стыкуются парами. Сколько разных пар получится?"),
    ("cyrillic_same_idea_new_setting", "Четыре корабля стыкуются парами. Сколько разных пар получится?",
     "Четыре котёнка играют парами. Сколько разных пар получится?"),
    ("cyrillic_declension", "У Маши три красных шара.", "У Маши было три красных шара."),
    ("cyrillic_unrelated", "Четыре корабля стыкуются парами. Сколько разных пар получится?",
     "Часы бьют три раза в три часа. Сколько они бьют в шесть часов?"),
    ("cyrillic_vs_latin", "Четыре корабля стыкуются парами.", "Four spaceships dock in pairs."),
    ("empty_vs_text", "", "Four spaceships dock in pairs."),
)


def anchors() -> list[dict]:
    """The readability and similarity anchors of the prototype's own tests, loaded from the test file itself.

    Copying the strings here would let them drift; the test module is the single source.
    """
    path = PROTOTYPE_ROOT / "tests" / "test_filters.py"
    spec = importlib.util.spec_from_file_location("prototype_test_filters", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    names = ("SIMPLE", "HARDER_WORDS", "LONG_SENTENCE", "VERY_LONG_SENTENCE", "HARD", "SPACESHIPS",
             "SPACESHIPS_AGAIN", "CLOCK")
    return [{"name": name, "text": getattr(module, name)} for name in names]


def measure(text: str, thresholds: filters.Thresholds) -> dict:
    """Readability measurements that do not depend on the student: the index, the sentences, the longest one."""
    result = filters.readability(text, grade=1, thresholds=thresholds)
    return {
        "fk_grade": result.fk_grade,
        "sentences": len(filters.sentences(text)),
        "longest_sentence_words": result.longest_sentence_words,
    }


def export_readability(thresholds: filters.Thresholds) -> dict:
    examples = [
        dict(
            id=task["id"],
            topic=task["topic"],
            grade_level=task["grade_level"],
            difficulty=task["difficulty"],
            **measure(task["question"], thresholds),
        )
        for task in load_examples()
    ]
    examples.sort(key=lambda row: row["id"])
    anchor_rows = []
    for anchor in anchors():
        row = {"name": anchor["name"], "text": anchor["text"], **measure(anchor["text"], thresholds)}
        row["ok_by_grade_en"] = {
            str(grade): filters.readability(anchor["text"], grade, thresholds).ok for grade in (1, 2, 3, 4)
        }
        row["ok_by_grade_ru"] = {
            str(grade): filters.readability(anchor["text"], grade, thresholds, language="ru").ok
            for grade in (1, 2, 3, 4)
        }
        anchor_rows.append(row)
    words = [row["longest_sentence_words"] for row in examples]
    indexes = [row["fk_grade"] for row in examples]
    return {
        "source": {
            "measured_by": "textstat.flesch_kincaid_grade and taskgen.filters.sentences",
            "thresholds": {
                "max_grade_margin": thresholds.max_grade_margin,
                "max_sentence_words": thresholds.max_sentence_words,
            },
            "note": "Flesch-Kincaid applies to English only (the prototype's D42); the sentence limit to every "
                    "language. The measurements here do not depend on the student's grade.",
        },
        "summary": {
            "examples": len(examples),
            "longest_sentence_words": {"min": min(words), "max": max(words),
                                       "median": statistics.median(words)},
            "fk_grade": {"min": min(indexes), "max": max(indexes), "median": statistics.median(indexes)},
        },
        "anchors": anchor_rows,
        "examples": examples,
    }


def export_trgm(conn: psycopg.Connection) -> dict:
    with conn.cursor() as cur:
        cur.execute("CREATE EXTENSION IF NOT EXISTS pg_trgm")
        conn.commit()
        cur.execute("SELECT current_setting('server_version'), extversion FROM pg_extension WHERE extname = 'pg_trgm'")
        server_version, extension_version = cur.fetchone()

        pairs = []
        for name, left, right in TRGM_PAIRS:
            cur.execute("SELECT similarity(%s, %s)", (left, right))
            pairs.append({"name": name, "a": left, "b": right, "similarity": cur.fetchone()[0]})

        examples = sorted(load_examples(), key=lambda task: task["id"])
        ids = [task["id"] for task in examples]
        questions = [task["question"] for task in examples]
        cur.execute(
            """
            WITH q(id, question) AS (SELECT unnest(%s::text[]), unnest(%s::text[]))
            SELECT a.id, b.id, similarity(a.question, b.question)
            FROM q a JOIN q b ON a.id < b.id
            WHERE similarity(a.question, b.question) > 0
            ORDER BY 3 DESC, 1, 2
            LIMIT %s
            """,
            (ids, questions, TOP_PAIRS),
        )
        top = [{"a": a, "b": b, "similarity": value} for a, b, value in cur.fetchall()]

        cur.execute(
            """
            WITH q(id, question) AS (SELECT unnest(%s::text[]), unnest(%s::text[]))
            SELECT a.id, max(similarity(a.question, b.question))
            FROM q a JOIN q b ON a.id <> b.id
            GROUP BY a.id
            ORDER BY 1
            """,
            (ids, questions),
        )
        nearest = [{"id": task_id, "similarity": value} for task_id, value in cur.fetchall()]

    buckets: dict[str, int] = {}
    for row in nearest:
        low = min(int(row["similarity"] * 20) / 20, 0.95)
        buckets[f"{low:.2f}-{low + 0.05:.2f}"] = buckets.get(f"{low:.2f}-{low + 0.05:.2f}", 0) + 1
    values = [row["similarity"] for row in nearest]
    return {
        "source": {
            "engine": "postgresql",
            "server_version": server_version,
            "pg_trgm_version": extension_version,
            "note": "similarity(a, b) is the Jaccard index of the two trigram sets after pg_trgm's own "
                    "normalisation: lowercase, non-alphanumerics to spaces, every word padded with two leading "
                    "spaces and one trailing space.",
        },
        "pairs": pairs,
        "examples_nearest_neighbour": {
            "count": len(nearest),
            "max": max(values),
            "median": statistics.median(values),
            "histogram": dict(sorted(buckets.items())),
        },
        "examples_top_pairs": top,
    }


def corridor_row(level: float, bounds: tuple[float, float]) -> dict:
    fit = rating.corridor(level, bounds)
    return {
        "level": round(level, 4),
        "beta_min": fit.beta_min,
        "beta_max": fit.beta_max,
        "probabilities": {str(difficulty): value for difficulty, value in fit.probabilities.items()},
        "inside": fit.inside,
        "recommended": fit.recommended,
        "fit": fit.fit,
        "elo": rating.elo(level),
    }


def replay_sequence(steps, params: rating.Params, bounds: tuple[float, float]) -> list[dict]:
    """One topic, one new task per step: theta, delta and beta after each answer, with the corridor that follows."""
    theta = delta = 0.0
    student_answers = topic_answers = 0
    rows = []
    for number, step in enumerate(steps, start=1):
        beta = rating.difficulty_to_beta(step["difficulty"])
        result = rating.update(
            theta,
            delta,
            beta,
            correct=step["correct"],
            student_answers=student_answers,
            topic_answers=topic_answers,
            task_answers=0,
            params=params,
        )
        rows.append({
            "step": number,
            "difficulty": step["difficulty"],
            "correct": step["correct"],
            "beta_before": beta,
            "theta_before": theta,
            "delta_before": delta,
            "probability": result.probability,
            "k_student": result.k_student,
            "k_topic": result.k_topic,
            "k_task": result.k_task,
            "theta_after": result.theta,
            "delta_after": result.delta,
            "beta_after": result.beta,
            "corridor_after": corridor_row(result.theta + result.delta, bounds),
        })
        theta, delta = result.theta, result.delta
        student_answers += 1
        topic_answers += 1
    return rows


def export_ratings(params: rating.Params, profiles: dict[str, dict]) -> dict:
    doc_params = rating.Params(DOC_EXAMPLE_K0, DOC_EXAMPLE_K0, DOC_EXAMPLE_K0, params.decay, params.corridor)
    grid, level = [], CORRIDOR_GRID_START
    while level <= CORRIDOR_GRID_STOP + 1e-9:
        grid.append(corridor_row(level, params.corridor))
        level += CORRIDOR_GRID_STEP

    replays = []
    for student in SEED_STUDENTS:
        ratings = rating.replay_history(profiles[student]["history"], params)
        replays.append({
            "student": student,
            "theta": ratings.theta,
            "answers": ratings.answers,
            "elo": rating.elo(ratings.theta),
            "topics": {
                topic: {"delta": delta, "answers": answers, "level": ratings.theta + delta,
                        "elo": rating.elo(ratings.theta + delta)}
                for topic, (delta, answers) in sorted(ratings.topics.items())
            },
        })

    return {
        "source": {
            "guess": rating.GUESS,
            "elo_scale": rating.ELO_SCALE,
            "config": {
                "k0_student": params.k0_student,
                "k0_topic": params.k0_topic,
                "k0_task": params.k0_task,
                "decay": params.decay,
                "corridor": list(params.corridor),
            },
            "note": "v1 keeps the theta and delta updates and the corridor unchanged, and stops updating beta: "
                    "every task is written for one child and never reused (SPEC 2.2). The beta_after column is "
                    "the prototype's and is exported for completeness.",
        },
        "doc_example": {
            "note": "The worked example of the prototype's docs/architecture/05-ratings.md: one k0 = 0.4 for all "
                    "three parameters, three new tasks of one topic, answers correct / wrong / '?'.",
            "params": {"k0_student": doc_params.k0_student, "k0_topic": doc_params.k0_topic,
                       "k0_task": doc_params.k0_task, "decay": doc_params.decay},
            "steps": replay_sequence(DOC_EXAMPLE_STEPS, doc_params, params.corridor),
        },
        "config_example": {
            "note": "The same three answers with the separate k0 of config.yaml (D37).",
            "steps": replay_sequence(DOC_EXAMPLE_STEPS, params, params.corridor),
        },
        "corridor_grid": grid,
        "seed_replays": replays,
    }


def export_rule(conn: psycopg.Connection, params: rating.Params) -> dict:
    topics = tutor_rule.load_catalog("topics")
    briefs = []
    for student_id in SEED_STUDENTS:
        student = db.load_student(conn, student_id, history_limit=None)
        if student is None:
            sys.exit(f"student {student_id!r} is not in the database: run the seed step first (see README.md)")
        briefs.append({
            "student": student_id,
            "profile": {
                "grade": student["grade"],
                "interests": list(student["interests"]),
                "excluded_skills": list(student["excluded_skills"]),
                "mastered_topics": sorted(student["mastered_topics"]),
                "consecutive_failures": student["consecutive_failures"],
                "history_rows": len(student["history"]),
                "theta": student["rating"],
                "topic_ratings": {
                    topic: {"delta": value["offset"], "answers": value["answers_count"]}
                    for topic, value in sorted(student["topic_ratings"].items())
                },
            },
            "brief": tutor_rule.make_brief(student, params, topics),
        })
    return {
        "source": {
            "note": "The prototype's rule (its SPEC 5.7, D40) on the five seed profiles. v1 changes three things "
                    "and these vectors do not cover them: the goal 'motivate' is gone (О-34), the setting rotates "
                    "by the total answer count rather than by the number of history rows, and grades 5-6 exist "
                    "(SPEC 3.3). Everything else — the goal, the topic, the difficulty and the traps — is the same.",
            "trap_count": tutor_rule.TRAP_COUNT,
        },
        "briefs": briefs,
    }


def write(path: Path, payload: dict) -> None:
    text = json.dumps(payload, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
    path.write_text(text, encoding="utf-8")
    print(f"{path.name}: {len(text.encode('utf-8')) / 1024:.1f} KB")


def main() -> None:
    parser = argparse.ArgumentParser(description="Export the prototype's golden vectors (RUN.md T16).")
    parser.add_argument("--out", type=Path, required=True, help="directory to write the vectors into")
    args = parser.parse_args()
    args.out.mkdir(parents=True, exist_ok=True)

    params = rating.load_params()
    thresholds = filters.load_thresholds()
    profiles = {
        path.stem: json.loads(path.read_text(encoding="utf-8"))
        for path in sorted((PROTOTYPE_ROOT / "data" / "seed").glob("*.json"))
    }

    write(args.out / "readability.json", export_readability(thresholds))
    write(args.out / "ratings.json", export_ratings(params, profiles))
    with psycopg.connect(os.environ.get("DATABASE_URL") or db.database_url()) as conn:
        write(args.out / "trgm_similarity.json", export_trgm(conn))
        write(args.out / "rule_briefs.json", export_rule(conn, params))


if __name__ == "__main__":
    main()
