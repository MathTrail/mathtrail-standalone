package account_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/tools/load/account"
)

// clientID is the identifier the stand-in registers every client under.
const clientID = "load-client"

// authority stands in for a deployed service's sign-in, as far as a client
// sees it: the documents it reads, a registration, the parent's consent, the
// tokens and their renewal, and an endpoint that lets in only an access token
// it issued. It checks what the protocol has a client prove — the code's
// verifier, the redirect, the resource — and counts what it was asked for.
type authority struct {
	t      *testing.T
	server *httptest.Server
	// refusal, when set, is the error the parent's consent sends the browser
	// back with; closed, when set, has the endpoint fail every request it
	// lets in, so that no session opens after a sign-in.
	refusal string
	closed  bool

	mu        sync.Mutex
	redirect  string
	challenge string
	issued    int
	renewals  int
	access    string
	refresh   string
	expiresIn int
	bearers   []string
	// renewed, when set, is called as a renewal of the tokens arrives, before
	// it is answered, with mu held: it must take nothing mu guards.
	renewed func()
}

// newAuthority starts the stand-in, whose access tokens last the seconds
// given, as the settings given leave it, and stops it when the test is over.
func newAuthority(t *testing.T, expiresIn int, settings ...func(*authority)) *authority {
	t.Helper()
	a := &authority{t: t, expiresIn: expiresIn}
	for _, set := range settings {
		set(a)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/oauth-protected-resource/mcp", a.resourceDocument)
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", a.serverDocument)
	mux.HandleFunc("POST /register", a.register)
	mux.HandleFunc("GET /authorize", a.authorize)
	mux.HandleFunc("POST /token", a.token)
	endpoint := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcp.NewServer(&mcp.Implementation{Name: "service", Version: "1"}, nil)
	}, &mcp.StreamableHTTPOptions{Stateless: true})
	mux.Handle("/mcp", a.signedIn(endpoint))
	a.server = httptest.NewServer(mux)
	t.Cleanup(a.server.Close)
	return a
}

// url is where the stand-in is reached.
func (a *authority) url() string { return a.server.URL }

func (a *authority) resourceDocument(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"resource":              a.url() + "/mcp",
		"authorization_servers": []string{a.url()},
		"scopes_supported":      []string{"mcp"},
	})
}

func (a *authority) serverDocument(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":                                         a.url(),
		"authorization_endpoint":                         a.url() + "/authorize",
		"token_endpoint":                                 a.url() + "/token",
		"registration_endpoint":                          a.url() + "/register",
		"response_types_supported":                       []string{"code"},
		"grant_types_supported":                          []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":               []string{"S256"},
		"token_endpoint_auth_methods_supported":          []string{"none"},
		"scopes_supported":                               []string{"mcp"},
		"authorization_response_iss_parameter_supported": true,
	})
}

