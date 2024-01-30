package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Give(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		response.Handled = true
		return response, nil
	}

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if strings.ToLower(args[1]) == "gold" {

		g, _ := strconv.ParseInt(args[0], 10, 32)
		giveGoldAmount = int(g)

		args = args[2:]

		if giveGoldAmount > mob.Character.Gold {
			response.Handled = true
			return response, nil
		}

		if len(args) < 1 {
			response.Handled = true
			return response, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = mob.Character.FindInBackpack(args[0])

		if !found {
			response.Handled = true
			return response, nil
		}

		args = args[1:]
	}

	playerId, mobId := room.FindByName(args[len(args)-1])

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			mob.Character.RemoveItem(giveItem)

			iSpec := giveItem.GetSpec()
			if iSpec.QuestToken != `` {
				cmdQueue.QueueQuest(targetUser.UserId, iSpec.QuestToken)
			}

			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, giveItem.Name()),
				true)

		} else if giveGoldAmount > 0 {

			targetUser.Character.Gold += giveGoldAmount
			mob.Character.Gold -= giveGoldAmount

			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, giveGoldAmount),
				true)

		}

		response.Handled = true
		return response, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		m := mobs.GetInstance(mobId)

		if m != nil {

			// Swap the item location
			if giveItem.ItemId > 0 {
				m.Character.StoreItem(giveItem)
				mob.Character.RemoveItem(giveItem)

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, giveItem.Name(), m.Character.Name),
					true)
			} else if giveGoldAmount > 0 {

				m.Character.Gold += giveGoldAmount
				mob.Character.Gold -= giveGoldAmount

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, m.Character.Name),
					true)
			}

		}

	}

	response.Handled = true
	return response, nil
}
