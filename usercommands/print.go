package usercommands

import (
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Print(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	events.AddToQueue(events.Message{
		UserId: user.UserId,
		Text:   rest,
	})

	return true, nil
}
