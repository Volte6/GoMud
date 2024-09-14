package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Read(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

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
		user.SendText(fmt.Sprintf(`You don't have a "%s" that can be read.`, rest))
	} else {
		user.SendText(
			fmt.Sprintf(`You look at <ansi fg="item">%s</ansi>...`, foundItemLongName),
		)

		if !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> looks at their <ansi fg="item">%s</ansi>...`, user.Character.Name, foundItemName),
				userId,
			)
		}

		user.SendText(foundItemDescription)
	}

	return true, nil
}
