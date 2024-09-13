package usercommands

import (
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

func Motd(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	response.SendUserMessage(userId, string(configs.GetConfig().Motd))

	response.Handled = true
	return response, nil
}
