package intent

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/parser"
)

func BenchmarkNewDefaultEngine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewDefaultEngine(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewEngineWithOverlay(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewEngineWithOverlay(ctx, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkResolve(b *testing.B, tokens []string) {
	eng, err := NewDefaultEngine()
	if err != nil {
		b.Fatal(err)
	}
	req := parser.Request{Tokens: tokens}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := eng.Resolve(ctx, req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkResolveFindFile(b *testing.B) {
	benchmarkResolve(b, []string{"locate", "help.txt"})
}

func BenchmarkResolvePWD(b *testing.B) {
	benchmarkResolve(b, []string{"pwd"})
}

func BenchmarkResolveLongNL(b *testing.B) {
	benchmarkResolve(b, []string{"show", "me", "the", "10", "largest", "files", "in", "this", "folder"})
}

func BenchmarkResolveMiss(b *testing.B) {
	eng, err := NewDefaultEngine()
	if err != nil {
		b.Fatal(err)
	}
	req := parser.Request{Tokens: []string{"unknown", "command", "xyz"}}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Resolve(ctx, req)
	}
}
