package store_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// An account travels through errors, log lines and anything else that prints
// or encodes it by its identifier alone. Its token opens somebody's files, and
// it never goes along — whatever verb the account is printed with, whatever it
// is printed inside, and in JSON too.
func TestAnAccountIsPrintedWithoutItsToken(t *testing.T) {
	t.Parallel()

	const token = "ya29.a-token-that-opens-a-drive"
	account := store.NewAccount("account-7", token, time.Time{})
	encoded, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	printed := map[string]string{
		"as JSON":                          string(encoded),
		"inside a struct":                  fmt.Sprintf("%+v", struct{ Account store.Account }{account}),
		"inside a field nobody else names": fmt.Sprintf("%+v", struct{ account store.Account }{account}),
		"inside a slice":                   fmt.Sprint([]store.Account{account}),
		"through a pointer":                fmt.Sprintf("%v", &account),
		"inside an error":                  fmt.Errorf("load for %v: %w", account, store.ErrNotFound).Error(),
	}
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d", "%12s"} {
		printed["with "+verb] = fmt.Sprintf(verb, account)
	}

	secrets := []string{token, hex.EncodeToString([]byte(token)), strings.ToUpper(hex.EncodeToString([]byte(token)))}
	for how, text := range printed {
		for _, secret := range secrets {
			if strings.Contains(text, secret) {
				t.Errorf("printed %s: %q carries the token", how, text)
			}
		}
	}
	for _, how := range []string{"as JSON", "inside a struct", "inside a field nobody else names", "inside a slice",
		"through a pointer", "inside an error", "with %v", "with %+v", "with %#v", "with %s", "with %q", "with %12s"} {
		if !strings.Contains(printed[how], account.ID) {
			t.Errorf("printed %s: %q does not name the account", how, printed[how])
		}
	}
}

// The verb and the flags an account is printed with are the identifier's, as
// if the identifier had been printed on its own.
func TestAnAccountIsPrintedAsItsIdentifierWouldBe(t *testing.T) {
	t.Parallel()

	account := store.NewAccount("account-7", "a token", time.Time{})
	for _, verb := range []string{"%v", "%s", "%q", "%#v", "%x", "%-12s|", "%12q"} {
		if got, want := fmt.Sprintf(verb, account), fmt.Sprintf(verb, account.ID); got != want {
			t.Errorf("Sprintf(%q) = %q, want %q", verb, got, want)
		}
	}
}

// The token an account was made with is the one it hands the store, until the
// moment it was made with; an account made without one has none, and nothing
// ends.
func TestAnAccountHandsOverTheTokenItWasMadeWith(t *testing.T) {
	t.Parallel()

	ends := time.Date(2026, time.September, 28, 10, 0, 0, 0, time.UTC)
	account := store.NewAccount("account-7", "a token", ends)
	if got := account.Token(); got != "a token" {
		t.Errorf("Token() = %q, want %q", got, "a token")
	}
	if got := account.Ends(); !got.Equal(ends) {
		t.Errorf("Ends() = %v, want %v", got, ends)
	}
	madeWithout := store.Account{ID: "account-7"}
	if got := madeWithout.Token(); got != "" {
		t.Errorf("Token() of an account made without one = %q, want none", got)
	}
	if got := madeWithout.Ends(); !got.IsZero() {
		t.Errorf("Ends() of an account made without a token = %v, want the zero time", got)
	}
}
