package intent

import "testing"

func TestMatchPatternParams(t *testing.T) {
	t.Parallel()
	params, ok := matchPattern("grep {{pattern}} {{file}}", []string{"grep", "errors", "logs.txt"})
	if !ok {
		t.Fatal("expected match")
	}
	if params["pattern"] != "errors" || params["file"] != "logs.txt" {
		t.Fatalf("got %v", params)
	}
}

func TestMatchPatternLiteral(t *testing.T) {
	t.Parallel()
	_, ok := matchPattern("pwd", []string{"pwd"})
	if !ok {
		t.Fatal("expected match")
	}
}

func TestMatchPatternPrintTextMultiWord(t *testing.T) {
	t.Parallel()
	params, ok := matchPattern("echo {{word1}} {{word2}}", []string{"echo", "Hello", "CLX"})
	if !ok {
		t.Fatal("expected match")
	}
	if params["word1"] != "Hello" || params["word2"] != "CLX" {
		t.Fatalf("got %v", params)
	}
}
