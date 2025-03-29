package mobcommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/util"
)

func Equip(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.HasBuffFlag(buffs.PermaGear) {
		mob.Command(`emote struggles with their gear for a while, then gives up.`)
		return true, nil
	}

	if rest == "all" {
		itemCopies := []items.Item{}
		itemCopies = append(itemCopies, mob.Character.Items...)

		for _, item := range itemCopies {
			iSpec := item.GetSpec()
			if iSpec.Subtype == items.Wearable || iSpec.Type == items.Weapon {
				Equip(item.Name(), mob, room)
			}
		}
		return true, nil
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
			return true, nil
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
							fmt.Sprintf(`<ansi fg="mobname">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, mob.Character.Name, oldItem.DisplayName()))

						mob.Character.StoreItem(oldItem)
					}
				}

				if iSpec.Subtype == items.Wearable {

					room.SendText(
						fmt.Sprintf(`<ansi fg="mobname">%s</ansi> puts on <ansi fg="item">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))
				} else {
					room.SendText(
						fmt.Sprintf(`<ansi fg="mobname">%s</ansi> wields <ansi fg="item">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()))
				}

				mob.Character.Validate()

				events.AddToQueue(events.EquipmentChange{
					MobInstanceId: mob.InstanceId,
					ItemsWorn:     []items.Item{matchItem},
					ItemsRemoved:  oldItems,
				})
			}
		}

	}

	return true, nil
}
