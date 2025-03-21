package hooks

import (
	"encoding/json"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

// Checks whether their level is too high for a guide
func GMCPOut_SendGMCP(e events.Event) events.ListenerReturn {

	gmcp, typeOk := e.(events.GMCPOut)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "GMCPOut", "Actual Type", e.Type())
		return events.Cancel
	}

	if gmcp.UserId < 1 {
		return events.Continue
	}

	connId := users.GetConnectionId(gmcp.UserId)
	if connId == 0 {
		return events.Continue
	}

	switch v := gmcp.Payload.(type) {
	case []byte:
		connections.SendTo(term.GmcpPayload.BytesWithPayload(v), connId)
	case string:
		connections.SendTo(term.GmcpPayload.BytesWithPayload([]byte(v)), connId)
	default:
		payload, err := json.Marshal(gmcp.Payload)
		if err != nil {
			mudlog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
			return events.Continue
		}
		connections.SendTo(term.GmcpPayload.BytesWithPayload(payload), connId)
	}

	return events.Continue
}
