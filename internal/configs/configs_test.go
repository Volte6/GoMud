package configs

import (
	"testing"
)

// As the config struct gets larger and larger
// This benchmark is to test whether a pointer becomes a
// better option than a copy of the struct in instances where
// we pass it around a lot.
//
// Results:
//
// goos: darwin
// goarch: arm64
// pkg: github.com/volte6/gomud/internal/configs
// cpu: Apple M3 Max
// Benchmark_Config_Pointer-14      	1000000000	         0.2590 ns/op	       0 B/op	       0 allocs/op
// Benchmark_Config_Copy-14         	1000000000	         0.2583 ns/op	       0 B/op	       0 allocs/op
// Benchmark_Config_Typed-14        	1000000000	         0.2535 ns/op	       0 B/op	       0 allocs/op
// Benchmark_Config_Interface-14    	88604206	        14.20 ns/op	       0 B/op	       0 allocs/op
//
// Copies are better.
// Assertions are costly with such a large, complex struct, though.

//
// Pointer vs Copy benchmarks
//

func Benchmark_Config_Pointer(b *testing.B) {

	c := GetConfig()
	for n := 0; n < b.N; n++ {
		ConfigPointer(&c)
	}
}

func Benchmark_Config_Copy(b *testing.B) {

	c := GetConfig()
	for n := 0; n < b.N; n++ {
		ConfigCopy(c)
	}

}

func ConfigPointer(c *Config) uint64 {
	return uint64(c.turnsPerRound)
}

func ConfigCopy(c Config) uint64 {
	return uint64(c.turnsPerRound)
}

//
// Type assertion benchmarks
//

func Benchmark_Config_Typed(b *testing.B) {

	c := GetConfig()
	for n := 0; n < b.N; n++ {
		Config_Typed(c)
	}

}

func Benchmark_Config_Interface(b *testing.B) {

	c := GetConfig()
	for n := 0; n < b.N; n++ {
		Config_Interface(c)
	}
}

func Config_Typed(c Config) uint64 {
	return uint64(c.turnsPerRound)
}

func Config_Interface(c interface{}) uint64 {
	return uint64(c.(Config).turnsPerRound)
}
