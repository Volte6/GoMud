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
	return uint64(c.Timing.turnsPerRound)
}

func ConfigCopy(c Config) uint64 {
	return uint64(c.Timing.turnsPerRound)
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
	return uint64(c.Timing.turnsPerRound)
}

func Config_Interface(c interface{}) uint64 {
	return uint64(c.(Config).Timing.turnsPerRound)
}

func TestConfigUInt64String(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigUInt64
		expected string
	}{
		{name: "Zero", value: 0, expected: "0"},
		{name: "SmallNumber", value: 123, expected: "123"},
		{name: "LargeNumber", value: 9223372036854775808, expected: "9223372036854775808"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigUInt64.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigIntString(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigInt
		expected string
	}{
		{name: "Zero", value: 0, expected: "0"},
		{name: "Positive", value: 42, expected: "42"},
		{name: "Negative", value: -100, expected: "-100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigInt.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigStringString(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigString
		expected string
	}{
		{name: "Empty", value: "", expected: ""},
		{name: "Hello", value: "Hello, world!", expected: "Hello, world!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigString.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigFloatString(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigFloat
		expected string
	}{
		{name: "Zero", value: 0.0, expected: "0"},
		{name: "Pi", value: 3.14, expected: "3.14"},
		{name: "WholeNumber", value: 2.0, expected: "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigFloat.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigBoolString(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigBool
		expected string
	}{
		{name: "True", value: true, expected: "true"},
		{name: "False", value: false, expected: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigBool.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigSliceStringString(t *testing.T) {
	tests := []struct {
		name     string
		value    ConfigSliceString
		expected string
	}{
		{name: "EmptySlice", value: ConfigSliceString{}, expected: `[]`},
		{name: "SingleElement", value: ConfigSliceString{"one"}, expected: `["one"]`},
		{name: "MultipleElements", value: ConfigSliceString{"one", "two", "three"}, expected: `["one", "two", "three"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.expected {
				t.Errorf("ConfigSliceString.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
