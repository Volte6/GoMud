package events

import (
	"sync"
)

type ListenerId uint64

var (
	listenerLock = sync.RWMutex{}
	// listeners that want to handle an event first.
	listenerCt          ListenerId = 0
	eventListeners      map[string][]ListenerWrapper
	hasWildcardListener bool = false
)

func ClearListeners() {
	listenerLock.Lock()
	defer listenerLock.Unlock()
	eventListeners = map[string][]ListenerWrapper{}
}

// Returns an ID for the listener which can be used to unregister later.
func RegisterListener(emptyEvent Event, cbFunc Listener, addToFront ...bool) ListenerId {
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

	if len(addToFront) > 0 && addToFront[0] {
		eventListeners[eType] = append([]ListenerWrapper{{listenerCt, cbFunc}}, eventListeners[eType]...)
	} else {
		eventListeners[eType] = append(eventListeners[eType], ListenerWrapper{listenerCt, cbFunc})
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

func DoListeners(e Event) bool {

	listenerLock.Lock()
	defer listenerLock.Unlock()

	if len(eventListeners) == 0 {
		return true
	}

	// wildcard listener is really for debugging purpose
	if hasWildcardListener {
		if vals, ok := eventListeners[`*`]; ok {
			for _, lw := range vals {
				if !lw.listner(e) {
					return false
				}
			}
		}
	}

	if vals, ok := eventListeners[e.Type()]; ok {
		for _, lw := range vals {
			if !lw.listner(e) {
				return false
			}
		}
	}

	return true
}

type ListenerWrapper struct {
	id      ListenerId
	listner Listener
}

// Return false to stop further handling of this event.
type Listener func(Event) bool
