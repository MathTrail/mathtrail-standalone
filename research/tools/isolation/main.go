// Command isolation runs a model's command-line client against a local
// endpoint that stands in for the model's service, and reports what the client
// would have put before the model. The endpoint records every request and
// answers each one with an error, so nothing leaves the machine, no
// subscription's limit is spent, and a made-up key is enough to sign in.
//
// The run passes when the client asked for at least one reply and no request
// it sent offered the model a tool: a model with no tool can neither read a
// file, nor run code, nor reach the network, and answers from itself alone.
//
// Usage:
//
//	isolation -dir <dir> [-timeout <duration>] -- <client> [args...]
//
// Every "{endpoint}" in the client's arguments becomes the endpoint's URL, so
// the client is pointed at it the way that client reads its endpoint: a flag,
// or a variable set through env(1). The endpoint listens on the loopback
// interface only, so the client runs beside this command: when the client
// lives in a container, this command runs in the same one. The client gets
// this command's standard input. The requests are written to the directory,
// which must be empty, one file each, beside the client's standard output and
// error.
//
// The exit status is 0 when the client is isolated, 1 when it is not or asked
// for no reply at all, and 2 when the check itself could not run.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// placeholder stands, in the client's arguments, for the endpoint's URL.
const placeholder = "{endpoint}"

func main() {
	isolated, err := run(context.Background(), os.Args[1:], os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "isolation:", err)
		os.Exit(2)
	}
	if !isolated {
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) (bool, error) {
	flags := flag.NewFlagSet("isolation", flag.ContinueOnError)
	dir := flags.String("dir", "", "the directory the requests and the client's output are written to")
	timeout := flags.Duration("timeout", 2*time.Minute, "how long the client may run before it is stopped")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	command := flags.Args()
	if *dir == "" || len(command) == 0 {
		return false, errors.New("usage: isolation -dir <dir> [-timeout <duration>] -- <client> [args...]")
	}
	if !namesEndpoint(command) {
		return false, fmt.Errorf("no argument of the client holds %s, so it would reach its real service", placeholder)
	}
	if err := isEmpty(*dir); err != nil {
		return false, err
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return false, fmt.Errorf("listen: %w", err)
	}
	endpoint := newRecorder(*dir)
	server := &http.Server{Handler: endpoint, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()

	url := "http://" + listener.Addr().String()
	outcome, err := runClient(ctx, *timeout, withEndpoint(command, url), stdin, *dir)
	if err != nil {
		return false, err
	}
	// Shutdown lets a handler that is still running finish, so every request
	// the client sent is kept before the requests are read.
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = server.Shutdown(stopCtx); err != nil {
		return false, fmt.Errorf("stop the endpoint: %w", err)
	}
	requests, err := endpoint.recorded()
	if err != nil {
		return false, err
	}

	fmt.Fprintf(stdout, "%s; its output is in %s and %s\n", outcome,
		filepath.Join(*dir, "client.out"), filepath.Join(*dir, "client.err"))
	verdict := Judge(requests)
	for _, line := range verdict.Lines {
		fmt.Fprintln(stdout, line)
	}
	fmt.Fprintln(stdout, verdict.Summary)
	return verdict.Isolated, nil
}

func namesEndpoint(command []string) bool {
	for _, arg := range command {
		if strings.Contains(arg, placeholder) {
			return true
		}
	}
	return false
}

func withEndpoint(command []string, url string) []string {
	out := make([]string, len(command))
	for i, arg := range command {
		out[i] = strings.ReplaceAll(arg, placeholder, url)
	}
	return out
}

// runClient runs the client to its end, or until the timeout stops it, with
// its output written to the directory. A client that fails is expected — every
// answer it gets is an error — so its status is described, not returned as an
// error.
func runClient(ctx context.Context, timeout time.Duration, command []string, stdin io.Reader, dir string) (string, error) {
	stdout, err := os.Create(filepath.Join(filepath.Clean(dir), "client.out"))
	if err != nil {
		return "", fmt.Errorf("client output: %w", err)
	}
	defer func() { _ = stdout.Close() }()
	stderr, err := os.Create(filepath.Join(filepath.Clean(dir), "client.err"))
	if err != nil {
		return "", fmt.Errorf("client output: %w", err)
	}
	defer func() { _ = stderr.Close() }()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...) //nolint:gosec // G204: running the client the person names is what the check is for
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	// A client stopped by the timeout is asked to end first: `docker run`
	// passes the signal on to its container, which a kill would leave running.
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second

	err = cmd.Run()
	var exit *exec.ExitError
	switch {
	case ctx.Err() != nil:
		return fmt.Sprintf("the client was stopped after %s", timeout), nil
	case err == nil:
		return "the client exited with status 0", nil
	case errors.As(err, &exit):
		return fmt.Sprintf("the client exited with status %d", exit.ExitCode()), nil
	default:
		return "", fmt.Errorf("run %s: %w", command[0], err)
	}
}

// isEmpty refuses a directory that already holds files: an earlier run's
// requests would sit beside this run's and be read as its own.
func isEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("the directory for the requests: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is not empty, and an earlier run's files would mix with this run's", dir)
	}
	return nil
}
