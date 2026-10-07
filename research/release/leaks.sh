#!/usr/bin/env bash
# The anonymous artifact holds no commit of the repositories it was made from,
# not even one that no branch, tag or pull request reaches any more but GitHub
# still serves by its hash. Every word of 7 to 40 hexadecimal digits with a
# letter among them, in the unpacked artifact, binary files included, is looked
# up in each repository through GitHub's API, a hundred words a request, and
# the artifact is refused when one names a commit, though not when it names a
# file. A word of digits alone reads as a number, as the artifact's assembly
# reads it, which refuses one that begins a commit the research names. The
# hashes the artifact masked, all zeros, are not looked up either.
#
# Usage: leaks.sh DIRECTORY OWNER/REPOSITORY...   (needs gh, signed in)
set -euo pipefail

usage="usage: leaks.sh DIRECTORY OWNER/REPOSITORY..."
dir=${1:?$usage}
shift
if [[ $# -eq 0 ]]; then
    echo "$usage" >&2
    exit 2
fi
if [[ ! -d "$dir" ]]; then
    echo "leaks.sh: $dir is no directory" >&2
    exit 2
fi

# grep says 1 when it finds nothing, and more when it cannot read a file: a
# file left unread is not a file that names no commit.
status=0
hits=$(grep -rhoaP '(?<![0-9A-Fa-f])[0-9A-Fa-f]{7,40}(?![0-9A-Fa-f])' "$dir") || status=$?
if [[ $status -gt 1 ]]; then
    echo "leaks.sh: the files under $dir cannot all be read" >&2
    exit 1
fi
mapfile -t words < <(printf '%s\n' "$hits" | tr 'A-F' 'a-f' | grep -x -E '[0-9a-f]{7,40}' | grep '[a-f]' | LC_ALL=C sort -u)
# Every artifact holds hashes of files, so a directory with no word of the
# kind is not the artifact.
if [[ ${#words[@]} -eq 0 ]]; then
    echo "leaks.sh: $dir holds no hexadecimal word with a letter, which no artifact does" >&2
    exit 1
fi

found=()
for repo in "$@"; do
    owner=${repo%%/*} name=${repo#*/}
    for ((i = 0; i < ${#words[@]}; i += 100)); do
        query="{ r: repository(owner: \"$owner\", name: \"$name\") {"
        for ((j = i; j < i + 100 && j < ${#words[@]}; j++)); do
            query+=" w$j: object(expression: \"${words[j]}\") { ... on Commit { oid } }"
        done
        query+=" } }"
        # A request that fails, or an answer that is not the one asked for,
        # stops the check: a word left unread is not a word that names nothing.
        answer=$(gh api graphql -f query="$query")
        if ! named=$(jq -r '.data.r | to_entries[] | select(.value.oid != null) | .key' <<<"$answer"); then
            echo "leaks.sh: GitHub's answer about $repo is not the one asked for" >&2
            exit 1
        fi
        while read -r alias; do
            if [[ -n "$alias" ]]; then found+=("$repo ${words[${alias#w}]}"); fi
        done <<<"$named"
    done
done

if [[ ${#found[@]} -gt 0 ]]; then
    printf '%s\n' "${found[@]}" >&2
    echo "leaks.sh: the words above name commits, and the artifact would lead to its repositories" >&2
    exit 1
fi
echo "leaks.sh: none of ${#words[@]} hexadecimal words in $dir names a commit of $*"
