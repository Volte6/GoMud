package inputhandlers

import (
	"encoding/json"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/term"
)

func TelnetIACHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// Check for Telnet IAC commands
	// If not, pass it on to next handler
	if !term.IsTelnetCommand(clientInput.DataIn) {
		return true
	}

	// Multiple Telnet IAC's can be stacked into one send, so useful to split them out
	iacCmds := [][]byte{}

	var lastIAC int = 0
	for i, b := range clientInput.DataIn {
		if i != 0 && b == term.TELNET_IAC {
			if i < len(clientInput.DataIn)-1 && clientInput.DataIn[i+1] != term.TELNET_SE {
				iacCmds = append(iacCmds, clientInput.DataIn[lastIAC:i])
				lastIAC = i
			}
		}
	}

	//mudlog.Debug("Received", "type", "IAC (TEST)", "data", term.BytesString(clientInput.DataIn))

	if lastIAC < len(clientInput.DataIn) {
		iacCmds = append(iacCmds, clientInput.DataIn[lastIAC:])
	}

	for _, iacCmd := range iacCmds {
		// Check incoming Telnet IAC commands for anything useful...

		if term.IsGMCPCommand(iacCmd) {

			if ok, payload := term.Matches(iacCmd, term.GmcpAccept); ok {
				mudlog.Debug("Received", "type", "IAC (Client-GMCP Accept)", "data", term.BytesString(payload))
				continue
			}

			if ok, payload := term.Matches(iacCmd, term.GmcpRefuse); ok {
				mudlog.Debug("Received", "type", "IAC (Client-GMCP Refuse)", "data", term.BytesString(payload))
				continue
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

				if _, ok := term.SupportedGMCP[command]; !ok {
					mudlog.Error("Received", "type", "GMCP (Ignored)", "command", command, "payload", string(payload))
					continue
				}

				mudlog.Debug("Received", "type", "GMCP (Handling)", "command", command, "payload", string(payload))

				switch command {

				case `External.Discord.Hello`:
					decoded := term.GMCPDiscord{}
					if err := json.Unmarshal(payload, &decoded); err == nil {
						cs := connections.GetClientSettings(clientInput.ConnectionId)
						cs.Discord.User = decoded.User
						cs.Discord.Private = decoded.Private
						connections.OverwriteClientSettings(clientInput.ConnectionId, cs)
					}
				case `Core.Hello`:
					decoded := term.GMCPHello{}
					if err := json.Unmarshal(payload, &decoded); err == nil {
						cs := connections.GetClientSettings(clientInput.ConnectionId)
						cs.Client.Name = decoded.Client
						cs.Client.Version = decoded.Version
						if strings.EqualFold(decoded.Client, `mudlet`) {
							cs.Client.IsMudlet = true
						}
						connections.OverwriteClientSettings(clientInput.ConnectionId, cs)
					}
				case `Core.Supports.Set`:
					decoded := term.GMCPSupportsSet{}
					if err := json.Unmarshal(payload, &decoded); err == nil {
						cs := connections.GetClientSettings(clientInput.ConnectionId)
						cs.GMCPModules = decoded.GetSupportedModules()
						connections.OverwriteClientSettings(clientInput.ConnectionId, cs)
					}
				case `Core.Supports.Remove`:
					decoded := term.GMCPSupportsRemove{}
					if err := json.Unmarshal(payload, &decoded); err == nil {
						cs := connections.GetClientSettings(clientInput.ConnectionId)
						if len(cs.GMCPModules) > 0 {
							for _, name := range decoded {
								delete(cs.GMCPModules, name)
							}
						}
						connections.OverwriteClientSettings(clientInput.ConnectionId, cs)
					}
				case `Char.Login`:
					decoded := term.GMCPLogin{}
					if err := json.Unmarshal(payload, &decoded); err == nil {
						mudlog.Debug("GMCP LOGIN", "username", decoded.Name, "password", decoded.Password)
					}
				}

				continue
			}

			// Unhanlded IAC command, log it
			mudlog.Debug("Received", "type", "GMCP?", "size", len(iacCmd), "data", string(iacCmd))

			continue
		}

		if term.IsMSPCommand(iacCmd) {

			if ok, payload := term.Matches(iacCmd, term.MspAccept); ok {
				mudlog.Debug("Received", "type", "IAC (Client-MSP Accept)", "data", term.BytesString(payload))

				cs := connections.GetClientSettings(clientInput.ConnectionId)
				cs.MSPEnabled = true
				connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

				connections.SendTo(
					term.MspCommand.BytesWithPayload([]byte("!!SOUND(Off U="+configs.GetFilePathsConfig().WebCDNLocation.String()+")")),
					clientInput.ConnectionId,
				)

				continue
			}

			if ok, payload := term.Matches(iacCmd, term.MspRefuse); ok {
				mudlog.Debug("Received", "type", "IAC (Client-MSP Refuse)", "data", term.BytesString(payload))

				cs := connections.GetClientSettings(clientInput.ConnectionId)
				cs.MSPEnabled = false
				connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

				continue
			}

			continue
		}

		if ok, payload := term.Matches(iacCmd, term.TelnetAcceptedChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetAcceptedChangeCharset)", "data", term.BytesString(payload))
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetRejectedChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetRejectedChangeCharset)")
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetAgreeChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetAgreeChangeCharset)")
			connections.SendTo(
				term.TelnetCharset.BytesWithPayload([]byte(" UTF-8")),
				clientInput.ConnectionId,
			)
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetDontSuppressGoAhead); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetDontSuppressGoAhead)")

			cs := connections.GetClientSettings(clientInput.ConnectionId)
			cs.SendTelnetGoAhead = true
			connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

			continue
		}

		// Is it a screen size report?
		if ok, payload := term.Matches(iacCmd, term.TelnetScreenSizeResponse); ok {

			w, h, err := term.TelnetParseScreenSizePayload(payload)
			if err != nil {
				mudlog.Debug("Received", "type", "IAC (Screensize)", "data", term.BytesString(payload), "error", err)
			} else {
				mudlog.Debug("Received", "type", "IAC (Screensize)", "width", w, "height", h)

				if err == nil {

					cs := connections.GetClientSettings(clientInput.ConnectionId)
					cs.Display.ScreenWidth = uint32(w)
					cs.Display.ScreenHeight = uint32(h)
					connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

				}

			}

			continue
		}

		// Unhanlded IAC command, log it
		mudlog.Debug("Received", "type", "IAC (Unhandled)", "size", len(clientInput.DataIn), "data", term.TelnetCommandToString(iacCmd))

	}

	// We handled it, so don't pass it on
	return false
}
