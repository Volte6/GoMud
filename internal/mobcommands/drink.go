package mobcommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
)

func Drink(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Check whether the user has an item in their inventory that matches
	if matchItem, found := mob.Character.FindInBackpack(rest); found {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Drinkable {
			return true, nil
		}

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		mob.Character.UseItem(matchItem)

		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> drinks <ansi fg="itemname">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))

		for _, buffId := range itemSpec.BuffIds {

			events.AddToQueue(events.Buff{
				UserId:        0,
				MobInstanceId: mob.InstanceId,
				BuffId:        buffId,
			})

		}
	}

	return true, nil
}
