package events

import "sync"

type dedupeQNode[T UniqueNode[R], R comparable] struct {
	data T
	next *dedupeQNode[T, R]
}

type UniqueNode[R comparable] interface {
	UniqueId() R
}

// A go-routine safe FIFO (first in first out) data stucture.
type DedupeQueue[T UniqueNode[R], R comparable] struct {
	head    *dedupeQNode[T, R]
	tail    *dedupeQNode[T, R]
	count   int
	lock    *sync.Mutex
	inQueue map[R]struct{}
}

// Creates a new pointer to a new queue.
func NewDedupeQueue[T UniqueNode[R], R comparable]() *DedupeQueue[T, R] {
	q := &DedupeQueue[T, R]{}
	q.lock = &sync.Mutex{}
	q.inQueue = map[R]struct{}{}
	return q
}

// Returns the number of elements in the queue (i.e. size/length)
// go-routine safe.
func (q *DedupeQueue[T, R]) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

// Pushes/inserts a value at the end/tail of the queue.
// Note: this function does mutate the queue.
// go-routine safe.
func (q *DedupeQueue[T, R]) Push(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Skip if already present
	if _, ok := q.inQueue[item.UniqueId()]; ok {
		return
	}

	n := &dedupeQNode[T, R]{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}

	q.count++
	q.inQueue[item.UniqueId()] = struct{}{}
}

// Returns the value at the front of the queue.
// i.e. the oldest value in the queue.
// Note: this function does mutate the queue.
// go-routine safe.
func (q *DedupeQueue[T, R]) Poll() T {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == nil {
		var nilVal T
		return nilVal
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}

	q.count--
	delete(q.inQueue, n.data.UniqueId())

	return n.data
}

// Returns a read value at the front of the queue.
// i.e. the oldest value in the queue.
// Note: this function does NOT mutate the queue.
// go-routine safe.
func (q *DedupeQueue[T, R]) Peek() T {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.head
	if n == nil {
		var nilVal T
		return nilVal
	}

	return n.data
}

// Shifts/inserts a value at the front/head of the queue.
// Note: this function does mutate the queue.
// go-routine safe.
func (q *DedupeQueue[T, R]) Shift(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Skip if already present
	if _, ok := q.inQueue[item.UniqueId()]; ok {
		return
	}

	n := &dedupeQNode[T, R]{data: item}

	if q.head == nil {
		// If the queue is empty, both head and tail should point to the new node
		q.head = n
		q.tail = n
	} else {
		// Insert the new node at the head
		n.next = q.head
		q.head = n
	}

	q.count++
	q.inQueue[item.UniqueId()] = struct{}{}
}
