#!/usr/bin/env bash
# The commit the claims of the service's real use are proven at. Real lessons
# began after the commit paper A describes, so these claims are pinned to a
# later commit of their own, and their proofs are lines of it alone: the stats
# name the commit and its date, and nothing is counted.
#
# Usage: use.sh <commit>   (prints key=value lines)
set -euo pipefail
# shellcheck source-path=SCRIPTDIR source=facts.sh
source "$(dirname "${BASH_SOURCE[0]}")/facts.sh"

commit=${1:?usage: use.sh <commit>}
full=$(git rev-parse --verify "$commit^{commit}")
date=$(git show -s --format=%cs "$full")

# The public name a reader finds the commit under.
emit repository github.com/MathTrail/mathtrail-standalone
emit commit "${full:0:12}"
emit commit_date "$date"
