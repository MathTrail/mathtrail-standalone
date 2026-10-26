package googleauth

import (
	"net/http"
	"testing"
	"time"
)

// The sign-in's calls go over a transport of their own rather than the one the
// whole process shares, so that nothing else closing that one's idle
// connections can fail a call that has already been answered.
func TestTheSignInHasATransportOfItsOwn(t *testing.T) {
	t.Parallel()

	signIn, err := New(&Settings{
		ClientID:     "a-client",
		ClientSecret: "a-secret",
		RedirectURL:  "https://mcp.example/oauth/callback",
		Endpoints: Endpoints{
			Auth:   "https://accounts.example/auth",
			Token:  "https://accounts.example/token",
			About:  "https://drive.example/about",
			Revoke: "https://accounts.example/revoke",
		},
		Now: time.Now,
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	g, ok := signIn.(*google)
	if !ok {
		t.Fatalf("New() = %T, want the sign-in at Google", signIn)
	}
	if g.client.Transport == nil || g.client.Transport == http.DefaultTransport {
		t.Errorf("the sign-in's transport is %T(%p), want one of its own rather than the standard one", g.client.Transport, g.client.Transport)
	}
}
