// Package account keeps the accounts the load signs in with where there is no
// development sign-in: a deployed service, which only a parent's own Google
// account gets into. A parent signs an account in once, in a browser, the way
// they sign a chat host in, and every run after that signs its children in
// with the account's tokens, renewed as they run out. The tokens reach the
// parent's Drive, so they are kept in a file of the parent's own, outside the
// repository, that nobody else on the computer can read.
package account

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// timeout is how long one request of a sign-in or a renewal may take.
const timeout = 30 * time.Second

// validName is what the name an account is kept under may be: the name of its
// file, held to characters a person reads at a glance.
var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

// Account is an account signed in on one service, as the load keeps it.
type Account struct {
	name   string
	tokens oauth2.TokenSource
}

// Name is the name the account is kept under.
func (a *Account) Name() string { return a.name }

// Token is the account's access token: the one it holds while that is good,
// and a new one, kept in its file, once it has run out.
func (a *Account) Token() (string, error) {
	token, err := a.tokens.Token()
	if err != nil {
		return "", fmt.Errorf("account: renew the tokens of %s, or sign it in again: %w", a.name, err)
	}
	return token.AccessToken, nil
}

// Dir is where accounts are kept unless another directory is named: the
// user's own configuration, never the repository.
func Dir() (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("account: find the user's configuration: %w", err)
	}
	return filepath.Join(config, "mathtrail-load"), nil
}

// Load is the account kept under the name given in the directory, for the
// service at the URL given, with an access token good for at least as long as
// asked. One that would run out sooner is renewed now, so that nothing needs
// renewing while a run is under way: a renewal goes to the sign-in, which
// holds every request to the pace of the address it came from, and a run may
// be spending that pace on purpose.
func Load(dir, name, service string, goodFor time.Duration) (*Account, error) {
	where, err := shelfOf(dir, name)
	if err != nil {
		return nil, err
	}
	k, err := where.read()
	if err != nil {
		return nil, err
	}
	if k.Service != serviceOf(service) {
		return nil, fmt.Errorf("account: %s was signed in on %s, not on %s: sign it in there", name, k.Service, serviceOf(service))
	}

	if time.Until(k.Token.Expiry) < goodFor {
		// A token with no access token in it is renewed before it is used.
		k.Token.AccessToken = ""
	}
	a := k.account(renewing(), where)
	if _, err = a.Token(); err != nil {
		return nil, err
	}
	return a, nil
}

// kept is an account as its file holds it: the service it was signed in on,
// the client the service registered for the load, where its tokens are
// renewed, and the tokens.
type kept struct {
	Service  string        `json:"service"`
	ClientID string        `json:"client_id"`
	TokenURL string        `json:"token_url"`
	Token    *oauth2.Token `json:"token"`
}

// whole says whether the file held all an account is: a service, a client, a
// place to renew at and a token to renew with.
func (k *kept) whole() bool {
	return k.Service != "" && k.ClientID != "" && k.TokenURL != "" && k.Token != nil && k.Token.RefreshToken != ""
}

// account is the account kept, whose tokens are renewed through the context
// given and kept where it is kept again whenever they are.
func (k *kept) account(ctx context.Context, where shelf) *Account {
	config := oauth2.Config{
		ClientID: k.ClientID,
		// The service registers the load as a public client, which proves
		// itself with its identifier in the body and nothing else.
		Endpoint: oauth2.Endpoint{TokenURL: k.TokenURL, AuthStyle: oauth2.AuthStyleInParams},
	}
	return &Account{name: where.name, tokens: &keeping{
		next:  config.TokenSource(ctx, k.Token),
		where: where,
		kept:  *k,
		last:  k.Token.AccessToken,
	}}
}

// keeping is the tokens of an account, which writes a renewed token into the
// account's file, so that the next run starts from it rather than from one
// that has run out.
type keeping struct {
	next  oauth2.TokenSource
	where shelf

	mu   sync.Mutex
	kept kept
	// last is the access token the file holds.
	last string
}

func (k *keeping) Token() (*oauth2.Token, error) {
	token, err := k.next.Token()
	if err != nil {
		return nil, err
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	if token.AccessToken != k.last {
		k.kept.Token = token
		if unkept := k.where.write(&k.kept); unkept != nil {
			return nil, unkept
		}
		k.last = token.AccessToken
	}
	return token, nil
}

// shelf is where one account is kept: a file of its name in the directory of
// accounts, reached only through that directory, so that nothing a name says
// can lead out of it.
type shelf struct {
	dir, name string
}

// shelfOf is where the account of the name given is kept in the directory.
func shelfOf(dir, name string) (shelf, error) {
	if !validName.MatchString(name) {
		return shelf{}, fmt.Errorf("account: %q is no name to keep an account under: "+
			"lowercase letters, digits and dashes, at most 32", name)
	}
	if dir == "" {
		return shelf{}, errors.New("account: no directory to keep accounts in")
	}
	return shelf{dir: dir, name: name}, nil
}

// file is the name of the account's file in the directory.
func (s shelf) file() string { return s.name + ".json" }

// read is the account kept here.
func (s shelf) read() (*kept, error) {
	var raw []byte
	root, err := os.OpenRoot(s.dir)
	if err == nil {
		raw, err = root.ReadFile(s.file())
		_ = root.Close()
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, fmt.Errorf("account: no account %s in %s: sign it in first", s.name, s.dir)
	case err != nil:
		return nil, fmt.Errorf("account: read %s: %w", s.name, err)
	}
	var k kept
	if json.Unmarshal(raw, &k) != nil || !k.whole() {
		return nil, fmt.Errorf("account: the file of %s holds no account: sign it in again", s.name)
	}
	return &k, nil
}

// write keeps the account here, readable by its owner alone, in a directory
// only its owner can enter when it is made for it. It is written whole or not
// at all: a run stopped halfway leaves the file it found.
func (s shelf) write(k *kept) error {
	raw, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return fmt.Errorf("account: encode %s: %w", s.name, err)
	}
	if err = os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("account: make %s: %w", s.dir, err)
	}
	root, err := os.OpenRoot(s.dir)
	if err != nil {
		return fmt.Errorf("account: open %s: %w", s.dir, err)
	}
	defer func() { _ = root.Close() }()
	draft := "." + s.name + "." + rand.Text() + ".draft"
	if err = root.WriteFile(draft, append(raw, '\n'), 0o600); err == nil {
		err = root.Rename(draft, s.file())
	}
	if err != nil {
		_ = root.Remove(draft)
		return fmt.Errorf("account: keep %s: %w", s.name, err)
	}
	return nil
}

// serviceOf is the URL of a service as an account names it: with no slash at
// its end, so that the same service is never two.
func serviceOf(url string) string { return strings.TrimSuffix(url, "/") }

// client is the HTTP client a sign-in and every renewal use.
func client() *http.Client { return &http.Client{Timeout: timeout} }

// renewing is the context a renewal runs in: the client above, and no end,
// since the tokens are renewed for as long as a run goes on.
func renewing() context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, client())
}
