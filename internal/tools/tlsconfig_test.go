package tools

import (
	"testing"
)

//go test -v ./internal/tools -bench=BenchmarkInitAutoMakeTLSConfig -benchmem -benchtime=10s
func BenchmarkInitAutoMakeTLSConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InitAutoMakeTLSConfig()
	}
	b.StopTimer()
}
