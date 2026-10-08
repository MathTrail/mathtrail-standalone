package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitLocation are the variables through which git is told which repository,
// working tree and index to use. A git hook sets them for the repository it
// runs in, so a test run from a hook would otherwise read, or commit to, that
// repository instead of its own.
var gitLocation = []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"}

// withoutGitLocation clears gitLocation from the process's environment for the
// rest of the test, both for the commands the test runs and for the ones the
// code under test runs.
func withoutGitLocation(t *testing.T) {
	t.Helper()
	for _, key := range gitLocation {
		t.Setenv(key, "") // restores the variable once the test ends
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

// git runs git in a repository and returns what it printed.
func git(t *testing.T, repo string, args ...string) string {
	t.Helper()
	identity := []string{"-c", "user.name=x", "-c", "user.email=x@example.invalid", "-c", "commit.gpgsign=false", "-C", repo}
	out, err := exec.CommandContext(t.Context(), "git", append(identity, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// committedRepository is a repository whose one commit holds a.go, with a.go
// changed in the working tree since; it returns the repository and the commit.
func committedRepository(t *testing.T) (repo, commit string) {
	t.Helper()
	repo = t.TempDir()
	git(t, repo, "init", "-q")
	if err := os.WriteFile(filepath.Join(repo, "a.go"), []byte("const Topics = 17\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "a.go")
	git(t, repo, "commit", "-q", "-m", "the pinned commit")
	commit = git(t, repo, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(repo, "a.go"), []byte("const Topics = 18\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return repo, commit
}

// A proof is read as the file stood at the pinned commit: an edit in the
// working tree since then must not stand in for it, or the ledger would vouch
// for a line no commit holds.
//
// Not parallel: the test clears git's location variables with t.Setenv.
func TestGitShowReadsTheCommitNotTheWorkingTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	withoutGitLocation(t)
	repo, commit := committedRepository(t)
	show := gitShow(t.Context(), repo, commit)

	content, err := show("a.go")

	if err != nil || string(content) != "const Topics = 17\n" {
		t.Errorf("show(a.go) = %q, %v; want the committed line", content, err)
	}

	_, err = show("missing.go")

	if err == nil || !strings.Contains(err.Error(), "git show "+commit+":missing.go: fatal:") {
		t.Errorf("show(missing.go) error = %v, want one naming %s:missing.go with git's own message", err, commit)
	}
}

// When git cannot be run at all, the error names the file it was asked for and
// keeps the cause, rather than passing for a file that is empty.
//
// Not parallel: the test clears git's location variables with t.Setenv.
func TestGitShowKeepsTheCauseWhenGitCannotRun(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	withoutGitLocation(t)
	repo, commit := committedRepository(t)
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()

	content, err := gitShow(cancelled, repo, commit)("a.go")

	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), commit+":a.go") || content != nil {
		t.Errorf("show(a.go) with git stopped = %q, %v; want no content and an error naming %s:a.go that wraps the cancellation", content, err, commit)
	}
}
