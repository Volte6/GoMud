package usercommands

import (
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Gearup(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

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
		cmdQueue.QueueCommand(userId, 0, fmt.Sprintf(`wear !%d`, itm.ItemId))
	}

	response.Handled = true
	return response, nil
}
