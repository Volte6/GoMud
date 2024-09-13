package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Throw(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. mob not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		return response, nil
	}

	throwWhat := args[0]
	args = args[1:]

	throwWhere := strings.Join(args, ` `)

	itemMatch, ok := mob.Character.FindInBackpack(throwWhat)
	if !ok {
		return response, nil
	}

	// check Exits and SecretExits for a string match. If found, move the player to that room.
	exitName, throwRoomId := room.FindExitByName(throwWhere)

	// If nothing found, consider directional aliases
	if throwRoomId == 0 {

		if alias := keywords.TryDirectionAlias(throwWhere); alias != throwWhere {
			exitName, throwRoomId = room.FindExitByName(alias)
			if throwRoomId != 0 {
				throwWhere = alias
			}
		}
	}

	if throwRoomId > 0 {

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {
			response.Handled = true
			return response, nil
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
		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, mob.Character.Name, itemMatch.DisplayName(), exitName),
			true)

		// Tell the new room the item arrived
		response.SendRoomMessage(throwToRoom.RoomId,
			fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
			true)

		response.Handled = true
	}

	// Still looking for an exit... try the temp ones
	if !response.Handled {
		if len(room.ExitsTemp) > 0 {
			// See if there's a close match
			exitNames := make([]string, 0, len(room.ExitsTemp))
			for exitName := range room.ExitsTemp {
				exitNames = append(exitNames, exitName)
			}

			exactMatch, closeMatch := util.FindMatchIn(throwWhere, exitNames...)

			var tempExit rooms.TemporaryRoomExit
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

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, mob.Character.Name, itemMatch.DisplayName(), tempExit.Title),
					true)

				// Tell the new room the item arrived
				response.SendRoomMessage(tempExit.RoomId,
					fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
					true)

				response.Handled = true

			}
		}
	}

	response.Handled = true
	return response, nil
}
