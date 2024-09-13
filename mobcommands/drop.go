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

func Drop(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if args[0] == "all" {

		iCopies := []items.Item{}

		if mob.Character.Gold > 0 {
			Drop(fmt.Sprintf("%d gold", mob.Character.Gold), mobId)
		}

		for _, item := range mob.Character.Items {
			iCopies = append(iCopies, item)
		}

		for _, item := range iCopies {
			Drop(item.Name(), mobId)
		}

		return true, ``, nil
	}

	// Drop 10 gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt < 1 {
			return true, ``, nil
		}

		if dropAmt <= mob.Character.Gold {

			room.Gold += dropAmt
			mob.Character.Gold -= dropAmt

			room.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> drops <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, dropAmt))

			return true, ``, nil
		}
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindInBackpack(rest)

	if found {

		// Swap the item location
		room.AddItem(matchItem, false)
		mob.Character.RemoveItem(matchItem)

		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> drops their <ansi fg="item">%s</ansi>...`, mob.Character.Name, matchItem.DisplayName()))
	}

	return true, ``, nil
}
