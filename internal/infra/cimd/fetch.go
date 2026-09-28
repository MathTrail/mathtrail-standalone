// Package cimd fetches the metadata documents OAuth clients name themselves
// by. A client whose identifier is an HTTPS address publishes, at that
// address, who it is and where a sign-in may send the parent back to.
//
// It is the one request the service makes to an address a stranger chose, so
// it is made as narrowly as it can be. The identifier has to be a proper
// client address before anything is resolved. The name is resolved here, and
// every address it stands for has to be public, or nothing is dialled at all.
// What is dialled is the address that was checked, never the name again. No
// redirect is followed, the answer is read to a small limit and no further, and
// a document is kept only when it is one. Every way a fetch can fail is one of
// the errors below, which is what a caller branches on and logs.
package cimd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// The ways a fetch fails. Each wraps whatever caused it, and a caller tells
// them apart with errors.Is.
var (
	// ErrClientURL is a client identifier that is not an address a document
	// may be fetched from. Nothing was resolved.
	ErrClientURL = errors.New("cimd: not a client address")
	// ErrAddress is a name that stands for an address that is not public: a
	// loopback, a private or link-local network, the platform's metadata
	// server. Nothing was dialled.
	ErrAddress = errors.New("cimd: not a public address")
	// ErrRedirect is an answer that sends the request elsewhere. It is not
	// followed.
	ErrRedirect = errors.New("cimd: redirected")
	// ErrStatus is an answer other than a document.
	ErrStatus = errors.New("cimd: answered without a document")
	// ErrTooLarge is a document longer than any client needs.
	ErrTooLarge = errors.New("cimd: document too large")
	// ErrUnreachable is an address that did not answer, or not in time, or not
	// over a connection whose certificate names it.
	ErrUnreachable = errors.New("cimd: unreachable")
	// ErrDocument is an answer that is not a client metadata document, or not
	// this client's.
	ErrDocument = errors.New("cimd: not a client metadata document")
)

// Fetcher reads the metadata document a client names itself by.
type Fetcher interface {
	// Fetch returns the document at the address a client identifier is. A
	// copy fetched before is returned while it is fresh.
	Fetch(ctx context.Context, clientID string) (Fetched, error)
}

// Fetched is a document, and whether it came from memory rather than the
// network.
type Fetched struct {
	Document Document
	Cached   bool
}

const (
	// maxDocument is the longest document read. A client's document is a few
	// hundred bytes; anything past this is not one, and is not read further.
	maxDocument = 64 << 10
	// maxHeaderBytes caps the headers of an answer the same way.
	maxHeaderBytes = 16 << 10
	// attemptTimeout is how long one attempt may take, from the name lookup to
	// the last byte.
	attemptTimeout = 10 * time.Second
	// attempts is the first attempt and one more. The first name lookup a
	// fresh process makes can take longer than the attempt allows, and a
	// second one then finds the answer waiting.
	attempts = 2
)

// fetcher reads documents over its own client, which dials only what its reach
// allows, and keeps the good ones.
type fetcher struct {
	client  *http.Client
	cache   *cache
	timeout time.Duration
	now     func() time.Time
}

// reach is how a fetcher gets to a document: how it finds the addresses a
// name stands for, which of them it may dial, how it dials one, and whose
// certificates it believes — the system's, when roots is nil.
type reach struct {
	resolve func(ctx context.Context, host string) ([]netip.Addr, error)
	allowed func(netip.Addr) bool
	dial    func(ctx context.Context, network, address string) (net.Conn, error)
	roots   *x509.CertPool
}

// NewFetcher builds the fetcher the service uses: the system's resolver and
// certificates, and public addresses only.
func NewFetcher() Fetcher {
	dialer := &net.Dialer{}
	return newFetcher(reach{
		resolve: func(ctx context.Context, host string) ([]netip.Addr, error) {
			return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		},
		allowed: public,
		dial:    dialer.DialContext,
	}, attemptTimeout, time.Now)
}

