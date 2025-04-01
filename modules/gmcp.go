package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

const (
	TELNET_GMCP term.IACByte = 201 // https://tintin.mudhalla.net/protocols/gmcp/
)

var (
	///////////////////////////
	// GMCP COMMANDS
	///////////////////////////
	GmcpEnable  = term.TerminalCommand{Chars: []byte{term.TELNET_IAC, term.TELNET_WILL, TELNET_GMCP}, EndChars: []byte{}} // Indicates the server wants to enable GMCP.
	GmcpDisable = term.TerminalCommand{Chars: []byte{term.TELNET_IAC, term.TELNET_WONT, TELNET_GMCP}, EndChars: []byte{}} // Indicates the server wants to disable GMCP.

	GmcpAccept = term.TerminalCommand{Chars: []byte{term.TELNET_IAC, term.TELNET_DO, TELNET_GMCP}, EndChars: []byte{}}   // Indicates the client accepts GMCP sub-negotiations.
	GmcpRefuse = term.TerminalCommand{Chars: []byte{term.TELNET_IAC, term.TELNET_DONT, TELNET_GMCP}, EndChars: []byte{}} // Indicates the client refuses GMCP sub-negotiations.

	GmcpPayload = term.TerminalCommand{Chars: []byte{term.TELNET_IAC, term.TELNET_SB, TELNET_GMCP}, EndChars: []byte{term.TELNET_IAC, term.TELNET_SE}} // Wrapper for sending GMCP payloads
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

	// connectionId to map[string]int
	g.cache, _ = lru.New[uint64, map[string]int](256)

	g.plug.ExportFunction(`SendGMCPEvent`, g.sendGMCPEvent)
	g.plug.Callbacks.SetIACHandler(g.HandleIAC)
	g.plug.Callbacks.SetOnNetConnect(g.onNetConnect)

	events.RegisterListener(GMCPOut{}, g.dispatchGMCP)
	events.RegisterListener(events.PlayerSpawn{}, g.handlePlayerJoin)

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
	plug  *plugins.Plugin
	cache *lru.Cache[uint64, map[string]int]
}

type GMCPHello struct {
	Client  string
	Version string
}

type GMCPSupportsSet []string

type GMCPSupportsRemove = []string

type GMCPLogin struct {
	Name     string
	Password string
}

func (g *GMCPModule) onNetConnect(n plugins.NetConnection) {
	if !n.IsWebSocket() {
		g.sendGMCPEnableRequest(n.ConnectionId())
	}
}

func (g *GMCPModule) getModules(connectionId uint64) map[string]int {
	data, ok := g.cache.Get(connectionId)

	// It may not be here, or it may have been evicted.
	// Re-request if this happens.
	if !ok {
		g.sendGMCPEnableRequest(connectionId)
	}

	return data
}

func (g *GMCPModule) setModules(connectionId uint64, modules map[string]int) {
	g.cache.Add(connectionId, modules)
}

