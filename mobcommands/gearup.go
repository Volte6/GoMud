package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Gearup(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	if rest != `` {
		// Check whether the user has an item in their inventory that matches
		matchItem, found := mob.Character.FindInBackpack(rest)

		if found {

			matchSpec := matchItem.GetSpec()
			for _, itm := range mob.Character.Equipment.GetAllItems() {
				itmSpec := itm.GetSpec()
				if itmSpec.Type == matchSpec.Type && matchSpec.Value > itmSpec.Value {
					cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`wear !%d`, matchItem.ItemId))
					cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`drop !%d`, itm.ItemId))
				}
			}

		}
	} else {
		wornItems := map[items.ItemType]items.Item{}
		wearNewItems := map[items.ItemType]items.Item{}

		allWornItems := mob.Character.Equipment.GetAllItems()

		for _, itm := range allWornItems {
			wornItems[itm.GetSpec().Type] = itm
		}

		allBackpackItems := mob.Character.GetAllBackpackItems()

		fmt.Println()
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

		isCharmed := mob.Character.IsCharmed()
		for _, itm := range wearNewItems {
			cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`wear !%d`, itm.ItemId))
			// drop the old one
			if isCharmed {
				if oldItm, ok := wornItems[itm.GetSpec().Type]; ok {
					cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`drop !%d`, oldItm.ItemId))
				}
			}
		}
		fmt.Println()
	}

	response.Handled = true
	return response, nil
}
