package inputhandlers

import (
	"log/slog"

	"github.com/volte6/gomud/connections"
	"github.com/volte6/gomud/term"
)

func AnsiHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// Check for ANSI commands
	if !term.IsAnsiCommand(clientInput.DataIn) {
		return true
	}

	// Multiple Ansi Commands's can be stacked into one send, so useful to split them out
	ansiCmds := [][]byte{}

	var lastAnsiEsc int = 0
	for i, b := range clientInput.DataIn {
		if i != 0 && b == term.ANSI_ESC {
			ansiCmds = append(ansiCmds, clientInput.DataIn[lastAnsiEsc:i])
			lastAnsiEsc = i
		}
	}
	if lastAnsiEsc < len(clientInput.DataIn) {
		ansiCmds = append(ansiCmds, clientInput.DataIn[lastAnsiEsc:])
	}

	for _, ansiCmds := range ansiCmds {
		// Check incoming ANSI commands for anything useful...

		// Is it a screen size report?
		if ok, payload := term.Matches(ansiCmds, term.AnsiClientScreenSize); ok {

			w, h, err := term.AnsiParseScreenSizePayload(payload)
			if err != nil {
				slog.Info("Received", "type", "ANSI (Screensize)", "data", term.BytesString(payload), "error", err)
			} else {
				slog.Info("Received", "type", "ANSI (Screensize)", "width", w, "height", h)

				if err != nil {

					cs := connections.GetClientSettings(clientInput.ConnectionId)
					cs.Display.ScreenWidth = uint32(w)
					cs.Display.ScreenHeight = uint32(h)
					connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

				}
			}

			continue
		}

		// Is it a mouse click report?
		if ok, _ := term.Matches(ansiCmds, term.AnsiClientMouseDown); ok {
			// Ignore the down click, wait for the up before processing the click
			continue
		}

		if ok, payload := term.Matches(ansiCmds, term.AnsiClientMouseUp); ok {

			x, y, err := term.AnsiParseMouseClickPayload(payload)
			if err != nil {
				slog.Info("Received", "type", "ANSI (MouseClick)", "data", term.BytesString(payload), "error", err)
			} else {
				slog.Info("Received", "type", "ANSI (MouseClick)", "x", x, "y", y)
			}

			continue
		}

		if ok, payload := term.Matches(ansiCmds, term.AnsiMouseWheelUp); ok {

			x, y, err := term.AnsiParseMouseWheelScroll(payload)
			if err != nil {
				slog.Info("Received", "type", "ANSI (MouseWheelUp)", "data", term.BytesString(payload), "error", err)
			} else {
				slog.Info("Received", "type", "ANSI (MouseWheelUp)", "x", x, "y", y)
			}

			continue
		}

		if ok, payload := term.Matches(ansiCmds, term.AnsiMouseWheelDown); ok {
			x, y, err := term.AnsiParseMouseWheelScroll(payload)
			if err != nil {
				slog.Info("Received", "type", "ANSI (MouseWheelDown)", "data", term.BytesString(payload), "error", err)
			} else {
				slog.Info("Received", "type", "ANSI (MouseWheelDown)", "x", x, "y", y)
			}
			continue
		}

		if ok, _ := term.Matches(ansiCmds, term.AnsiMoveCursorUp); ok {
			slog.Info("Received", "type", "ANSI (MoveCursorUp)", "currentInput", string(clientInput.Buffer), "LastSubmitted", string(clientInput.LastSubmitted))

			// For each character in the buffer, backspace it out
			// Then add whatever was last submitted
			clientInput.DataIn = []byte{}

			bsSequence := []byte{}
			spaceSequence := []byte{}
			for i := 0; i < len(clientInput.Buffer); i++ {
				bsSequence = append(bsSequence, term.ASCII_BACKSPACE)
				spaceSequence = append(spaceSequence, term.ASCII_SPACE)
			}

			slog.Info("Received", "type", "ANSI (MoveCursorUp)", "bsSequence", len(bsSequence), "spaceSequence", len(spaceSequence))

			connections.SendTo(bsSequence, clientInput.ConnectionId)
			connections.SendTo(spaceSequence, clientInput.ConnectionId)
			connections.SendTo(bsSequence, clientInput.ConnectionId)

			clientInput.History.Previous()
			historicInput := clientInput.History.Get()
			clientInput.DataIn = make([]byte, len(historicInput))
			copy(clientInput.DataIn, historicInput)

			//clientInput.DataIn = make([]byte, len(clientInput.LastSubmitted))
			//copy(clientInput.DataIn, clientInput.LastSubmitted)

			clientInput.Buffer = []byte{}
			clientInput.EnterPressed = false
			nextHandler = true
			continue
		}

		if ok, _ := term.Matches(ansiCmds, term.AnsiMoveCursorDown); ok {
			slog.Info("Received", "type", "ANSI (MoveCursorDown)", "currentInput", string(clientInput.Buffer), "LastSubmitted", string(clientInput.LastSubmitted))

			// For each character in the buffer, backspace it out
			// Then add whatever was last submitted
			clientInput.DataIn = []byte{}

			bsSequence := []byte{}
			spaceSequence := []byte{}
			for i := 0; i < len(clientInput.Buffer); i++ {
				bsSequence = append(bsSequence, term.ASCII_BACKSPACE)
				spaceSequence = append(spaceSequence, term.ASCII_SPACE)
			}

			slog.Info("Received", "type", "ANSI (MoveCursorUp)", "bsSequence", len(bsSequence), "spaceSequence", len(spaceSequence))

			connections.SendTo(bsSequence, clientInput.ConnectionId)
			connections.SendTo(spaceSequence, clientInput.ConnectionId)
			connections.SendTo(bsSequence, clientInput.ConnectionId)

			clientInput.History.Next()
			historicInput := clientInput.History.Get()
			clientInput.DataIn = make([]byte, len(historicInput))
			copy(clientInput.DataIn, historicInput)

			//clientInput.DataIn = make([]byte, len(clientInput.LastSubmitted))
			//copy(clientInput.DataIn, clientInput.LastSubmitted)

			clientInput.Buffer = []byte{}
			clientInput.EnterPressed = false
			nextHandler = true
			continue
		}

		isF1, _ := term.Matches(ansiCmds, term.AnsiF1)
		if !isF1 { // check for Alternate F1
			isF1, _ = term.Matches(ansiCmds, term.AnsiF1b)
		}
		if isF1 {
			clientInput.DataIn = []byte("=1")
			clientInput.Buffer = []byte{}
			clientInput.EnterPressed = true
			// Since we are transforming this, pass it on
			nextHandler = true
			continue
		}

		// Unhanlded ANSI command, log it
		slog.Info("Received", "type", "ANSI", "size", len(ansiCmds), "data", term.AnsiCommandToString(ansiCmds))
	}

	return nextHandler
}
