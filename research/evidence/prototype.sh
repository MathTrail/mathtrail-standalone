#!/usr/bin/env bash
# Facts about the prototype at one commit of its public repository, read from
# a local clone of it, never from the local copy in prototype/.
#
# Usage: prototype.sh <clone> <commit>   (prints key=value lines)
set -euo pipefail
# shellcheck source-path=SCRIPTDIR source=facts.sh
source "$(dirname "${BASH_SOURCE[0]}")/facts.sh"

clone=${1:?usage: prototype.sh <clone> <commit>}
commit=${2:?usage: prototype.sh <clone> <commit>}
unpack "$clone" "$commit"

# The public name a reader finds the commit under.
emit repository github.com/MathTrail/llm-taskgen-prototype
emit commit "${full:0:12}"
emit commit_date "$date"

# Catalogs: each file is one JSON array.
for catalog in topics traps skills; do
    n=$(jq length "data/catalogs/$catalog.json")
    emit "$catalog" "$n"
done

# Reference tasks: every examples file is an array of tasks, meant to hold five
# for each topic, grade level and difficulty.
n=$(jq -s 'map(length) | add' data/examples/*.json)
emit reference_tasks "$n"
cells=$(jq -s -r 'add | group_by([.topic, .grade_level, .difficulty])[] | length' data/examples/*.json)
n=$(printf '%s\n' "$cells" | count_matching .)
emit reference_cells "$n"
n=$(printf '%s\n' "$cells" | count_matching '^[^5]$|^..+$')
emit reference_cells_not_five "$n"
n=$(find tests/example_checks -name '*.py' ! -name '__init__.py' | count_matching .)
emit reference_check_modules "$n"

# The decision log, the research notes — files 00 to 13, of which 00 is their
# index and summary — the seed profiles, and the evaluation data the live runs
# left behind.
n=$(count_matching '^\*\*D[0-9]+\. ' < docs/decisions.md)
emit decisions "$n"
n=$(find research -name '[0-9][0-9]-*.md' | count_matching .)
emit research_notes "$n"
n=$(find data/seed -name '*.json' | count_matching .)
emit seed_profiles "$n"
n=$(jq length data/eval/live_scenarios.json)
emit live_scenarios "$n"
n=$(jq length data/eval/reviews.json)
emit eval_reviews "$n"
