package mcpserver_test

import (
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
)

// fileWithHistory is a profile file in the Drive whose latest state is the
// damage and whose state before it is the profile given.
func fileWithHistory(t *testing.T, fake *drivetest.Drive, earlier *profile.Profile) {
	t.Helper()

	raw, err := profile.Marshal(earlier)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	fake.Put(googleToken, &drivetest.File{
		Name: "mathtrail-profile.json", MimeType: "application/json", AppProperties: maps.Clone(profileMarker),
		Revisions: []drivetest.Revision{{Content: raw}, {Content: damaged}},
	})
}

// otter is a profile as a first sign-in makes it.
func otter() *profile.Profile {
	return profile.New(profile.Student{Pseudonym: "Otter", Grade: 2, Interests: []string{}, ExcludedSkills: []string{}},
		"restore_test", time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC))
}

// A damaged file is told, and nothing is written to the parent's Drive until
// the adult agrees: then the file is put back to its latest earlier version
// that reads, and the profile it holds is the one handed back.
func TestADamagedProfileIsPutBackOnlyWhenTheAdultAgrees(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	fileWithHistory(t, fake, otter())
	d := instanceOverDrive(t, fake)

	fake.ResetCalls()
	for _, step := range []struct {
		tool string
		args map[string]any
	}{
		{"get_profile", map[string]any{}},
		{"get_progress", map[string]any{}},
		{"next_task", map[string]any{"language": "en"}},
	} {
		told := call(t, d.session, step.tool, step.args)
		wantOurSentence(t, told, "The child's profile file in the adult's Google Drive is damaged")
	}
	if calls := fake.Calls(); calls["update"] != 0 || calls["keep"] != 0 || calls["delete"] != 0 || calls["revisions"] != 0 {
		t.Errorf("finding the damage cost %v, want nothing written and the history left alone", calls)
	}

	restored := call(t, d.session, "save_profile", map[string]any{"restore": true})
	if text := textOf(t, restored); restored.IsError || !strings.HasPrefix(text, "The damaged profile file was put back") {
		t.Fatalf("save_profile with restore = %q, want the file put back", text)
	}
	if got := payloadOf[profilePayload](t, restored); got.Profile == nil || got.Profile.Pseudonym != "Otter" {
		t.Errorf("save_profile with restore = %+v, want the profile put back", got)
	}
	if got := payloadOf[profilePayload](t, call(t, d.session, "get_profile", map[string]any{})); got.Profile == nil || got.Profile.Pseudonym != "Otter" {
		t.Errorf("get_profile after the file was put back = %+v, want the profile put back", got)
	}
}

// Putting the file back is all a call that asks for it may ask for: beside
// anything else it is refused whole, and nothing of the Drive is touched.
func TestARestoreBesideAnythingElseIsRefusedWhole(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	fileWithHistory(t, fake, otter())
	d := instanceOverDrive(t, fake)

	fake.ResetCalls()
	result := call(t, d.session, "save_profile", map[string]any{"restore": true, "pseudonym": "Badger"})
	if text := textOf(t, result); result.IsError || !strings.HasPrefix(text, "Nothing was done: restore is passed alone") {
		t.Errorf("save_profile with restore and a pseudonym = %q, want it refused", text)
	}
	if got := payloadOf[profilePayload](t, result); got.Status != "rejected" || got.Code != "restore_alone" {
		t.Errorf("save_profile with restore and a pseudonym = %+v, want status rejected, code restore_alone", got)
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("a refused restore cost %v, want nothing asked of Drive", calls)
	}
}

// A profile that reads is never put back over: the call says so, and writes
// nothing.
func TestARestoreOfAReadableProfilePutsNothingBack(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	d := instanceOverDrive(t, fake)
	call(t, d.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2})

	fake.ResetCalls()
	result := call(t, d.session, "save_profile", map[string]any{"restore": true})
	if text := textOf(t, result); result.IsError || !strings.HasPrefix(text, "The profile can be read, so nothing was put back.") {
		t.Errorf("save_profile with restore over a readable profile = %q, want nothing put back", text)
	}
	if calls := fake.Calls(); calls["update"] != 0 || calls["keep"] != 0 || calls["delete"] != 0 {
		t.Errorf("a restore not made cost %v, want nothing written", calls)
	}
}
