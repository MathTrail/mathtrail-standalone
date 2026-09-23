package profile_test

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// The sealer these tests use is the real one, with a key made on the spot.
// It reaches out of the process for nothing, so there is nothing here worth
// faking — and a fake that returned a fixed string would prove nothing about
// a value whose whole job is to be tied to the child it belongs to.
//
// The import is the test package's, not the profile package's: what this
// package declares is the two methods it needs, and who provides them is
// somebody else's business.
func newSealer(t *testing.T) profile.Sealer {
	t.Helper()

	return ringNamed(t, "the key of these tests")
}

// newOtherSealer is a ring that never sealed anything of ours: the shape a
// retired key leaves behind.
func newOtherSealer(t *testing.T) profile.Sealer {
	t.Helper()

	return ringNamed(t, "a key nobody has any more")
}

func ringNamed(t *testing.T, name string) profile.Sealer {
	t.Helper()

	key := sha256.Sum256([]byte(name))
	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(key[:]), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	return ring.For(seal.PurposeTaskAnswer)
}

func secret() profile.TaskSecret {
	return profile.TaskSecret{
		Answer: "C",
		Distractors: map[string]profile.Distractor{
			"A": {Trap: "off_by_one", Text: "You counted the posts instead of the gaps."},
			"B": {Trap: "missed_case", Text: "One of the orders was left out."},
		},
		Solution: "Count the gaps: there are eight.",
		Solver:   "def solve(): return 'C'",
	}
}

func TestWhatIsSealedComesBack(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	answering(t, p, "counting.gaps", 3)
	sealer := newSealer(t)

	if err := p.SealTask(sealer, secret()); err != nil {
		t.Fatalf("SealTask() error = %v, want nil", err)
	}

	sealed := p.CurrentTask.Sealed
	if strings.Contains(sealed, "C\"") || strings.Contains(sealed, "off_by_one") ||
		strings.Contains(sealed, "gaps: there are eight") {
		t.Errorf("the sealed value carries what it is meant to hide: %q", sealed)
	}

	opened, err := p.OpenTask(sealer)
	if err != nil {
		t.Fatalf("OpenTask() error = %v, want nil", err)
	}
	if opened.Answer != "C" || opened.Solution != secret().Solution || opened.Solver != secret().Solver {
		t.Errorf("OpenTask() = %+v, want what was sealed", opened)
	}
	if len(opened.Distractors) != 2 || opened.Distractors["A"].Trap != "off_by_one" {
		t.Errorf("the distractors came back as %+v", opened.Distractors)
	}
	if opened.Version != profile.SecretVersion {
		t.Errorf("version = %d, want %d", opened.Version, profile.SecretVersion)
	}
}

// What is sealed belongs to one child and one task. Moved to another profile,
// or set against another task, it stops opening — which is the only thing
// standing between one child's answer and another child's screen.
func TestASealedTaskDoesNotTravel(t *testing.T) {
	t.Parallel()

	sealer := newSealer(t)
	mine := parseFixture(t, "dima")
	answering(t, mine, "counting.gaps", 3)
	if err := mine.SealTask(sealer, secret()); err != nil {
		t.Fatalf("SealTask() error = %v, want nil", err)
	}

	t.Run("to another child", func(t *testing.T) {
		t.Parallel()

		theirs := parseFixture(t, "olya")
		answering(t, theirs, "counting.gaps", 3)
		theirs.CurrentTask.ID = mine.CurrentTask.ID
		theirs.CurrentTask.Sealed = mine.CurrentTask.Sealed

		if _, err := theirs.OpenTask(sealer); !errors.Is(err, profile.ErrSealed) {
			t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrSealed)
		}
	})

	t.Run("to another task", func(t *testing.T) {
		t.Parallel()

		moved := parseFixture(t, "dima")
		answering(t, moved, "counting.gaps", 3)
		moved.CurrentTask.ID = "tsk_something_else"
		moved.CurrentTask.Sealed = mine.CurrentTask.Sealed

		if _, err := moved.OpenTask(sealer); !errors.Is(err, profile.ErrSealed) {
			t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrSealed)
		}
	})
}

