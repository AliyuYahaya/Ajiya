package cli

import (
	"fmt"
	"path/filepath"

	"github.com/AliyuYahaya/Ajiya/internal/gitx"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

// filesByTicket returns, for every ticket named by a commit in any of the
// project's repositories, the files those commits changed, from one pass over
// each repository's history. Plan files (the ajiya/ folder) are left out: they
// say nothing about the code a ticket touches. Tickets without commits are
// absent. Callers that need many tickets, such as change impact, use this
// rather than one lookup per ticket.
func (pr *project) filesByTicket() (map[string]*gitx.TicketFiles, error) {
	m, err := gitx.FilesByTicket(pr.repos(), []string{filepath.Join(pr.cfg.Root, plan.Dir)})
	if err == gitx.ErrNotRepo {
		return map[string]*gitx.TicketFiles{}, nil
	}
	return m, err
}

// displayPath names a file for people: its path from the project root when the
// file is in an app repository, and the bare path in the root repository.
func (pr *project) displayPath(f gitx.RepoFile) string {
	if l := pr.repoLabel(f.Repo); l != "" {
		return l + "/" + f.Path
	}
	return f.Path
}

func runTicketFiles(e *env, args []string) error {
	fs := newFlags("ticket files")
	asJSON := fs.Bool("json", false, "print JSON")
	pos, err := parse(fs, args, 1)
	if err != nil {
		return err
	}
	pr, err := load(e)
	if err != nil {
		return err
	}
	t, err := pr.ticket(pos[0])
	if err != nil {
		return err
	}
	all, err := pr.filesByTicket()
	if err != nil {
		return err
	}
	tf := all[t.ID]
	if tf == nil {
		tf = &gitx.TicketFiles{}
	}
	if *asJSON {
		type file struct {
			Path string `json:"path"` // from the project root
			Repo string `json:"repo,omitempty"`
		}
		files := []file{}
		for _, f := range tf.Files {
			files = append(files, file{pr.displayPath(f), pr.repoLabel(f.Repo)})
		}
		return writeJSON(e, struct {
			ID      string `json:"id"`
			Commits int    `json:"commits"`
			Files   []file `json:"files"`
		}{t.ID, tf.Commits, files})
	}
	if tf.Commits == 0 {
		fmt.Fprintf(e.stdout, "No commit references %s, so there are no files to list.\n", t.ID)
		return nil
	}
	if len(tf.Files) == 0 {
		fmt.Fprintf(e.stdout, "%d commit(s) reference %s, but none changed a file outside the plan folder.\n", tf.Commits, t.ID)
		return nil
	}
	for _, f := range tf.Files {
		fmt.Fprintln(e.stdout, pr.displayPath(f))
	}
	return nil
}
