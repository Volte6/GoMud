package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Quit(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

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
