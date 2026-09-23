package starlark

import (
	"runtime"
	"testing"

	"go.starlark.net/starlark"
)

// TestLettersOfRefusesWithoutReading holds the reader of a result to refusing a
// long one without taking a copy of it first. The proof has to be in bytes
// rather than in the status, because both a reader that copies and a reader
// that measures end at the same refusal — and only one of them doubles a list
// the program has already allocated.
//
// It runs alone and reads the allocator, so it takes no part in the parallel
// tests around it.
func TestLettersOfRefusesWithoutReading(t *testing.T) {
	const elements = 1_000_000
	long := starlark.NewList(make([]starlark.Value, elements))

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	letters, problem := lettersOf(long)
	runtime.ReadMemStats(&after)

	if problem == "" {
		t.Fatalf("lettersOf: got %v and no complaint, want %d elements refused", letters, elements)
	}
	if taken := after.TotalAlloc - before.TotalAlloc; taken > 1<<20 {
		t.Errorf("refusing %d elements allocated %d bytes, want next to nothing", elements, taken)
	}
}
