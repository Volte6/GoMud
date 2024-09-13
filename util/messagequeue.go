package util

// Response expected from user commands
type MessageQueue struct {
	Handled     bool   // Was this command recognized and handled? If not, display help etc.
	NextCommand string // Force another command to run right after this one
}

func NewMessageQueue(uId int, mId int) MessageQueue {
	return MessageQueue{}
}
