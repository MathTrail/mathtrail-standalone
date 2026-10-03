package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The test binary stands in for a model's client when its first argument is
// fakeClient; the next is what it does, then the endpoint it does it to.
const fakeClient = "fake-client"

func TestMain(m *testing.M) {
	if len(os.Args) > 3 && os.Args[1] == fakeClient {
		os.Exit(actAsClient(os.Args[2], os.Args[3], os.Args[4:]))
	}
	os.Exit(m.Run())
}

// actAsClient posts a body to the endpoint's /messages and fails as a client
// does when it is refused, or stays silent, or hangs, as the action says.
func actAsClient(action, endpoint string, args []string) int {
	var body io.Reader
	switch action {
	case "post":
		body = strings.NewReader(args[0])
	case "post-stdin":
		body = os.Stdin
	case "silent":
		return 0
	case "hang":
		time.Sleep(time.Minute)
		return 0
	case "hang-until-term":
		return noteTermination(args[0])
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint+"/messages", body)
	if err != nil {
		return 3
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 3
	}
	_ = response.Body.Close()
	return 1
}

// client is the command that runs the fake client against the endpoint's /v1.
func client(action string, args ...string) []string {
	return append([]string{os.Args[0], fakeClient, action, placeholder + "/v1"}, args...)
}

func TestRunJudgesWhatTheClientSent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		client       []string
		wantIsolated bool
		wantOutput   []string
	}{
		{
			name:         "a reply asked for with no tools",
			client:       client("post", `{"model":"m","messages":[]}`),
			wantIsolated: true,
			wantOutput:   []string{"the client exited with status 1", "POST /v1/messages: model m, no tools", "isolated: 1 request"},
		},
		{
			name:       "a reply asked for with a tool",
			client:     client("post", `{"model":"m","messages":[],"tools":[{"name":"Bash"}]}`),
			wantOutput: []string{"model m, 1 tool: Bash", "not isolated"},
		},
		{
			name:       "nothing sent",
			client:     client("silent"),
			wantOutput: []string{"the client exited with status 0", "not shown"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			var stdout bytes.Buffer
			args := append([]string{"-dir", dir, "--"}, tt.client...)

			isolated, err := run(context.Background(), args, strings.NewReader(""), &stdout)

			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if isolated != tt.wantIsolated {
				t.Errorf("isolated = %v, want %v; output:\n%s", isolated, tt.wantIsolated, &stdout)
			}
			wantInOutput(t, stdout.String(), tt.wantOutput...)
			wantFiles(t, dir, "client.out", "client.err")
		})
	}
}

func TestRunGivesTheClientItsStandardInput(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	args := append([]string{"-dir", t.TempDir(), "--"}, client("post-stdin")...)

	isolated, err := run(context.Background(), args, strings.NewReader(`{"model":"from-stdin","messages":[]}`), &stdout)

	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !isolated || !strings.Contains(stdout.String(), "POST /v1/messages: model from-stdin") {
		t.Errorf("isolated = %v, want the request read from standard input; output:\n%s", isolated, &stdout)
	}
}

func TestRunStopsAClientThatHangs(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	args := append([]string{"-dir", t.TempDir(), "-timeout", "100ms", "--"}, client("hang")...)

	isolated, err := run(context.Background(), args, strings.NewReader(""), &stdout)

	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if isolated || !strings.Contains(stdout.String(), "the client was stopped after 100ms") {
		t.Errorf("isolated = %v, want the client stopped and nothing shown; output:\n%s", isolated, &stdout)
	}
}

func TestRunRefusesToStartWithoutWhatItNeeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := t.TempDir()
	if err := os.WriteFile(filepath.Join(full, "request-001.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
	}{
		{"no directory", []string{"--", "env", "URL=" + placeholder, "true"}},
		{"no client", []string{"-dir", dir}},
		{"a client not pointed at the endpoint", []string{"-dir", dir, "--", "true"}},
		{"a client that does not exist", []string{"-dir", dir, "--", filepath.Join(dir, "no-such-client"), placeholder}},
		{"a directory that does not exist", []string{"-dir", filepath.Join(dir, "missing"), "--", "env", "URL=" + placeholder, "true"}},
		{"a directory that is not empty", []string{"-dir", full, "--", "env", "URL=" + placeholder, "true"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := run(context.Background(), tt.args, strings.NewReader(""), io.Discard)

			if err == nil {
				t.Error("run started, want an error")
			}
		})
	}
}

func wantInOutput(t *testing.T, output string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(output, want) {
			t.Errorf("output lacks %q:\n%s", want, output)
		}
	}
}

func wantFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("the client's output: %v", err)
		}
	}
}

// noteTermination waits for SIGTERM and writes a file when it comes, so a test
// can tell a client asked to end from one that was killed.
func noteTermination(marker string) int {
	terms := make(chan os.Signal, 1)
	signal.Notify(terms, syscall.SIGTERM)
	select {
	case <-terms:
		if os.WriteFile(marker, nil, 0o600) != nil {
			return 3
		}
		return 0
	case <-time.After(time.Minute):
		return 4
	}
}

func TestRunAsksAClientToEndBeforeKillingIt(t *testing.T) {
	t.Parallel()
	marker := filepath.Join(t.TempDir(), "terminated")
	args := append([]string{"-dir", t.TempDir(), "-timeout", "200ms", "--"}, client("hang-until-term", marker)...)

	if _, err := run(context.Background(), args, strings.NewReader(""), io.Discard); err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("the client got no SIGTERM before it was stopped: %v", err)
	}
}
