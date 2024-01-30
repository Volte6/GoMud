package util

import (
	"errors"
	"sync"
)

var (
	errTooManyInputs = errors.New("too many inputs")
)

type Identifiable interface {
	Id() int // The id of the data source
}

type LimitQueue[T Identifiable] struct {
	lock        sync.RWMutex
	queue       []T
	inputCount  map[int]int
	inputLimit  int
	len         int
	initialized bool
}

func (lq *LimitQueue[T]) SetLimit(inputLimit int) {
	lq.inputLimit = inputLimit
}

func (lq *LimitQueue[T]) init() {
	lq.queue = make([]T, 0, 100)
	lq.inputCount = make(map[int]int)
	lq.initialized = true
}

// If provided fromId is 0, then it will return the total size
// Otherwise, it will return the number of inputs from that user
func (lq *LimitQueue[T]) Len(fromId int) int {
	lq.lock.RLock()
	defer lq.lock.RUnlock()

	if fromId == 0 {
		return len(lq.queue)
	}

	if val, ok := lq.inputCount[fromId]; ok {
		return val
	}

	return 0
}

// Should return an error if the fromId has too many inputs already pending.
func (lq *LimitQueue[T]) Push(wi Identifiable) error {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	if !lq.initialized {
		lq.init()
	}

	if _, ok := lq.inputCount[wi.Id()]; !ok {
		lq.inputCount[wi.Id()] = 0
	}

	if lq.inputLimit > 0 {
		if lq.inputCount[wi.Id()] >= lq.inputLimit {
			return errTooManyInputs
		}
	}

	lq.inputCount[wi.Id()]++
	lq.len++
	lq.queue = append(lq.queue, wi.(T))

	return nil
}

func (lq *LimitQueue[T]) Pop() (T, bool) {
	lq.lock.Lock()
	defer lq.lock.Unlock()

	if len(lq.queue) == 0 {
		var emptyVal T
		return emptyVal, false
	}

	if !lq.initialized {
		lq.init()
	}

	wi := lq.queue[0]
	lq.queue = lq.queue[1:]
	lq.len--

	lq.inputCount[wi.Id()]--
	if lq.inputCount[wi.Id()] < 1 {
		delete(lq.inputCount, wi.Id())
	}

	return wi, true
}
