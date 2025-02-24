package usercommands

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Prepare(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
