#!/usr/bin/env bash
# Runs one experiment of paper A into an empty directory and compares what it
# writes, byte for byte, with the results the paper's numbers come from: every
# file of E-A1 and E-A3, and of E-A4 the package sizes and the numbers drawn
# from them, since its times belong to the machine that measured them.
#
# Usage: reproduce.sh faultinject|learnersim|perf <shipped> <run> <command>...
#   <shipped>  the results the run is compared with
#   <run>      an empty or missing directory the experiment writes into
#   <command>  the experiment's program; -out <run> is added to it
set -euo pipefail

usage="usage: reproduce.sh faultinject|learnersim|perf <shipped> <run> <command>..."
if [[ $# -lt 4 ]]; then
    echo "$usage" >&2
    exit 2
fi
experiment=$1 shipped=$2 run=$3
shift 3
case $experiment in
faultinject | learnersim | perf) ;;
*)
    echo "reproduce.sh: there is no experiment $experiment; $usage" >&2
    exit 2
    ;;
esac
# What an earlier run left would be compared as if this run had written it.
if [[ -e "$run" && -n "$(ls -A "$run")" ]]; then
    echo "reproduce.sh: $run is not empty" >&2
    exit 2
fi
mkdir -p "$run"
# The verdicts of the case-by-case reading are recorded by a reader, not made
# by the run; the run reads them from where it writes.
if [[ "$experiment" == faultinject ]]; then cp "$shipped/reading.csv" "$run/"; fi
"$@" -out "$run"

differ=()
# same compares a file the run wrote with the one shipped, by name.
same() {
    local file=$1
    if ! cmp -s "$shipped/$file" "$run/$file"; then differ+=("$file"); fi
}
case $experiment in
perf)
    same packages.csv
    # The sizes the paper prints are numbers drawn from packages.csv, kept
    # beside the times in numbers.txt.
    if ! grep -q '^package' "$shipped/numbers.txt"; then
        echo "reproduce.sh: $shipped/numbers.txt holds no package sizes to compare" >&2
        exit 1
    fi
    if ! cmp -s <(grep '^package' "$shipped/numbers.txt") <(grep '^package' "$run/numbers.txt" || true); then
        differ+=("numbers.txt (its package sizes)")
    fi
    what="its package sizes"
    ;;
*)
    # provenance.txt names the commit and the toolchain of a run, not a result.
    mapfile -t names < <(find "$shipped" "$run" -maxdepth 1 -type f ! -name provenance.txt -printf '%f\n' | LC_ALL=C sort -u)
    for name in "${names[@]}"; do same "$name"; done
    what="every file of its results"
    ;;
esac
if [[ ${#differ[@]} -ne 0 ]]; then
    echo "reproduce.sh: $experiment wrote other bytes than $shipped holds, in:" >&2
    printf '  %s\n' "${differ[@]}" >&2
    exit 1
fi
echo "reproduce.sh: $experiment reproduces $what"
