package usercommands

import (
	"github.com/volte6/mud/events"
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

	events.AddToQueue(events.Message{
		UserId: userId,
		Text:   `It is now ` + gd.String() + `. It is ` + dayNight + `.`,
	})

	response.Handled = true
	return response, nil
}