func (a *authority) register(w http.ResponseWriter, r *http.Request) {
	var asked struct {
		RedirectURIs []string `json:"redirect_uris"`
		AuthMethod   string   `json:"token_endpoint_auth_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&asked); err != nil || len(asked.RedirectURIs) != 1 || asked.AuthMethod != "none" {
		http.Error(w, "a public client with one redirect", http.StatusBadRequest)
		return
	}
	a.mu.Lock()
	a.redirect = asked.RedirectURIs[0]
	a.mu.Unlock()
	writeJSON(w, http.StatusCreated, map[string]any{
		"client_id":                  clientID,
		"redirect_uris":              asked.RedirectURIs,
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}

// authorize is the parent's consent and their sign-in at once: the browser is
// sent straight back with a code, or with the refusal the case set.
func (a *authority) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	a.mu.Lock()
	redirect := a.redirect
	a.challenge = q.Get("code_challenge")
	a.mu.Unlock()
	if q.Get("client_id") != clientID || q.Get("redirect_uri") != redirect || q.Get("code_challenge_method") != "S256" ||
		q.Get("resource") != a.url()+"/mcp" || q.Get("response_type") != "code" {
		http.Error(w, "an authorization request this client did not register for", http.StatusBadRequest)
		return
	}
	back := url.Values{"state": {q.Get("state")}, "iss": {a.url()}}
	if a.refusal != "" {
		back.Set("error", a.refusal)
	} else {
		back.Set("code", "the-code")
	}
	http.Redirect(w, r, redirect+"?"+back.Encode(), http.StatusFound)
}

func (a *authority) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil || r.PostForm.Get("client_id") != clientID {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_client"})
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		verified := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
		if r.PostForm.Get("code") != "the-code" || r.PostForm.Get("redirect_uri") != a.redirect ||
			base64.RawURLEncoding.EncodeToString(verified[:]) != a.challenge {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_grant"})
			return
		}
	case "refresh_token":
		if r.PostForm.Get("refresh_token") != a.refresh || a.refresh == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_grant"})
			return
		}
		a.renewals++
		if a.renewed != nil {
			a.renewed()
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "unsupported_grant_type"})
		return
	}
	a.issued++
	a.access, a.refresh = fmt.Sprintf("access-%d", a.issued), fmt.Sprintf("refresh-%d", a.issued)
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": a.access, "token_type": "Bearer", "expires_in": a.expiresIn, "refresh_token": a.refresh,
	})
}

// signedIn lets a request through to the endpoint only with an access token
// the stand-in issued last, and asks for a sign-in otherwise, as the service
// does.
func (a *authority) signedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		a.mu.Lock()
		good := bearer != "" && bearer == a.access
		if good {
			a.bearers = append(a.bearers, bearer)
		}
		a.mu.Unlock()
		if good && a.closed {
			http.Error(w, "no session today", http.StatusInternalServerError)
			return
		}
		if !good {
			w.Header().Set("WWW-Authenticate",
				fmt.Sprintf(`Bearer resource_metadata=%q`, a.url()+"/.well-known/oauth-protected-resource/mcp"))
			http.Error(w, "sign in", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// counted is what the stand-in was asked for: the access tokens it issued,
// the renewals among them, and the bearers its endpoint let in.
func (a *authority) counted() (issued, renewals int, bearers []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.issued, a.renewals, append([]string(nil), a.bearers...)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// following is a parent whose browser opens the address shown and follows
// the service all the way back to this computer.
func following(t *testing.T) account.Browser {
	t.Helper()
	return account.Browser{Show: func(address string) {
		go func() {
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
			if err != nil {
				t.Errorf("the browser could not read %s: %v", address, err)
				return
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Errorf("the browser could not open %s: %v", address, err)
				return
			}
			said, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.Request.URL.Path != "/callback" {
				t.Errorf("the browser stopped at %s with %d: %s", response.Request.URL, response.StatusCode, said)
			}
		}()
	}}
}

// signedInAt signs an account named parent in on the stand-in, keeping it in
// a directory of the test's, and is that directory.
func signedInAt(t *testing.T, a *authority) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := account.SignIn(soon(t), dir, "parent", a.url(), following(t), 0); err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	return dir
}

// A parent who signs in leaves an account the service let in: its token is the
// one the service issued, the session opened with it, and the file it is kept
// in is the owner's alone to read.
func TestSignInKeepsAnAccountTheServiceLetIn(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	// The directory is made by the first sign-in, as the user's own is.
	dir := filepath.Join(t.TempDir(), "accounts")
	signed, err := account.SignIn(soon(t), dir, "parent", a.url()+"/", following(t), 0)
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	token, err := signed.Token()
	if err != nil || token != "access-1" || signed.Name() != "parent" {
		t.Errorf("the account is %q with the token %q, %v; want parent with access-1", signed.Name(), token, err)
	}
	if _, _, bearers := a.counted(); len(bearers) == 0 || bearers[0] != "access-1" {
		t.Errorf("the endpoint let in %q, want the session opened with access-1", bearers)
	}
	info, err := os.Stat(filepath.Join(dir, "parent.json"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("the account's file is %v, %v; want one only its owner can read and write", info, err)
	}
	if folder, err := os.Stat(dir); err != nil || folder.Mode().Perm()&0o077 != 0 {
		t.Errorf("the accounts' directory is %v, %v; want one only its owner can enter", folder, err)
	}
}

// A browser that cannot come back to this computer leaves the address it was
// sent to in its address bar, and a parent who pastes it in signs in all the
// same.
func TestSignInTakesAnAddressPasted(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	pasted, paste := io.Pipe()
	defer func() { _ = paste.Close() }()
	browser := account.Browser{
		Show: func(address string) {
			go func() {
				// The browser follows the service, but not back to this computer.
				noFollow := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
				request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
				if err != nil {
					t.Errorf("the browser could not read %s: %v", address, err)
					return
				}
				response, err := noFollow.Do(request)
				if err != nil {
					t.Errorf("the browser could not open %s: %v", address, err)
					return
				}
				_ = response.Body.Close()
				_, _ = fmt.Fprintf(paste, "not an address\n%s\n", response.Header.Get("Location"))
			}()
		},
		Pasted: pasted,
	}
	signed, err := account.SignIn(soon(t), t.TempDir(), "parent", a.url(), browser, 0)
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if token, err := signed.Token(); err != nil || token != "access-1" {
		t.Errorf("Token() = %q, %v; want access-1", token, err)
	}
}

// A parent who refuses the load leaves no account, and is told the reason in
// the protocol's words.
func TestSignInRefusedKeepsNothing(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900, func(a *authority) { a.refusal = "access_denied" })
	dir := t.TempDir()
	if _, err := account.SignIn(soon(t), dir, "parent", a.url(), following(t), 0); err == nil ||
		!strings.Contains(err.Error(), "access_denied") {
		t.Fatalf("SignIn() error = %v, want the refusal named", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "parent.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refused sign-in left the file of an account: %v", err)
	}
}

// A run starts with tokens good for as long as it lasts: a token that would
// run out sooner is renewed when the account is loaded, and the renewed one is
// kept, so that the next run starts from it; one good for long enough is not.
func TestLoadRenewsATokenThatWouldRunOutDuringTheRun(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	dir := signedInAt(t, a)

	renewed, err := account.Load(dir, "parent", a.url(), 20*time.Minute)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if until := time.Until(renewed.Until()); until < 14*time.Minute || until > 16*time.Minute {
		t.Errorf("the renewed token runs out in %v, want the fifteen minutes the service gave it", until)
	}
	token, err := renewed.Token()
	if err != nil || token != "access-2" {
		t.Errorf("Token() = %q, %v; want the renewed access-2", token, err)
	}
	again, err := account.Load(dir, "parent", a.url(), time.Minute)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if token, err := again.Token(); err != nil || token != "access-2" {
		t.Errorf("Token() = %q, %v; want the access-2 kept from before", token, err)
	}
	if issued, renewals, _ := a.counted(); issued != 2 || renewals != 1 {
		t.Errorf("the service issued %d tokens in %d renewals, want two in one", issued, renewals)
	}
}

// An access token that has run out is renewed when it is next asked for, and
// the renewed one is kept.
func TestAnAccountRenewsATokenThatRanOut(t *testing.T) {
	t.Parallel()

	// A token that lasts a second is one the renewal treats as run out.
	a := newAuthority(t, 1)
	dir := signedInAt(t, a)

	loaded, err := account.Load(dir, "parent", a.url(), 0)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	token, err := loaded.Token()
	if err != nil || token == "access-1" {
		t.Errorf("Token() = %q, %v; want one renewed past access-1", token, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "parent.json"))
	if err != nil || strings.Contains(string(raw), `"refresh-1"`) {
		t.Errorf("the account's file holds %s, %v; want the renewed tokens kept", raw, err)
	}
}

// What an account cannot be loaded from is refused before a run starts, with
// what to do about it.
func TestLoadRefusesWhatItCannotUse(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	dir := signedInAt(t, a)
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{"service":"x"}`), 0o600); err != nil {
		t.Fatalf("write a broken account: %v", err)
	}
	for _, test := range []struct {
		name, account, service, says string
	}{
		{"never signed in", "nobody", a.url(), "sign it in first"},
		{"a name that is no file's", "../parent", a.url(), "no name to keep an account under"},
		{"another service", "parent", "https://elsewhere.example", "sign it in there"},
		{"a file that holds no account", "broken", a.url(), "holds no account"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := account.Load(dir, test.account, test.service, time.Minute); err == nil ||
				!strings.Contains(err.Error(), test.says) {
				t.Errorf("Load(%q, %q) error = %v, want one that says %q", test.account, test.service, err, test.says)
			}
		})
	}
}

