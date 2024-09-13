package usercommands

import (
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/util"
)

func Time(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	gd := gametime.GetDate()

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	response.SendUserMessage(userId, `It is now `+gd.String()+`. It is `+dayNight+`.`, true)

	response.Handled = true
	return response, nil
}
