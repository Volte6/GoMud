package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Equip(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	if rest == "all" {
		itemCopies := []items.Item{}
		itemCopies = append(itemCopies, mob.Character.Items...)

		for _, item := range itemCopies {
			iSpec := item.GetSpec()
			if iSpec.Subtype == items.Wearable || iSpec.Type == items.Weapon {
				Equip(item.Name(), mobId)
			}
		}
		response.Handled = true
		return response, nil
	}

	var matchItem items.Item = items.Item{}
	var found bool = false

	if rest == `random` {
		if len(mob.Character.Items) > 0 {
			matchItem = mob.Character.Items[util.Rand(len(mob.Character.Items))]
			found = true
		}
	}

	if !found {
		// Check whether the user has an item in their inventory that matches
		matchItem, found = mob.Character.FindInBackpack(rest)
	}

	if found {

		iSpec := matchItem.GetSpec()
		if iSpec.Type != items.Weapon && iSpec.Subtype != items.Wearable {
			response.Handled = true
			return response, nil
		}

		// Swap the item location
		oldItems, wearSuccess, _ := mob.Character.Wear(matchItem)

		if wearSuccess {

			mob.Character.RemoveItem(matchItem)

			// if there is only one item removed, and it's the same as the one put on, don't bother with the rest.
			// This is to address blind commands where mobs wear the same item over and over.
			if len(oldItems) == 1 && matchItem.Equals(oldItems[0]) {

				mob.Character.StoreItem(oldItems[0])

			} else {

				mob.Character.CancelBuffsWithFlag(buffs.Hidden)

				for _, oldItem := range oldItems {
					if oldItem.ItemId != 0 {

						room.SendText(
							fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, mob.Character.Name, oldItem.DisplayName()))

						mob.Character.StoreItem(oldItem)
					}
				}

				if iSpec.Subtype == items.Wearable {

					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> puts on <ansi fg="item">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))
				} else {
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> wields <ansi fg="item">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))
				}

				mob.Character.Validate()
			}
		}

	}

	response.Handled = true
	return response, nil
}
