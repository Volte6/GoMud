package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Drop(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	if args[0] == "all" {

		iCopies := []items.Item{}

		if mob.Character.Gold > 0 {
			r, _ := Drop(fmt.Sprintf("%d gold", mob.Character.Gold), mobId, cmdQueue)
			response.AbsorbMessages(r)
		}

		for _, item := range mob.Character.Items {
			iCopies = append(iCopies, item)
		}

		for _, item := range iCopies {
			r, _ := Drop(item.Name(), mobId, cmdQueue)
			response.AbsorbMessages(r)
		}

		response.Handled = true
		return response, nil
	}

	// Drop 10 gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt < 1 {
			response.Handled = true
			return response, nil
		}

		if dropAmt <= mob.Character.Gold {

			room.Gold += dropAmt
			mob.Character.Gold -= dropAmt

			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> drops <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, dropAmt),
				true)

			response.Handled = true
			return response, nil
		}
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindInBackpack(rest)

	if found {

		// Swap the item location
		room.AddItem(matchItem, false)
		mob.Character.RemoveItem(matchItem)

		response.SendRoomMessage(mob.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> drops their <ansi fg="item">%s</ansi>...`, mob.Character.Name, matchItem.DisplayName()),
			true)
	}

	response.Handled = true
	return response, nil
}
