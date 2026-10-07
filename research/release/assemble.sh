#!/usr/bin/env bash
# The artifact of paper A: the product at the ledger's pin, which holds the
# reference tasks with their solvers, and the research the paper's numbers come
# from — the experiments with their results, the protocol with its timestamp
# and its deviations, the evidence ledger and every number the paper prints.
#
# The anonymous flavour is the one a double-blind review may link to. It keeps
# of the product only the packages the experiments build on, replaces every name
# and commit that would lead to the system, its repositories or its authors, and
# fails if one is left anywhere in it.
#
# Usage: assemble.sh named|anonymous   (from research/; prints the tarball)
set -euo pipefail

flavour=${1:?usage: assemble.sh named|anonymous}
case $flavour in
named) name=paper-a-artifact ;;
anonymous) name=paper-a-artifact-anonymous ;;
*)
    echo "assemble.sh: the flavour is named or anonymous, not $flavour" >&2
    exit 2
    ;;
esac

prototype_clone=.cache/llm-taskgen-prototype.git
if [[ "$flavour" == anonymous && ! -d "$prototype_clone" ]]; then
    # Without the clone the prototype's commits could not be told from other
    # hexadecimal words, and would ship unmasked.
    echo "assemble.sh: the anonymous artifact needs the prototype's clone: run just research prototype-fetch" >&2
    exit 1
fi

pin=$(sed -n 's/^commit=//p' evidence/product-stats.txt)
full=$(git -C .. rev-parse --verify "$pin^{commit}")
module=$(go list -m)
product_module=${module%/research}

build=release/build
out=$build/$name
# The service's tree stays as the commit holds it, research beside its code as
# in the repository, and the artifact's own README above both.
tree=$out/service
rm -rf "$out" "$build/$name.tar.gz"
mkdir -p "$tree"

# The product, as the pinned commit holds it. The research plan the commit
# carries is the plan of that day, in Russian, and not part of the artifact.
product=(.)
if [[ "$flavour" == anonymous ]]; then
    # Only the packages the experiments and their tests import: the rest of the
    # product — its server, site, cloud project and documents — names its makers
    # in ways no replacement can be trusted to catch. A failing go list must not
    # leave a tree with no code, so it runs on its own before its output is read.
    deps=$(go list -deps -test -f '{{with .Module}}{{.Path}}{{end}} {{.ImportPath}}' ./experiments/...)
    mapfile -t product < <(awk -v m="$product_module" '$1 == m && $2 != m { print substr($2, length(m) + 2) }' <<<"$deps" | LC_ALL=C sort -u)
    if [[ "${#product[@]}" -eq 0 ]]; then
        echo "assemble.sh: go list found no package of the product the experiments import" >&2
        exit 1
    fi
    packages=$(printf "\`service/%s/\`, " "${product[@]}")
    packages=${packages%, }
    product+=(go.mod go.sum LICENSE THIRD_PARTY_LICENSES)
fi
git -C .. archive "$full" -- "${product[@]}" | tar -x -C "$tree"
rm -f "$tree/research/RUN.md"
rmdir "$tree/research" 2>/dev/null || true

# The research the paper's numbers come from.
research=(
    go.mod go.sum
    experiments/PROTOCOL-A-offline.md experiments/PROTOCOL-A-offline.md.ots
    experiments/faultinject experiments/learnersim experiments/perf experiments/reviewing
    evidence/ledger.md evidence/claims-product.json evidence/product-stats.txt evidence/ledger-numbers.txt
    evidence/prototype.md evidence/claims-prototype.json evidence/prototype-stats.txt
    evidence/facts.sh evidence/product.sh evidence/prototype.sh
    literature/numbers.txt
)
mkdir -p "$tree/research/paper-a"
tar -c "${research[@]}" | tar -x -C "$tree/research"
# The module requires the product at the pin's version; in the artifact that
# version sits beside it, and the module builds from it with no network.
(cd "$tree/research" && go mod edit -replace="$product_module=../")
cp paper-a/generated/numbers.tex "$tree/research/paper-a/numbers.tex"
cp release/DEVIATIONS.md "$tree/research/experiments/DEVIATIONS.md"
cp release/README.md "$out/README.md"
git -C .. show "$full:LICENSE" >"$tree/research/LICENSE"

# The research files are copied from the working tree rather than from git, so
# nothing has kept a stray file — a test binary, a profile, an editor's
# backup — out of what was just copied.
stray=$(find "$tree/research" -type f ! \( -name '*.go' -o -name '*.md' -o -name '*.ots' -o -name '*.json' \
    -o -name '*.txt' -o -name '*.csv' -o -name '*.tex' -o -name '*.sh' -o -name go.mod -o -name go.sum -o -name LICENSE \))
if [[ -n "$stray" ]]; then
    echo "assemble.sh: files of no kind the artifact ships:" >&2
    echo "$stray" >&2
    exit 1
fi

# The modules the experiments link beside the product's own code, as the
# product lists its own: licence, module, and the licence's text at the
# version built.
tool=$(just --justfile ../justfile --evaluate GO_LICENSES)
{
    printf 'Third-party licenses of the research module\n'
    printf '===========================================\n\n'
    printf 'The research code is MIT; see LICENSE. Beside the service'"'"'s own code\n'
    printf '(MIT, with its own third-party list at the root of the service), the\n'
    printf 'experiments link the Go modules below and the standard library of the Go\n'
    printf 'release that builds them, listed as std.\n\n'
    go_version=$(go list -m -f '{{.GoVersion}}')
    printf '%-13s %-45s %s\n' BSD-3-Clause std "https://github.com/golang/go/blob/go$go_version/LICENSE"
    go run "$tool" report ./experiments/faultinject ./experiments/learnersim ./experiments/perf \
        --ignore "$module" --ignore "$product_module" 2>/dev/null |
        LC_ALL=C sort | awk -F, '{ printf "%-13s %-45s %s\n", $3, $1, $2 }'
} >"$tree/research/THIRD_PARTY_LICENSES"

