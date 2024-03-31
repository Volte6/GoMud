package usercommands

import (
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

func Motd(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	response.SendUserMessage(userId, string(configs.GetConfig().Motd), true)

	response.Handled = true
	return response, nil
}
