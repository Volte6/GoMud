package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
)

func Stash(rest string, userId int) (bool, error) {

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
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to stash.", rest))
	} else {
		// Swap the item location
		room.AddItem(matchItem, true)
		user.Character.RemoveItem(matchItem)

		isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

		user.SendText(
			fmt.Sprintf(`You stash the <ansi fg="itemname">%s</ansi>. To get it back, try <ansi fg="command">get %s from stash</ansi>`, matchItem.DisplayName(), matchItem.DisplayName()))

		if !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is attempting to look unsuspicious.`, user.Character.Name),
				userId)
		}

		// Trigger lost event
		scripting.TryItemScriptEvent(`onLost`, matchItem, userId)
	}

	return true, nil
}
