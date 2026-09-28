package mcpserver_test

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// googleToken is the Google token the vouched account reaches its Drive with.
const googleToken = "a-google-access-token"

// damaged is what a profile file edited by hand and saved half-way holds.
var damaged = []byte(`{"schema_version": 1, "student": {"pseud`)

// profileMarker is how the service marks the profile's file in Drive.
var profileMarker = map[string]string{"mathtrail": "profile"}

// startOver is the arguments of a new start from the details given.
func startOver(pseudonym string, grade int) map[string]any {
	return map[string]any{"start_over": true, "pseudonym": pseudonym, "grade": grade}
}

// setAsideIn is what the files set aside in the Drive hold.
func setAsideIn(fake *drivetest.Drive) [][]byte {
	var aside [][]byte
	for _, file := range fake.Files(googleToken) {
		if file.AppProperties["mathtrail"] == "set-aside" {
			aside = append(aside, file.Content)
		}
	}
	return aside
}

// A file nothing can read, with nothing earlier to put it back from, is told
// as damage with its two ways out; the adult asks for a new start, and a new
// profile starts while the damaged file is set aside.
func TestAProfileNothingCanReadIsStartedOverWhenTheAdultAsks(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	fake.Put(googleToken, &drivetest.File{
		Name: "mathtrail-profile.json", MimeType: "application/json", AppProperties: maps.Clone(profileMarker), Content: damaged,
	})
	d := instanceOverDrive(t, fake)

	told := call(t, d.session, "get_profile", map[string]any{})
	wantOurSentence(t, told, "The child's profile file in the adult's Google Drive cannot be read")
	if text := textOf(t, told); !strings.Contains(text, "start_over") {
		t.Errorf("the damage is told as %q, want the new start offered", text)
	}

	started := call(t, d.session, "save_profile", startOver("Otter", 2))
	if text := textOf(t, started); started.IsError || !strings.HasPrefix(text, "A new profile is started.") {
		t.Fatalf("save_profile with start_over = %q, want a new profile started", text)
	}
	if got := payloadOf[profilePayload](t, call(t, d.session, "get_profile", map[string]any{})); got.Profile == nil || got.Profile.Pseudonym != "Otter" {
		t.Errorf("get_profile after the new start = %+v, want the new profile", got)
	}
	if aside := setAsideIn(fake); len(aside) != 1 || !bytes.Equal(aside[0], damaged) {
		t.Errorf("the files set aside hold %q, want the damaged file as it was", aside)
	}
}

// A profile that reads is never started over: the call says so, and writes
// nothing.
func TestAReadableProfileIsNotStartedOver(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	d := instanceOverDrive(t, fake)
	call(t, d.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2})

	fake.ResetCalls()
	result := call(t, d.session, "save_profile", startOver("Badger", 5))
	if text := textOf(t, result); result.IsError || !strings.HasPrefix(text, "The profile can be read, so it was not started over") {
		t.Errorf("save_profile with start_over over a readable profile = %q, want it not started over", text)
	}
	if calls := fake.Calls(); calls["update"] != 0 || calls["create"] != 0 {
		t.Errorf("a new start not made cost %v, want nothing written", calls)
	}
	if got := payloadOf[profilePayload](t, result); got.Profile == nil || got.Profile.Pseudonym != "Otter" {
		t.Errorf("save_profile with start_over over a readable profile = %+v, want the profile as it was", got)
	}
}

// A profile in the bin is the adult's to restore, or to start over from: the
// call says both, and a new start sets the file in the bin aside.
func TestAProfileInTheBinIsRestoredOrStartedOver(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	d := instanceOverDrive(t, fake)
	call(t, d.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2})
	for _, file := range fake.Files(googleToken) {
		if file.AppProperties["mathtrail"] == "profile" {
			fake.Edit(googleToken, file.ID, func(binned *drivetest.File) { binned.Trashed = true })
		}
	}
	cold := instanceOverDrive(t, fake)

	wantOurSentence(t, call(t, cold.session, "get_profile", map[string]any{}),
		"The child's profile file is in the adult's Google Drive bin")
	wantFailed(t, cold.h, "get_profile", "in_bin")
	if started := call(t, cold.session, "save_profile", startOver("Badger", 5)); started.IsError {
		t.Fatalf("save_profile with start_over from the bin: %s", textOf(t, started))
	}
	if aside := setAsideIn(fake); len(aside) != 1 {
		t.Errorf("%d files were set aside, want the one in the bin", len(aside))
	}
}

