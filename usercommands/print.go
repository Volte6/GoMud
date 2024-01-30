package usercommands

import (
	"github.com/volte6/mud/util"
)

func Print(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	response.SendUserMessage(userId, rest, true)

	response.Handled = true
	return response, nil
}
