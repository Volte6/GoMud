package events

import "sync"

type EventType string

var (
	qLock     = sync.RWMutex{}
	allQueues = map[string]*Queue{}
)

type Event interface {
	Type() string
}

func AddToQueue(e Event, shiftToFront ...bool) {

	qLock.Lock()
	defer qLock.Unlock()

	eventType := e.Type()

	q, ok := allQueues[eventType]

	if !ok {
		q = NewQueue()
		allQueues[eventType] = q
	}

	if len(shiftToFront) > 0 && shiftToFront[0] {
		q.Shift(e)
	} else {
		q.Push(e)
	}
}

func GetQueue(e Event) *Queue {

	qLock.Lock()
	defer qLock.Unlock()

	eventType := e.Type()

	if _, ok := allQueues[eventType]; !ok {
		allQueues[eventType] = NewQueue()
	}
	return allQueues[eventType]
}
