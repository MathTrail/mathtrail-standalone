package mcpserver_test

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The development sign-in acts for the account a request's credential names:
// an account of the name a bearer credential gives, which is how several
// children are told apart on one machine, and the development account for
// anything else. It refuses nobody — not a request with no credential, and not
// one whose credential names nobody, such as a token a client kept from a
// server with the real sign-in at the same address. The account reaches the
// tool, and the line of the call names it.
func TestTheDevelopmentSignInActsForTheAccountTheCredentialNames(t *testing.T) {
	t.Parallel()

	longest := strings.Repeat("n", 64)
	for _, test := range []struct {
		name, credential, account string
	}{
		{"a name", "Bearer greedy", "dev-greedy"},
		{"a name of every sign it may hold, the scheme in lower case", "bearer a.b_c-9", "dev-a.b_c-9"},
		{"the longest name", "Bearer " + longest, "dev-" + longest},
		{"no credential", "", mcpserver.DevAccount},
		{"another scheme", "Basic dXNlcjpwYXNz", mcpserver.DevAccount},
		{"a bearer with nothing after it", "Bearer", mcpserver.DevAccount},
		{"a bearer and a space", "Bearer ", mcpserver.DevAccount},
		{"two words", "Bearer two words", mcpserver.DevAccount},
		{"a name one sign too long", "Bearer " + longest + "n", mcpserver.DevAccount},
		{"a sign a name may not hold", "Bearer masha@school", mcpserver.DevAccount},
		{"letters outside the alphabet of names", "Bearer ünï", mcpserver.DevAccount},
		{"a token kept from the real sign-in", "Bearer " + strings.Repeat("Ab0-_", 60), mcpserver.DevAccount},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			h := serve(t, mcpserver.DevSignIn)
			header := http.Header{}
			if test.credential != "" {
				header.Set("Authorization", test.credential)
			}
			got := h.post(t, legacyCall("echo", `{"say":"hello"}`), header)
			if got.status != http.StatusOK {
				t.Fatalf("status = %d, want %d", got.status, http.StatusOK)
			}
			result, _ := message(t, got)["result"].(map[string]any)
			payload, _ := result["structuredContent"].(map[string]any)
			if payload["for"] != test.account {
				t.Errorf("the tool acted for %v, want %q", payload["for"], test.account)
			}

			h.settle()
			if user := field(t, h.lineOf(t, "echo"), "user"); user != test.account {
				t.Errorf("the line names user %q, want %q", user, test.account)
			}
		})
	}
}

// Every account a credential names keeps a pace of its own: one child's flood
// holds that child back and nobody else.
func TestEveryNamedDevelopmentAccountKeepsItsOwnPace(t *testing.T) {
	t.Parallel()

	var reached atomic.Int32
	h := pacedHarness(t, mcpserver.DevSignIn, pacesOf(t, 3, 1<<20), &reached)

	if first := h.countAs(t, "greedy"); first["isError"] == true {
		t.Fatalf("the greedy child's first call = %v, want it answered", first)
	}
	if again := h.countAs(t, "greedy"); !heldBack(again) {
		t.Errorf("the greedy child's call past its pace = %v, want it held back", again)
	}
	if patient := h.countAs(t, "patient"); patient["isError"] == true {
		t.Errorf("another child's call = %v, want it answered: the flood was not theirs", patient)
	}

	h.settle()
	if hits := hitsOf(h); !slices.Equal(hits, [][2]string{{"user_rate", "dev-greedy"}}) {
		t.Errorf("limit_hit lines = %v, want one, naming user_rate and the greedy child", hits)
	}
}

// Every account a credential names keeps a profile of its own: what one child
// saves, another does not read.
func TestEveryNamedDevelopmentAccountKeepsItsOwnProfile(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil)
	h.start(t, mcpserver.DevSignIn, service.ProfileTools()...)

	children := map[string]string{"ann": "Otter", "ben": "Badger"}
	for name, pseudonym := range children {
		session, err := h.connectWith(t, name)
		if err != nil {
			t.Fatalf("connect as %s: %v", name, err)
		}
		call(t, session, "save_profile", map[string]any{"pseudonym": pseudonym, "grade": 2})
	}

	for name, pseudonym := range children {
		session, err := h.connectWith(t, name)
		if err != nil {
			t.Fatalf("connect as %s: %v", name, err)
		}
		read := payloadOf[struct {
			Profile struct {
				Pseudonym string `json:"pseudonym"`
			} `json:"profile"`
		}](t, call(t, session, "get_profile", map[string]any{}))
		if read.Profile.Pseudonym != pseudonym {
			t.Errorf("%s reads the profile of %q, want their own, %q", name, read.Profile.Pseudonym, pseudonym)
		}

		p, _, err := kept.Load(context.Background(), store.NewAccount(mcpserver.DevAccount+"-"+name, "", time.Time{}))
		if err != nil {
			t.Fatalf("the store holds no profile for %s: %v", name, err)
		}
		if p.Student.Pseudonym != pseudonym {
			t.Errorf("the store keeps %q for %s, want %q", p.Student.Pseudonym, name, pseudonym)
		}
	}
	if _, _, err := kept.Load(context.Background(), devAccount); err == nil {
		t.Error("the development account has a profile, want none: every call named a child")
	}
}
