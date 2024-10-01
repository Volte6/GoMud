package usercommands

import (
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Quit(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Character.Aggro != nil {
		user.SendText("You're too busy to quit right now!")
		return true, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        user.UserId,
		MobInstanceId: 0,
		BuffId:        0,
	})

	return true, nil
}
