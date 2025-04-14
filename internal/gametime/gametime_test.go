package gametime

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/util"
)

func Benchmark_GetDate_Uncached(b *testing.B) {

	util.IncrementRoundCount()
	for n := 0; n < b.N; n++ {
		getDate(uint64(n))
	}
}

func Benchmark_GetDate_Cached(b *testing.B) {

	for n := 0; n < b.N; n++ {
		GetDate()
	}
}
