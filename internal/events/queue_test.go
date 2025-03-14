package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewQueue verifies that a new queue is initialized properly.
func TestNewQueue(t *testing.T) {
	q := NewQueue[int]()

	assert.NotNil(t, q, "NewQueue should return a non-nil Queue pointer")
	assert.Equal(t, 0, q.Len(), "Newly created queue should have length 0")
}

// TestPushAndLen verifies pushing elements increases the queue length as expected.
func TestPushAndLen(t *testing.T) {
	q := NewQueue[int]()

	q.Push(1)
	assert.Equal(t, 1, q.Len(), "Queue length should be 1 after first push")

	q.Push(2)
	q.Push(3)
	assert.Equal(t, 3, q.Len(), "Queue length should be 3 after pushing three items")
}

// TestPoll verifies that Poll returns elements in FIFO order and decreases the queue length.
func TestPoll(t *testing.T) {
	q := NewQueue[int]()

	// Poll on empty queue should return zero-value for int (i.e., 0)
	val := q.Poll()
	assert.Equal(t, 0, val, "Polling an empty queue should return the zero-value")
	assert.Equal(t, 0, q.Len(), "Queue length remains 0 after polling an empty queue")

	q.Push(10)
	q.Push(20)
	q.Push(30)

	// Poll in FIFO order
	v1 := q.Poll()
	assert.Equal(t, 10, v1, "First poll should return 10")
	assert.Equal(t, 2, q.Len(), "Queue length should be 2 after polling once")

	v2 := q.Poll()
	assert.Equal(t, 20, v2, "Second poll should return 20")
	assert.Equal(t, 1, q.Len(), "Queue length should be 1")

	v3 := q.Poll()
	assert.Equal(t, 30, v3, "Third poll should return 30")
	assert.Equal(t, 0, q.Len(), "Queue should be empty after polling all items")

	// Poll again should return zero-value
	v4 := q.Poll()
	assert.Equal(t, 0, v4, "Polling an empty queue again should return zero-value")
	assert.Equal(t, 0, q.Len(), "Queue remains empty after polling")
}

// TestPeek verifies that Peek returns the front element without removing it.
func TestPeek(t *testing.T) {
	q := NewQueue[int]()

	// Peeking on empty queue should return zero-value (0)
	val := q.Peek()
	assert.Equal(t, 0, val, "Peeking an empty queue should return zero-value")
	assert.Equal(t, 0, q.Len(), "Queue length remains 0 after peeking an empty queue")

	q.Push(5)
	q.Push(6)

	// First peek
	peekVal := q.Peek()
	assert.Equal(t, 5, peekVal, "Peek should return the front item (5)")
	assert.Equal(t, 2, q.Len(), "Peek does not remove item, length should stay 2")

	// Second peek should return the same item
	peekVal2 := q.Peek()
	assert.Equal(t, 5, peekVal2, "Consecutive peeks should return the same item")
	assert.Equal(t, 2, q.Len(), "Length is unchanged by peek")

	// Poll to verify the item is actually still in queue
	pollVal := q.Poll()
	assert.Equal(t, 5, pollVal, "Poll after peek should still give the original front item (5)")
	assert.Equal(t, 1, q.Len(), "Queue length should now be 1 after polling once")
}

// TestShift verifies that Shift adds elements to the front of the queue.
func TestShift(t *testing.T) {
	q := NewQueue[int]()

	// Shift on empty queue
	q.Shift(100)
	assert.Equal(t, 1, q.Len(), "Queue should have length 1 after shifting into empty queue")
	assert.Equal(t, 100, q.Peek(), "Peek should return the shifted item (100)")

	// Shift on non-empty queue
	q.Push(200) // Now front=100, back=200
	q.Push(300) // Now front=100, next=200, back=300

	q.Shift(50) // Insert 50 at front
	// The queue is now front=50 -> 100 -> 200 -> 300

	assert.Equal(t, 4, q.Len(), "Queue should have length 4 after shifting onto non-empty queue")

	// Verify order with Poll
	p1 := q.Poll()
	assert.Equal(t, 50, p1, "First poll should return the newly shifted item (50)")

	p2 := q.Poll()
	assert.Equal(t, 100, p2, "Next poll should return 100")

	p3 := q.Poll()
	assert.Equal(t, 200, p3, "Then 200")

	p4 := q.Poll()
	assert.Equal(t, 300, p4, "Finally 300")

	assert.Equal(t, 0, q.Len(), "Queue should be empty after polling all items")
}

// Example test verifying concurrency safety (optional demonstration).
// This test spawns multiple goroutines pushing and polling items.
// Adjust the iteration counts as needed for your environment.
func TestQueueConcurrency(t *testing.T) {
	q := NewQueue[int]()
	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func(start int) {
			for j := start; j < start+100; j++ {
				q.Push(j)
			}
			done <- true
		}(i * 1000)
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			// We don't know how many items will be available each time we poll,
			// but we'll do a few polls.
			for k := 0; k < 100; k++ {
				q.Poll() // It's fine if it returns 0 sometimes if queue is temporarily empty.
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// We can't know exactly how many items remain without more careful synchronization,
	// but we can at least ensure the queue is in a valid state (no panics).
	t.Logf("Final queue length (could be anything >= 0): %d", q.Len())
	assert.True(t, q.Len() >= 0, "Queue should have non-negative length")
}

// Benchmark using the custom Queue implementation.
func BenchmarkQueuePushPoll(b *testing.B) {
	q := NewQueue[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		for j := 0; j < 100; j++ {

			q.Push(i)
			if j%3 == 0 {
				_ = q.Poll()
			}
		}

		for q.Len() > 0 {
			_ = q.Poll()
		}
	}
}
