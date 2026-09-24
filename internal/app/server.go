package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Server is the HTTP server and its graceful shutdown.
type Server struct {
	http            *http.Server
	logger          *zap.Logger
	shutdownTimeout time.Duration

	// mu guards listener: Run binds it on one goroutine while Addr may be
	// read from another, which is how a caller finds out which port the
	// kernel picked.
	mu       sync.RWMutex
	listener net.Listener
}

// NewServer builds the server from a wired container.
func NewServer(c *Container) *Server {
	return &Server{
		logger:          c.Logger,
		shutdownTimeout: c.Config.ShutdownTimeout,
		http: &http.Server{
			// Empty host, so the server listens on every interface. The
			// configuration guarantees a bare port number, and JoinHostPort is
			// what keeps the address well formed if that ever stops being true.
			Addr:              net.JoinHostPort("", c.Config.Port),
			Handler:           c.Router,
			ReadHeaderTimeout: c.Config.ReadHeaderTimeout,
			ReadTimeout:       c.Config.ReadTimeout,
			WriteTimeout:      c.Config.WriteTimeout,
			IdleTimeout:       c.Config.IdleTimeout,
		},
	}
}

// Listen binds the port without serving on it yet. Run does this itself, so
// only a test that needs the real address before the server starts has to call
// it; binding early is also how a port that is already taken is reported at
// startup rather than swallowed by the goroutine inside Run.
func (s *Server) Listen(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil
	}
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("app: listen on %s: %w", s.http.Addr, err)
	}
	s.listener = listener
	return nil
}

// Addr is the address the server is bound to, empty before Listen. It is safe
// to call while the server is starting or serving.
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// Run serves until ctx is cancelled, then shuts down gracefully. Shutdown is
// driven by the context rather than by a signal handler of its own, so that
// every part of the process stops for the same reason at the same moment.
//
// It returns only once the goroutine it serves on has finished, so every line
// Run itself logs is written by then. A request still being handled when the
// shutdown deadline passes is cut off rather than waited for.
func (s *Server) Run(ctx context.Context) error {
	if err := s.Listen(ctx); err != nil {
		return err
	}

	s.mu.RLock()
	listener := s.listener
	s.mu.RUnlock()

	// The port is bound already, so this is true before serving starts, and
	// written here it cannot land after Run has returned.
	s.logger.Info("listening", zap.String("addr", listener.Addr().String()))
	served := make(chan error, 1)
	go func() { served <- s.http.Serve(listener) }()

	select {
	case err := <-served:
		// Nothing had asked it to stop, so whatever it returned is a failure.
		return fmt.Errorf("app: serve: %w", err)
	case <-ctx.Done():
		s.logger.Info("shutdown requested")
	}

	// A fresh context, so the shutdown deadline starts now rather than at the
	// moment the signal arrived and cancelled the parent.
	shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.shutdownTimeout)
	defer cancel()

	err := s.http.Shutdown(shutdown)
	if err != nil {
		_ = s.http.Close() // the deadline passed: drop what is left rather than hang
	}
	// Shutdown and Close both make Serve return, even one that had not begun
	// yet, so this wait is short.
	<-served
	if err != nil {
		return fmt.Errorf("app: graceful shutdown: %w", err)
	}

	s.logger.Info("stopped")
	return nil
}
