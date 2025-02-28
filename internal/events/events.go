package events

import (
	"sync"
)

type EventType string

var (
	qLock     = sync.RWMutex{}
	allQueues = map[string]*Queue[Event]{}
	requeues  = map[string][]Event{}
)

type Event interface {
	Type() string
}

// events added via Requeue() will only show up in the queue after a call to GetQueue()
func Requeue(e Event) {
	qLock.Lock()
	defer qLock.Unlock()

	t := e.Type()
	if _, ok := requeues[t]; !ok {
		requeues[t] = []Event{}
	}

	requeues[t] = append(requeues[t], e)
}

func AddToQueue(e Event, shiftToFront ...bool) {

	qLock.Lock()
	defer qLock.Unlock()

	eventType := e.Type()

	q, ok := allQueues[eventType]

	if !ok {
		q = NewQueue[Event]()
		allQueues[eventType] = q
	}

	if len(shiftToFront) > 0 && shiftToFront[0] {
		q.Shift(e)
	} else {
		q.Push(e)
	}
}

func GetQueue(e Event) *Queue[Event] {

	qLock.Lock()
	defer qLock.Unlock()

	eventType := e.Type()

	if _, ok := allQueues[eventType]; !ok {
		allQueues[eventType] = NewQueue[Event]()
		requeues[eventType] = []Event{}
	}

	for _, e := range requeues[eventType] {
		allQueues[eventType].Shift(e)
	}

	requeues[eventType] = requeues[eventType][:0]

	return allQueues[eventType]
}

// Iterator
/*
	for q := range events.Queues {
		events.DoListeners(q.Poll())
	}
*/
func Queues(yield func(value *Queue[Event]) bool) {
	qLock.Lock()
	defer qLock.Unlock()
	for _, eQueue := range allQueues {
		if !yield(eQueue) {
			return
		}
	}
}
