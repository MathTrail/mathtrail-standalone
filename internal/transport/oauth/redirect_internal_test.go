package oauthserver

import (
	"net/url"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
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

// A host is written in ASCII and in one case, a name as IDNA spells it: a
// letter of another script that looks like a Latin one is spelled out as what
// it is, a mark that turns text around, which an address may carry escaped, is
// escaped again rather than let reorder the line or the page it is read in,
// and a capital of another script stays itself rather than lowered into an
// ASCII letter.
func TestAHostIsWrittenInASCII(t *testing.T) {
	t.Parallel()

	backslash := string(rune(92))
	for _, tc := range []struct {
		name, address, want string
	}{
		{"a host", "https://claude.ai/api/mcp/auth_callback", "claude.ai"},
		{"a host in capitals", "https://Claude.AI/cb", "claude.ai"},
		{"a host that passes for another", "https://" + string(rune(0x0441)) + "laude.ai/cb", "xn--laude-0ye.ai"},
		{"a host that turns text around", "https://a%E2%80%AEb.example/cb", "a" + backslash + "u202eb.example"},
		{"a host with no ASCII form, in capitals", "https://My_Host.Example/cb", "my_host.example"},
		{"a capital of another script in a host with no ASCII form", "https://_x.opena%C4%B0.com/cb", "_x.opena" + backslash + "u0130.com"},
		{"an IPv4 address", "http://127.0.0.1:33418/cb", "127.0.0.1"},
		{"an IPv6 address", "http://[::1]:33418/cb", "::1"},
		{"no address", "not an address", ""},
		{"an address that does not parse", "https://[::1/cb", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := hostOf(tc.address); got != tc.want {
				t.Errorf("hostOf(%q) = %q, want %q", tc.address, got, tc.want)
			}
		})
	}
}

// A host reaches a line of the log and the consent screen in printable ASCII
// alone, whatever the address it came from spells: no line break to start a
// line of its own, and no mark to reorder the text it stands in.
func TestHostsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	printableASCII := func(text string) bool {
		return !strings.ContainsFunc(text, func(c rune) bool { return c <= ' ' || c >= 0x7f })
	}
	properties := gopter.NewProperties(nil)

	properties.Property("whatever follows the scheme of an address, its host is written in printable ASCII", prop.ForAll(
		func(scheme, rest string) bool { return printableASCII(hostOf(scheme + "://" + rest)) },
		gen.OneConstOf("https", "http"), gen.AnyString(),
	))

	// The address is put together from its parts, so that whatever the host
	// is spelled with stays the host rather than ending it.
	properties.Property("whatever a host is spelled with, it is written in printable ASCII", prop.ForAll(
		func(host string) bool {
			address := url.URL{Scheme: "https", Host: host + ".example", Path: "/cb"}
			return printableASCII(hostOf(address.String()))
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}
