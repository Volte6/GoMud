package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
)

func Trash(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to trash.`, rest))
	} else {

		isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

		user.Character.RemoveItem(matchItem)

		user.SendText(
			fmt.Sprintf(`You trash the <ansi fg="item">%s</ansi> for good.`, matchItem.DisplayName()))

		if !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> destroys <ansi fg="item">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
				user.UserId)
		}

		iSpec := matchItem.GetSpec()

		grantXP, xpScale := user.Character.GrantXP(int(float64(iSpec.Value) / 10))

		xpMsgExtra := ``
		if xpScale != 100 {
			xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
		}

		user.SendText(
			fmt.Sprintf(`You gained <ansi fg="yellow-bold">%d experience points</ansi>%s!`, grantXP, xpMsgExtra))

		// Trigger lost event
		scripting.TryItemScriptEvent(`onLost`, matchItem, user.UserId)

	}

	return true, nil
}
