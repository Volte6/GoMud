package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Give(rest string, mobId int) (util.MessageQueue, error) {

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

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		response.Handled = true
		return response, nil
	}

	var giveWho string = args[len(args)-1]
	args = args[:len(args)-1]
	var giveWhat string = strings.Join(args, " ")

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if len(giveWhat) > 4 && giveWhat[len(giveWhat)-4:] == "gold" {

		g, _ := strconv.ParseInt(giveWhat[0:len(giveWhat)-5], 10, 32)
		giveGoldAmount = int(g)

		if giveGoldAmount > mob.Character.Gold {
			response.Handled = true
			return response, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = mob.Character.FindInBackpack(giveWhat)

		if !found {
			response.Handled = true
			return response, nil
		}

	}

	playerId, mobId := room.FindByName(giveWho)

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			mob.Character.RemoveItem(giveItem)

			iSpec := giveItem.GetSpec()
			if iSpec.QuestToken != `` {

				events.AddToQueue(events.Quest{
					UserId:     targetUser.UserId,
					QuestToken: iSpec.QuestToken,
				})

			}

			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName()),
			)

		} else if giveGoldAmount > 0 {

			targetUser.Character.Gold += giveGoldAmount
			mob.Character.Gold -= giveGoldAmount

			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, giveGoldAmount),
			)

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
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName(), m.Character.Name),
				)
			} else if giveGoldAmount > 0 {

				m.Character.Gold += giveGoldAmount
				mob.Character.Gold -= giveGoldAmount

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, m.Character.Name),
				)
			}

		}

	}

	response.Handled = true
	return response, nil
}
