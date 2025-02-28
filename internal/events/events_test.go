package events

import (
	"testing"
)

// goos: darwin
// goarch: arm64
// pkg: github.com/volte6/gomud/internal/events
// cpu: Apple M3 Max
// Benchmark_Typed-14         	1000000000	         0.2562 ns/op	       0 B/op	       0 allocs/op
// Benchmark_Interface-14     	1000000000	         0.2477 ns/op	       0 B/op	       0 allocs/op
// Benchmark_QueueLoops-14    	  690573	      1622 ns/op	    2428 B/op	      17 allocs/op
// Benchmark_Iterator-14      	 5961668	       200.9 ns/op	       0 B/op	       0 allocs/op

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

//
// Test iterator vs. individual queue loops
//

func SetupTestData() {

	for i := 0; i < 10; i++ {
		AddToQueue(Buff{})
		AddToQueue(Broadcast{})
		AddToQueue(Quest{})
		AddToQueue(Input{})
		AddToQueue(Broadcast{})
		AddToQueue(Message{})
		AddToQueue(WebClientCommand{})
		AddToQueue(GMCPIn{})
		AddToQueue(GMCPOut{})
		AddToQueue(System{})
		AddToQueue(MSP{})
		AddToQueue(RoomChange{})
		AddToQueue(NewRound{})
		AddToQueue(NewTurn{})
		AddToQueue(ItemOwnership{})
		AddToQueue(ScriptedEvent{})
		AddToQueue(PlayerSpawn{})
		AddToQueue(PlayerDespawn{})
		AddToQueue(Log{})
	}
}

func Benchmark_QueueLoops(b *testing.B) {

	SetupTestData()

	for n := 0; n < b.N; n++ {

		eq := GetQueue(Buff{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Broadcast{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Quest{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Input{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Broadcast{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Message{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(WebClientCommand{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(GMCPIn{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(GMCPOut{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(System{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(MSP{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(RoomChange{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(NewRound{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(NewTurn{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(ItemOwnership{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(ScriptedEvent{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(PlayerDespawn{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(PlayerDespawn{})
		for eq.Len() > 0 {
			eq.Poll()
		}
		eq = GetQueue(Log{})
		for eq.Len() > 0 {
			eq.Poll()
		}

	}

}

func Benchmark_Iterator(b *testing.B) {

	SetupTestData()

	for n := 0; n < b.N; n++ {

		for q := range Queues {
			for q.Len() > 0 {
				q.Poll()
			}
		}

	}
}
