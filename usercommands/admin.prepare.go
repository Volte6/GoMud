package usercommands

import (
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/util"
)

func Prepare(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.prepare", nil)
		response.Handled = true
		response.SendUserMessage(userId, infoOutput, false)
		return response, nil
	}

	allRoomIds := rooms.GetAllRoomIds()
	for _, roomId := range allRoomIds {
		room := rooms.LoadRoom(roomId)
		room.Prepare(false) // we are preparing all rooms, no need to check adjacent rooms
	}

	response.SendUserMessage(userId,
		"All rooms have been Prepare()'ed",
		true)

	response.Handled = true
	return response, nil
}
