package hooks

import (
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
)

// Checks whether their level is too high for a guide
func WebClientCommand_SendWebClientCommand(e events.Event) events.EventReturn {

	cmd, typeOk := e.(events.WebClientCommand)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "WebClientCommand", "Actual Type", e.Type())
		return events.Continue
	}

	if !connections.IsWebsocket(cmd.ConnectionId) {
		return events.Cancel
	}

	connections.SendTo([]byte(cmd.Text), cmd.ConnectionId)

	return events.Continue
}
