package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Gearup(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	wornItems := map[items.ItemType]items.Item{}
	wearNewItems := map[items.ItemType]items.Item{}

	allWornItems := user.Character.Equipment.GetAllItems()

	for _, itm := range allWornItems {
		wornItems[itm.GetSpec().Type] = itm
	}

	allBackpackItems := user.Character.GetAllBackpackItems()
	wearableCount := 0

	for _, itm := range allBackpackItems {
		itmSpec := itm.GetSpec()

		if itmSpec.Type != items.Weapon && itmSpec.Subtype != items.Wearable {
			continue
		}

		if itmSpec.Type == items.Weapon {
			// If it requires 2 hands, make sure it won't remove an offhand item!
			if user.Character.HandsRequired(itm) == 2 {
				if _, ok := wornItems[items.Offhand]; ok {
					continue
				}
			}
		}

		if itmSpec.Type == items.Offhand {
			// If it's offhand, make sure it won't remove an equipped two handed weapon
			if currentWeapon, ok := wornItems[items.Weapon]; ok {
				if user.Character.HandsRequired(currentWeapon) == 2 {
					continue

				}
			}
		}

		// Keep track of how many wearble items they hold
		wearableCount++

		// Skip items if something is already in that slot.
		if _, ok := wornItems[itmSpec.Type]; ok {
			continue
		}

		// If we've chosen something to wear in that slot already, consider this as a better option.
		if plannedItem, ok := wearNewItems[itmSpec.Type]; ok {
			if itmSpec.Value > plannedItem.GetSpec().Value {
				wearNewItems[itmSpec.Type] = itm
			}
			continue
		}

		// Getting here means there's nothing currently worn, so just accept the offering.
		wearNewItems[itmSpec.Type] = itm
	}

	if len(wearNewItems) == 0 {
		if wearableCount == 0 {
			user.SendText("You have nothing to wear.")
		} else {
			user.SendText("You're already wearing everything you can!")
		}
		return true, nil
	}

	for _, itm := range wearNewItems {
		user.Command(fmt.Sprintf(`wear !%d`, itm.ItemId), -1)
	}

	return true, nil
}
