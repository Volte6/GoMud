package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
)

func Trash(rest string, mobId int) (bool, error) {

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindInBackpack(rest)

	if found {
		mob.Character.RemoveItem(matchItem)

		// Trashing items may be useful for quest stuff
		// So don't wanna tell players mob is doing it
		/*
			isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> destroys <ansi fg="item">%s</ansi>...`, mob.Character.Name, matchItem.DisplayName()),
					true)
			}
		*/

	}

	return true, nil
}
