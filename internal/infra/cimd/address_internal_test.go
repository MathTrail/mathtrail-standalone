package cimd

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Every kind of address a stranger's name could point the service at inside
// somebody's own network is refused, however it is written; the addresses of
// servers on the internet are not.
func TestOnlyPublicAddressesAreReached(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		address string
		public  bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"2606:4700:4700::1111", true},
		{"2001:4860:4860::8888", true},
		{"::ffff:8.8.8.8", true},

		{"127.0.0.1", false},
		{"127.1.2.3", false},
		{"::1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"172.31.255.255", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false},
		{"169.254.0.1", false},
		{"fe80::1", false},
		{"fc00::1", false},
		{"fd00::1", false},
		{"fd20:ce::254", false},
		{"224.0.0.1", false},
		{"239.255.255.250", false},
		{"ff02::1", false},
		{"0.0.0.0", false},
		{"::", false},
		{"255.255.255.255", false},
		{"0.1.2.3", false},
		{"100.64.0.1", false},
		{"192.0.0.1", false},
		{"198.18.0.1", false},
		{"240.0.0.1", false},
		{"::ffff:10.0.0.1", false},
		{"::ffff:127.0.0.1", false},
		{"::ffff:169.254.169.254", false},
		{"::a00:1", false},
		{"64:ff9b::a00:1", false},
		{"64:ff9b:1::a00:1", false},
		{"2001::a00:1", false},
		{"2001:db8::1", false},
		{"2002:a00:1::", false},
		{"4000::1", false},
	} {
		t.Run(tc.address, func(t *testing.T) {
			t.Parallel()

			if got := public(netip.MustParseAddr(tc.address)); got != tc.public {
				t.Errorf("public(%s) = %v, want %v", tc.address, got, tc.public)
			}
		})
	}
}

// closedBlocks are the blocks no address inside may pass, whatever its last
// bits are.
var closedBlocks = []string{
	"127.0.0.0/8", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7",
	"169.254.0.0/16", "fe80::/10", "224.0.0.0/4", "ff00::/8", "0.0.0.0/8", "100.64.0.0/10",
	"192.0.0.0/24", "198.18.0.0/15", "240.0.0.0/4", "64:ff9b::/96", "2001::/23", "2001:db8::/32",
	"2002::/16", "::ffff:10.0.0.0/104", "::ffff:127.0.0.0/104", "::ffff:169.254.0.0/112",
}

func TestTheAddressRuleHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("an IPv4 address written inside IPv6 is judged as the IPv4 address it is", prop.ForAll(
		func(bytes []uint8) bool {
			four := netip.AddrFrom4([4]byte(bytes))
			return public(four) == public(netip.AddrFrom16(four.As16()))
		},
		gen.SliceOfN(4, gen.UInt8()),
	))

	properties.Property("no address inside a closed block passes", prop.ForAll(
		func(block string, bytes []uint8) bool {
			return !public(inside(netip.MustParsePrefix(block), bytes))
		},
		gen.OneConstOf(anyOf(closedBlocks)...), gen.SliceOfN(16, gen.UInt8()),
	))

	properties.TestingRun(t)
}

// inside is an address of the block whose bits past the block's own are the
// ones given.
func inside(block netip.Prefix, bits []uint8) netip.Addr {
	base := block.Masked().Addr()
	if base.Is4() {
		four := base.As4()
		for i := range four {
			four[i] |= bits[i] &^ mask(block.Bits(), i)
		}
		return netip.AddrFrom4(four)
	}
	sixteen := base.As16()
	for i := range sixteen {
		sixteen[i] |= bits[i] &^ mask(block.Bits(), i)
	}
	return netip.AddrFrom16(sixteen)
}

// mask is the part of the i-th byte a block of the given length keeps.
func mask(length, i int) uint8 {
	switch kept := length - 8*i; {
	case kept >= 8:
		return 0xff
	case kept <= 0:
		return 0
	default:
		return uint8(0xff << (8 - kept))
	}
}

func anyOf(values []string) []interface{} {
	out := make([]interface{}, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

// An address the HTTP client hands the dial without a port is refused before
// its name is looked up: nothing about it is known to be a client's.
func TestAnAddressWithNoPortIsNotLookedUp(t *testing.T) {
	t.Parallel()

	var lookups atomic.Int32
	dial := reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) {
			lookups.Add(1)
			return []netip.Addr{netip.MustParseAddr("192.0.2.1")}, nil
		},
		allowed: public,
		dial:    (&net.Dialer{}).DialContext,
	}.guardedDial(time.Second)

	if conn, err := dial(t.Context(), "tcp", "client.example.com"); conn != nil || err == nil {
		t.Errorf("guardedDial() = %v, %v; want a refusal", conn, err)
	}
	if got := lookups.Load(); got != 0 {
		t.Errorf("%d lookups, want none", got)
	}
}

// A connection that will not take the deadline of its attempt is not used: it
// would be a connection nothing bounds, which is what the deadline is for. It
// is closed, and the dial refused.
func TestAConnectionThatWillNotTakeItsDeadlineIsClosed(t *testing.T) {
	t.Parallel()

	ours, theirs := net.Pipe()
	t.Cleanup(func() { _ = theirs.Close() })
	stubborn := &unbounded{Conn: ours}
	dial := reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("192.0.2.1")}, nil
		},
		allowed: public,
		dial:    func(context.Context, string, string) (net.Conn, error) { return stubborn, nil },
	}.guardedDial(time.Second)

	conn, err := dial(t.Context(), "tcp", "client.example.com:443")
	if conn != nil || err == nil {
		t.Errorf("guardedDial() = %v, %v; want a refusal", conn, err)
	}
	if !stubborn.closed.Load() {
		t.Error("the connection that would not take its deadline was left open")
	}
}

// unbounded is a connection that will not take a deadline, and says whether it
// was closed.
type unbounded struct {
	net.Conn
	closed atomic.Bool
}

func (*unbounded) SetDeadline(time.Time) error { return errors.New("no deadline taken") }

func (c *unbounded) Close() error {
	c.closed.Store(true)
	return c.Conn.Close()
}
