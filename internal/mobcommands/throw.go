package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/exit"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/util"
)

func Throw(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		return false, nil
	}

	throwWhat := args[0]
	args = args[1:]

	throwWhere := strings.Join(args, ` `)

	itemMatch, ok := mob.Character.FindInBackpack(throwWhat)
	if !ok {
		return false, nil
	}

	// check Exits and SecretExits for a string match. If found, move the player to that room.
	exitName, throwRoomId := room.FindExitByName(throwWhere)

	// If nothing found, consider directional aliases
	if exitName == `` {

		if alias := keywords.TryDirectionAlias(throwWhere); alias != throwWhere {
			exitName, throwRoomId = room.FindExitByName(alias)
			if exitName != `` {
				throwWhere = alias
			}
		}
	}

	if exitName != `` {

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {
			return true, nil
		}

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		throwToRoom := rooms.LoadRoom(throwRoomId)
		returnExitName := throwToRoom.FindExitTo(mob.Character.RoomId)

		if len(returnExitName) < 1 {
			returnExitName = "somewhere"
		} else {
			returnExitName = fmt.Sprintf("the %s exit", returnExitName)
		}

		mob.Character.RemoveItem(itemMatch)
		throwToRoom.AddItem(itemMatch, false)

		// Tell the old room they are leaving
		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, mob.Character.Name, itemMatch.DisplayName(), exitName),
		)

		// Tell the new room the item arrived
		throwToRoom.SendText(
			fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
		)

		return true, nil
	}

	// Still looking for an exit... try the temp ones

	if len(room.ExitsTemp) > 0 {
		// See if there's a close match
		exitNames := make([]string, 0, len(room.ExitsTemp))
		for exitName := range room.ExitsTemp {
			exitNames = append(exitNames, exitName)
		}

		exactMatch, closeMatch := util.FindMatchIn(throwWhere, exitNames...)

		var tempExit exit.TemporaryRoomExit
		var tempExitFound bool = false
		if len(exactMatch) > 0 {
			tempExit = room.ExitsTemp[exactMatch]
			tempExitFound = true
		} else if len(closeMatch) > 0 && len(rest) >= 3 {
			tempExit = room.ExitsTemp[closeMatch]
			tempExitFound = true
		}

		if tempExitFound {

			mob.Character.CancelBuffsWithFlag(buffs.Hidden)

			// do something with tempExit
			throwToRoom := rooms.LoadRoom(tempExit.RoomId)
			returnExitName := throwToRoom.FindExitTo(mob.Character.RoomId)

			if len(returnExitName) < 1 {
				returnExitName = "somewhere"
			} else {
				returnExitName = fmt.Sprintf("the %s exit", returnExitName)
			}

			mob.Character.RemoveItem(itemMatch)
			throwToRoom.AddItem(itemMatch, false)

			room.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, mob.Character.Name, itemMatch.DisplayName(), tempExit.Title),
			)

			// Tell the new room the item arrived
			exitRoom := rooms.LoadRoom(tempExit.RoomId)
			exitRoom.SendText(
				fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
			)

			return true, nil

		}
	}

	return false, nil
}
