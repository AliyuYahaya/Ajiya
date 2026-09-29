package serve

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/dashboard"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

const phaseHead = `<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Core

Goal: the core works

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
`

func phase(tickets ...string) string {
	var b strings.Builder
	b.WriteString(phaseHead)
	for i, t := range tickets {
		fmt.Fprintf(&b, "| TP-%04d | infra | %s | it works | - | 🟥 Pending |\n", i+1, t)
	}
	return b.String()
}

// syncBuffer is a log that tests read while the watcher writes.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

type fixture struct {
	root string
	base string // http://127.0.0.1:<port>
	port int
	srv  *Server
	log  *syncBuffer
}

func (f *fixture) write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.root, filepath.FromSlash(name)), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// start makes a project in a temp directory and serves it on 127.0.0.1:0.
func start(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{root: t.TempDir(), log: &syncBuffer{}}
	if err := os.Mkdir(filepath.Join(f.root, plan.Dir), 0o755); err != nil {
		t.Fatal(err)
	}
	f.write(t, config.FileName, config.Template("Test Project", "TP"))
	f.write(t, "ajiya/core.md", phase("First ticket"))

	srv, err := New(Options{
		Root:     f.root,
		Interval: 20 * time.Millisecond,
		Log:      f.log,
		Load: func() (*config.Config, *plan.Plan, []*plan.Phase, error) {
			cfg, err := config.Load(f.root)
			if err != nil {
				return nil, nil, nil, err
			}
			p, err := plan.Load(f.root)
			if err != nil {
				return nil, nil, nil, err
			}
			return cfg, p, nil, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.srv = srv
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f.port = ln.Addr().(*net.TCPAddr).Port
	f.base = fmt.Sprintf("http://127.0.0.1:%d", f.port)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx, ln) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Serve: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("Serve did not stop")
		}
	})
	return f
}

func (f *fixture) do(t *testing.T, method, path, host string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, f.base+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if host != "" {
		req.Host = host
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func (f *fixture) get(t *testing.T, path string) (*http.Response, string) {
	return f.do(t, http.MethodGet, path, "")
}

func TestHost(t *testing.T) {
	f := start(t)
	p := fmt.Sprint(f.port)
	tests := []struct {
		host string
		want int
	}{
		{"127.0.0.1:" + p, 200},
		{"localhost:" + p, 200},
		{"LocalHost:" + p, 200},
		{"[::1]:" + p, 200},
		{"evil.example:" + p, 403},
		{"evil.example", 403},
		{"127.0.0.1", 403},
		{"localhost:1", 403},
		{"127.0.0.1.evil.example:" + p, 403},
	}
	for _, tt := range tests {
		resp, _ := f.do(t, http.MethodGet, "/", tt.host)
		if resp.StatusCode != tt.want {
			t.Errorf("Host %q: status %d, want %d", tt.host, resp.StatusCode, tt.want)
		}
	}
	// A forbidden host gets 403 even on a route that does not exist.
	if resp, _ := f.do(t, http.MethodGet, "/nope", "evil.example:"+p); resp.StatusCode != 403 {
		t.Errorf("unknown route with bad host: status %d, want 403", resp.StatusCode)
	}
}

func TestMissingHost(t *testing.T) {
	f := start(t)
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", f.port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	fmt.Fprint(conn, "GET / HTTP/1.0\r\n\r\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Errorf("no Host: status %d, want 403", resp.StatusCode)
	}
}

func TestMethod(t *testing.T) {
	f := start(t)
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions} {
		resp, _ := f.do(t, m, "/", "")
		if resp.StatusCode != 405 {
			t.Errorf("%s: status %d, want 405", m, resp.StatusCode)
		}
		if resp.Header.Get("Allow") != "GET, HEAD" {
			t.Errorf("%s: Allow = %q", m, resp.Header.Get("Allow"))
		}
	}
	resp, body := f.do(t, http.MethodHead, "/data.js", "")
	if resp.StatusCode != 200 || body != "" {
		t.Errorf("HEAD: status %d, body %q", resp.StatusCode, body)
	}
}

func TestRoutes(t *testing.T) {
	f := start(t)
	tests := []struct {
		path, ctype, contains string
	}{
		{"/", "text/html; charset=utf-8", string(dashboard.HTML()[:min(40, len(dashboard.HTML()))])},
		{"/index.html", "text/html; charset=utf-8", ""},
		{"/data.js", "text/javascript; charset=utf-8", `"name": "Test Project"`},
		{"/PROGRESS.md", "text/markdown; charset=utf-8", "First ticket"},
		{"/version", "text/plain; charset=utf-8", f.srv.Version()},
	}
	for _, tt := range tests {
		resp, body := f.get(t, tt.path)
		if resp.StatusCode != 200 {
			t.Errorf("%s: status %d", tt.path, resp.StatusCode)
			continue
		}
		if got := resp.Header.Get("Content-Type"); got != tt.ctype {
			t.Errorf("%s: Content-Type %q, want %q", tt.path, got, tt.ctype)
		}
		if !strings.Contains(body, tt.contains) {
			t.Errorf("%s: body does not contain %q", tt.path, tt.contains)
		}
		if resp.Header.Get("Cache-Control") != "no-store" || resp.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: headers %v", tt.path, resp.Header)
		}
	}
	resp, _ := f.get(t, "/")
	if c := resp.Header.Get("Content-Security-Policy"); !strings.Contains(c, "default-src 'self'") || !strings.Contains(c, "connect-src 'self'") {
		t.Errorf("CSP = %q", c)
	}
	for _, p := range []string{"/nope", "/ajiya.toml", "/core.md", "/../ajiya.toml", "/data.js/"} {
		if resp, _ := f.get(t, p); resp.StatusCode != 404 {
			t.Errorf("%s: status %d, want 404", p, resp.StatusCode)
		}
	}
	// The build output was written, as 'ajiya build' does.
	if _, err := os.Stat(filepath.Join(f.root, "ajiya", "data.js")); err != nil {
		t.Error(err)
	}
}

