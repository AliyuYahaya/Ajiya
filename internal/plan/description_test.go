package plan

import (
	"os"
	"strings"
	"testing"
)

func TestDescription(t *testing.T) {
	data, _ := os.ReadFile("testdata/golden/described.md")
	ph, err := ParsePhase("bookings", data)
	if err != nil {
		t.Fatal(err)
	}
	want := "Patients pick a clinic, a day and a slot.\nCancelling frees the slot at once.\n\nOut of scope: payments, which come in a later phase."
	if ph.Description != want || len(ph.Tickets) != 1 {
		t.Fatalf("description = %q, tickets %d", ph.Description, len(ph.Tickets))
	}
	ph.Description = ""
	if strings.Contains(string(ph.Format()), "Patients pick") {
		t.Error("a cleared description is still written")
	}
	plain, _ := os.ReadFile("testdata/golden/bookings.md")
	if p2, _ := ParsePhase("x", plain); p2.Description != "" {
		t.Errorf("a file without a description reads %q", p2.Description)
	}
}

func TestNormalizeDescription(t *testing.T) {
	got, err := NormalizeDescription("\n\n a  \n\n\n\nb \n\n")
	if err != nil || got != "a\n\nb" {
		t.Errorf("got %q, %v", got, err)
	}
	for _, bad := range []string{"x\n| a | b |", "Goal: y", "a\rb"} {
		if _, err := NormalizeDescription(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
