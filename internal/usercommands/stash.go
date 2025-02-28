package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Stash(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf("You don't have a %s to stash.", rest))
	} else {
		// Swap the item location

		matchItem.StashedBy = user.UserId

		room.AddItem(matchItem, true)
		user.Character.RemoveItem(matchItem)

		events.AddToQueue(events.ItemOwnership{
			UserId: user.UserId,
			Item:   matchItem,
			Gained: false,
		})

		isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

		user.SendText(
			fmt.Sprintf(`You stash the <ansi fg="itemname">%s</ansi>. To get it back, try <ansi fg="command">get %s from stash</ansi>`, matchItem.DisplayName(), matchItem.DisplayName()))

		if !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is attempting to look unsuspicious.`, user.Character.Name),
				user.UserId)
		}

	}

	return true, nil
}
