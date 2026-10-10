package mcpserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A call given up on while it waits for an answer still on its way ends the
// wait at once: a pause of an hour is not slept out over a call nobody waits
// for, and what ended it is told.
func TestAWaitForAnAnswerEndsWithItsCall(t *testing.T) {
	t.Parallel()

	waiting := newOtter()
	waiting.CurrentTask = &profile.CurrentTask{ID: "tsk_waiting"}
	ctx, giveUp := context.WithCancel(t.Context())
	defer giveUp()
	s := &Service{store: givesUpAfterARead{p: waiting, giveUp: giveUp}, answerPause: time.Hour}

	ended := make(chan error, 1)
	go func() {
		_, _, err := s.toldOnceWritten(ctx, masha, waiting.CurrentTask.ID)
		ended <- err
	}()
	select {
	case err := <-ended:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("toldOnceWritten() error = %v, want the call given up on", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("toldOnceWritten() still waits for the answer, want the wait over with its call")
	}
}

// givesUpAfterARead is a store that finds the task on the card with no answer
// yet, and gives the call up as soon as it has read it.
type givesUpAfterARead struct {
	store.Storage
	p      *profile.Profile
	giveUp context.CancelFunc
}

func (g givesUpAfterARead) Load(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	g.giveUp()
	return g.p, "1", nil
}
