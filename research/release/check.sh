#!/usr/bin/env bash
# Shows that the artifact reproduces the results of paper A on its own: the
# tarball is unpacked into an empty directory, each experiment is run there
# from the artifact's own sources, and what it writes is compared byte for byte
# with the results the artifact ships — every file of E-A1 and E-A3, and the
# package sizes of E-A4, whose times belong to the machine that measured them.
#
# Usage: check.sh named|anonymous   (from research/, after assemble.sh)
set -euo pipefail

flavour=${1:?usage: check.sh named|anonymous}
case $flavour in
named) name=paper-a-artifact ;;
anonymous) name=paper-a-artifact-anonymous ;;
*)
    echo "check.sh: the flavour is named or anonymous, not $flavour" >&2
    exit 2
    ;;
esac
tarball=$PWD/release/build/$name.tar.gz
if [ ! -f "$tarball" ]; then
    echo "check.sh: $tarball is missing; assemble the artifact first" >&2
    exit 1
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
tar -xzf "$tarball" -C "$work"
cd "$work/$name/service/research"
# The artifact is no git checkout, and no workspace outside it may stand in
# for its modules.
export GOWORK=off GOFLAGS=-buildvcs=false

differ=0
# same compares the named files of a run with those the artifact ships.
same() {
    local shipped=$1 run=$2 base
    shift 2
    for base in "$@"; do
        if ! cmp -s "$shipped/$base" "$run/$base"; then
            echo "check.sh: $shipped/$base differs between the artifact and the run" >&2
            differ=1
        fi
    done
}
# files lists the results of a directory, provenance aside: it names the
# commit and the machine's toolchain, not a result.
files() {
    find "$@" -maxdepth 1 -type f ! -name provenance.txt -printf '%f\n' | LC_ALL=C sort -u
}

shipped=experiments/faultinject/results
mkdir -p "$work/ea1"
# The verdicts of the case-by-case reading are recorded by a reader, not made
# by the run; the run reads them from where it writes.
cp "$shipped/reading.csv" "$work/ea1/"
go run ./experiments/faultinject -out "$work/ea1"
mapfile -t names < <(files "$shipped" "$work/ea1")
same "$shipped" "$work/ea1" "${names[@]}"

shipped=experiments/learnersim/results
mkdir -p "$work/ea3"
go run ./experiments/learnersim -out "$work/ea3"
mapfile -t names < <(files "$shipped" "$work/ea3")
same "$shipped" "$work/ea3" "${names[@]}"

mkdir -p "$work/ea4"
go run ./experiments/perf -out "$work/ea4"
same experiments/perf/results "$work/ea4" packages.csv

if [ "$differ" -ne 0 ]; then exit 1; fi
echo "the $flavour artifact reproduces E-A1 and E-A3 byte for byte, and E-A4's package sizes"
