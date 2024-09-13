package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Get(rest string, mobId int) (bool, string, error) {

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

	if len(args) == 0 {
		return true, ``, nil
	}

	if args[0] == "all" {
		if room.Gold > 0 {
			Get("gold", mobId)
		}

		if len(room.Items) > 0 {
			iCopies := []items.Item{}
			for _, item := range room.Items {
				iCopies = append(iCopies, item)
			}

			for _, item := range iCopies {
				Get(item.Name(), mobId)
			}
		}

		return true, ``, nil
	}

	if args[0] == "gold" {

		if room.Gold > 0 {

			mob.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			goldAmt := room.Gold
			mob.Character.Gold += goldAmt
			room.Gold -= goldAmt

			room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> picks up <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, goldAmt))
		}

		return true, ``, nil
	}

	getFromStash := false

	if len(args) >= 2 {
		// Detect "stash" or "from stash" at end and remove it
		if args[len(args)-1] == "stash" {
			getFromStash = true
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		if args[len(args)-1] == "ground" {
			getFromStash = false
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := room.FindOnFloor(rest, getFromStash)

	if found {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

		// Swap the item location
		room.RemoveItem(matchItem, getFromStash)
		mob.Character.StoreItem(matchItem)

		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, mob.Character.Name, matchItem.DisplayName()))
	}

	return true, ``, nil
}
