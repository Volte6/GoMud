package usercommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Gearup(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	wornItems := map[items.ItemType]items.Item{}
	wearNewItems := map[items.ItemType]items.Item{}

	allWornItems := user.Character.Equipment.GetAllItems()

	for _, itm := range allWornItems {
		wornItems[itm.GetSpec().Type] = itm
	}

	allBackpackItems := user.Character.GetAllBackpackItems()

	for _, itm := range allBackpackItems {
		itmSpec := itm.GetSpec()

		// Is there already a new item ready for that slot? Compare to that.
		if plannedItem, ok := wearNewItems[itmSpec.Type]; ok {
			if itmSpec.Value > plannedItem.GetSpec().Value {
				wearNewItems[itmSpec.Type] = itm
			}
			continue
		}

		// If we get here, there hasn't been anything to replace the current gear yet.
		if wornItem, ok := wornItems[itmSpec.Type]; ok {
			if itmSpec.Value > wornItem.GetSpec().Value {
				wearNewItems[itmSpec.Type] = itm
			}
			continue
		}

		// Getting here means there's nothing currently worn, so just accept the offering.
		wearNewItems[itmSpec.Type] = itm
	}

	for _, itm := range wearNewItems {
		user.Command(fmt.Sprintf(`wear !%d`, itm.ItemId), -1)
	}

	return true, nil
}
