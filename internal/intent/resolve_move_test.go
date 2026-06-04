package intent

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/parser"
)

func TestResolveMoveCopyArgv(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	cases := []struct {
		tokens     []string
		wantIntent string
	}{
		{[]string{"move", "a.txt", "b.txt"}, "move_file"},
		{[]string{"copy", "a.txt", "b.txt"}, "copy_file"},
	}
	for _, tc := range cases {
		got, err := eng.Resolve(context.Background(), parser.Request{Tokens: tc.tokens})
		if err != nil {
			t.Fatalf("resolve %v: %v", tc.tokens, err)
		}
		if got.Intent != tc.wantIntent || got.Source != SourceRule {
			t.Fatalf("got %+v want %s SourceRule", got, tc.wantIntent)
		}
	}
}
