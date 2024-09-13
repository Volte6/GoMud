package usercommands

import (
	"github.com/volte6/mud/util"
)

func Print(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	response.SendUserMessage(userId, rest)

	response.Handled = true
	return response, nil
}