# commits_in prints, lower-cased, every run of 7 to 40 hexadecimal characters
# in the files under a directory, binary ones included, that names a commit of
# one of the repositories given. A run is bounded by any character that is no
# hexadecimal digit, so a hash joined to a word is found too.
commits_in() {
    local dir=$1 words repo
    shift
    words=$(grep -rhoaP '(?<![0-9A-Fa-f])[0-9A-Fa-f]{7,40}(?![0-9A-Fa-f])' "$dir" |
        tr 'A-F' 'a-f' | grep '[a-f]' | LC_ALL=C sort -u || true)
    if [[ -z "$words" ]]; then return 0; fi
    for repo in "$@"; do
        awk '{ print $0 "^{commit} " $0 }' <<<"$words" |
            git -C "$repo" cat-file --batch-check='%(objectname) %(rest)' |
            awk '$1 ~ /^[0-9a-f]{40}$/ { print $2 }'
    done | LC_ALL=C sort -u
}

stamp=$(git -C .. show -s --format=%ct "$full")
if [[ "$flavour" == anonymous ]]; then
    mapfile -t commits < <(commits_in "$out" .. "$prototype_clone")
    mapfile -t texts < <(grep -rlI '' "$out")
    COMMITS="${commits[*]}" perl -CSD -pi -e '
        BEGIN { %mask = map { $_ => 1 } split " ", $ENV{COMMITS} }
        s/(?<![0-9A-Fa-f])([0-9A-Fa-f]{7,40})(?![0-9A-Fa-f])/$mask{lc $1} ? "0" x length $1 : $1/ge;
        s/llm[-_]taskgen[-_]prototype/anonymous-prototype/gi;
        s/(math[-_ ]?trail)/my $m = $1; $m eq uc $m ? "ANONYMOUS" : $m =~ m{^M} ? "Anonymous" : "anonymous"/gie;
    ' "${texts[@]}"
    protocol=$(sha256sum experiments/PROTOCOL-A-offline.md | cut -d' ' -f1)
    PROTOCOL=$protocol PACKAGES=$packages perl -pe 's/\{\{PROTOCOL_SHA256\}\}/$ENV{PROTOCOL}/g; s/\{\{PACKAGES\}\}/$ENV{PACKAGES}/g' release/ANONYMISED.md >"$out/ANONYMISED.md"
    # The README describes the named copy; the reader of this one is sent to
    # what it leaves out first.
    {
        printf '> This copy is anonymised for double-blind review: [ANONYMISED.md](ANONYMISED.md) says what it replaces and what it leaves out.\n\n'
        cat "$out/README.md"
    } >"$out/README.md.new"
    mv "$out/README.md.new" "$out/README.md"

    # The checks fail closed: binary files are searched too, file names as well
    # as contents, and a name or commit that survived stops the artifact.
    # Cyrillic is matched byte for byte, so each spelling is listed as it is.
    if grep -rnaiP 'math[-_ ]?trail|llm[-_]taskgen|r[iy]azanov|rjazanov' "$out" ||
        grep -rnaF -e Матрейл -e Маттрейл -e МатТрейл -e матрейл -e маттрейл -e Рязанов -e рязанов "$out"; then
        echo "assemble.sh: the anonymous artifact still names the system or its authors (above)" >&2
        exit 1
    fi
    if find "$out" \( -iname '*math*trail*' -o -iname '*taskgen*' \) -print | grep .; then
        echo "assemble.sh: file names in the anonymous artifact still name the system (above)" >&2
        exit 1
    fi
    left=$(commits_in "$out" .. "$prototype_clone")
    if [[ -n "$left" ]]; then
        echo "assemble.sh: the anonymous artifact still names commits: ${left//$'\n'/ }" >&2
        exit 1
    fi
    # A run of digits alone is read as a number, so the search above passes it
    # over; it may still begin the hash of a commit the research names — the
    # pin, the prototype's pin, a run's commit — and then it is refused.
    known=$({
        git -C .. rev-parse "$full"
        sed -n 's/^\(commit\|pin\)=//p' experiments/*/results/provenance.txt | xargs -r -I{} git -C .. rev-parse "{}^{commit}"
        git -C "$prototype_clone" rev-parse "$(sed -n 's/^commit=//p' evidence/prototype-stats.txt)^{commit}"
    } | LC_ALL=C sort -u)
    digits=$(grep -rhoaP '(?<![0-9A-Fa-f])[0-9]{7,40}(?![0-9A-Fa-f])' "$out" | LC_ALL=C sort -u || true)
    prefixes=$(awk 'NR == FNR { known[$0]; next } NF { for (k in known) if (index(k, $0) == 1) { print; break } }' \
        <(echo "$known") <(echo "$digits"))
    if [[ -n "$prefixes" ]]; then
        echo "assemble.sh: these numbers begin the hash of a commit the research names: ${prefixes//$'\n'/ }" >&2
        exit 1
    fi
    # The exact time of the pinned commit would point to it as surely as its
    # hash, so the anonymous tarball's files carry the first day of 2000.
    stamp=946684800
fi

# Every file of the tarball carries one time, so the same sources make the
# same tarball.
# One mode for every file as well, whatever the umask the tree was made under.
tar --sort=name --mtime="@$stamp" --owner=0 --group=0 --numeric-owner --format=gnu --mode='u+rwX,go=rX' \
    -C "$build" -cf - "$name" | gzip -n -9 >"$build/$name.tar.gz"
sha256sum "$build/$name.tar.gz"
