#!/usr/bin/env bash
# Facts about the product at one commit, computed from that commit itself.
# The working tree is shared with other work in progress, so nothing here reads
# it: the commit is unpacked into a temporary directory, and every count and
# the test coverage come from there.
#
# Usage: product.sh <commit>   (prints key=value lines)
set -euo pipefail
# shellcheck source-path=SCRIPTDIR source=facts.sh
source "$(dirname "${BASH_SOURCE[0]}")/facts.sh"

commit=${1:?usage: product.sh <commit>}
unpack "$(git rev-parse --absolute-git-dir)" "$commit"

# The public name a reader finds the commit under.
emit repository github.com/MathTrail/mathtrail-standalone
emit commit "${full:0:12}"
emit commit_date "$date"

# Catalogs: each file is one JSON array.
for catalog in topics traps skills; do
    n=$(jq length "content/catalogs/$catalog.json")
    emit "$catalog" "$n"
done

# Reference tasks: every examples file is an array of tasks.
n=$(jq -s 'map(length) | add' content/examples/*.json)
emit reference_tasks "$n"
for level in 1-2 3-4 5-6; do
    n=$(jq -r '.[].grade_level' content/examples/*.json | count_matching "^$level\$")
    emit "reference_tasks_${level/-/_}" "$n"
done
n=$(find content/examples/solvers -maxdepth 1 -name '*.star' | count_matching .)
emit reference_solvers "$n"
n=$(jq -r '.[].id' content/examples/*.json | while read -r id; do
    [ -f "content/examples/solvers/$id.star" ] || echo "$id"
done | count_matching .)
emit reference_tasks_without_solver "$n"

# Reference tasks whose every wrong option names a trap of the catalog: the
# distractors are exactly the options other than the key, and each names a
# trap the catalog has.
n=$(jq -s --slurpfile traps content/catalogs/traps.json '
    ($traps[0] | map(.id)) as $known
    | [.[][] | select(
        ((.options | keys) - [.correct_answer]) == (.distractors | keys)
        and all(.distractors[]; .trap as $t | any($known[]; . == $t))
      )] | length' content/examples/*.json)
emit reference_tasks_with_a_trap_per_wrong_option "$n"
n=$(find content/solvers -mindepth 2 -maxdepth 2 -name '*.star' | count_matching .)
emit solver_templates "$n"
n=$(find content/drawings -maxdepth 1 -name '*.json' | count_matching .)
emit drawing_frames "$n"
for f in content/instructions/*.md; do
    n=$(wc -w < "$f")
    emit "words_$(basename "$f" .md)" "${n// /}"
done

# The checks' refusal codes, one constant each.
n=$(count_matching '^[[:space:]]+Code[A-Za-z]+[[:space:]]+Code = "' < internal/domain/checks/checks.go)
emit check_codes "$n"

# Lines of the rule's code, tests aside, that name the answer's pace or its
# "I don't know" in any form: the pace as Pace, PaceSlow, paceOf or lastPace,
# and "I don't know" as the flag it sets (Confused, isConfused) or as the
# answer itself (DontKnow). "Space" and its kin are no pace. None means the
# rule that chooses the next task reads neither by name; a count above zero
# means it started to.
pace_or_confused='[a-z]Pace|(^|[^A-Za-z])[Pp]ace|[Cc]onfused|DontKnow'
n=$(find internal/domain/tutor -name '*.go' ! -name '*_test.go' -exec cat {} + | count_matching "$pace_or_confused")
emit pace_or_confused_lines_in_rule "$n"

# Model providers. The chat's model writes every task and the service calls no
# model of its own: no module of a model provider's SDK is among the product's
# dependencies, direct or not, and no model API's host is named in any Go file
# of the commit, tests included. A count above zero means a call to a model
# may have come in.
n=$(tr '[:upper:]' '[:lower:]' < go.mod |
    count_matching 'anthropic|openai|generative-?ai|genai|aiplatform|vertexai|bedrock|langchain|ollama|mistral|cohere|huggingface')
emit llm_sdk_modules "$n"
n=$(find . -name '*.go' -exec cat {} + | tr '[:upper:]' '[:lower:]' |
    count_matching 'api\.anthropic\.com|api\.openai\.com|generativelanguage\.googleapis\.com|aiplatform\.googleapis\.com|bedrock-runtime')
emit llm_api_hosts_in_code "$n"

# The widget's dictionaries: one file for each language its cards speak.
n=$(find web/locales -maxdepth 1 -name '*.json' | count_matching .)
emit widget_languages "$n"

# The Cloud Run settings as Terraform applies them. It reads terraform.tfvars,
# then every *.auto.tfvars in lexical order, a later value winning; a file it
# does not load by itself, such as staging.tfvars, sets nothing. A variable no
# file sets keeps its default, read inside that variable's own block so that
# another variable's default can never stand in. This repository has no
# .tfvars.json files, so they are not read.
terraform_value() {
    local LC_ALL=C name=$1 file value set=""
    for file in infra/terraform/terraform.tfvars infra/terraform/*.auto.tfvars; do
        [ -f "$file" ] || continue
        # A quoted value, or a bare one up to a space or a comment.
        value=$(sed -n -e "s/^$name *= *\"\([^\"]*\)\".*$/\1/p" \
            -e "s/^$name *= *\([^ \"#][^ #]*\).*$/\1/p" "$file" | tail -n 1)
        if [ -n "$value" ]; then
            set=$value
        fi
    done
    if [ -n "$set" ]; then
        echo "$set"
        return
    fi
    awk -v name="$1" '
        $0 == "variable \"" name "\" {" { inside = 1; next }
        inside && /^}/ { exit }
        inside && /^  default *=/ { sub(/^  default *= */, ""); gsub(/"/, ""); print; exit }
    ' infra/terraform/variables.tf
}
for variable in max_instances concurrency cpu memory request_timeout; do
    n=$(terraform_value "$variable")
    emit "cloud_run_$variable" "$n"
