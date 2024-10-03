package inputhandlers

import (
	"log/slog"

	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/term"
)

func TelnetIACHandler(clientInput *connection.ClientInput, connectionPool *connection.ConnectionTracker, sharedState map[string]any) (nextHandler bool) {

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

	//slog.Info("Received", "type", "IAC (TEST)", "data", term.BytesString(clientInput.DataIn))

	if lastIAC < len(clientInput.DataIn) {
		iacCmds = append(iacCmds, clientInput.DataIn[lastIAC:])
	}

	for _, iacCmd := range iacCmds {
		// Check incoming Telnet IAC commands for anything useful...

		if ok, payload := term.Matches(iacCmd, term.TelnetAcceptedChangeCharset); ok {
			slog.Info("Received", "type", "IAC (TelnetAcceptedChangeCharset)", "data", term.BytesString(payload))
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetRejectedChangeCharset); ok {
			slog.Info("Received", "type", "IAC (TelnetRejectedChangeCharset)")
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetAgreeChangeCharset); ok {
			slog.Info("Received", "type", "IAC (TelnetAgreeChangeCharset)")
			connectionPool.SendTo(
				term.TelnetCharset.BytesWithPayload([]byte(" UTF-8")),
				clientInput.ConnectionId,
			)
			continue
		}

		// Is it a screen size report?
		if ok, payload := term.Matches(iacCmd, term.TelnetScreenSizeResponse); ok {

			w, h, err := term.TelnetParseScreenSizePayload(payload)
			if err != nil {
				slog.Info("Received", "type", "IAC (Screensize)", "data", term.BytesString(payload), "error", err)
			} else {
				slog.Info("Received", "type", "IAC (Screensize)", "width", w, "height", h)

				if err == nil {

					events.AddToQueue(events.ClientSettings{
						ConnectionId: clientInput.ConnectionId,
						ScreenWidth:  uint32(w),
						ScreenHeight: uint32(h),
					})

				}

			}

			continue
		}

		// Unhanlded IAC command, log it
		slog.Info("Received", "type", "IAC", "size", len(clientInput.DataIn), "data", term.TelnetCommandToString(iacCmd))

	}

	// We handled it, so don't pass it on
	return false
}