// newFetcher builds a fetcher over a reach, taking an attempt's time limit and
// the clock the cache keeps time by.
func newFetcher(r reach, timeout time.Duration, now func() time.Time) *fetcher {
	transport := &http.Transport{
		// A proxy the environment names would be dialled in place of the
		// address that was checked.
		Proxy:                  nil,
		DialContext:            r.guardedDial(timeout),
		TLSClientConfig:        &tls.Config{RootCAs: r.roots, MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:      true,
		MaxResponseHeaderBytes: maxHeaderBytes,
		// A document is fetched once and then kept, so a connection left open
		// for the next fetch would only hold a socket to a host a stranger
		// chose.
		DisableKeepAlives: true,
	}
	return &fetcher{
		client: &http.Client{
			Transport: transport,
			// A redirect is an answer, and it is never followed: the address
			// it names was never checked, and a document that lives
			// elsewhere is not at the address the client is.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		cache:   newCache(),
		timeout: timeout,
		now:     now,
	}
}

// Fetch reads the document a client identifier is the address of, and keeps
// it while its answer allows.
func (f *fetcher) Fetch(ctx context.Context, clientID string) (Fetched, error) {
	if err := clientURL(clientID); err != nil {
		return Fetched{}, err
	}
	if document, fresh := f.cache.get(clientID, f.now()); fresh {
		return Fetched{Document: document, Cached: true}, nil
	}

	body, lifetime, err := f.get(ctx, clientID)
	if err != nil {
		return Fetched{}, err
	}
	document, err := parseDocument(body, clientID)
	if err != nil {
		return Fetched{}, err
	}
	now := f.now()
	f.cache.put(clientID, document, now, now.Add(lifetime))
	return Fetched{Document: document}, nil
}

// get reads the document at an address, and tries once more after a failure
// the next attempt might not meet: an address that did not answer, or not in
// time, or a server that failed. A refusal of ours, a redirect, a document too
// large and a certificate that does not name the address would all be the same
// the second time, and are not tried again.
func (f *fetcher) get(ctx context.Context, address string) (body []byte, lifetime time.Duration, err error) {
	for range attempts {
		var again bool
		body, lifetime, again, err = f.attempt(ctx, address)
		if err == nil || !again || ctx.Err() != nil {
			break
		}
	}
	return body, lifetime, err
}

// attempt reads the document once, within the time an attempt has, and says
// whether a failure is worth another attempt.
func (f *fetcher) attempt(ctx context.Context, address string) (body []byte, lifetime time.Duration, again bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return nil, 0, false, fmt.Errorf("%w: %w", ErrClientURL, err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := f.client.Do(request)
	if err != nil {
		return nil, 0, !final(err), unreached(err)
	}
	defer func() { _ = response.Body.Close() }()

	switch status := response.StatusCode; {
	case status == http.StatusOK:
	case status >= http.StatusMultipleChoices && status < http.StatusBadRequest:
		return nil, 0, false, fmt.Errorf("%w: %d", ErrRedirect, status)
	default:
		return nil, 0, status >= http.StatusInternalServerError, fmt.Errorf("%w: %d", ErrStatus, status)
	}
	body, err = io.ReadAll(io.LimitReader(response.Body, maxDocument+1))
	switch {
	case err != nil:
		return nil, 0, true, fmt.Errorf("%w: %w", ErrUnreachable, err)
	case len(body) > maxDocument:
		return nil, 0, false, ErrTooLarge
	}
	return body, lifetimeOf(response.Header), false, nil
}

// unreached is a request that got no answer. A refusal of ours stays what it
// is; anything else is an address that could not be reached.
func unreached(err error) error {
	if errors.Is(err, ErrAddress) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrUnreachable, err)
}

// final reports whether a request that got no answer would fail the same way
// the next time: refused by us, or over a connection whose certificate does
// not name the address.
func final(err error) bool {
	var certificate *tls.CertificateVerificationError
	return errors.Is(err, ErrAddress) || errors.As(err, &certificate)
}
