package events

import (
	"sync"

	"github.com/volte6/gomud/internal/mudlog"
)

type ListenerReturn int8

type ListenerId uint64

type ListenerWrapper struct {
	id       ListenerId
	listener Listener
	isFinal  bool
}

// Return false to stop further handling of this event.
type Listener func(Event) ListenerReturn
type QueueFlag int

var (
	listenerLock = sync.RWMutex{}
	// listeners that want to handle an event first.
	listenerCt          ListenerId = 0
	eventListeners      map[string][]ListenerWrapper
	hasWildcardListener bool = false

	eventsWithoutListeners map[string]int = map[string]int{}
)

const (
	NoListenerSampleSize = 100

	First QueueFlag = 1
	Last  QueueFlag = 2
	//
	// Event return codes
	//
	// Allows the event to continu to the next listener
	Continue ListenerReturn = 0b00000001
	// Cancels any further processing of the event
	Cancel ListenerReturn = 0b00000010
	// Cancels processing, but adds back into the queue for the next event loop.
	CancelAndRequeue ListenerReturn = 0b00000100
)

func ClearListeners() {
	listenerLock.Lock()
	defer listenerLock.Unlock()
	eventListeners = map[string][]ListenerWrapper{}
}

// Returns an ID for the listener which can be used to unregister later.
func RegisterListener(emptyEvent Event, cbFunc Listener, qFlag ...QueueFlag) ListenerId {
	listenerLock.Lock()
	defer listenerLock.Unlock()

	if eventListeners == nil {
		eventListeners = map[string][]ListenerWrapper{}
	}

	listenerCt++

	eType := `*`
	if emptyEvent != nil {
		eType = emptyEvent.Type()
	}

	if _, ok := eventListeners[eType]; !ok {
		eventListeners[eType] = []ListenerWrapper{}
	}

	listenerDetails := ListenerWrapper{
		id:       listenerCt,
		listener: cbFunc,
		isFinal:  len(qFlag) > 0 && qFlag[0] == Last,
	}

	frontOfQueue := len(qFlag) > 0 && qFlag[0] == First

	if frontOfQueue {
		eventListeners[eType] = append([]ListenerWrapper{listenerDetails}, eventListeners[eType]...)

	} else if listenerDetails.isFinal {
		eventListeners[eType] = append(eventListeners[eType], listenerDetails)

	} else {

		insertPosition := 0
		for idx := 0; idx < len(eventListeners[eType]); idx++ {
			// If we're looking at a "final" listener, we can't go any farther down the list
			if !eventListeners[eType][idx].isFinal {
				insertPosition = idx
				continue
			}
			break
		}

		eventListeners[eType] = append(eventListeners[eType], ListenerWrapper{})
		copy(eventListeners[eType][insertPosition+1:], eventListeners[eType][insertPosition:])
		eventListeners[eType][insertPosition] = listenerDetails
	}

	// Write it to debug out
	//mudlog.Debug("Listener Registered", "Event", eType, "Function", runtime.FuncForPC(reflect.ValueOf(cbFunc).Pointer()).Name())

	if eType == `*` {
		hasWildcardListener = true
	}

	return listenerCt
}

// Returns true if listener found and removed.
func UnregisterListener(emptyEvent Event, id ListenerId) bool {

	listenerLock.Lock()
	defer listenerLock.Unlock()

	eType := `*`
	if emptyEvent != nil {
		eType = emptyEvent.Type()
	}

	if vals, ok := eventListeners[eType]; ok {

		for idx, wrapper := range vals {
			if wrapper.id == id {
				vals = append(vals[:idx], vals[idx+1:]...)
				eventListeners[eType] = vals
				return true
			}
		}
	}

	if eType == `*` {
		hasWildcardListener = len(eventListeners[eType]) > 0
	}

	return false

}

func DoListeners(e Event) ListenerReturn {

	listenerLock.Lock()
	defer listenerLock.Unlock()

	if len(eventListeners) == 0 {
		return Continue
	}

	listenerFound := false
	// wildcard listener is really for debugging purpose
	if hasWildcardListener {
		if vals, ok := eventListeners[`*`]; ok {
			listenerFound = true
			for _, lw := range vals {
				if result := lw.listener(e); result != Continue {
					return result
				}
			}
		}

	}

	if vals, ok := eventListeners[e.Type()]; ok {
		listenerFound = true
		for _, lw := range vals {
			if result := lw.listener(e); result != Continue {
				return result
			}
		}
	}

	if !listenerFound {
		t := e.Type()
		eventsWithoutListeners[t] = eventsWithoutListeners[t] + 1
		if eventsWithoutListeners[t]%NoListenerSampleSize == 0 {
			mudlog.Error(`DoListeners`, "Event", t, "error", "no listener for event", "sample-size", NoListenerSampleSize)
		}
	}

	return Continue
}
