package parser

import (
	"strings"
	"testing"
)

func TestTokenizeQuotes(t *testing.T) {
	t.Parallel()
	res, err := tokenizeInput(`grep "error message" logs.txt`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grep", "error message", "logs.txt"}
	if len(res.tokens) != len(want) {
		t.Fatalf("got %v", res.tokens)
	}
	for i := range want {
		if res.tokens[i] != want[i] {
			t.Fatalf("[%d] got %q want %q", i, res.tokens[i], want[i])
		}
	}
}

func TestTokenizeSingleQuotes(t *testing.T) {
	t.Parallel()
	res, err := tokenizeInput(`grep 'error message' logs.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if res.tokens[1] != "error message" {
		t.Fatalf("got %v", res.tokens)
	}
}

func TestTokenizeEscapeDoubleQuote(t *testing.T) {
	t.Parallel()
	res, err := tokenizeInput(`echo "say \"hi\""`)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.tokens) != 2 || res.tokens[1] != `say "hi"` {
		t.Fatalf("got %v", res.tokens)
	}
}

func TestTokenizeAssignment(t *testing.T) {
	t.Parallel()
	res, err := tokenizeInput("FOO=bar BAZ=qux cmd")
	if err != nil {
		t.Fatal(err)
	}
	if res.args["FOO"] != "bar" || res.args["BAZ"] != "qux" {
		t.Fatalf("args %v", res.args)
	}
	if len(res.tokens) != 1 || res.tokens[0] != "cmd" {
		t.Fatalf("tokens %v", res.tokens)
	}
}

func TestTokenizeInputTooLong(t *testing.T) {
	t.Parallel()
	_, err := tokenizeInput(strings.Repeat("a", maxInputBytes+1))
	if err != errInputTooLong {
		t.Fatalf("got %v", err)
	}
}

func FuzzTokenize(f *testing.F) {
	f.Add("grep errors logs.txt")
	f.Add(`"hello world"`)
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > maxInputBytes {
			return
		}
		_, _ = tokenizeInput(s)
	})
}
