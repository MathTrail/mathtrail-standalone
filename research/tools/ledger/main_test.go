package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceFileKeepsPermissionsAndLeavesNothingBehind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.md")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, []byte("new"), 0o640); err != nil {
		t.Fatalf("replaceFile: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want \"new\"", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("permissions = %v, want -rw-r-----", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only the ledger", len(entries))
	}
}

func TestReplaceFileLeavesNothingBehindWhenItFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A directory where the file should be: the final rename cannot succeed.
	path := filepath.Join(dir, "ledger.md")
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, []byte("new"), 0o640); err == nil {
		t.Fatal("replaceFile renamed a file over a directory, want an error")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries after a failed write, want only what was there", len(entries))
	}
}
