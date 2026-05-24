package executor

import "testing"

func TestQuotePOSIX(t *testing.T) {
	t.Parallel()
	if got := QuotePOSIX("hello"); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got := QuotePOSIX("a b"); got != "'a b'" {
		t.Fatalf("got %q", got)
	}
}

func TestQuoteCmd(t *testing.T) {
	t.Parallel()
	if got := QuoteCmd(`a"b`); got != `"a""b"` {
		t.Fatalf("got %q", got)
	}
}

func TestQuoteArgv(t *testing.T) {
	t.Parallel()
	got := QuoteArgv("bash", []string{"grep", "x y"})
	if got == "" {
		t.Fatal("empty")
	}
}
