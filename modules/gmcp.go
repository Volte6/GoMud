package modules

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {

	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	g := GMCPModule{
		plug: plugins.New(`gmcp`, `1.0`),
	}

	g.plug.ExportFunction(`SendGMCPEvent`, g.sendGMCPEvent)

	events.RegisterListener(GMCPOut{}, g.dispatchGMCP)

}

// GMCP Commands from server to client
type GMCPOut struct {
	UserId  int
	Module  string
	Payload any
}

func (g GMCPOut) Type() string { return `GMCPOut` }

type GMCPModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
}

func (g *GMCPModule) sendGMCPEvent(userId int, payload any, moduleName ...string) {

	evt := GMCPOut{
		UserId:  userId,
		Payload: payload,
	}

	if len(moduleName) > 0 {
		evt.Module = moduleName[0]
	}

	events.AddToQueue(evt)
}

// Checks whether their level is too high for a guide
func (g *GMCPModule) dispatchGMCP(e events.Event) events.ListenerReturn {

	gmcp, typeOk := e.(GMCPOut)
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

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			if len(gmcp.Module) > 0 {
				fmt.Print(gmcp.Module + ` `)
			}
			fmt.Println(string(v))
		}

		if len(gmcp.Module) > 0 {
			v = append([]byte(gmcp.Module), v...)
		}

		connections.SendTo(term.GmcpPayload.BytesWithPayload(v), connId)
	case string:

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			if len(gmcp.Module) > 0 {
				fmt.Print(gmcp.Module + ` `)
			}
			fmt.Println(string(v))
		}

		if len(gmcp.Module) > 0 {
			v = gmcp.Module + v
		}

		connections.SendTo(term.GmcpPayload.BytesWithPayload([]byte(v)), connId)
	default:
		payload, err := json.Marshal(gmcp.Payload)
		if err != nil {
			mudlog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
			return events.Continue
		}

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, payload, "", "\t")

			if len(gmcp.Module) > 0 {
				fmt.Print(gmcp.Module + ` `)
			}
			fmt.Println(string(prettyJSON.Bytes()))
		}

		if len(gmcp.Module) > 0 {
			payload = append([]byte(gmcp.Module), payload...)
		}

		connections.SendTo(term.GmcpPayload.BytesWithPayload(payload), connId)
	}

	return events.Continue
}
