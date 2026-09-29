// Package serve serves the dashboard on 127.0.0.1 and rebuilds it when the
// plan changes.
//
// The server listens on the loopback address only and answers 403 to any
// request whose Host header is not 127.0.0.1, localhost or [::1] with the
// server's port, so a web page on another site cannot reach it through DNS
// rebinding. Only GET and HEAD are allowed.
//
// Routes:
//
//	/, /index.html  the dashboard page
//	/data.js        the latest build data (window.AJIYA)
//	/PROGRESS.md    the latest PROGRESS.md
//	/version        a short string that changes whenever data.js changes
//
// The page may poll /version (about once a second) and reload data.js when the
// string changes; that gives live refresh without a socket or network access
// beyond the server itself.
//
// A watcher polls the phase files, ajiya.toml and git HEAD for changes to
// their modification time or size, and rebuilds on a change. A rebuild also
// writes the build output into ajiya/, as 'ajiya build' does. When the config
// or a phase file cannot be read, the server keeps serving the last good data
// and logs the error.
package serve

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/build"
	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/dashboard"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// DefaultPort is the port 'ajiya serve' uses unless told otherwise.
const DefaultPort = 7390

// DefaultInterval is how often the watcher looks for changes.
const DefaultInterval = time.Second

// csp keeps the page from reaching anything but the server itself.
const csp = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"

// LoadFunc reads the config and plan and returns them with the phases in
// display order (nil for slug order).
type LoadFunc func() (*config.Config, *plan.Plan, []*plan.Phase, error)

// Options configure a Server.
type Options struct {
	Root     string        // project root
	Load     LoadFunc      // reads the project for each build
	Interval time.Duration // watcher poll interval; DefaultInterval if zero
	Log      io.Writer     // one line per rebuild or error; io.Discard if nil
}

// Server holds the latest build and serves it.
type Server struct {
	opts Options

	// The watcher and its last snapshot, taken before the first build so
	// that a change made during or after it is seen.
	w    *watcher
	last map[string]stamp

	mu       sync.RWMutex
	name     string // project name
	dataJS   []byte
	progress []byte
	version  string
}

// New builds the project once and returns a server for it. It fails if the
// first build fails, since there is nothing to serve.
func New(opts Options) (*Server, error) {
	if opts.Interval <= 0 {
		opts.Interval = DefaultInterval
	}
	if opts.Log == nil {
		opts.Log = io.Discard
	}
	s := &Server{opts: opts, w: newWatcher(opts.Root)}
	s.last = s.w.snapshot()
	if err := s.rebuild(true); err != nil {
		return nil, err
	}
	return s, nil
}

// Project returns the project name of the latest build.
func (s *Server) Project() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// Version returns the current version string.
func (s *Server) Version() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// rebuild reads the project, builds it and writes the output. A phase file
// that does not parse is an error unless first is set: then there is no
// earlier data to keep, so the rest of the plan is served as 'ajiya build'
// would write it.
func (s *Server) rebuild(first bool) error {
	cfg, p, order, err := s.opts.Load()
	if err != nil {
		return err
	}
	for slug, pe := range p.Broken {
		msg := fmt.Sprintf("%s:%d: %s", plan.Path(slug), pe.Line, pe.Msg)
		if !first {
			return errors.New(msg + "; still serving the last good data")
		}
		fmt.Fprintf(s.opts.Log, "Warning: %s; that phase is left out\n", msg)
	}
	d, err := build.Collect(cfg, p, order)
	if err != nil {
		return err
	}
	progress, dataJS, err := build.Files(d)
	if err != nil {
		return err
	}
	if _, err := build.Write(s.opts.Root, d); err != nil {
		return err
	}
	sum := sha256.Sum256(dataJS)
	s.mu.Lock()
	s.name, s.dataJS, s.progress, s.version = cfg.Project.Name, dataJS, progress, hex.EncodeToString(sum[:6])
	s.mu.Unlock()
	return nil
}

// Serve serves on ln until ctx is cancelled, then shuts down and returns nil.
// It closes ln.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	tcp, ok := ln.Addr().(*net.TCPAddr)
	if !ok || !tcp.IP.IsLoopback() {
		ln.Close()
		return fmt.Errorf("refusing to serve on %s: only a loopback address is allowed", ln.Addr())
	}
	hs := &http.Server{
		Handler:           s.Handler(tcp.Port),
		ReadHeaderTimeout: 10 * time.Second,
	}
	watchCtx, stopWatch := context.WithCancel(ctx)
	defer stopWatch()
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.watch(watchCtx)
	}()

	errc := make(chan error, 1)
	go func() { errc <- hs.Serve(ln) }()
	select {
	case err := <-errc:
		stopWatch()
		<-done
		return err
	case <-ctx.Done():
	}
	shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := hs.Shutdown(shutCtx)
	<-errc
	stopWatch()
	<-done
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Handler returns the HTTP handler for a server listening on port.
func (s *Server) Handler(port int) http.Handler {
	p := strconv.Itoa(port)
	allowed := map[string]bool{"127.0.0.1:" + p: true, "localhost:" + p: true, "[::1]:" + p: true}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")
		if !allowed[strings.ToLower(r.Host)] {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			h.Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body []byte
		switch r.URL.Path {
		case "/", "/index.html":
			h.Set("Content-Type", "text/html; charset=utf-8")
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Frame-Options", "DENY")
			body = dashboard.HTML()
		case "/data.js":
			h.Set("Content-Type", "text/javascript; charset=utf-8")
			s.mu.RLock()
			body = s.dataJS
			s.mu.RUnlock()
		case "/PROGRESS.md":
			h.Set("Content-Type", "text/markdown; charset=utf-8")
			s.mu.RLock()
			body = s.progress
			s.mu.RUnlock()
		case "/version":
			h.Set("Content-Type", "text/plain; charset=utf-8")
			body = []byte(s.Version())
		default:
			http.NotFound(w, r)
			return
		}
		h.Set("Content-Length", strconv.Itoa(len(body)))
		if r.Method == http.MethodHead {
			return
		}
		w.Write(body)
	})
}
