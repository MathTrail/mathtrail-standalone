package drivestore

import (
	"strconv"
	"testing"
)

// The memory never holds more than its bound, and what was just learnt is
// always among what it holds: making room forgets somebody else.
func TestTheMemoryKeepsToItsBound(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	for i := range maxRemembered + 10 {
		account := "account-" + strconv.Itoa(i)
		ids.remember(account, "file-"+strconv.Itoa(i))
		if id, remembered := ids.recall(account); !remembered || id != "file-"+strconv.Itoa(i) {
			t.Fatalf("recall(%s) = %q, %t right after it was learnt, want its file", account, id, remembered)
		}
	}
	if held := len(ids.ids); held != maxRemembered {
		t.Errorf("the memory holds %d entries, want its bound of %d", held, maxRemembered)
	}

	// An account already remembered changes its file without forgetting
	// anyone.
	ids.remember("account-"+strconv.Itoa(maxRemembered+9), "another-file")
	if held := len(ids.ids); held != maxRemembered {
		t.Errorf("after a known account's file changed the memory holds %d entries, want %d", held, maxRemembered)
	}
}

// Forgetting a file that was found wrong leaves alone what was learnt about
// the account since.
func TestForgettingLeavesWhatWasLearntSince(t *testing.T) {
	t.Parallel()

	ids := newFileIDs()
	ids.remember("mia", "the-file-found-wrong")
	ids.remember("mia", "the-file-learnt-since")
	ids.forget("mia", "the-file-found-wrong")
	if id, remembered := ids.recall("mia"); !remembered || id != "the-file-learnt-since" {
		t.Errorf("recall() = %q, %t, want the file learnt since", id, remembered)
	}

	ids.forget("mia", "the-file-learnt-since")
	if id, remembered := ids.recall("mia"); remembered {
		t.Errorf("recall() = %q after it was forgotten, want nothing", id)
	}
}
