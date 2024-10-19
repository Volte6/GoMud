package usercommands

import (
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
)

func Exits(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	details := room.GetRoomDetails(user)

	exitTxt, _ := templates.Process("descriptions/exits", details)
	user.SendText(exitTxt)

	return true, nil
}
