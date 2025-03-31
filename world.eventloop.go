package main

import (
	"time"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/util"
)

// Should only handle sending messages out to users
func (w *World) EventLoop() {

	start := time.Now()
	defer func() {
		util.TrackTime(`World::EventLoop()`, time.Since(start).Seconds())
	}()

	w.eventRequeue = w.eventRequeue[:0]

	events.ProcessEvents()

	for _, e := range w.eventRequeue {
		events.AddToQueue(e)
	}

	clear(w.eventTracker)
}
