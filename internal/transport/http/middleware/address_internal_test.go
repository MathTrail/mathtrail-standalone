package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A request is counted by the last hop of X-Forwarded-For, whatever came
// before it and however many times the header was sent; by the address of its
// connection when it carries no header, or a last hop that is no address; and
// an IPv6 address by its /64.
func TestARequestIsCountedByTheHopThePlatformAppended(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		forwarded []string
		peer      string
		want      string
	}{
		{name: "one hop", forwarded: []string{"203.0.113.7"}, peer: "10.1.2.3:4567", want: "203.0.113.7"},
		{name: "a client's hops before it", forwarded: []string{"10.0.0.1, 198.51.100.4, 203.0.113.7"}, peer: "10.1.2.3:4567", want: "203.0.113.7"},
		{name: "no spaces", forwarded: []string{"10.0.0.1,203.0.113.7"}, peer: "10.1.2.3:4567", want: "203.0.113.7"},
		{name: "the header sent twice", forwarded: []string{"203.0.113.7", "10.0.0.1, 198.51.100.4"}, peer: "10.1.2.3:4567", want: "198.51.100.4"},
		{name: "a hop with its port", forwarded: []string{"203.0.113.7:443"}, peer: "10.1.2.3:4567", want: "203.0.113.7"},
		{name: "an IPv6 hop", forwarded: []string{"2001:db8:1:2:3:4:5:6"}, peer: "10.1.2.3:4567", want: "2001:db8:1:2::/64"},
		{name: "an IPv6 hop with its port", forwarded: []string{"[2001:db8:1:2::9]:443"}, peer: "10.1.2.3:4567", want: "2001:db8:1:2::/64"},
		{name: "an IPv6 hop with a zone", forwarded: []string{"fe80::1%eth0"}, peer: "10.1.2.3:4567", want: "fe80::/64"},
		{name: "IPv4 written as IPv6", forwarded: []string{"::ffff:203.0.113.7"}, peer: "10.1.2.3:4567", want: "203.0.113.7"},
		{name: "a last hop that is no address", forwarded: []string{"203.0.113.7, unknown"}, peer: "10.1.2.3:4567", want: "10.1.2.3"},
		{name: "an empty last hop", forwarded: []string{"203.0.113.7, "}, peer: "10.1.2.3:4567", want: "10.1.2.3"},
		{name: "no header", peer: "192.0.2.10:4567", want: "192.0.2.10"},
		{name: "no header, an IPv6 connection", peer: "[2001:db8:9::1]:4567", want: "2001:db8:9::/64"},
		{name: "no header, and a connection of no address", peer: "a pipe", want: "a pipe"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
			req.RemoteAddr = tc.peer
			for _, line := range tc.forwarded {
				req.Header.Add("X-Forwarded-For", line)
			}
			if got := clientAddress(req); got != tc.want {
				t.Errorf("clientAddress() = %q, want %q", got, tc.want)
			}
		})
	}
}
