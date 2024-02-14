package usercommands

import (
	"fmt"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

func Time(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	_, hour, minute, ampm, isNight := configs.GetConfig().GetDate(util.GetRoundCount(), 0)
	dayNight := `day`
	if isNight {
		dayNight = `night`
	}

	response.SendUserMessage(userId, fmt.Sprintf("<ansi fg=\"%s\">It is now %d:%02d%s (%s).</ansi>", dayNight, hour, minute, ampm, dayNight), true)

	response.Handled = true
	return response, nil
}
