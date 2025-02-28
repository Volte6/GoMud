package events

import (
	"sync"
	"testing"
)

// testNode is a simple struct that implements UniqueNode[int].
type testNode struct {
	id int
}

// UniqueId returns the ID of this node as the unique key.
func (t testNode) UniqueId() int {
	return t.id
}

func TestDedupeQueue_PushAndPoll(t *testing.T) {
	q := NewDedupeQueue[testNode]()
	if q.Len() != 0 {
		t.Errorf("Expected queue length 0, got %d", q.Len())
	}

	// Push items
	q.Push(testNode{id: 1})
	q.Push(testNode{id: 2})
	q.Push(testNode{id: 3})

	if q.Len() != 3 {
		t.Errorf("Expected queue length 3, got %d", q.Len())
	}

	// Poll them in FIFO order
	n := q.Poll()
	if n.id != 1 {
		t.Errorf("Expected poll=1, got %d", n.id)
	}
	n = q.Poll()
	if n.id != 2 {
		t.Errorf("Expected poll=2, got %d", n.id)
	}
	n = q.Poll()
	if n.id != 3 {
		t.Errorf("Expected poll=3, got %d", n.id)
	}

	// Now the queue should be empty
	if q.Len() != 0 {
		t.Errorf("Expected empty queue, got length %d", q.Len())
	}
	// Polling from empty queue returns zero-value
	emptyPoll := q.Poll()
	if emptyPoll.id != 0 {
		t.Errorf("Expected zero-value poll from empty queue, got %v", emptyPoll.id)
	}
}

func TestDedupeQueue_Deduplication(t *testing.T) {
	q := NewDedupeQueue[testNode]()

	q.Push(testNode{id: 10})
	q.Push(testNode{id: 10}) // Duplicate on purpose
	q.Push(testNode{id: 11})

	if q.Len() != 2 {
		t.Errorf("Expected queue length 2 after duplicate push, got %d", q.Len())
	}

	first := q.Poll()
	second := q.Poll()

	if first.id != 10 || second.id != 11 {
		t.Errorf("Unexpected poll order or deduplication failure. Polled: %d, %d", first.id, second.id)
	}
	// Queue now empty
	if q.Len() != 0 {
		t.Errorf("Expected queue to be empty, got %d", q.Len())
	}
}

func TestDedupeQueue_Peek(t *testing.T) {
	q := NewDedupeQueue[testNode]()

	q.Push(testNode{id: 1})
	q.Push(testNode{id: 2})

	peekVal := q.Peek()
	if peekVal.id != 1 {
		t.Errorf("Expected peek = 1, got %d", peekVal.id)
	}
	// Confirm that peek does not remove the item
	if q.Len() != 2 {
		t.Errorf("Expected length = 2 after peek, got %d", q.Len())
	}

	// Double-check actual FIFO ordering
	n := q.Poll()
	if n.id != 1 {
		t.Errorf("Expected poll=1, got %d", n.id)
	}
}

func TestDedupeQueue_Shift(t *testing.T) {
	q := NewDedupeQueue[testNode]()

	// Shift inserts at the head
	q.Shift(testNode{id: 5}) // first in queue
	q.Push(testNode{id: 6})
	q.Push(testNode{id: 7})

	if q.Len() != 3 {
		t.Errorf("Expected length=3, got %d", q.Len())
	}

	// The order should be 5 -> 6 -> 7
	peekVal := q.Peek()
	if peekVal.id != 5 {
		t.Errorf("Expected head=5, got %d", peekVal.id)
	}

	// Shift an item not in queue
	q.Shift(testNode{id: 4})
	if q.Len() != 4 {
		t.Errorf("Expected length=4, got %d", q.Len())
	}
	peekVal = q.Peek()
	if peekVal.id != 4 {
		t.Errorf("Expected head=4 after shift, got %d", peekVal.id)
	}

	// Shift a duplicate (already in queue). No change in length or order.
	q.Shift(testNode{id: 5})
	if q.Len() != 4 {
		t.Errorf("Expected length=4 after shifting duplicate, got %d", q.Len())
	}
	peekVal = q.Peek()
	if peekVal.id != 4 {
		t.Errorf("Expected head to remain=4, got %d", peekVal.id)
	}

	// Validate poll order: 4 -> 5 -> 6 -> 7
	if p := q.Poll(); p.id != 4 {
		t.Errorf("Expected poll=4, got %d", p.id)
	}
	if p := q.Poll(); p.id != 5 {
		t.Errorf("Expected poll=5, got %d", p.id)
	}
	if p := q.Poll(); p.id != 6 {
		t.Errorf("Expected poll=6, got %d", p.id)
	}
	if p := q.Poll(); p.id != 7 {
		t.Errorf("Expected poll=7, got %d", p.id)
	}
	if q.Len() != 0 {
		t.Errorf("Expected queue empty at end, got length=%d", q.Len())
	}
}

func TestDedupeQueue_ConcurrentAccess(t *testing.T) {
	q := NewDedupeQueue[testNode]()

	var wg sync.WaitGroup
	pushCount := 100

	// Push a bunch of items concurrently
	for i := 0; i < pushCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q.Push(testNode{id: i})
		}(i)
	}

	wg.Wait()

	if q.Len() != pushCount {
		t.Errorf("Expected queue length=%d, got %d", pushCount, q.Len())
	}

	// Poll all concurrently
	pollResults := make(chan int, pushCount)

	for i := 0; i < pushCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := q.Poll()
			pollResults <- n.id
		}()
	}
	wg.Wait()
	close(pollResults)

	// We only check total number polled (not the exact FIFO order, due to concurrency).
	polledCount := 0
	for range pollResults {
		polledCount++
	}

	if polledCount != pushCount {
		t.Errorf("Expected polled count = %d, got %d", pushCount, polledCount)
	}
	if q.Len() != 0 {
		t.Errorf("Expected queue empty after concurrent polls, got length=%d", q.Len())
	}
}
