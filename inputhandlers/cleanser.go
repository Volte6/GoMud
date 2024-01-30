package inputhandlers

import (
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/term"
)

// CleanserInputHandler's job is to remove any bad characters from the input stream
// before passing it down the chain.
// For this reason, it's important it happen before other text processing handlers
func CleanserInputHandler(clientInput *connection.ClientInput, connectionPool *connection.ConnectionTracker, sharedState map[string]any) (nextHandler bool) {

	if len(clientInput.DataIn) < 1 {
		return true
	}

	// backspace
	dIn := clientInput.DataIn[len(clientInput.DataIn)-1]

	if dIn == term.ASCII_DELETE || dIn == term.ASCII_BACKSPACE {
		//connectionPool.SendTo([]byte(term.AnsiMoveCursorBackward.String()+" "+term.AnsiMoveCursorBackward.String()), connDetails.UniqueId())
		// send backspace, space, backspace
		if len(clientInput.Buffer) > 0 {
			connectionPool.SendTo([]byte{term.ASCII_BACKSPACE, term.ASCII_SPACE, term.ASCII_BACKSPACE}, clientInput.ConnectionId)
			clientInput.Buffer = clientInput.Buffer[:len(clientInput.Buffer)-1]
		}
		clientInput.DataIn = clientInput.DataIn[:len(clientInput.DataIn)-1]
		return true
	}

	// Check if the last byte is a CR or LF or NULL
	if dIn <= 13 {
		if clientInput.DataIn[len(clientInput.DataIn)-1] == 0 || clientInput.DataIn[len(clientInput.DataIn)-1] == 10 || clientInput.DataIn[len(clientInput.DataIn)-1] == 13 {
			clientInput.EnterPressed = true
		}
	}

	// Remove non printing chars
	clientInput.DataIn = trimNonPrintingBytes(clientInput.DataIn)

	// Add all input to the currentBuffer
	clientInput.Buffer = append(clientInput.Buffer, clientInput.DataIn...)

	return true
}

// Trims non printing bytes and SPACE from front/back of a byte slice
func trimNonPrintingBytes(b []byte) []byte {
	start := 0
	for ; start < len(b); start++ {
		c := b[start]
		if c > 31 && c < 127 {
			break
		}
	}

	stop := len(b)
	for ; stop > start; stop-- {
		c := b[stop-1]
		if c > 31 && c < 127 {
			break
		}
	}

	return b[start:stop]
}
