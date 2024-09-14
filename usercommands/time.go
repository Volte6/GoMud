package usercommands

import (
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/gametime"
)

func Time(rest string, userId int) (bool, error) {

	gd := gametime.GetDate()

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	}

	events.AddToQueue(events.Message{
		UserId: userId,
		Text:   `It is now ` + gd.String() + `. It is ` + dayNight + `.`,
	})

	return true, nil
}
