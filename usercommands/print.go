package usercommands

import (
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/util"
)

func Print(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	events.AddToQueue(events.Message{
		UserId: userId,
		Text:   rest,
	})

	response.Handled = true
	return response, nil
}
