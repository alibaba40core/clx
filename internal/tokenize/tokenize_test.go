package tokenize

import "testing"

func TestTokenizeQuotes(t *testing.T) {
	t.Parallel()
	toks, err := Tokenize(`grep "error message" logs.txt`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grep", "error message", "logs.txt"}
	if len(toks) != len(want) {
		t.Fatalf("got %v", toks)
	}
	for i := range want {
		if toks[i] != want[i] {
			t.Fatalf("[%d] got %q want %q", i, toks[i], want[i])
		}
	}
}
