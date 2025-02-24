package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
)

func Eat(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to eat.`, rest))
	} else {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Edible {
			user.SendText(
				fmt.Sprintf(`You can't eat <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
			)
			return true, nil
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		user.SendText(fmt.Sprintf(`You eat some of the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> eats some <ansi fg="itemname">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()), user.UserId)

		// If no more uses, will be lost, so trigger event
		if usesLeft := user.Character.UseItem(matchItem); usesLeft < 1 {
			scripting.TryItemScriptEvent(`onLost`, matchItem, user.UserId)
		}

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
