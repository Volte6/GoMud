package usercommands

import (
	"github.com/volte6/mud/events"
)

func Print(rest string, userId int) (bool, string, error) {

	events.AddToQueue(events.Message{
		UserId: userId,
		Text:   rest,
	})

	return true, ``, nil
}
