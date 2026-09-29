package check

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/config"
	"github.com/AliyuYahaya/Ajiya/internal/plan"
	"github.com/rogpeppe/go-internal/txtar"
)

var update = flag.Bool("update", false, "rewrite the want section of each fixture")

// Each testdata/*.txtar is a small project and, in its "want" file, the
// findings it must produce, one per line.
func TestFixtures(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.txtar")
	if len(files) == 0 {
		t.Fatal("no fixtures")
	}
	for _, f := range files {
		t.Run(strings.TrimSuffix(filepath.Base(f), ".txtar"), func(t *testing.T) {
			ar, err := txtar.ParseFile(f)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			want, wantIdx := "", -1
			for i, file := range ar.Files {
				if file.Name == "want" {
					want, wantIdx = string(file.Data), i
					continue
				}
				path := filepath.Join(root, filepath.FromSlash(file.Name))
				os.MkdirAll(filepath.Dir(path), 0o755)
				if err := os.WriteFile(path, file.Data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			p, err := plan.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			var b strings.Builder
			for _, finding := range Run(cfg, p) {
				b.WriteString(finding.String() + "\n")
			}
			got := b.String()
			if *update {
				if wantIdx < 0 {
					ar.Files = append(ar.Files, txtar.File{Name: "want"})
					wantIdx = len(ar.Files) - 1
				}
				ar.Files[wantIdx].Data = []byte(got)
				os.WriteFile(f, txtar.Format(ar), 0o644)
				return
			}
			if got != want {
				t.Errorf("findings:\n%s\nwant:\n%s", got, want)
			}
			// The fixture named after a code (E012, or E012-milestones for a
			// second case) must produce that code.
			code, _, _ := strings.Cut(strings.TrimSuffix(filepath.Base(f), ".txtar"), "-")
			if code != "clean" && !strings.Contains(got, code+" ") {
				t.Errorf("fixture %s does not produce %s", f, code)
			}
		})
	}
}
