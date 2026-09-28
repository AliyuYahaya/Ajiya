package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/detect"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"no args", nil, ExitUsage, "", "usage: ajiya"},
		{"version", []string{"version"}, ExitOK, "ajiya dev", ""},
		{"help", []string{"help"}, ExitOK, "usage: ajiya", ""},
		{"unknown", []string{"frobnicate"}, ExitUsage, "", `unknown command "frobnicate"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := Run(tt.args, &out, &errOut)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(out.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", out.String(), tt.wantStdout)
			}
			if !strings.Contains(errOut.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", errOut.String(), tt.wantStderr)
			}
		})
	}
}

func TestChooseAppsInteractive(t *testing.T) {
	cands := []detect.Candidate{{Name: "web", Path: "apps/web"}, {Name: "api", Path: "apps/api"}, {Name: "ui", Path: "packages/ui"}}
	tests := []struct {
		answer string
		want   string
		err    string
	}{
		{"\n", "web api ui", ""},
		{"all\n", "web api ui", ""},
		{"none\n", "", ""},
		{"3, 1 1\n", "ui web", ""},
		{"4\n", "", "not a number from 1 to 3"},
		{"", "web api ui", ""}, // end of input: the default
	}
	for _, tt := range tests {
		var out bytes.Buffer
		e := &env{stdin: strings.NewReader(tt.answer), stdout: &out, stderr: &out, interactive: true}
		got, err := chooseApps(e, cands, false)
		if tt.err != "" {
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("%q: err = %v, want %q", tt.answer, err, tt.err)
			}
			continue
		}
		var names []string
		for _, c := range got {
			names = append(names, c.Name)
		}
		if err != nil || strings.Join(names, " ") != tt.want {
			t.Errorf("%q: got %v, %v; want %s", tt.answer, names, err, tt.want)
		}
		if !strings.Contains(out.String(), "Register which?") {
			t.Errorf("%q: no prompt in %q", tt.answer, out.String())
		}
	}
}
