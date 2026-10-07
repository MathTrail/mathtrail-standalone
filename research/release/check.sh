#!/usr/bin/env bash
# Shows that the artifact reproduces the results of paper A on its own: the
# tarball is unpacked into an empty directory, each experiment is run there
# from the artifact's own sources, and what it writes is compared byte for byte
# with the results the artifact ships — every file of E-A1 and E-A3, and the
# package sizes of E-A4, whose times belong to the machine that measured them.
# The comparison is experiments/reproduce.sh, the one the repository's own
# results are held to.
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
if [[ ! -f "$tarball" ]]; then
    echo "check.sh: $tarball is missing; assemble the artifact first" >&2
    exit 1
fi

reproduce=$PWD/experiments/reproduce.sh
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
tar -xzf "$tarball" -C "$work"
cd "$work/$name/service/research"
# The artifact is no git checkout, and no workspace outside it may stand in
# for its modules.
export GOWORK=off GOFLAGS=-buildvcs=false

differ=0
for experiment in faultinject learnersim perf; do
    if ! "$reproduce" "$experiment" "experiments/$experiment/results" "$work/$experiment" go run "./experiments/$experiment"; then
        differ=1
    fi
done
if [[ "$differ" -ne 0 ]]; then exit 1; fi
echo "the $flavour artifact reproduces E-A1 and E-A3 byte for byte, and E-A4's package sizes"
