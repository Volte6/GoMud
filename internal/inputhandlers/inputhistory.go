package inputhandlers

import "github.com/GoMudEngine/GoMud/internal/connections"

// All this does is manage the input history stack
func HistoryInputHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {
	// Save whatever was in the buffer when enter was hit as the last submitted
	if clientInput.EnterPressed {
		// copy the bytes over (If not just an enter press)
		if len(clientInput.Buffer) > 0 {
			clientInput.History.Add(clientInput.Buffer)
		}
	}

	return true
}
