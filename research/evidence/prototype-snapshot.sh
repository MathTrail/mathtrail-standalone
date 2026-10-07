#!/usr/bin/env bash
# Compares the local copies of the prototype — prototype/ and docs/prototype/ —
# with one commit of its public repository, byte for byte, and records what
# a claim about the prototype may rest on:
#
#   evidence/prototype-snapshot.txt   every local file: its public path and
#                                     whether it matches the commit
#   evidence/prototype-files.sha256   the SHA-256 of every file at the commit
#   archive/                          a copy of every local file that does not
#                                     match, with archive/MANIFEST.sha256
#
# The local copies are deleted once the product has ported what it needs, so
# this runs while they exist; what it wrote stays, and so does the commit.
# Everything is computed in a temporary directory first and put in place only
# at the end, the archive's new copies just before the manifest that lists them.
#
# Usage: prototype-snapshot.sh <clone> <commit>   (run from research/)
set -euo pipefail

clone=${1:?usage: prototype-snapshot.sh <clone> <commit>}
commit=$(git --git-dir "$clone" rev-parse --verify "${2:?usage: prototype-snapshot.sh <clone> <commit>}^{commit}")
root=$(git rev-parse --show-toplevel)

# The copies are what a run compares. Once one is deleted there is nothing to
# compare it by, and a run would put a report of what is left in place of the
# evidence the copies once gave.
for copy in prototype docs/prototype; do
    if [[ ! -d "$root/$copy" ]]; then
        echo "prototype-snapshot: $copy/ is gone; the evidence it gave stands as it was written" >&2
        exit 1
    fi
done

# public_path names the file of the public repository a local copy stands for,
# or nothing when the file is not the prototype's own.
public_path() {
    case $1 in
        prototype/CLAUDE.prototype.md) echo CLAUDE.md ;; # renamed so the product's sessions ignore it
        prototype/*) echo "${1#prototype/}" ;;
        docs/prototype/SPEC.md) echo SPEC.md ;;
        docs/prototype/decisions.md) echo docs/decisions.md ;;
        docs/prototype/research/*) echo "${1#docs/prototype/}" ;;
        *) ;;
    esac
}

# The blob of every path at the commit, read once. -z keeps a path as it is,
# where git would otherwise quote one that holds a space or a non-ASCII letter.
declare -A blob_at
while IFS= read -r -d '' entry; do
    meta=${entry%%$'\t'*}
    blob_at[${entry#*$'\t'}]=${meta##* }
done < <(git --git-dir "$clone" ls-tree -r -z "$commit")

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/kept" "$work/commit"
: > "$work/report"

matching=0 total=0 kept=0
while IFS= read -r -d '' local; do
    total=$((total + 1))
    public=$(public_path "$local")
    # --no-filters: the product's .gitattributes must not normalise line
    # endings, or a copy that differs only in them would pass as a match.
    blob=$(git hash-object --no-filters "$root/$local")
    expected=${public:+${blob_at[$public]:-}}
    if [[ -n "$expected" && "$expected" == "$blob" ]]; then
        matching=$((matching + 1))
        printf '%s\t%s\tmatch\n' "$local" "$public" >> "$work/report"
        continue
    fi
    result=${expected:+differs}
    result=${result:-absent}
    # Python's byte-code cache is rebuilt from sources that do match, and no
    # claim can rest on it; anything else that does not match is kept.
    case $local in
        */__pycache__/*) printf '%s\t%s\t%s, byte-code cache, not kept\n' "$local" "${public:--}" "$result" >> "$work/report" ;;
        *)
            mkdir -p "$work/kept/$(dirname "$local")"
            cp "$root/$local" "$work/kept/$local"
            kept=$((kept + 1))
            printf '%s\t%s\t%s, kept in archive/\n' "$local" "${public:--}" "$result" >> "$work/report"
            ;;
    esac
done < <(cd "$root" && find prototype docs/prototype -type f -print0 | LC_ALL=C sort -z)

{
    echo "# Local copies of the prototype against its public repository, byte for byte."
    echo "repository=github.com/MathTrail/llm-taskgen-prototype"
    echo "commit=$commit"
    echo "local_files=$total"
    echo "matching=$matching"
    echo "kept_in_archive=$kept"
    echo "# local path<TAB>path at the commit<TAB>result"
    cat "$work/report"
} > "$work/snapshot"

# The SHA-256 of every file at the commit, from the commit itself.
git --git-dir "$clone" archive "$commit" | tar -x -C "$work/commit"
(cd "$work/commit" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum | sed 's|  \./|  |') > "$work/files"

chmod 644 "$work/snapshot" "$work/files"
mv "$work/snapshot" evidence/prototype-snapshot.txt
mv "$work/files" evidence/prototype-files.sha256

# The archive's manifest covers everything kept there, whoever kept it.
mkdir -p archive
cp -R "$work/kept/." archive/
(cd archive && find . -type f ! -name MANIFEST.sha256 -print0 | LC_ALL=C sort -z | xargs -0 -r sha256sum | sed 's|  \./|  |') > "$work/manifest"
chmod 644 "$work/manifest"
mv "$work/manifest" archive/MANIFEST.sha256

echo "$matching of $total local files match ${commit:0:12}; $kept kept in archive/"
