package tools

import (
	"test"
)

func BenchmarkInitAutoMakeTLSConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InitAutoMakeTLSConfig()
	}
	b.StopTimer()
}