done

# Tests of the product module. A function counts by the signature go test runs:
# a Test, Fuzz or Benchmark name taking *testing.T, *testing.F or *testing.B,
# where what follows the prefix does not start with a lower-case letter.
# TestMain takes *testing.M and is no test of its own.
test_sources() { find internal cmd content -name '*_test.go' -exec cat {} +; }
functions() {
    test_sources | count_matching "^func $1([^a-z( ][^( ]*)?\([A-Za-z_][A-Za-z0-9_]* \*testing\.$2\)"
}
n=$(functions Test T)
emit test_functions "$n"
n=$(functions Fuzz F)
emit fuzz_functions "$n"
n=$(functions Benchmark B)
emit benchmark_functions "$n"
n=$(test_sources | count_matching '\.Property\(')
emit gopter_properties "$n"
n=$(find internal cmd content -name '*_test.go' | count_matching .)
emit test_files "$n"
n=$(find internal cmd content -name '*.go' ! -name '*_test.go' -exec cat {} + | count_matching .)
emit go_nonblank_lines_code "$n"
n=$(test_sources | count_matching .)
emit go_nonblank_lines_tests "$n"

# Statement coverage of the product's tests, measured the way the product's own
# test recipe measures it: with the race detector, atomic counters, no cache.
# A test that fails stops the facts, and its output says which one.
if ! go test ./... -race -count=1 -coverprofile="$work/cover.out" -covermode=atomic > "$work/test.log" 2>&1; then
    echo "${0##*/}: the product's tests fail at ${full:0:12}; the last lines of their output:" >&2
    tail -n 40 "$work/test.log" >&2
    exit 1
fi
n=$(go tool cover -func="$work/cover.out" | awk '/^total:/ { print $NF }')
emit coverage_statements "$n"

# The tools the MCP endpoint defines: each is one Define(Spec{...}), and the
# card's page, a resource, is not among them.
n=$(find internal/transport/mcp -name '*.go' ! -name '*_test.go' -exec cat {} + | count_matching 'Define\(Spec\{')
emit mcp_tools "$n"

# The model's own numbers: the constants of the rating, the checks, the solver,
# the sandbox and the content, most of them unexported, and what follows from
# them, such as the corridor's bounds on the difficulty scale. A test written
# into each package of the unpacked commit prints them as key=value lines. It is
# written only now, after every count above, so that no count includes it, and
# it goes with the temporary tree.
probe() {
    local dir=$1 package=$2 imports=$3 body=$4 lines
    printf 'package %s\n\nimport (\n%s\n)\n\nfunc TestResearchFacts(t *testing.T) {\n%s\n}\n' \
        "$package" "$imports" "$body" > "$dir/zz_research_facts_test.go"
    lines=$(go test -count=1 -run '^TestResearchFacts$' -v "./$dir" | grep -E '^[a-z0-9_]+=' || true)
    if [ -z "$lines" ]; then
        echo "${0##*/}: the probe in $dir printed nothing" >&2
        exit 1
    fi
    while IFS='=' read -r key value; do
        emit "$key" "$value"
    done <<<"$lines"
}

