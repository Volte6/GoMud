package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/util"
)

func Alchemy(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if args[0] == "random" {
		// select a random item
		if len(mob.Character.Items) > 0 {
			matchItem := mob.Character.Items[util.Rand(len(mob.Character.Items))]
			Alchemy(matchItem.Name(), mob, room)

		}
		return true, nil
	}

	if args[0] == "all" {

		iCopies := []items.Item{}
		for _, item := range mob.Character.Items {
			iCopies = append(iCopies, item)
		}

		for _, item := range iCopies {
			Alchemy(item.Name(), mob, room)
		}

		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindInBackpack(rest)

	if found {

		mob.Character.RemoveItem(matchItem)
		mob.Character.Gold += 1
		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> chants softly. Their <ansi fg="item">%s</ansi> slowly levitates in the air, trembles briefly and then in a flash of light becomes a gold coin!`, mob.Character.Name, matchItem.DisplayName()))
	}

	return true, nil
}
