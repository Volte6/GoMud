package usercommands

import (
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Prepare(rest string, user *users.UserRecord) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.prepare", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	allRoomIds := rooms.GetAllRoomIds()
	for _, roomId := range allRoomIds {
		room := rooms.LoadRoom(roomId)
		room.Prepare(false) // we are preparing all rooms, no need to check adjacent rooms
	}

	user.SendText(
		"All rooms have been Prepare()'ed",
	)

	return true, nil
}
