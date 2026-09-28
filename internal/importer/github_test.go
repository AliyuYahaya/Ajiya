package importer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/AliyuYahaya/Ajiya/internal/plan"
)

func TestGitHubDecode(t *testing.T) {
	for _, tc := range []struct {
		name, in string
		want     []int  // issue numbers, in order
		err      string // part of the error, or ""
	}{
		{"empty", `[]`, nil, ""},
		{"sorted", `[
			{"number":3,"title":"C","state":"CLOSED","url":"https://github.com/o/r/issues/3"},
			{"number":1,"title":"A","state":"OPEN","url":"https://github.com/o/r/issues/1"},
			{"number":2,"title":"B","state":"OPEN","url":"https://github.com/o/r/issues/2","extra":true}
		]`, []int{1, 2, 3}, ""},
		{"not json", `oops`, nil, "could not read the issue list from gh"},
		{"not an array", `{"number":1}`, nil, "could not read the issue list from gh"},
		{"no number", `[{"title":"A","state":"OPEN","url":"u"}]`, nil, "without a number"},
		{"no url", `[{"number":4,"title":"A","state":"OPEN"}]`, nil, "issue #4 without a URL"},
		{"bad state", `[{"number":5,"title":"A","state":"MERGED","url":"u"}]`, nil, `state "MERGED"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GitHubDecode([]byte(tc.in))
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var nums []int
			for _, is := range got {
				nums = append(nums, is.Number)
			}
			if !reflect.DeepEqual(nums, tc.want) {
				t.Errorf("numbers = %v, want %v", nums, tc.want)
			}
		})
	}
}

func TestGitHubDecodeKeepsTitle(t *testing.T) {
	got, err := GitHubDecode([]byte(`[{"number":1,"title":"Fix | pipes  and \"quotes\"","state":"OPEN","url":"https://x/1"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Title != `Fix | pipes  and "quotes"` {
		t.Errorf("title = %q", got[0].Title)
	}
}

func TestGitHubStatus(t *testing.T) {
	for _, tc := range []struct {
		is   GitHubIssue
		want string
	}{
		{GitHubIssue{7, "Open one", "OPEN", "https://github.com/o/r/issues/7"},
			"🟥 Pending · Link: https://github.com/o/r/issues/7"},
		{GitHubIssue{8, "Closed one", "CLOSED", "https://github.com/o/r/issues/8"},
			"🟩 Done · issue #8 · Link: https://github.com/o/r/issues/8"},
	} {
		s := GitHubStatus(tc.is)
		if got := s.String(); got != tc.want {
			t.Errorf("GitHubStatus(#%d) = %q, want %q", tc.is.Number, got, tc.want)
		}
		back, err := plan.ParseStatus(s.String())
		if err != nil || !reflect.DeepEqual(back, s) {
			t.Errorf("round trip of %q = %+v, %v", s, back, err)
		}
		if s.State == plan.Done && !s.HasEvidence() {
			t.Errorf("closed issue %d has no evidence", tc.is.Number)
		}
	}
}

func TestGitHubCheck(t *testing.T) {
	for _, tc := range []struct {
		is  GitHubIssue
		err string
	}{
		{GitHubIssue{1, "Fine title", "OPEN", "https://x/1"}, ""},
		{GitHubIssue{2, "Two\nlines", "OPEN", "https://x/2"}, "one line"},
		{GitHubIssue{3, "", "OPEN", "https://x/3"}, "empty"},
		{GitHubIssue{4, " padded", "OPEN", "https://x/4"}, "spaces"},
		{GitHubIssue{5, "Fine", "OPEN", "https://x/a · b"}, "'·'"},
	} {
		err := GitHubCheck(tc.is)
		if tc.err == "" {
			if err != nil {
				t.Errorf("#%d: %v", tc.is.Number, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.err) {
			t.Errorf("#%d: err = %v, want %q", tc.is.Number, err, tc.err)
		}
	}
}
