package hooks

import (
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

// Checks whether their level is too high for a guide
func Broadcast_SendToAll(e events.Event) events.ListenerReturn {

	broadcast, typeOk := e.(events.Broadcast)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
		return events.Continue
	}

	messageColorized := templates.AnsiParse(broadcast.Text)

	var sentToConnectionIds []connections.ConnectionId

	// If it's communication, respect deafeaning rules
	skipConnectionIds := []connections.ConnectionId{}
	if broadcast.IsCommunication {
		for _, u := range users.GetAllActiveUsers() {
			if u.Deafened && !broadcast.SourceIsMod {
				skipConnectionIds = append(skipConnectionIds, u.ConnectionId())
			}
		}
	}

	if broadcast.SkipLineRefresh {

		sentToConnectionIds = connections.Broadcast(
			[]byte(messageColorized),
			skipConnectionIds...,
		)

	} else {

		sentToConnectionIds = connections.Broadcast(
			[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized),
			skipConnectionIds...,
		)

	}

	for _, connId := range sentToConnectionIds {
		if user := users.GetByConnectionId(connId); user != nil {
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)
		}
	}

	return events.Continue
}
