package mcpserver

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// readingLate is a store whose read, once armed, finds the file and then lets
// something else happen before it hands back what it found.
type readingLate struct {
	store.Storage
	armed  atomic.Bool
	during func()
}

func (r *readingLate) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	p, revision, err := r.Storage.Load(ctx, account)
	if r.armed.CompareAndSwap(true, false) {
		r.during()
	}
	return p, revision, err
}

// A write that lands while the question reads the file is heard: the question
// listens before it reads, so it is woken at once rather than at the end of
// its hold.
func TestAWriteThatLandsWhileTheQuestionReadsIsHeard(t *testing.T) {
	t.Parallel()

	s, kept, request := questionOver(t)
	kept.during = func() {
		p, revision, err := s.store.Load(t.Context(), masha)
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
			return
		}
		if _, err := p.Refuse(lessonMoment); err != nil {
			t.Errorf("Refuse() error = %v, want nil", err)
			return
		}
		p.Touch("test", lessonMoment)
		if _, err := s.store.Save(t.Context(), masha, p, revision); err != nil {
			t.Errorf("Save() error = %v, want nil", err)
		}
	}
	kept.armed.Store(true)

	told := askedWithin(t, s, request, 5*time.Second)
	if told.Screen != screenComing || told.Refused != 1 {
		t.Errorf("read_task = %+v, want the try turned down while it read the file", told)
	}
}

// A question whose call has ended waits no longer, and says how the task stood.
func TestAQuestionWhoseCallHasEndedWaitsNoLonger(t *testing.T) {
	t.Parallel()

	s, kept, request := questionOver(t)
	ctx, end := context.WithCancel(t.Context())
	kept.during = end
	kept.armed.Store(true)

	answered := make(chan Reply[awaitedOut], 1)
	go func() {
		reply, err := s.readTask(ctx, masha, readTaskIn{RequestID: request, Refused: new(int)})
		if err != nil {
			t.Errorf("readTask() error = %v, want nil", err)
		}
		answered <- reply
	}()
	select {
	case reply := <-answered:
		if reply.Payload.Screen != screenComing || reply.Payload.Refused != 0 {
			t.Errorf("read_task = %+v, want the task still being written", reply.Payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("read_task still waits once its call has ended, want it answered")
	}
}

// masha is the account the questions are asked for, and lessonMoment the
// moment they are asked at.
var (
	masha        = store.Account{ID: "masha"}
	lessonMoment = time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
)

// questionOver is just as much of the service as a card's question needs,
// over a store that reads late, holding a question for a minute; the profile
// it keeps has a task being written, the request named.
func questionOver(t *testing.T) (*Service, *readingLate, string) {
	t.Helper()

	kept := &readingLate{Storage: memory.New()}
	heard := newNews()
	s := &Service{
		store:  telling{Storage: kept, news: heard},
		news:   heard,
		hold:   time.Minute,
		window: time.Hour,
		now:    func() time.Time { return lessonMoment },
	}
	p := newOtter()
	request := p.Ask(&profile.Brief{
		Constraints: []string{}, Difficulty: 3, ExcludedSkills: []string{}, GradeLevel: rating.Grades12,
		PedagogicalGoal: profile.GoalNewTopic, Rationale: "A topic the child has not met yet.", Setting: "space",
		TargetConcept: "counting.gaps", TrapsToUse: []string{},
	}, profile.TutorRule, "en", lessonMoment)
	if _, err := kept.Create(t.Context(), masha, p); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	return s, kept, request.ID
}

// askedWithin is the answer to a card's question that has heard the task is
// being written with no try turned down, which has to come within the bound.
func askedWithin(t *testing.T, s *Service, request string, bound time.Duration) awaitedOut {
	t.Helper()

	answered := make(chan Reply[awaitedOut], 1)
	go func() {
		reply, err := s.readTask(t.Context(), masha, readTaskIn{RequestID: request, Refused: new(int)})
		if err != nil {
			t.Errorf("readTask() error = %v, want nil", err)
		}
		answered <- reply
	}()
	select {
	case reply := <-answered:
		return reply.Payload
	case <-time.After(bound):
		t.Fatalf("read_task did not answer within %v, want the news heard", bound)
		return awaitedOut{}
	}
}
