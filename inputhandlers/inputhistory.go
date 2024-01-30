package inputhandlers

import "github.com/volte6/mud/connection"

// All this does is manage the input history stack
func HistoryInputHandler(clientInput *connection.ClientInput, connectionPool *connection.ConnectionTracker, sharedState map[string]any) (nextHandler bool) {
	// Save whatever was in the buffer when enter was hit as the last submitted
	if clientInput.EnterPressed {
		// copy the bytes over

		clientInput.History.Add(clientInput.Buffer)

		//clientInput.LastSubmitted = make([]byte, len(clientInput.Buffer))
		//copy(clientInput.LastSubmitted, clientInput.Buffer)
	}

	return true
}
