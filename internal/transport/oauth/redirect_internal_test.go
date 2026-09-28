package oauthserver

import (
	"strings"
	"testing"
)

// A sign-in may send the parent back over https to anywhere, over plain http
// to their own computer alone, and never to an address that means something
// other than it shows.
func TestOnlyAnAddressAParentMayBeSentToIsRedirectable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		uri          string
		redirectable bool
	}{
		{"https://claude.ai/api/mcp/auth_callback", true},
		{"https://chatgpt.com/connector_platform_oauth_redirect", true},
		{"https://a.example:8443/cb?state=kept", true},
		{"http://localhost:3000/callback", true},
		{"http://LOCALHOST:3000/callback", true},
		{"http://127.0.0.1:33418/", true},
		{"http://[::1]:8080/cb", true},
		{"HTTPS://a.example/cb", true},

		{"http://a.example/cb", false},
		{"ftp://a.example/cb", false},
		{"http://localhost.a.example/cb", false},
		{"http://127.0.0.2/cb", false},
		{"https://a.example/cb#part", false},
		{"https://a.example/cb#", false},
		{"https://user:secret@a.example/cb", false},
		{"https:///cb", false},
		{"/cb", false},
		{"com.example.app:/oauth", false},
		{"javascript:alert(1)", false},
		{"data:text/html,a", false},
		{"https://a.example/" + strings.Repeat("a", maxRedirectURI), false},
		{"", false},
	} {
		if got := redirectable(tc.uri); got != tc.redirectable {
			t.Errorf("redirectable(%q) = %v, want %v", tc.uri, got, tc.redirectable)
		}
	}
}