// The accounts are kept in the user's own configuration unless told
// otherwise, never in the directory the tool runs from.
func TestAccountsAreKeptInTheUsersConfiguration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/home/someone/.config")

	if dir, err := account.Dir(); err != nil || dir != "/home/someone/.config/mathtrail-load" {
		t.Errorf("Dir() = %q, %v; want the load's folder of the user's configuration", dir, err)
	}
}

// soon is a context that ends long before the sign-in would give up by itself,
// so that a sign-in nobody came back to fails the test rather than hanging it.
func soon(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// A service that lets the parent sign in but opens no session leaves no
// account: the tokens it gave are kept only once a session opened with them.
func TestASignInWhoseSessionNeverOpensKeepsNothing(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900, func(a *authority) { a.closed = true })
	dir := t.TempDir()
	if _, err := account.SignIn(soon(t), dir, "parent", a.url(), following(t), 0); err == nil {
		t.Fatal("SignIn() error = nil, want the session that never opened told")
	}
	if issued, _, _ := a.counted(); issued != 1 {
		t.Errorf("the service issued %d tokens, want the one of the sign-in", issued)
	}
	if _, err := os.Stat(filepath.Join(dir, "parent.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a sign-in whose session never opened left the file of an account: %v", err)
	}
}

// A token the service gave no end to lasts as long as the service lets it,
// and is not renewed before every run for want of one.
func TestATokenWithNoEndIsNotRenewedBeforeARun(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	dir := t.TempDir()
	kept := `{"service": "` + a.url() + `", "client_id": "` + clientID + `", "token_url": "` + a.url() + `/token",` +
		` "token": {"access_token": "for-good", "token_type": "Bearer", "refresh_token": "never-asked"}}`
	if err := os.WriteFile(filepath.Join(dir, "parent.json"), []byte(kept), 0o600); err != nil {
		t.Fatalf("keep an account: %v", err)
	}

	loaded, err := account.Load(dir, "parent", a.url(), 20*time.Minute)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	token, err := loaded.Token()
	if err != nil || token != "for-good" {
		t.Errorf("Token() = %q, %v; want the kept for-good", token, err)
	}
	if _, renewals, _ := a.counted(); renewals != 0 {
		t.Errorf("the service was asked for %d renewals, want none", renewals)
	}
	if !loaded.Until().IsZero() {
		t.Errorf("Until() = %v, want no end for a token the service gave none", loaded.Until())
	}
}

