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

	gmcpModule GMCPModule = GMCPModule{}
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
	gmcpModule = GMCPModule{
		plug: plugins.New(`gmcp`, `1.0`),
	}

	// connectionId to map[string]int
	gmcpModule.cache, _ = lru.New[uint64, GMCPSettings](128)

	gmcpModule.plug.ExportFunction(`SendGMCPEvent`, gmcpModule.sendGMCPEvent)
	gmcpModule.plug.ExportFunction(`IsMudlet`, gmcpModule.IsMudletExportedFunction)

	gmcpModule.plug.Callbacks.SetIACHandler(gmcpModule.HandleIAC)
	gmcpModule.plug.Callbacks.SetOnNetConnect(gmcpModule.onNetConnect)

	events.RegisterListener(GMCPOut{}, gmcpModule.dispatchGMCP)
	events.RegisterListener(events.PlayerSpawn{}, gmcpModule.handlePlayerJoin)

}

func isGMCPEnabled(connectionId uint64) bool {
	if gmcpData, ok := gmcpModule.cache.Get(connectionId); ok {
		return gmcpData.GMCPAccepted
	}

	return false
}

// ///////////////////
// EVENTS
// ///////////////////

type GMCPOut struct {
	UserId  int
	Module  string
	Payload any
}

func (g GMCPOut) Type() string { return `GMCPOut` }

// ///////////////////
// END EVENTS
// ///////////////////
type GMCPModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug  *plugins.Plugin
	cache *lru.Cache[uint64, GMCPSettings]
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

// / SETTINGS
type GMCPSettings struct {
	Client struct {
		Name     string
		Version  string
		IsMudlet bool // Knowing whether is a mudlet client can be useful, since Mudlet hates certain ANSI/Escape codes.
	}
	GMCPAccepted   bool           // Do they accept GMCP data?
	EnabledModules map[string]int // What modules/versions are accepted? Might not be used properly by clients.
}

func (gs *GMCPSettings) IsMudlet() bool {
	return gs.Client.IsMudlet
}

/// END SETTINGS

func (g *GMCPModule) IsMudletExportedFunction(connectionId uint64) bool {
	gmcpData, ok := g.cache.Get(connectionId)
	if !ok {
		return false
	}
	return gmcpData.IsMudlet()
}

func (g *GMCPModule) onNetConnect(n plugins.NetConnection) {

	if n.IsWebSocket() {
		setting := GMCPSettings{}
		setting.Client.Name = `WebClient`
		setting.Client.Version = `1.0.0`
		g.cache.Add(n.ConnectionId(), setting)
		return
	}

	g.cache.Add(n.ConnectionId(), GMCPSettings{})

	g.sendGMCPEnableRequest(n.ConnectionId())
}

func (g *GMCPModule) isGMCPCommand(b []byte) bool {
	return len(b) > 2 && b[0] == term.TELNET_IAC && b[2] == TELNET_GMCP
}

func (g *GMCPModule) sendGMCPEvent(userId int, moduleName string, payload any) {

	evt := GMCPOut{
		UserId:  userId,
		Module:  moduleName,
		Payload: payload,
	}

	events.AddToQueue(evt)
}

func (g *GMCPModule) handlePlayerJoin(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerSpawn", "Actual Type", e.Type())
		return events.Cancel
	}

	// Send request to enable GMCP
	g.sendGMCPEnableRequest(evt.ConnectionId)

	return events.Continue
}

