package mobcommands

import (
	"fmt"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/events"
	"github.com/volte6/gomud/items"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
)

func Eat(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if matchItem, found := mob.Character.FindInBackpack(rest); found {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Edible {
			return true, nil
		}

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		mob.Character.UseItem(matchItem)

		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> eats some <ansi fg="itemname">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))

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