// A new start asked for where there is no profile at all — the bin emptied —
// is a first profile, and one without the details a first profile needs is
// refused for them.
func TestANewStartWithNothingToSetAsideIsAFirstProfile(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	d := instanceOverDrive(t, fake)
	refused := payloadOf[profilePayload](t, call(t, d.session, "save_profile", map[string]any{"start_over": true}))
	if refused.Status != "rejected" || refused.Code != "invalid_profile" {
		t.Errorf("save_profile with start_over and no details = %+v, want the missing details refused", refused)
	}
	created := call(t, d.session, "save_profile", startOver("Otter", 2))
	if text := textOf(t, created); created.IsError || !strings.HasPrefix(text, "The profile is created.") {
		t.Errorf("save_profile with start_over and no profile = %q, want a first profile", text)
	}
}

// The profile's tool says where the adult finds the file for themselves — the
// file is the export — and names the other files that hold a profile too.
func TestTheProfileSaysWhereItIsKept(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	d := instanceOverDrive(t, fake)
	call(t, d.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2})
	older := fake.Put(googleToken, &drivetest.File{
		Name: "mathtrail-profile (1).json", MimeType: "application/json", AppProperties: maps.Clone(profileMarker),
		Content: []byte("{}"), ModifiedTime: lessonDay.Add(-24 * time.Hour),
	})

	result := call(t, d.session, "get_profile", map[string]any{})
	payload := payloadOf[struct {
		Location *struct {
			Folder string `json:"folder"`
			File   string `json:"file"`
			Link   string `json:"link"`
			Others []struct {
				File string `json:"file"`
				Link string `json:"link"`
			} `json:"others"`
		} `json:"location"`
	}](t, result)
	where := payload.Location
	switch {
	case where == nil:
		t.Fatal("get_profile carries no location, want where the file is")
	case where.Folder != "MathTrail" || where.File != "mathtrail-profile.json" || !strings.HasPrefix(where.Link, "https://drive.google.com/"):
		t.Errorf("location = %+v, want the folder, the file and its link", where)
	case len(where.Others) != 1 || where.Others[0].Link != drivetest.WebViewLink(older):
		t.Errorf("the other files = %+v, want the older profile file", where.Others)
	}
	if text := textOf(t, result); !strings.Contains(text, "That file is the export") || !strings.Contains(text, "mathtrail-profile (1).json") {
		t.Errorf("get_profile says %q, want where the file is and the other one named", text)
	}
}

// accessGone is how a case takes the parent's Google access away: taken back
// at Google, or a Google token that ends before a call could.
var accessGone = []struct {
	name     string
	reader   mcpserver.TokenReader
	spoil    func(*drivetest.Drive)
	sentence string
	kind     string
}{
	{
		name:     "access taken back",
		reader:   readerEndingIn(time.Hour),
		spoil:    func(fake *drivetest.Drive) { fake.Revoke(googleToken) },
		sentence: "MathTrail can no longer reach the child's profile",
		kind:     "revoked",
	},
	{
		name: "a Google token that ends too soon",
		reader: func(context.Context, string) (store.Account, time.Time, error) {
			return store.NewAccount(vouchedUser, googleToken, time.Now().Add(time.Second)), time.Now().Add(time.Hour), nil
		},
		spoil:    func(*drivetest.Drive) {},
		sentence: "MathTrail's access to the adult's Google Drive has to be renewed",
		kind:     "expired",
	},
}

