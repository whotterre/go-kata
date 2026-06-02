package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
)

func BenchmarkNDJSONParsers(b *testing.B) {
	input := mustReadFixture(b, "input-64kb.ndjson")
	b.SetBytes(int64(len(input)))

	benchmarks := []struct {
		name   string
		parser func(context.Context, io.Reader, func([]byte) error) error
	}{
		{name: "ReadNDJSON", parser: ReadNDJSON},
		{name: "slowParser", parser: slowParser},
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			ctx := context.Background()
			handle := func([]byte) error { return nil }

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := benchmark.parser(ctx, bytes.NewReader(input), handle); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func mustReadFixture(b *testing.B, path string) []byte {
	b.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}

	return data
}
