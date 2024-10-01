package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Who(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	details := room.GetRoomDetails(user)

	whoTxt, _ := templates.Process("descriptions/who", details)
	user.SendText(whoTxt)

	return true, nil
}