// When the parent's Google access is gone, the call that finds it is answered
// 401 with the challenge a request with no token gets, and why — the answer
// that has a host sign the parent in again — and its result carries the
// sentence and the same challenge.
func TestAccessGoneIsAnsweredWithAChallenge(t *testing.T) {
	t.Parallel()

	for _, tc := range accessGone {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake := drivetest.New(t)
			d := instanceOverDriveReading(t, fake, tc.reader)
			tc.spoil(fake)

			resp := d.h.post(t, legacyCall("get_profile", `{}`), http.Header{"Authorization": {"Bearer " + vouchedToken}})
			if resp.status != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", resp.status, http.StatusUnauthorized)
			}
			wantChallengeIn(t, resp.header.Get("WWW-Authenticate"))
			result, _ := message(t, resp)["result"].(map[string]any)
			if isError, _ := result["isError"].(bool); !isError || !strings.Contains(fmt.Sprint(result["content"]), tc.sentence) {
				t.Errorf("result = %v, want the failure told in words", result)
			}
			meta, _ := result["_meta"].(map[string]any)
			challenges, _ := meta["mcp/www_authenticate"].([]any)
			if len(challenges) != 1 {
				t.Fatalf("_meta = %v, want one challenge under mcp/www_authenticate", meta)
			}
			wantChallengeIn(t, fmt.Sprint(challenges[0]))
			d.h.settle()
			wantFailed(t, d.h, "get_profile", tc.kind)
		})
	}
}

// ChatGPT starts its sign-in from the challenge in a failed result, and not
// from a 401 in the middle of a call: it is answered with the result alone.
func TestChatGPTIsAskedToSignInAgainInTheResult(t *testing.T) {
	t.Parallel()

	for _, tc := range accessGone {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake := drivetest.New(t)
			d := instanceOverDriveReading(t, fake, tc.reader)
			chatGPT, err := d.h.connectAs(t, "openai-mcp", vouchedToken)
			if err != nil {
				t.Fatalf("connectAs() error = %v, want nil", err)
			}
			tc.spoil(fake)

			result := call(t, chatGPT, "get_profile", map[string]any{})
			wantOurSentence(t, result, tc.sentence)
			challenges, given := result.Meta["mcp/www_authenticate"].([]any)
			if !given || len(challenges) != 1 {
				t.Fatalf("_meta = %v, want one challenge under mcp/www_authenticate", result.Meta)
			}
			wantChallengeIn(t, fmt.Sprint(challenges[0]))
		})
	}
}

// wantChallengeIn holds a challenge to naming where the resource's metadata
// is, the scope, and that the token no longer does.
func wantChallengeIn(t *testing.T, challenge string) {
	t.Helper()

	for _, part := range []string{"Bearer ", `resource_metadata="` + metadata + `"`, `error="invalid_token"`, `scope="mcp"`} {
		if !strings.Contains(challenge, part) {
			t.Errorf("challenge = %q, want it to carry %q", challenge, part)
		}
	}
}

// revoked is a store whose access has been taken back.
type revoked struct{ store.Storage }

func (revoked) Load(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	return nil, "", store.ErrAccessRevoked
}

// A call under the development sign-in, which gives no challenge, is told the
// sentence alone.
func TestWithoutASignInThereIsNoChallenge(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, revoked{keptWith(t, "masha")})
	result := call(t, session, "get_profile", map[string]any{})
	wantOurSentence(t, result, "MathTrail can no longer reach the child's profile")
	if _, given := result.Meta["mcp/www_authenticate"]; given {
		t.Errorf("_meta = %v, want no challenge where the sign-in gave none", result.Meta)
	}
}

// damagedNoMore is a store whose profile reads as damage, and whose new start
// cannot reach where profiles are kept.
type damagedNoMore struct{ store.Storage }

func (damagedNoMore) Load(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	return nil, "", store.ErrCorrupted
}

func (damagedNoMore) StartOver(context.Context, store.Account, *profile.Profile) (store.Revision, error) {
	return "", store.ErrUnavailable
}

// A new start the store could not make is told as the store's failure, and
// nothing is said to have started.
func TestANewStartTheStoreCouldNotMakeIsToldSo(t *testing.T) {
	t.Parallel()

	h, session := lesson(t, damagedNoMore{keptWith(t, "masha")})
	wantOurSentence(t, call(t, session, "save_profile", startOver("Otter", 2)), "Google Drive is not answering at the moment")
	h.settle()
	wantFailed(t, h, "save_profile", "drive_unavailable")
}