// A key that has been retired takes the tasks it sealed with it. That is the
// accepted cost of one key ring for everything, and the answer is to tell the
// child this task can no longer be checked rather than to guess at it.
func TestATaskSealedByAKeyThatIsGoneCannotBeRead(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	answering(t, p, "counting.gaps", 3)
	if err := p.SealTask(newSealer(t), secret()); err != nil {
		t.Fatalf("SealTask() error = %v, want nil", err)
	}

	if _, err := p.OpenTask(newOtherSealer(t)); !errors.Is(err, profile.ErrSealed) {
		t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrSealed)
	}
}

// Nothing to seal, and nothing to open, when there is no task in flight.
func TestSealingNeedsATaskInFlight(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.CurrentTask = nil
	sealer := newSealer(t)

	if err := p.SealTask(sealer, secret()); !errors.Is(err, profile.ErrNoTask) {
		t.Errorf("SealTask() error = %v, want %v", err, profile.ErrNoTask)
	}
	if _, err := p.OpenTask(sealer); !errors.Is(err, profile.ErrNoTask) {
		t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrNoTask)
	}
}

// A sealed value this build cannot make sense of is refused the same way a
// retired key is: a task in flight is worth less than a profile.
func TestASealedValueOfAnotherShapeIsRefused(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	answering(t, p, "counting.gaps", 3)
	sealer := newSealer(t)

	plaintext := []byte(`{"answer":"C","version":0}`)
	value, err := sealer.Seal(plaintext, p.StudentID, p.CurrentTask.ID)
	if err != nil {
		t.Fatalf("seal a value of another shape: %v", err)
	}
	p.CurrentTask.Sealed = value

	if _, err := p.OpenTask(sealer); !errors.Is(err, profile.ErrSealed) {
		t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrSealed)
	}
}

// refusingSealer is a sealer that will not: the one thing this package can do
// about that is say so and leave the profile as it was.
type refusingSealer struct{ err error }

func (s refusingSealer) Seal([]byte, ...string) (string, error) { return "", s.err }

func (s refusingSealer) Open(string, ...string) ([]byte, error) { return nil, s.err }

func TestASealerThatRefusesLeavesTheTaskAlone(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	answering(t, p, "counting.gaps", 3)
	was := p.CurrentTask.Sealed
	refused := refusingSealer{err: errors.New("no key today")}

	err := p.SealTask(refused, secret())
	if err == nil {
		t.Fatal("SealTask() error = nil, want the refusal to be passed on")
	}
	if !strings.Contains(err.Error(), "no key today") {
		t.Errorf("SealTask() said %q, want it to carry what the sealer said", err)
	}
	if p.CurrentTask.Sealed != was {
		t.Error("a refused sealing still changed the task")
	}

	if _, err := p.OpenTask(refused); !errors.Is(err, profile.ErrSealed) {
		t.Errorf("OpenTask() error = %v, want %v", err, profile.ErrSealed)
	}
}

// Something opened, and it was not a task's secret. It is refused exactly as a
// retired key is, because the answer is the same either way: this task can no
// longer be checked.
func TestSomethingElseSealedInItsPlaceIsRefused(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	answering(t, p, "counting.gaps", 3)
	sealer := newSealer(t)

	for _, plaintext := range []string{`[1,2,3]`, `"a string"`, `{"answer":5}`, `not json at all`} {
		value, err := sealer.Seal([]byte(plaintext), p.StudentID, p.CurrentTask.ID)
		if err != nil {
			t.Fatalf("seal %q: %v", plaintext, err)
		}
		p.CurrentTask.Sealed = value

		if _, err := p.OpenTask(sealer); !errors.Is(err, profile.ErrSealed) {
			t.Errorf("OpenTask() on %q error = %v, want %v", plaintext, err, profile.ErrSealed)
		}
	}
}
