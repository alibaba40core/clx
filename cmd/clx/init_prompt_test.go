package main

import "testing"

func TestParseInitChoice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		line    string
		n       int
		def     int
		want    int
		wantOK  bool
	}{
		{"", 4, 0, 0, true},
		{"3", 4, 0, 2, true},
		{"99", 4, 0, 0, false},
		{"AQ.secret", 4, 0, 0, false},
		{"abc", 4, 0, 0, false},
	}
	for _, tc := range tests {
		got, ok := parseInitChoice(tc.line, tc.n, tc.def)
		if ok != tc.wantOK || got != tc.want {
			t.Fatalf("parseInitChoice(%q) = %d,%v want %d,%v", tc.line, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestParseYesNoInput(t *testing.T) {
	t.Parallel()
	yes, ok := parseYesNoInput("y", true)
	if !ok || !yes {
		t.Fatalf("y: got %v,%v", yes, ok)
	}
	yes, ok = parseYesNoInput("", true)
	if !ok || yes {
		t.Fatalf("empty defaultNo: got %v,%v", yes, ok)
	}
	_, ok = parseYesNoInput("AQ.Ab8RN6I9XdyVULaAY_4C5w", true)
	if ok {
		t.Fatal("api key paste must be rejected on yes/no prompt")
	}
}
