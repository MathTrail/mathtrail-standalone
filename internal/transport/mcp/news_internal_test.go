package mcpserver

import (
	"errors"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// News of an account reaches the questions listening for it, each handed the
// profile written, and none of another account's; a question taken away,
// before the news or after, leaves nothing behind.
func TestNewsReachesTheQuestionsOfItsAccountAlone(t *testing.T) {
	t.Parallel()

	heard := newNews()
	first, stopFirst := heard.listen("masha")
	second, stopSecond := heard.listen("masha")
	other, stopOther := heard.listen("petya")
	gone, stopGone := heard.listen("masha")
	stopGone()

	written := newOtter()
	written.Student.Notes = "Likes puzzles."
	heard.tell("masha", written)
	for name, ear := range map[string]<-chan []byte{"the first": first, "the second": second} {
		raw, told := heardOf(ear)
		if !told {
			t.Errorf("%s question of the account did not hear the news, want it heard", name)
			continue
		}
		if p, err := store.Parse(raw); err != nil || p.Student.Notes != written.Student.Notes {
			t.Errorf("%s question was handed %q (%v), want the profile written", name, raw, err)
		}
	}
	if _, told := heardOf(other); told {
		t.Error("another account's question heard the news, want it still waiting")
	}
	if _, told := heardOf(gone); told {
		t.Error("a question taken away before the news heard it, want it forgotten")
	}

	stopFirst()
	stopSecond()
	stopOther()
	if len(heard.waiting) != 0 {
		t.Errorf("questions left listening = %v, want none once every one is taken away", heard.waiting)
	}
}

// A write that lands is news to the questions of its account; a write that
// fails, and a write of another account, are not.
func TestAWriteThatLandsIsNewsToItsAccount(t *testing.T) {
	t.Parallel()

	heard := newNews()
	kept := telling{Storage: memory.New(), news: heard}
	masha, petya := store.Account{ID: "masha"}, store.Account{ID: "petya"}
	revisions := map[string]store.Revision{}
	for _, account := range []store.Account{masha, petya} {
		revision, err := kept.Create(t.Context(), account, newOtter())
		if err != nil {
			t.Fatalf("Create(%s) error = %v, want nil", account.ID, err)
		}
		revisions[account.ID] = revision
	}
	ear, stop := heard.listen(masha.ID)
	defer stop()

	written := newOtter()
	written.Touch("test", time.Now())
	if _, err := kept.Save(t.Context(), masha, written, "a revision never written"); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("Save() over another revision error = %v, want store.ErrConflict", err)
	}
	if _, told := heardOf(ear); told {
		t.Fatal("a write that failed was news, want none")
	}
	if _, err := kept.Save(t.Context(), petya, written, revisions[petya.ID]); err != nil {
		t.Fatalf("Save() of another account error = %v, want nil", err)
	}
	if _, told := heardOf(ear); told {
		t.Fatal("another account's write was news, want none")
	}
	if _, err := kept.Save(t.Context(), masha, written, revisions[masha.ID]); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	raw, told := heardOf(ear)
	if !told {
		t.Fatal("a write of the account that landed was no news, want its question told")
	}
	if p, err := store.Parse(raw); err != nil || p.Revision != written.Revision {
		t.Errorf("the question was handed %q (%v), want the profile written, at revision %d", raw, err, written.Revision)
	}
}

// newOtter is a first profile.
func newOtter() *profile.Profile {
	return profile.New(profile.Student{Pseudonym: "Otter", Grade: 2, Interests: []string{}, ExcludedSkills: []string{}},
		"test", time.Now())
}

// heardOf is what an ear was handed, and whether it was handed anything yet.
func heardOf(ear <-chan []byte) ([]byte, bool) {
	select {
	case raw := <-ear:
		return raw, true
	default:
		return nil, false
	}
}

// A profile that cannot be written down is handed to the questions as nothing,
// and they hear it all the same: a write landed, and a question that waited for
// one is not left waiting out its hold for news that came.
func TestAProfileThatCannotBeWrittenDownIsStillNews(t *testing.T) {
	t.Parallel()

	heard := newNews()
	ear, stop := heard.listen("masha")
	defer stop()
	unwritable := newOtter()
	unwritable.Student.Grade = profile.MaxGrade + 1

	heard.tell("masha", unwritable)
	if raw, told := heardOf(ear); !told || raw != nil {
		t.Errorf("the question was handed %q, told %t, want it told, with nothing", raw, told)
	}
}
