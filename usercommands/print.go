package usercommands

import (
	"github.com/volte6/gomud/events"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Print(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	events.AddToQueue(events.Message{
		UserId: user.UserId,
		Text:   rest,
	})

	return true, nil
}
