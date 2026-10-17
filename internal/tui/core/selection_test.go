package tui

import "testing"

func TestSelectionColRange(t *testing.T) {
	s := selection{anchor: point{3, 5}, head: point{1, 2}, active: true}
	cases := []struct {
		line, s, e int
		ok         bool
	}{
		{0, 0, 0, false},
		{1, 2, 10, true},
		{2, 0, 10, true},
		{3, 0, 6, true},
		{4, 0, 0, false},
	}
	for _, c := range cases {
		s0, e0, ok := s.colRange(c.line, 10)
		if ok != c.ok || (ok && (s0 != c.s || e0 != c.e)) {
			t.Errorf("line %d: got %d,%d,%v want %d,%d,%v", c.line, s0, e0, ok, c.s, c.e, c.ok)
		}
	}
}

func TestSoftWrapped(t *testing.T) {
	if !softWrapped(" the quick brown fox", " jumps over", 24) {
		t.Error("full line followed by a word that did not fit should be joined")
	}
	if softWrapped(" short", " next paragraph", 24) {
		t.Error("short line is a real break")
	}
	if softWrapped("", " word", 24) || softWrapped(" word", "", 24) {
		t.Error("blank lines are real breaks")
	}
}
