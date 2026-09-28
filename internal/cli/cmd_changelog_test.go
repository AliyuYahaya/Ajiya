package cli

import "testing"

func TestChangelogJSON(t *testing.T) {
	dir := t.TempDir()
	for k, v := range map[string]string{"HOME": dir, "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "T",
		"GIT_AUTHOR_EMAIL": "t@example.com", "GIT_COMMITTER_NAME": "T", "GIT_COMMITTER_EMAIL": "t@example.com"} {
		t.Setenv(k, v)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	ajiya(t, dir, 0, "init", "--name", "Demo", "--prefix", "DE")

	type commit struct{ Hash, Date, Subject string }
	type changelog struct {
		Phases []struct {
			Slug, Title string
			Tickets     []struct {
				ID, Title, State string
				Commits          []commit
			}
		}
		Other []commit
	}

	// No commits: empty lists, not null.
	if got := string(ajiya(t, dir, 0, "changelog", "--json")); got != "{\n  \"phases\": [],\n  \"other\": []\n}\n" {
		t.Errorf("changelog --json with no commits = %q", got)
	}

	ajiya(t, dir, 0, "phase", "add", "core", "Core works")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "core", "--app", "infra", "--done-when", "x", "One")
	ajiya(t, dir, 0, "ticket", "add", "--phase", "core", "--app", "infra", "--done-when", "x", "Two")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "Plan", "-m", "Ajiya: chore")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "Both", "-m", "Ajiya: DE-0001, DE-0002")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "Second", "-m", "Ajiya: DE-0002")

	var cl changelog
	decode(t, ajiya(t, dir, 0, "changelog", "--json"), &cl)
	if len(cl.Phases) != 1 || cl.Phases[0].Slug != "core" || cl.Phases[0].Title != "Core" || len(cl.Phases[0].Tickets) != 2 {
		t.Fatalf("changelog --json = %+v", cl)
	}
	one, two := cl.Phases[0].Tickets[0], cl.Phases[0].Tickets[1]
	if one.ID != "DE-0001" || one.Title != "One" || one.State != "pending" || len(one.Commits) != 1 || one.Commits[0].Subject != "Both" {
		t.Errorf("first ticket = %+v", one)
	}
	if two.ID != "DE-0002" || len(two.Commits) != 2 || two.Commits[0].Subject != "Both" || two.Commits[1].Subject != "Second" ||
		len(two.Commits[1].Hash) != 40 || len(two.Commits[1].Date) != 10 {
		t.Errorf("second ticket = %+v", two)
	}
	if len(cl.Other) != 1 || cl.Other[0].Subject != "Plan" {
		t.Errorf("other = %+v", cl.Other)
	}

	// A range with only a chore has no phases.
	cl = changelog{}
	decode(t, ajiya(t, dir, 0, "changelog", "HEAD~2", "--json"), &cl)
	if cl.Phases == nil || len(cl.Phases) != 0 || len(cl.Other) != 1 {
		t.Errorf("changelog --json on a chore range = %+v", cl)
	}
	ajiya(t, dir, 2, "changelog", "nope..HEAD", "--json")
}
