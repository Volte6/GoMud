package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Prepare(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.prepare", nil)
		response.Handled = true
		user.SendText(infoOutput)
		return response, nil
	}

	allRoomIds := rooms.GetAllRoomIds()
	for _, roomId := range allRoomIds {
		room := rooms.LoadRoom(roomId)
		room.Prepare(false) // we are preparing all rooms, no need to check adjacent rooms
	}

	user.SendText(
		"All rooms have been Prepare()'ed",
	)

	response.Handled = true
	return response, nil
}
