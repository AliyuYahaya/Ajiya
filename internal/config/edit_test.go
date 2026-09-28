package config

import "testing"

func TestSetLaunchText(t *testing.T) {
	const head = "[project]\nname = \"X\"\nprefix = \"X\"\n"
	tests := []struct{ name, in, want string }{
		{"replace, keeping comment",
			head + "\n[launch]\ntarget = 'old'  # the go-live phase\n",
			head + "\n[launch]\ntarget = \"v0-1\"  # the go-live phase\n"},
		{"table without target",
			head + "\n[launch]\n\n[test]\ncommand = \"go test\"\n",
			head + "\n[launch]\ntarget = \"v0-1\"\n\n[test]\ncommand = \"go test\"\n"},
		{"no table, apps follow",
			head + "\n[[apps]]\nname = \"a\"\npath = \"a\"\n",
			head + "\n[launch]\ntarget = \"v0-1\"\n\n[[apps]]\nname = \"a\"\npath = \"a\"\n"},
		{"no table, nothing follows",
			head,
			head + "\n[launch]\ntarget = \"v0-1\"\n"},
		{"commented template is left alone",
			head + "\n# [launch]\n# target = \"go-live\"\n",
			head + "\n# [launch]\n# target = \"go-live\"\n\n[launch]\ntarget = \"v0-1\"\n"},
		{"target in another table is not touched",
			head + "\n[test]\ntarget = \"x\"\n",
			head + "\n[test]\ntarget = \"x\"\n\n[launch]\ntarget = \"v0-1\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetLaunchText(tt.in, "v0-1")
			if got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
