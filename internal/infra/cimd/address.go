package cimd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"
)

// globalIPv6 is the IPv6 space the internet routes to servers today. Anything
// outside it is either special by definition — the loopback, link-local,
// unique-local and multicast ranges — or reaches an IPv4 address through a
// translator: NAT64, and the IPv4-compatible form.
var globalIPv6 = netip.MustParsePrefix("2000::/3")

// notPublic are the blocks refused although the address alone looks routable.
// Each is either a network of somebody's own, or a way into IPv4 from IPv6
// that could land on one.
var notPublic = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // "this network"
	netip.MustParsePrefix("100.64.0.0/10"), // shared address space behind a carrier's NAT
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking networks
	netip.MustParsePrefix("240.0.0.0/4"),   // reserved, with the broadcast address in it
	netip.MustParsePrefix("2001::/23"),     // IETF protocol assignments, Teredo among them
	netip.MustParsePrefix("2001:db8::/32"), // documentation
	netip.MustParsePrefix("2002::/16"),     // 6to4, an IPv4 address inside
}

// public reports whether a document may be fetched from an address: one the
// internet routes to somebody else's server. Refused are the loopback, the
// private and unique-local networks, the link-local ones — the platform's
// metadata server, 169.254.169.254, among them — multicast, the unspecified
// address, and the blocks above. An IPv4 address written inside IPv6 is judged
// as the IPv4 address it is.
func public(addr netip.Addr) bool {
	addr = addr.Unmap()
	switch {
	case !addr.IsGlobalUnicast(), addr.IsPrivate():
		return false
	case addr.Is6() && !globalIPv6.Contains(addr):
		return false
	}
	for _, block := range notPublic {
		if block.Contains(addr) {
			return false
		}
	}
	return true
}

// mostAddresses is how many of the addresses a name stands for are dialled.
// The name a client publishes its document under stands for a few, and the
// resolver puts first the ones this machine can reach; dialling more would
// only let a stranger's name have the service knock on one address after
// another of the name's choosing.
const mostAddresses = 4

// guardedDial connects to a name only through addresses that were checked. It
// resolves the name itself and refuses the whole name when any of its
// addresses is not allowed: a name that stands for a public address and a
// private one is not a public name. Then it dials the first few of those
// addresses, never the name, so that no second lookup can put another address
// in their place.
//
// Everything it starts is over within the time an attempt has, counted from
// the moment it is called: the lookup, each dial, and — through the deadline
// the connection is given — the handshake and whatever is read and written
// after it. The request's own deadline reaches none of these: the HTTP client
// lets a dial outlive its request, so that a later one may use the
// connection, and a server that takes the connection and never says a word
// would otherwise hold it, and a goroutine with it, for as long as it liked.
//
// Each address is dialled in an equal share of the time that is left, so that
// one that never answers — an IPv6 route that swallows every packet, say —
// leaves the ones after it their turn.
func (r reach) guardedDial(within time.Duration) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		deadline := time.Now().Add(within)
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("cimd: split the address: %w", err)
		}
		addrs, err := r.checked(ctx, host)
		if err != nil {
			return nil, err
		}
		addrs = addrs[:min(len(addrs), mostAddresses)]

		failures := make([]error, 0, len(addrs))
		for i, addr := range addrs {
			share := time.Until(deadline) / time.Duration(len(addrs)-i)
			conn, err := r.dialWithin(ctx, network, net.JoinHostPort(addr.Unmap().String(), port), share)
			if err != nil {
				failures = append(failures, err)
				continue
			}
			if err := conn.SetDeadline(deadline); err != nil {
				_ = conn.Close()
				return nil, fmt.Errorf("cimd: bound the connection: %w", err)
			}
			return conn, nil
		}
		return nil, errors.Join(failures...)
	}
}

// dialWithin dials one address, giving up once its share of the time is spent.
func (r reach) dialWithin(ctx context.Context, network, address string, share time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, share)
	defer cancel()
	return r.dial(ctx, network, address)
}

// checked are the addresses a name stands for, when every one of them may be
// dialled.
func (r reach) checked(ctx context.Context, host string) ([]netip.Addr, error) {
	addrs, err := r.resolve(ctx, host)
	switch {
	case err != nil:
		return nil, fmt.Errorf("cimd: resolve: %w", err)
	case len(addrs) == 0:
		return nil, errors.New("cimd: resolve: no address")
	}
	for _, addr := range addrs {
		if !r.allowed(addr) {
			return nil, ErrAddress
		}
	}
	return addrs, nil
}
