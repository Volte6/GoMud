package events

import "sync"

type queuenode[T any] struct {
	data T
	next *queuenode[T]
}

// A go-routine safe FIFO (first in first out) data stucture.
type Queue[T any] struct {
	head  *queuenode[T]
	tail  *queuenode[T]
	count int
	lock  *sync.Mutex
}

// Creates a new pointer to a new queue.
func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{}
	q.lock = &sync.Mutex{}
	return q
}

// Returns the number of elements in the queue (i.e. size/length)
// go-routine safe.
func (q *Queue[T]) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

// Pushes/inserts a value at the end/tail of the queue.
// Note: this function does mutate the queue.
// go-routine safe.
func (q *Queue[T]) Push(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &queuenode[T]{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
}

// Returns the value at the front of the queue.
// i.e. the oldest value in the queue.
// Note: this function does mutate the queue.
// go-routine safe.
func (q *Queue[T]) Poll() T {
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

	return n.data
}

// Returns a read value at the front of the queue.
// i.e. the oldest value in the queue.
// Note: this function does NOT mutate the queue.
// go-routine safe.
func (q *Queue[T]) Peek() T {
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
func (q *Queue[T]) Shift(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &queuenode[T]{data: item}

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
}
