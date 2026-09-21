package main

import (
	"context"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
)

// run must stop because its context was cancelled — the same way a signal
// reaches it — and it must stop without an error.
func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	port := freePort(t)
	t.Setenv("PORT", port)
	t.Setenv("MATHTRAIL_LOG_LEVEL", "error")
	t.Setenv("MATHTRAIL_SHUTDOWN_TIMEOUT", "2s")

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- run(ctx) }()

	// Cancel only once the service is actually listening, so that a pass
	// means "it started and then stopped" rather than "it never started".
	waitUntilListening(t, port)
	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("run() error = %v, want nil", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run() did not return within 10s of the context being cancelled")
	}
}

// A configuration it cannot satisfy is reported, not worked around: nothing is
// started and the error says which variable is wrong.
func TestRunRefusesBadConfiguration(t *testing.T) {
	t.Setenv("PORT", "eighty")

	err := run(t.Context())
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("run() error = %v, want it to wrap config.ErrInvalid", err)
	}
}

// freePort asks the kernel for a port and gives it straight back, so the test
// binds a port nobody else is using rather than a number somebody guessed.
func freePort(t *testing.T) string {
	t.Helper()

	var lc net.ListenConfig
	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for a free port: %v", err)
	}
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address = %T, want *net.TCPAddr", listener.Addr())
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close the probe listener: %v", err)
	}
	return strconv.Itoa(addr.Port)
}

func waitUntilListening(t *testing.T, port string) {
	t.Helper()

	var dialer net.Dialer
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := dialer.DialContext(t.Context(), "tcp", net.JoinHostPort("127.0.0.1", port))
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("nothing is listening on port %s after 10s", port)
}
