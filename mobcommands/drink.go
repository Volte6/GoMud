package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Drink(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	// Check whether the user has an item in their inventory that matches
	if matchItem, found := mob.Character.FindInBackpack(rest); found {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Drinkable {
			response.Handled = true
			return response, nil
		}

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		mob.Character.UseItem(matchItem)

		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> drinks <ansi fg="itemname">%s</ansi>.`, mob.Character.Name, matchItem.DisplayName()), true)

		for _, buffId := range itemSpec.BuffIds {
			cmdQueue.QueueBuff(0, mob.InstanceId, buffId)
		}
	}

	response.Handled = true
	return response, nil
}
