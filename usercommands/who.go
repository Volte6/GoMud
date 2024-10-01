package usercommands

import (
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Who(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	details := room.GetRoomDetails(user)

	whoTxt, _ := templates.Process("descriptions/who", details)
	user.SendText(whoTxt)

	return true, nil
}
