package hooks

import (
	"fmt"
	"log/slog"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
)

// Tee's log output to admins following
var (
	logFollowUniques       = map[connections.ConnectionId]struct{}{}
	logFollowConnectionIds = []connections.ConnectionId{}
)

func FollowLogs(e events.Event) bool {

	evt, typeOk := e.(events.Log)
	if !typeOk {
		slog.Error("Event", "Expected Type", "Log", "Actual Type", e.Type())
		return false
	}

	if evt.FollowAdd > 0 {

		if _, ok := logFollowUniques[evt.FollowAdd]; !ok {
			logFollowUniques[evt.FollowAdd] = struct{}{}
			logFollowConnectionIds = append(logFollowConnectionIds, evt.FollowAdd)
		}

	} else if evt.FollowRemove > 0 {

		if _, ok := logFollowUniques[evt.FollowRemove]; ok {
			delete(logFollowUniques, evt.FollowAdd)
			for idx, connId := range logFollowConnectionIds {
				if connId == evt.FollowRemove {
					logFollowConnectionIds = append(logFollowConnectionIds[:idx], logFollowConnectionIds[idx+1:]...)
				}
			}
		}

		// do some general cleanup
		for idx := len(logFollowConnectionIds) - 1; idx >= 0; idx++ {
			if connections.Get(logFollowConnectionIds[idx]) == nil {
				delete(logFollowUniques, logFollowConnectionIds[idx])
				if logFollowConnectionIds[idx] == evt.FollowRemove {
					logFollowConnectionIds = append(logFollowConnectionIds[:idx], logFollowConnectionIds[idx+1:]...)
				}
			}
		}
	} else {
		connections.SendTo([]byte(fmt.Sprintln(evt.Data...)), logFollowConnectionIds...)
	}

	return true
}
