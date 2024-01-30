package roundinfo

import (
	"time"

	"github.com/volte6/mud/term"
)

type Message struct {
	ToRoom bool
	Msg    string
}

// Hypthetical structure to hold information about the round that happened
// Could be returned by whoever is processing the round
type RoundInfo struct {
	ProcessingTime time.Duration
	Messages       []Message
}

func (r *RoundInfo) SendUserMessage(message string, newLine bool) {
	if newLine {
		message += term.CRLFStr
	}
	r.Messages = append(r.Messages, Message{
		ToRoom: false,
		Msg:    message,
	})
}

func (r *RoundInfo) SendRoomMessage(message string, newLine bool) {
	if newLine {
		message += term.CRLFStr
	}
	r.Messages = append(r.Messages, Message{
		ToRoom: true,
		Msg:    message,
	})
}

func (r *RoundInfo) AbsorbMessages(other *RoundInfo) {
	r.Messages = append(r.Messages, other.Messages...)
}

func New() *RoundInfo {
	return &RoundInfo{
		Messages: make([]Message, 0),
	}
}
