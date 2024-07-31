package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Read(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	/*
		room := rooms.LoadRoom(user.Character.RoomId)
	*/

	// Check whether the user has an item in their inventory that matches

	foundItemName := ""
	foundItemLongName := ""
	foundItemDescription := ""
	// Search for an exact match first
	if readItem, found := user.Character.FindInBackpack(rest); found {
		iSpec := readItem.GetSpec()
		if iSpec.Type == items.Readable {
			foundItemName = readItem.DisplayName()
			foundItemLongName = readItem.DisplayName()
			foundItemDescription = string(readItem.GetBlob())
			if len(foundItemDescription) == 0 {
				foundItemDescription = iSpec.Description
			}
		}
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	if len(foundItemName) == 0 {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't have a "%s" that can be read.`, rest), true)
	} else {
		response.SendUserMessage(userId,
			fmt.Sprintf(`You look at <ansi fg="item">%s</ansi>...`, foundItemLongName),
			true)

		if !isSneaking {
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> looks at their <ansi fg="item">%s</ansi>...`, user.Character.Name, foundItemName),
				true)
		}

		response.SendUserMessage(userId, foundItemDescription, true)
	}

	response.Handled = true
	return response, nil
}
