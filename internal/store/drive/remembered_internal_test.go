package drivestore

import (
	"strconv"
	"testing"
	"time"
)

// learnt is when the files of these cases are learnt.
var learnt = time.Date(2026, time.September, 28, 9, 30, 0, 0, time.UTC)

// The memory never holds more than its bound, and what was just learnt is
// always among what it holds: making room forgets somebody else.
func TestTheMemoryKeepsToItsBound(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	for i := range maxRemembered + 10 {
		account := "account-" + strconv.Itoa(i)
		ids.remember(account, "file-"+strconv.Itoa(i), learnt)
		if known, remembered := ids.recall(account, learnt); !remembered || known.file != "file-"+strconv.Itoa(i) {
			t.Fatalf("recall(%s) = %+v, %t right after it was learnt, want its file", account, known, remembered)
		}
	}
	if held := len(ids.entries); held != maxRemembered {
		t.Errorf("the memory holds %d entries, want its bound of %d", held, maxRemembered)
	}

	// An account already remembered changes its file without forgetting
	// anyone.
	ids.remember("account-"+strconv.Itoa(maxRemembered+9), "another-file", learnt)
	if held := len(ids.entries); held != maxRemembered {
		t.Errorf("after a known account's file changed the memory holds %d entries, want %d", held, maxRemembered)
	}
}

// Forgetting a file that was found wrong leaves alone what was learnt about
// the account since.
func TestForgettingLeavesWhatWasLearntSince(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	ids.remember("mia", "the-file-found-wrong", learnt)
	ids.remember("mia", "the-file-learnt-since", learnt)
	ids.forget("mia", "the-file-found-wrong")
	if known, remembered := ids.recall("mia", learnt); !remembered || known.file != "the-file-learnt-since" {
		t.Errorf("recall() = %+v, %t, want the file learnt since", known, remembered)
	}

	ids.forget("mia", "the-file-learnt-since")
	if known, remembered := ids.recall("mia", learnt); remembered {
		t.Errorf("recall() = %+v after it was forgotten, want nothing", known)
	}
}

// A file is trusted for as long as a file found by a search is, and then the
// memory has nothing to say: a search finds where the profile is now, the bin
// included. Writing to it does not make it any younger.
func TestAFileIsTrustedForALimitedTime(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	ids.remember("mia", "the-file", learnt)
	ids.wrote("mia", "the-file", 7)
	for _, tc := range []struct {
		after time.Duration
		known bool
	}{
		{0, true},
		{maxAge - time.Second, true},
		{maxAge, false},
		{maxAge + time.Hour, false},
	} {
		if _, remembered := ids.recall("mia", learnt.Add(tc.after)); remembered != tc.known {
			t.Errorf("recall() %v after the file was learnt: remembered = %t, want %t", tc.after, remembered, tc.known)
		}
	}
}

// The number written is held to the file it was written to: a number written
// to another file changes nothing, finding the same file again keeps it,
// finding another file in its place lets it go, and letting go of it keeps the
// file.
func TestTheNumberWrittenBelongsToItsFile(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	ids.remember("mia", "the-file", learnt)
	ids.wrote("mia", "the-file", 7)
	ids.wrote("mia", "another-file", 9)
	if known, _ := ids.recall("mia", learnt); known.written != 7 {
		t.Errorf("written = %d after a write to another file, want 7", known.written)
	}

	later := learnt.Add(time.Minute)
	ids.remember("mia", "the-file", later)
	if known, _ := ids.recall("mia", later); known.written != 7 || !known.learnt.Equal(later) {
		t.Errorf("recall() after the same file was found again = %+v, want 7 written, learnt anew", known)
	}

	ids.wrote("mia", "the-file", 0)
	if known, remembered := ids.recall("mia", later); !remembered || known.file != "the-file" || known.written != 0 {
		t.Errorf("recall() after the number was let go of = %+v, %t, want the file and nothing written", known, remembered)
	}

	ids.wrote("mia", "the-file", 8)
	ids.remember("mia", "another-file", later)
	if known, _ := ids.recall("mia", later); known.written != 0 {
		t.Errorf("written = %d after another file was found in its place, want nothing", known.written)
	}
}
