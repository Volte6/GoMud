package inputhandlers

import (
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/term"
)

func EchoInputHandler(clientInput *connection.ClientInput, connectionPool *connection.ConnectionTracker, sharedState map[string]any) (nextHandler bool) {

	// If no actual input, for now just do/change nothing
	if len(clientInput.DataIn) > 0 {
		// echo it back
		connectionPool.SendTo(clientInput.DataIn, clientInput.ConnectionId)
	}

	// if they didn't hit enter, just keep buffering, go next.
	if !clientInput.EnterPressed {
		return false
	}

	// Echo back their Enter press
	connectionPool.SendTo(term.CRLF, clientInput.ConnectionId)

	return true
}