// Sends a telnet IAC request to enable GMCP
func (g *GMCPModule) sendGMCPEnableRequest(connectionId uint64) {
	connections.SendTo(
		GmcpEnable.BytesWithPayload(nil),
		connectionId,
	)
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

func (g *GMCPModule) HandleIAC(connectionId uint64, iacCmd []byte) bool {

	if !g.isGMCPCommand(iacCmd) {
		return false
	}

	if ok, payload := term.Matches(iacCmd, GmcpAccept); ok {

		gmcpData, ok := g.cache.Get(connectionId)
		if !ok {
			gmcpData = GMCPSettings{}
		}
		gmcpData.GMCPAccepted = true
		g.cache.Add(connectionId, gmcpData)

		mudlog.Debug("Received", "type", "IAC (Client-GMCP Accept)", "data", term.BytesString(payload))
		return true
	}

	if ok, payload := term.Matches(iacCmd, GmcpRefuse); ok {

		gmcpData, ok := g.cache.Get(connectionId)
		if !ok {
			gmcpData = GMCPSettings{}
		}
		gmcpData.GMCPAccepted = false
		g.cache.Add(connectionId, gmcpData)

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

				gmcpData, ok := g.cache.Get(connectionId)
				if !ok {
					gmcpData = GMCPSettings{}
					gmcpData.GMCPAccepted = true
				}

				gmcpData.Client.Name = decoded.Client
				gmcpData.Client.Version = decoded.Version

				if strings.EqualFold(decoded.Client, `mudlet`) {

					gmcpData.Client.IsMudlet = true

				}

				g.cache.Add(connectionId, gmcpData)
			}
		case `Core.Supports.Set`:
			decoded := GMCPSupportsSet{}
			if err := json.Unmarshal(payload, &decoded); err == nil {

				gmcpData, ok := g.cache.Get(connectionId)
				if !ok {
					gmcpData = GMCPSettings{}
					gmcpData.GMCPAccepted = true
				}

				gmcpData.EnabledModules = map[string]int{}

				for name, value := range decoded.GetSupportedModules() {

					// Break it down into:
					// Char.Inventory.Backpack
					// Char.Inventory
					// Char
					for {
						gmcpData.EnabledModules[name] = value
						idx := strings.LastIndex(name, ".")
						if idx == -1 {
							break
						}
						name = name[:idx]
					}

				}

				g.cache.Add(connectionId, gmcpData)

			}
		case `Core.Supports.Remove`:
			decoded := GMCPSupportsRemove{}
			if err := json.Unmarshal(payload, &decoded); err == nil {

				gmcpData, ok := g.cache.Get(connectionId)
				if !ok {
					gmcpData = GMCPSettings{}
					gmcpData.GMCPAccepted = true
				}

				if len(gmcpData.EnabledModules) > 0 {
					for _, name := range decoded {
						delete(gmcpData.EnabledModules, name)
					}
				}

				g.cache.Add(connectionId, gmcpData)

			}
		case `Char.Login`:
			decoded := GMCPLogin{}
			if err := json.Unmarshal(payload, &decoded); err == nil {
				mudlog.Debug("GMCP LOGIN", "username", decoded.Name, "password", strings.Repeat(`*`, len(decoded.Password)))
			}
		}

		return true
	}

	// Unhanlded IAC command, log it
	mudlog.Debug("Received", "type", "GMCP?", "data-size", len(iacCmd), "data-string", string(iacCmd), "data-bytes", iacCmd)

	return true
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

	if !isGMCPEnabled(connId) {
		gmcpSettings, ok := g.cache.Get(connId)
		if !ok {

			gmcpSettings = GMCPSettings{}
			g.cache.Add(connId, gmcpSettings)

			g.sendGMCPEnableRequest(connId)

			return events.Continue
		}

		// Get enabled modules... if none, skip out.
		if !gmcpSettings.GMCPAccepted {
			return events.Continue
		}
	}

	switch v := gmcp.Payload.(type) {
	case []byte:

		if len(gmcp.Module) > 0 {
			v = append([]byte(gmcp.Module+` `), v...)
		}

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			fmt.Println(string(v))
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(v), connId)
	case string:
		if len(gmcp.Module) > 0 {
			v = gmcp.Module + ` ` + v
		}

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			fmt.Println(string(v))
		}

		connections.SendTo(GmcpPayload.BytesWithPayload([]byte(v)), connId)
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
			fmt.Print(gmcp.Module + ` `)
			fmt.Println(string(prettyJSON.Bytes()))
		}

		if len(gmcp.Module) > 0 {
			payload = append([]byte(gmcp.Module+` `), payload...)
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(payload), connId)
	}

	return events.Continue
}
