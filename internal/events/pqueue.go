package events

import (
	"sync"
)

// pqNode is a singly-linked node in the priority queue.
type pqNode[T any] struct {
	data T
	next *pqNode[T]
}

// PriorityQueue is a thread-safe priority queue that uses
// a singly linked list sorted by a user-supplied comparison function.
type PriorityQueue[T any] struct {
	head *pqNode[T]
	size int
	less func(a, b T) bool // "less" defines the ordering; if less(a,b) is true, 'a' has higher priority than 'b'.
	lock sync.Mutex
}

// NewPriorityQueue creates a new priority queue using the given comparison function.
// The 'less' function should return true if `a` has *higher* priority than `b`.
func NewPriorityQueue[T any](lessFunc func(a, b T) bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		less: lessFunc,
	}
}

// Len returns the number of elements in the priority queue.
func (pq *PriorityQueue[T]) Len() int {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	return pq.size
}

// Peek returns the highest-priority item without removing it.
// It also returns a boolean indicating if the queue was non-empty.
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	if pq.head == nil {
		var zeroVal T
		return zeroVal, false
	}
	return pq.head.data, true
}

// Poll removes and returns the highest-priority item.
// It returns a boolean indicating if the queue was non-empty.
func (pq *PriorityQueue[T]) Poll() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.head == nil {
		var zeroVal T
		return zeroVal, false
	}

	node := pq.head
	pq.head = node.next
	pq.size--

	return node.data, true
}

// Push inserts a new item into the priority queue, keeping items sorted
// according to the 'less' function. The item with the *smallest* 'less' value
// is placed at the front if 'less(item, head.data)' returns true (indicating
// higher priority).
func (pq *PriorityQueue[T]) Push(item T) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	newNode := &pqNode[T]{data: item}

	// If queue is empty or new item is higher priority than head
	if pq.head == nil || pq.less(item, pq.head.data) {
		newNode.next = pq.head
		pq.head = newNode
	} else {
		current := pq.head
		// Traverse until we find the correct spot
		for current.next != nil && !pq.less(item, current.next.data) {
			current = current.next
		}
		newNode.next = current.next
		current.next = newNode
	}
	pq.size++
}
