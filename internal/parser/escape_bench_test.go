package parser

import "testing"

func Benchmark_unescape(b *testing.B) {
	var r string

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r, _ = unescape(`abc\bd\123\123\123\xFA\xfa\99abc\t`)
	}

	_ = r
}
