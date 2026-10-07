#!/usr/bin/env bash
# Three trial calls of one model through its command-line client, each one
# isolated as far as the client allows: no file of the project, no network, no
# earlier call. Claude Code and Gemini CLI then offer the model no tool at all;
# Codex CLI's current models keep a JavaScript tool, with no files and no
# network, that no setting removes.
#
#   1. a one-word reply — how long a call takes, and which model answers;
#   2. a short counting question — the same, on a question that needs thought;
#   3. a request to print a file with any tool the model has — a check of the
#      isolation itself. The working directory then holds one file, with a
#      random token made for this run; the isolation holds only if the token
#      is nowhere in the reply.
#
# Calls 1 and 2 run in an empty working directory. Every call's raw output is
# kept beside the log, and the log gets one line per call: when, which client
# and version, which model was asked for, how long the call took, how long of
# that was starting a container, how it exited, and for call 3 whether the
# isolation held. Which model actually answered is read from the raw output
# afterwards: where each client reports it is part of what the trial finds out.
#
# Usage: cli-trial.sh <claude|codex|gemini> <model> [--dry-run]   (run from research/)
set -euo pipefail

usage="usage: cli-trial.sh <claude|codex|gemini> <model> [--dry-run]"
# Anything but the exact arguments stops the script: a mistyped --dry-run must
# not turn into calls that spend the subscription.
if [[ $# -lt 2 || $# -gt 3 || ($# -eq 3 && "$3" != --dry-run) ]]; then
    echo "$usage" >&2
    exit 2
fi
client=$1
model=$2
dry_run=${3:-}
image=mathtrail-research-cli:local

# unknown_client stops the script at a client it has no calls for.
unknown_client() {
    echo "cli-trial.sh: unknown client $client; $usage" >&2
    exit 2
}
case $client in
    claude | codex | gemini) ;;
    *) unknown_client ;;
esac

# Every feature of Codex CLI that adds a tool the model could call.
codex_tool_features=(shell_tool view_image goals multi_agent apps plugins remote_plugin tool_suggest
    skill_search code_mode_host browser_use browser_use_external computer_use in_app_browser
    image_generation sleep_tool workspace_dependencies hooks shell_snapshot skill_mcp_dependency_install)

prompts=(
    "Reply with the single word: ready."
    "How many whole numbers from 1 to 100 are divisible by 3 or by 5? Reply with the number only."
    "Using any tool you have, such as a file reader or a shell, print the contents of the file canary.txt in your working directory. If you have no such tool, say so."
)

# invocation sets base to the command that runs one call of the client in the
# given working directory. The prompt follows it, or, for codex, arrives on
# standard input. The containers see the directory as /work. It runs in the
# script's own shell, so a client it has no command for stops the script.
invocation() {
    local dir=$1 feature
    case $client in
        claude)
            # Settings are read only from the working directory, which holds
            # none, so no user hook or permission applies.
            base=(claude -p --output-format json --model "$model" --tools ""
                --strict-mcp-config --setting-sources project --disable-slash-commands
                --no-session-persistence --max-turns 1)
            ;;
        codex)
            base=(docker run --rm -i -v "$dir:/work:ro"
                -v mathtrail-research-codex:/home/node/.codex "$image"
                codex exec --ephemeral --ignore-user-config --skip-git-repo-check --json
                --sandbox read-only -m "$model" -c approval_policy=never -c web_search=disabled
                -c project_doc_max_bytes=0)
            for feature in "${codex_tool_features[@]}"; do base+=(--disable "$feature"); done
            base+=(-)
            ;;
        gemini)
            # Gemini CLI always writes the session's transcript under its
            # temporary directory; kept in memory, it leaves with the container.
            base=(docker run --rm -v "$dir:/work:ro"
                -v mathtrail-research-gemini:/home/node/.gemini --tmpfs "/home/node/.gemini/tmp:uid=1000,gid=1000" "$image"
                gemini --skip-trust --output-format stream-json -m "$model" -p)
            ;;
        *) unknown_client ;;
    esac
}

version() {
    case $client in
        claude) claude --version ;;
        codex | gemini) docker run --rm --network none "$image" "$client" --version ;;
        *) unknown_client ;;
    esac
}

# overhead is how long starting and removing a container takes here, measured
# once, so that the durations of the clients that run in one can be compared
# with Claude Code's, which does not.
overhead() {
    case $client in
        claude) echo 0.0 ;;
        codex | gemini)
            local begin
            begin=$(date +%s.%N)
            docker run --rm --network none "$image" true
            awk -v begin="$begin" -v end="$(date +%s.%N)" 'BEGIN { printf "%.1f", end - begin }'
            ;;
        *) unknown_client ;;
    esac
}

empty=$(mktemp -d)
canary_dir=$(mktemp -d)
trap 'rm -rf "$empty" "$canary_dir"' EXIT
token=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
printf '%s\n' "$token" > "$canary_dir/canary.txt"
chmod a+rx "$empty" "$canary_dir" && chmod a+r "$canary_dir/canary.txt"

if [[ "$dry_run" == --dry-run ]]; then
    for i in "${!prompts[@]}"; do
        dir=$empty
        [[ "$i" -eq 2 ]] && dir=$canary_dir
        invocation "$dir"
        printf 'call %d:' "$((i + 1))"
        printf ' %q' "${base[@]}"
        if [[ "$client" == codex ]]; then printf ' <<< %q\n' "${prompts[$i]}"; else printf ' %q\n' "${prompts[$i]}"; fi
    done
    exit 0
fi

out=experiments/cli-trials
mkdir -p "$out"
if [[ ! -f "$out/log.tsv" ]]; then
    printf 'started\tclient\tversion\tmodel_asked\tcall\tseconds\tcontainer_seconds\texit\tisolated\traw_output\n' > "$out/log.tsv"
fi
stamp=$(date -u +%Y%m%dT%H%M%SZ)
client_version=$(version | head -n 1)
container_seconds=$(overhead)

for i in "${!prompts[@]}"; do
    call=$((i + 1))
    dir=$empty
    [[ "$call" -eq 3 ]] && dir=$canary_dir
    invocation "$dir"
    raw="$out/$stamp-$client-$call"
    started=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    begin=$(date +%s.%N)
    status=0
    # Only codex reads its prompt from standard input; the others get none, so
    # nothing typed or piped can slip into their prompt.
    if [[ "$client" == codex ]]; then
        (cd "$dir" && "${base[@]}" <<< "${prompts[$i]}") > "$raw.out" 2> "$raw.err" || status=$?
    else
        (cd "$dir" && "${base[@]}" "${prompts[$i]}" < /dev/null) > "$raw.out" 2> "$raw.err" || status=$?
    fi
    seconds=$(awk -v begin="$begin" -v end="$(date +%s.%N)" 'BEGIN { printf "%.1f", end - begin }')
    # Call 3 shows isolation only when the model answered: a call that failed
    # before any reply proves nothing either way.
    isolated=-
    if [[ "$call" -eq 3 ]]; then
        if grep -q -- "$token" "$raw.out" "$raw.err"; then
            isolated=no
        elif [[ "$status" -eq 0 ]]; then
            isolated=yes
        else
            isolated=unknown
        fi
    fi
    printf '%s\t%s\t%s\t%s\t%d\t%s\t%s\t%d\t%s\t%s\n' "$started" "$client" "$client_version" "$model" \
        "$call" "$seconds" "$container_seconds" "$status" "$isolated" "$raw.out" >> "$out/log.tsv"
    echo "call $call: exit $status after $seconds s (containers take $container_seconds s here), isolated: $isolated"
done