// waitVersion polls /version until it differs from old.
func (f *fixture) waitVersion(t *testing.T, old string) string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, v := f.get(t, "/version"); v != old {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("/version still %q; log:\n%s", old, f.log.String())
	return ""
}

// waitLog polls the log until it contains s.
func (f *fixture) waitLog(t *testing.T, s string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(f.log.String(), s) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("log has no %q:\n%s", s, f.log.String())
}

func TestRebuild(t *testing.T) {
	f := start(t)
	_, v1 := f.get(t, "/version")
	f.write(t, "ajiya/core.md", phase("First ticket", "Second ticket"))
	f.waitVersion(t, v1)
	if _, body := f.get(t, "/data.js"); !strings.Contains(body, "Second ticket") {
		t.Errorf("data.js has no new ticket")
	}
	f.waitLog(t, "Rebuilt: ajiya/core.md changed")
	disk, err := os.ReadFile(filepath.Join(f.root, "ajiya", "data.js"))
	if err != nil || !bytes.Contains(disk, []byte("Second ticket")) {
		t.Errorf("ajiya/data.js not rewritten: %v", err)
	}
	if strings.Contains(f.log.String(), "PROGRESS.md") || strings.Contains(f.log.String(), "data.js") {
		t.Errorf("build output triggered a rebuild:\n%s", f.log.String())
	}
}

func TestBrokenKeepsLastData(t *testing.T) {
	f := start(t)
	_, v1 := f.get(t, "/version")
	_, data1 := f.get(t, "/data.js")

	f.write(t, "ajiya/core.md", phaseHead+"| TP-0001 | not a row\n")
	f.waitLog(t, "Rebuild failed: ajiya/core.md:")
	if _, v := f.get(t, "/version"); v != v1 {
		t.Errorf("version changed after a broken phase file")
	}
	if _, d := f.get(t, "/data.js"); d != data1 {
		t.Errorf("data.js changed after a broken phase file")
	}

	f.write(t, config.FileName, "[project\nname = ")
	f.waitLog(t, "Rebuild failed: ")
	if _, d := f.get(t, "/data.js"); d != data1 {
		t.Errorf("data.js changed after a broken config")
	}

	// Fixing both brings rebuilds back.
	f.write(t, config.FileName, config.Template("Test Project", "TP"))
	f.write(t, "ajiya/core.md", phase("First ticket", "Fixed ticket"))
	f.waitVersion(t, v1)
	if _, d := f.get(t, "/data.js"); !strings.Contains(d, "Fixed ticket") {
		t.Errorf("data.js not rebuilt after the fix")
	}
}

func TestServeRefusesNonLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Skip(err)
	}
	s := &Server{opts: Options{Interval: time.Second}}
	if err := s.Serve(context.Background(), ln); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Errorf("Serve on %s: err = %v", ln.Addr(), err)
	}
}

func TestChanged(t *testing.T) {
	now := time.Now()
	old := map[string]stamp{"a": {now, 1}, "b": {now, 2}}
	cur := map[string]stamp{"a": {now, 1}, "b": {now, 3}, "c": {now, 1}}
	if got := strings.Join(changed(old, cur), " "); got != "b c" {
		t.Errorf("changed = %q", got)
	}
	if got := strings.Join(changed(cur, old), " "); got != "b c" {
		t.Errorf("changed back = %q", got)
	}
	if got := reason([]string{"a", "b", "c", "d", "e"}); got != "a, b, c and 2 more changed" {
		t.Errorf("reason = %q", got)
	}
}

func TestWatcherGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=T", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("commit", "-q", "--allow-empty", "-m", "one")
	w := newWatcher(root)
	before := w.snapshot()
	if _, ok := before["git refs/heads/main"]; !ok {
		t.Fatalf("HEAD's ref not watched: %v", before)
	}
	git("commit", "-q", "--allow-empty", "-m", "two")
	if got := changed(before, w.snapshot()); !slices.Contains(got, "git refs/heads/main") {
		t.Errorf("a commit changed %v", got)
	}
	before = w.snapshot()
	git("checkout", "-q", "-b", "other")
	if got := changed(before, w.snapshot()); !slices.Contains(got, "git HEAD") {
		t.Errorf("a checkout changed %v", got)
	}
}