func (g *GMCPModule) HandleIAC(connectionId uint64, iacCmd []byte) bool {

	if !g.isGMCPCommand(iacCmd) {
		return false
	}

	if ok, payload := term.Matches(iacCmd, GmcpAccept); ok {
		mudlog.Debug("Received", "type", "IAC (Client-GMCP Accept)", "data", term.BytesString(payload))
		return true
	}

	if ok, payload := term.Matches(iacCmd, GmcpRefuse); ok {
		mudlog.Debug("Received", "type", "IAC (Client-GMCP Refuse)", "data", term.BytesString(payload))
		return true
	}

	if len(iacCmd) >= 5 && iacCmd[len(iacCmd)-2] == term.TELNET_IAC && iacCmd[len(iacCmd)-1] == term.TELNET_SE {
		// Unhanlded IAC command, log it

		requestBody := iacCmd[3 : len(iacCmd)-2]
		//mudlog.Debug("Received", "type", "GMCP", "size", len(iacCmd), "data", string(requestBody))

		spaceAt := 0
		for i := 0; i < len(requestBody); i++ {
			if requestBody[i] == 32 {
				spaceAt = i
				break
			}
		}

		command := ``
		payload := []byte{}

		if spaceAt > 0 && spaceAt < len(requestBody) {
			command = string(requestBody[0:spaceAt])
			payload = requestBody[spaceAt+1:]
		} else {
			command = string(requestBody)
		}

		mudlog.Debug("Received", "type", "GMCP (Handling)", "command", command, "payload", string(payload))

		switch command {

		case `Core.Hello`:
			decoded := GMCPHello{}
			if err := json.Unmarshal(payload, &decoded); err == nil {
				cs := connections.GetClientSettings(connectionId)
				cs.Client.Name = decoded.Client
				cs.Client.Version = decoded.Version
				if strings.EqualFold(decoded.Client, `mudlet`) {
					cs.Client.IsMudlet = true
				}
				connections.OverwriteClientSettings(connectionId, cs)
			}
		case `Core.Supports.Set`:
			decoded := GMCPSupportsSet{}
			if err := json.Unmarshal(payload, &decoded); err == nil {
				enabledModules := decoded.GetSupportedModules()
				g.setModules(connectionId, enabledModules)
			}
		case `Core.Supports.Remove`:
			decoded := GMCPSupportsRemove{}
			if err := json.Unmarshal(payload, &decoded); err == nil {

				enabledModules := g.getModules(connectionId)
				if len(enabledModules) > 0 {
					for _, name := range decoded {
						delete(enabledModules, name)
					}
				}

				g.setModules(connectionId, enabledModules)

			}
		case `Char.Login`:
			decoded := GMCPLogin{}
			if err := json.Unmarshal(payload, &decoded); err == nil {
				mudlog.Debug("GMCP LOGIN", "username", decoded.Name, "password", decoded.Password)
			}
		}

		return true
	}

	// Unhanlded IAC command, log it
	mudlog.Debug("Received", "type", "GMCP?", "data-size", len(iacCmd), "data-string", string(iacCmd), "data-bytes", iacCmd)

	return true
}

func (g *GMCPModule) isGMCPCommand(b []byte) bool {
	return len(b) > 2 && b[0] == term.TELNET_IAC && b[2] == TELNET_GMCP
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

func (g *GMCPModule) handlePlayerJoin(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerSpawn", "Actual Type", e.Type())
		return events.Cancel
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	// Send request to enable GMCP
	g.getModules(user.ConnectionId())

	return events.Continue
}

// Sends a telnet IAC request to enable GMCP
func (g *GMCPModule) sendGMCPEnableRequest(connectionId uint64) {
	connections.SendTo(
		GmcpEnable.BytesWithPayload(nil),
		connectionId,
	)
	data := map[string]int{}
	g.cache.Add(connectionId, data)
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

	// Get enabled modules... if none, skip out.

	enabledModules := g.getModules(connId)
	if len(enabledModules) == 0 {
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
			v = append([]byte(gmcp.Module+` `), v...)
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(v), connId)
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
			v = gmcp.Module + ` ` + v
		}

		connections.SendTo(GmcpPayload.BytesWithPayload([]byte(v)), connId)
	default:

		payload, err := json.Marshal(gmcp.Payload)
		if err != nil {
			mudlog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
			return events.Continue
		}

		if len(gmcp.Module) > 0 {
			payload = append([]byte(gmcp.Module+` `), payload...)
		}

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, payload, "", "\t")
			fmt.Println(string(prettyJSON.Bytes()))
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(payload), connId)
	}

	return events.Continue
}

// Returns a map of module name to version number
func (s GMCPSupportsSet) GetSupportedModules() map[string]int {

	ret := map[string]int{}

	for _, entry := range s {

		parts := strings.Split(entry, ` `)
		if len(parts) == 2 {
			ret[parts[0]], _ = strconv.Atoi(parts[1])
		}

	}

	return ret
}
