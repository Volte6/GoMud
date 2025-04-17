package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/GoMudEngine/GoMud/internal/configs"
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

	// Set up a callback to run after the system is fully loaded
	gmcpModule.plug.Callbacks.SetOnLoad(func() {
		ensureExternalModulesRegistered()
	})
}

func isGMCPEnabled(connectionId uint64) bool {

	//return true

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

	if _, ok := g.cache.Get(evt.ConnectionId); !ok {
		g.cache.Add(evt.ConnectionId, GMCPSettings{})
		// Send request to enable GMCP
		g.sendGMCPEnableRequest(evt.ConnectionId)
	}

	// Only try to send Discord info when a player joins with a Mudlet client
	// We'll check inside sendExternalDiscordInfo if it's actually a Mudlet client
	// and only log if it is
	if user := users.GetByUserId(evt.UserId); user != nil {
		g.sendExternalDiscordInfo(user)
		
		// Check if this is a Mudlet client and send UI notification if appropriate
		connectionId := uint64(user.ConnectionId())
		gmcpData, ok := g.cache.Get(connectionId)
		if ok && gmcpData.Client.IsMudlet {
			// Check if the user has disabled the UI notice
			if suppress, ok := user.GetConfigOption("ui_suppress_mudlet_notice").(bool); !ok || !suppress {
				// Send the notification about the UI
				user.SendText("\n<ansi fg=\"highlight\">Mudlet client detected.</ansi> If you want to use our Mudlet UI, you can use the '<ansi fg=\"command\">ui install</ansi>' command to install it into Mudlet automatically.")
			}
		}
	}

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
		
		// We'll check for Mudlet later in Core.Hello, not here
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

				// Record if this is a Mudlet client
				isMudlet := strings.EqualFold(decoded.Client, `mudlet`)
				gmcpData.Client.IsMudlet = isMudlet
				
				// Only send Discord and client info for Mudlet clients
				if isMudlet {
					mudlog.Info("Client", "status", "Mudlet client detected", "connectionId", connectionId)
					userId := g.findUserIdForConnection(connectionId)
					if userId > 0 {
						user := users.GetByUserId(userId)
						if user != nil {
							g.sendExternalDiscordInfo(user)
							
							// Send UI notification if appropriate
							if suppress, ok := user.GetConfigOption("ui_suppress_mudlet_notice").(bool); !ok || !suppress {
								// Send the notification about the UI
								user.SendText("\n<ansi fg=\"highlight\">Mudlet client detected.</ansi> If you want to use our Mudlet UI, you can use the '<ansi fg=\"command\">ui install</ansi>' command to install it into Mudlet automatically.")
							}
						}
					}
				} else {
					mudlog.Info("GMCP", "status", "Non-Mudlet client detected", "client", decoded.Client)
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

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, v, "", "\t")
			fmt.Print(gmcp.Module + ` `)
			fmt.Println(string(prettyJSON.Bytes()))
		}

		// Regular code follows...
		if len(gmcp.Module) > 0 {
			v = append([]byte(gmcp.Module+` `), v...)
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(v), connId)
	case string:

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, []byte(v), "", "\t")
			fmt.Print(gmcp.Module + ` `)
			fmt.Println(string(prettyJSON.Bytes()))
		}

		// Regular code follows...
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

		// DEBUG ONLY
		// TODO: REMOVE
		if gmcp.UserId == 1 {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, payload, "", "\t")
			fmt.Print(gmcp.Module + ` `)
			fmt.Println(string(prettyJSON.Bytes()))
		}

		// Regular code follows...
		if len(gmcp.Module) > 0 {
			payload = append([]byte(gmcp.Module+` `), payload...)
		}

		connections.SendTo(GmcpPayload.BytesWithPayload(payload), connId)
	}

	return events.Continue
}

// Helper method to find a user ID for a connection ID
func (g *GMCPModule) findUserIdForConnection(connectionId uint64) int {
	// Try to find the user that has this connection ID
	for _, user := range users.GetAllActiveUsers() {
		if uint64(user.ConnectionId()) == connectionId {
			return user.UserId
		}
	}
	return 0
}

// This function ensures Discord info is sent to all Mudlet clients
func ensureExternalModulesRegistered() {
	// Only log when we actually find a Mudlet client
	mudletFound := false
	
	// Check all active users with Mudlet clients
	for _, user := range users.GetAllActiveUsers() {
		// Only send to Mudlet clients
		if gmcpData, ok := gmcpModule.cache.Get(uint64(user.ConnectionId())); ok && gmcpData.Client.IsMudlet {
			if !mudletFound {
				mudlog.Info("GMCP", "status", "Detected existing Mudlet client(s)")
				mudletFound = true
			}
			
			// Send all GMCP modules - the sendExternalDiscordInfo function will log each module
			gmcpModule.sendExternalDiscordInfo(user)
		}
	}
}

// Helper function to directly send Discord info for Mudlet clients
func (g *GMCPModule) sendExternalDiscordInfo(user *users.UserRecord) {
	// Check if this is a Mudlet client first
	connectionId := uint64(user.ConnectionId())
	gmcpData, ok := g.cache.Get(connectionId)
	if !ok || !gmcpData.Client.IsMudlet {
		// Not a Mudlet client, don't send Discord info or log anything
		return
	}
	
	// Get config for MUD name
	c := configs.GetConfig()
	mudName := "GoMud"
	if c.Server.MudName != "" {
		mudName = string(c.Server.MudName)
	}
	
	// Send Discord Info
	infoPayload := `{ 
		"inviteurl": "https://discord.gg/FaauSYej3n",	
		"applicationid": "1234",
		"largeImageKey": "server-icon"
	}`

	// Log specific module being sent
	mudlog.Info("GMCP", "action", "Sending External.Discord.Info to Mudlet user", "userId", user.UserId)
	
	events.AddToQueue(GMCPOut{
		UserId:  user.UserId,
		Module:  `External.Discord.Info`,
		Payload: infoPayload,
	})

	// Send Discord Status
	statusPayload := `{ 
		"game": "` + mudName + `",
		"startTimestamp": ` + strconv.FormatInt(user.GetConnectTime().Unix(), 10) + `,
		"state": "Playing GoMud",
		"details": "using GoMud Engine"
	}`

	// Log specific module being sent
	mudlog.Info("GMCP", "action", "Sending External.Discord.Status to Mudlet user", "userId", user.UserId)
	
	events.AddToQueue(GMCPOut{
		UserId:  user.UserId,
		Module:  `External.Discord.Status`,
		Payload: statusPayload,
	})
	
	// Also send Mudlet client info
	mudletClientModule.SendClientInfoExported(user.UserId)
}
