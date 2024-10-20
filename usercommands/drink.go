package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/events"
	"github.com/volte6/gomud/items"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Drink(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

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

			events.AddToQueue(events.Buff{
				UserId:        user.UserId,
				MobInstanceId: 0,
				BuffId:        buffId,
			})

		}
	}

	return true, nil
}
