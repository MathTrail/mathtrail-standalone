# shellcheck shell=bash
# What the scripts that compute facts at a commit share. Sourced, not run.
#
# Every value is computed into a variable before it is printed: a command that
# fails inside an assignment stops the script, where the same command inside
# the argument of echo would leave a blank value behind.

# unpack extracts a commit of a git repository into a temporary directory that
# is removed on exit, and moves there. It sets full (the commit's whole hash)
# and date (its commit date), and nothing it reads comes from a working tree.
unpack() {
    local git_dir=$1 commit=$2
    full=$(git --git-dir "$git_dir" rev-parse --verify "$commit^{commit}")
    # shellcheck disable=SC2034 # read by the scripts that source this file
    date=$(git --git-dir "$git_dir" show -s --format=%cs "$full")
    work=$(mktemp -d)
    trap 'rm -rf "$work"' EXIT
    git --git-dir "$git_dir" archive "$full" | tar -x -C "$work"
    cd "$work" || exit 1
}

# emit prints one fact, and refuses one that came out empty.
emit() {
    if [ -z "$2" ]; then
        echo "${0##*/}: no value for $1" >&2
        exit 1
    fi
    printf '%s=%s\n' "$1" "$2"
}

# count_matching counts the lines of its input that match; no line matching is
# a count of zero, not a failure.
count_matching() { grep -cE -- "$1" || true; }
