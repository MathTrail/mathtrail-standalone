#!/usr/bin/env bash
# The version of paper A a release names: the first twelve hexadecimal digits
# of a hash of what the release would hold and of the paper's text. What it
# holds is the named and the anonymous artifact and the paper as a review reads
# it, which prints no version of its own and so is the same at every commit of
# the same sources; the text is the tree of paper-a/ as git holds it, since
# the authors the anonymous paper leaves out change none of those files.
#
# Usage: version.sh TREE   (from research/, after the artifacts and the paper
# are built; TREE is the tree of paper-a/ at the commit, as git rev-parse
# HEAD:research/paper-a gives it)
set -euo pipefail

tree=${1:?usage: version.sh TREE}
{
    sha256sum <release/build/paper-a-artifact.tar.gz
    sha256sum <release/build/paper-a-artifact-anonymous.tar.gz
    sha256sum <paper-a/build/paper-a-anonymous.pdf
    printf '%s\n' "$tree"
} | sha256sum | cut -c1-12
