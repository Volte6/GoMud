package main

import (
	"github.com/volte6/gomud/internal/events"
)

// Should only handle sending messages out to users
func (w *World) EventLoop() {

	w.eventRequeue = w.eventRequeue[:0]

	events.ProcessEvents()

	for _, e := range w.eventRequeue {
		events.AddToQueue(e)
	}

	clear(w.eventTracker)
}
