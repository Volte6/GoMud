package events

import (
	"testing"
)

//
// Benchmark using so much type assertion in events
//

func Benchmark_Typed(b *testing.B) {

	e := Message{}
	for n := 0; n < b.N; n++ {
		FuncType(e)
	}

}

func Benchmark_Interface(b *testing.B) {

	e := Message{}
	for n := 0; n < b.N; n++ {
		FuncInterface(e)
	}
}

func FuncType(m Message) uint64 {
	return uint64(m.RoomId)
}

func FuncInterface(m interface{}) uint64 {
	return uint64(m.(Message).RoomId)
}
