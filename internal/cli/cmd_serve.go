package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/AliyuYahaya/Ajiya/internal/serve"
)

// runServe serves the dashboard on 127.0.0.1 and rebuilds it when files change.
func runServe(e *env, args []string) error {
	fs := newFlags("serve")
	port := fs.Int("port", serve.DefaultPort, "port on 127.0.0.1 (0 picks a free one)")
	if _, err := parse(fs, args, 0); err != nil {
		return err
	}
	if *port < 0 || *port > 65535 {
		return usageErr("--port must be from 0 to 65535")
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	root := pr.cfg.Root
	s, err := serve.New(serve.Options{
		Root: root,
		Load: func() (*config.Config, *plan.Plan, []*plan.Phase, error) {
			// Read from the root found at start, so a missing ajiya.toml is an
			// error rather than a switch to a parent project.
			cfg, err := config.Load(root)
			if err != nil {
				return nil, nil, nil, err
			}
			p, err := plan.Load(root)
			if err != nil {
				return nil, nil, nil, err
			}
			pr := &project{cfg: cfg, plan: p}
			return cfg, p, pr.displayPhases(), nil
		},
		Log: e.stdout,
	})
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(*port)))
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			return refused("port %d is in use; pick another with --port <n>, or --port 0 for a free one", *port)
		}
		return err
	}
	fmt.Fprintf(e.stdout, "Serving %s at http://%s/ (Ctrl+C to stop)\n", s.Project(), ln.Addr())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := s.Serve(ctx, ln); err != nil {
		return err
	}
	fmt.Fprintln(e.stdout, "Stopped.")
	return nil
}
