package instance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// label marks every container a load run starts: one a run was killed before
// it could remove is found by it, with docker ps --filter label=mathtrail-load=1,
// and removed by hand.
const label = "mathtrail-load=1"

// servicePort is the port the service listens on inside its container.
const servicePort = "8080"

// sealKey is the variable of the key the service seals its tokens with. It is
// named on the command line and its value given in the command's own
// environment, so that no list of processes shows it.
const sealKey = "MATHTRAIL_SEAL_KEY_CURRENT"

// reserved are the variables every instance of a load run is given, which a
// spec may not change: the port its container publishes, the development
// sign-in its children sign in through, and the key made for it.
var reserved = []string{"PORT", "MATHTRAIL_DEV_AUTH", sealKey}

// runArgs are the arguments of the docker command that starts a container of
// the spec: detached, marked as a load run's, with the processors and the
// memory of an instance and no swap beside it, the service's port published
// on this machine alone, and the environment every instance has before the
// spec's own.
func runArgs(spec *Spec) []string {
	args := []string{
		"run", "--detach", "--label", label,
		"--cpus", strconv.FormatFloat(spec.CPUs, 'f', -1, 64),
		"--memory", spec.Memory, "--memory-swap", spec.Memory,
		"--publish", "127.0.0.1::" + servicePort,
		"--env", "PORT=" + servicePort, "--env", "MATHTRAIL_DEV_AUTH=true", "--env", sealKey,
	}
	for _, variable := range spec.Env {
		args = append(args, "--env", variable)
	}
	return append(args, spec.Image)
}

// publicHost is the host the service of the spec knows itself by: the one of
// the public URL the spec gives it, or the one the service falls back to when
// it is given none.
func publicHost(env []string) string {
	for _, variable := range env {
		if value, given := strings.CutPrefix(variable, "MATHTRAIL_PUBLIC_URL="); given {
			if public, err := url.Parse(value); err == nil && public.Host != "" {
				return public.Host
			}
		}
	}
	return "localhost:" + servicePort
}

// portOf is the port of this machine a container's port is published on, as
// docker port writes it: an address and a port on each line.
func portOf(published string) (string, error) {
	line, _, _ := strings.Cut(strings.TrimSpace(published), "\n")
	_, port, err := net.SplitHostPort(strings.TrimSpace(line))
	switch {
	case err != nil:
		return "", fmt.Errorf("instance: read the published port of %q: %w", published, err)
	case port == "":
		return "", fmt.Errorf("instance: %q publishes no port", published)
	}
	return port, nil
}

// process is a container's process as docker inspect writes it: its id on
// this machine and when it began.
func process(inspected string) (pid int, began time.Time, err error) {
	id, at, found := strings.Cut(strings.TrimSpace(inspected), " ")
	if !found {
		return 0, time.Time{}, fmt.Errorf("instance: read the process of %q", inspected)
	}
	if pid, err = strconv.Atoi(id); err != nil {
		return 0, time.Time{}, fmt.Errorf("instance: read the process id of %q: %w", inspected, err)
	}
	if began, err = time.Parse(time.RFC3339Nano, at); err != nil {
		return 0, time.Time{}, fmt.Errorf("instance: read when %q began: %w", inspected, err)
	}
	return pid, began, nil
}

// state is what docker says of a container as it stands: whether it runs, and
// if not, how it ended.
type state struct {
	Running   bool
	OOMKilled bool
	ExitCode  int
}

// stateOf reads a container's state as docker inspect writes it.
func stateOf(inspected string) (state, error) {
	var s state
	if err := json.Unmarshal([]byte(inspected), &s); err != nil {
		return state{}, fmt.Errorf("instance: read the state of a container: %w", err)
	}
	return s, nil
}

// docker runs the docker command with the arguments given, the variables
// given added to its environment, and is what it wrote.
func docker(ctx context.Context, env []string, args ...string) (string, error) {
	command := dockerCommand(ctx, args...)
	command.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("instance: docker %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// logOf is what a container wrote, both of its streams in the order docker
// kept them: all of it, or as docker's own flags given narrow it.
func logOf(ctx context.Context, id string, narrowed ...string) (string, error) {
	written, err := dockerCommand(ctx, append(append([]string{"logs"}, narrowed...), id)...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("instance: docker logs: %w", err)
	}
	return string(written), nil
}

// dockerCommand is the docker command with the arguments given, in a group of
// processes of its own: an interrupt typed at the terminal goes to the tool,
// which stops what it runs as it decides, and not straight to a command that
// has to finish — the one that asks for a container, whose container would
// otherwise be left with nobody to remove it.
func dockerCommand(ctx context.Context, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, "docker", args...) //nolint:gosec // the docker command, with arguments the tool makes itself
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return command
}