// Anything else that asks at the address the browser comes back to — another
// program, a page that probes it — is turned away, and the parent's own
// browser coming back after it still signs the account in.
func TestStrayRequestsAtTheWayBackChangeNothing(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	parent := following(t)
	browser := account.Browser{Show: func(address string) {
		sent, err := url.Parse(address)
		if err != nil {
			t.Errorf("the address %q does not read: %v", address, err)
			return
		}
		for _, stray := range []string{"", "?error=access_denied&state=another", "?code=forged&state=another"} {
			turnedAway(t, sent.Query().Get("redirect_uri")+stray)
		}
		parent.Show(address)
	}}

	signed, err := account.SignIn(soon(t), t.TempDir(), "parent", a.url(), browser, 0)
	if err != nil {
		t.Fatalf("SignIn() error = %v, want the parent's own sign-in to go through", err)
	}
	if token, err := signed.Token(); err != nil || token != "access-1" {
		t.Errorf("Token() = %q, %v; want access-1", token, err)
	}
}

// A renewed token that cannot be written is still used: the run goes on with
// it, and the account says why it was not kept. Here the account's file has
// turned into a directory by the time the renewal comes back, and nobody can
// write a file over a directory — not even root, whom a directory that only
// forbids writing would not stop.
func TestARenewedTokenThatCannotBeKeptIsUsedAndTold(t *testing.T) {
	t.Parallel()

	a := newAuthority(t, 900)
	dir := signedInAt(t, a)
	file := filepath.Join(dir, "parent.json")
	a.mu.Lock()
	a.renewed = func() {
		if err := os.Remove(file); err != nil {
			t.Errorf("take the account's file away: %v", err)
		}
		if err := os.Mkdir(file, 0o700); err != nil {
			t.Errorf("put a directory where the account's file was: %v", err)
		}
	}
	a.mu.Unlock()

	loaded, err := account.Load(dir, "parent", a.url(), 20*time.Minute)
	if err != nil {
		t.Fatalf("Load() error = %v, want the renewed token used though it cannot be kept", err)
	}
	token, err := loaded.Token()
	if err != nil || token != "access-2" {
		t.Errorf("Token() = %q, %v; want the renewed access-2", token, err)
	}
	if unkept := loaded.Unkept(); unkept == nil || !strings.Contains(unkept.Error(), "account: keep parent") {
		t.Errorf("Unkept() = %v, want why the renewed token of parent could not be kept", unkept)
	}
	if left := namesIn(t, dir); !slices.Equal(left, []string{"parent.json"}) {
		t.Errorf("the accounts' directory holds %q, want only parent.json: a draft left behind keeps the renewed tokens where nothing looks after them", left)
	}
}

// namesIn are the names of everything in dir, in order.
func namesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v, want nil", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// turnedAway asks at the address given, as something other than the parent's
// browser would, and fails the case unless the request is turned away.
func turnedAway(t *testing.T, address string) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
	if err != nil {
		t.Errorf("the request %q does not read: %v", address, err)
		return
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("the request %q could not be sent: %v", address, err)
		return
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("the request %q came to %d, want it turned away", address, response.StatusCode)
	}
}
