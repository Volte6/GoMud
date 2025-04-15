package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
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
			mob.AddBuff(buffId, `food`)
		}
	}

	return true, nil
}