probe internal/domain/rating rating '"fmt"; "math"; "strings"; "testing"' '
	further := func(p float64) float64 { return p*(1-Guess)/(p-Guess) - 1 }
	fmt.Printf("guess_floor=%g\nstep_theta=%g\nstep_delta=%g\nstep_decay=%g\n", Guess, k0Theta, k0Delta, decay)
	fmt.Printf("middle_difficulty=%d\ndifficulties=%d\nlevel_shift=%g\n", middleDifficulty, Difficulties, levelShift)
	for i, level := range gradeLevels {
		fmt.Printf("grade_level_%d=%s\n", i+1, strings.Replace(string(level), "-", "–", 1))
		fmt.Printf("level_shift_%d=%g\n", i+1, Point{GradeLevel: level, Difficulty: middleDifficulty}.Beta())
	}
	fmt.Printf("ladder_points=%d\n", len(gradeLevels)*Difficulties)
	fmt.Printf("difficulty_step=%g\n", Point{GradeLevel: Grades12, Difficulty: 2}.Beta()-Point{GradeLevel: Grades12, Difficulty: 1}.Beta())
	fmt.Printf("corridor_low=%g\ncorridor_high=%g\ncorridor_middle=%g\n", corridorLow, corridorHigh, CorridorMiddle)
	fmt.Printf("corridor_easiest_offset=%.4f\ncorridor_hardest_offset=%.4f\n", logit(corridorHigh), logit(corridorLow))
	fmt.Printf("corridor_middle_offset=%.4f\ncorridor_width=%.4f\n", logit(CorridorMiddle), logit(corridorHigh)-logit(corridorLow))
	fmt.Printf("step_further_at_corridor_high_percent=%.2f\n", 100*further(corridorHigh))
	fmt.Printf("step_further_at_corridor_low_percent=%.2f\n", 100*further(corridorLow))
	const even, hard = 0.5, 0.3
	fmt.Printf("step_point_even=%g\nstep_further_at_even_percent=%.2f\n", even, 100*further(even))
	fmt.Printf("step_point_hard=%g\nstep_ratio_at_hard=%.2f\n", hard, 1+further(hard))
	fmt.Printf("trial_answers=%d\ntrial_prior_spread=%g\n", TrialAnswers, startSpread)
	fmt.Printf("elo_base=%d\nelo_points=%.0f\n", eloBase, eloScale*math.Ln10)'

probe internal/domain/checks checks '"fmt"; "math"; "testing"; "github.com/MathTrail/mathtrail-standalone/internal/domain/rating"' '
	for i, level := range []rating.GradeLevel{rating.Grades12, rating.Grades34, rating.Grades56} {
		fmt.Printf("sentence_words_%d=%d\n", i+1, sentenceLimits[level].words)
	}
	fmt.Printf("fk_grade_margin=%d\nnear_duplicate_threshold=%g\n", gradeMargin, Threshold)
	fmt.Printf("sketch_positions=%d\nsketch_bits=%g\nsketch_bytes=%d\n", sketchSize, math.Log2(positionValues), sketchBytes)
	fmt.Printf("drawing_width=%d\ndrawing_height=%d\n", DefaultDrawingLimits().Width, DefaultDrawingLimits().Height)'

probe internal/domain/solver solver '"fmt"; "testing"' '
	fmt.Printf("options=%d\nlabel_shift=%d\n", Count, relabel)'

probe internal/infra/starlark starlark '"fmt"; "testing"' '
	fmt.Printf("bytes_per_step=%d\n", bytesPerStep)'

probe internal/config config '"fmt"; "testing"' '
	fmt.Printf("solver_steps=%d\nsolver_seconds=%g\nsolver_wait_seconds=%g\n",
		DefaultSolverSteps, DefaultSolverTimeout.Seconds(), DefaultSolverWait.Seconds())'

probe internal/domain/profile profile '"fmt"; "testing"' '
	fmt.Printf("grade_min=%d\ngrade_max=%d\ndifficulty_min=%d\n", MinGrade, MaxGrade, MinDifficulty)
	fmt.Printf("max_attempts=%d\nfingerprints_kept=%d\n", MaxAttempts, MaxFingerprints)
	fmt.Printf("mastery_answers=%d\nmastery_streak=%d\nmastery_lost_after=%d\n", MasteryAnswers, MasteryStreak, MasteryLostAfter)'

probe internal/domain/tutor tutor '"fmt"; "testing"' '
	fmt.Printf("brief_traps=%d\n", Traps)'

probe content content '"fmt"; "testing"' '
	fmt.Printf("package_examples=%d\npackage_budget_kib=%d\n", examplesPerPackage, PackageBudget/1024)'

probe internal/transport/mcp mcpserver '"fmt"; "testing"' '
	fmt.Printf("mcp_revision=%s\n", supportedVersions[0])'
