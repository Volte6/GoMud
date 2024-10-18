package inputhandlers

import (
	"github.com/volte6/mud/connections"
	"github.com/volte6/mud/term"
)

func EchoInputHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// If no actual input, for now just do/change nothing
	if len(clientInput.DataIn) > 0 {
		// echo it back
		connections.SendTo(clientInput.DataIn, clientInput.ConnectionId)
	}

	// if they didn't hit enter, just keep buffering, go next.
	if !clientInput.EnterPressed {
		return false
	}

	// Echo back their Enter press
	connections.SendTo(term.CRLF, clientInput.ConnectionId)

	return true
}
