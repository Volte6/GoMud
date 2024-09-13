package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Remove(rest string, mobId int) (util.MessageQueue, error) {

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
		for _, item := range mob.Character.Equipment.GetAllItems() {
			r, _ := Remove(item.Name(), mobId)
			response.AbsorbMessages(r)
		}
		response.Handled = true
		return response, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindOnBody(rest)

	if found && matchItem.ItemId > 0 {

		if mob.Character.RemoveFromBody(matchItem) {

			mob.Character.CancelBuffsWithFlag(buffs.Hidden)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, mob.Character.Name, matchItem.DisplayName()),
			)

			mob.Character.StoreItem(matchItem)
		}

		mob.Character.Validate()

	}

	response.Handled = true
	return response, nil
}
