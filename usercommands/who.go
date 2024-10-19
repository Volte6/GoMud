package usercommands

import (
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
)

func Who(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	details := room.GetRoomDetails(user)

	whoTxt, _ := templates.Process("descriptions/who", details)
	user.SendText(whoTxt)

	return true, nil
}
