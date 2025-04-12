package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Drink(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to drink.`, rest))
	} else {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Drinkable {
			user.SendText(
				fmt.Sprintf(`You can't drink <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
			)
			return true, nil
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		user.Character.UseItem(matchItem)

		user.SendText(fmt.Sprintf(`You drink the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> drinks <ansi fg="itemname">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()), user.UserId)

		for _, buffId := range itemSpec.BuffIds {
			user.AddBuff(buffId, `drink`)

		}
	}

	return true, nil
}
