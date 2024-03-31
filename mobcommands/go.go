package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Go(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// If has a buff that prevents combat, skip the player
	if mob.Character.HasBuffFlag(buffs.NoMovement) {
		response.Handled = true
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	exitName := ``
	goRoomId := 0

	if rest == `home` {

		if mob.Character.RoomId == mob.HomeRoomId {

			if len(mob.RoomStack) > 0 {
				mob.RoomStack = make([]int, 0)
			}
			mob.GoingHome = false

			response.Handled = true
			return response, nil

		} else {

			mob.GoingHome = true

			if len(mob.RoomStack) == 0 {

				if util.Rand(50) == 0 {
					goRoomId = mob.HomeRoomId
					exitName = `mysterious`
				} else {
					cmdQueue.QueueCommand(0, mobId, `say I'm lost`)
					response.Handled = true
					return response, nil
				}
			} else {

				targetRoomId := mob.RoomStack[len(mob.RoomStack)-1]
				mob.RoomStack = mob.RoomStack[:len(mob.RoomStack)-1]
				exitName = room.FindExitTo(targetRoomId)
				goRoomId = targetRoomId
				if len(exitName) < 1 {
					exitName = fmt.Sprintf(`%d room`, goRoomId)
				}

			}

		}

	} else {
		exitName, goRoomId = room.FindExitByName(rest)

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {

			cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`emote tries to go the <ansi fg="exit">%s</ansi> exit, but it's locked.`, exitName))

			response.Handled = true
			return response, nil
		}

	}

	if goRoomId > 0 {

		// It does so we won't need to continue down the logic after this chunk
		response.Handled = true

		// Load current room details
		destRoom := rooms.LoadRoom(goRoomId)
		if destRoom == nil {
			return response, fmt.Errorf(`room %d not found`, goRoomId)
		}

		// Grab the exit in the target room that leads to this room (if any)
		enterFromExit := destRoom.FindExitTo(room.RoomId)

		if len(enterFromExit) < 1 {
			enterFromExit = "somewhere"
		} else {

			// Entering through the other side unlocks this side
			exitInfo := destRoom.Exits[enterFromExit]

			if exitInfo.Lock.IsLocked() {

				// For now, mobs won't go through doors if it unlocks them.
				response.Handled = true
				return response, nil
				//exitInfo.Unlock()
				//destRoom.Exits[enterFromExit] = exitInfo
			}

			enterFromExit = fmt.Sprintf(`the <ansi fg="exit">%s</ansi>`, enterFromExit)
		}

		if rest != `home` {
			// track the room we are leaving
			repeatRoom := false
			stackSize := len(mob.RoomStack)
			for i := 0; i < stackSize; i++ {
				if mob.RoomStack[i] == room.RoomId {
					mob.RoomStack = mob.RoomStack[:i]
					repeatRoom = true
					break
				}
			}
			if !repeatRoom && mob.MaxWander > -1 { // If they can wander forever, don't track it
				mob.RoomStack = append(mob.RoomStack, room.RoomId)
			}
		}

		room.RemoveMob(mobId)
		destRoom.AddMob(mobId)

		c := configs.GetConfig()

		// Tell the old room they are leaving
		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(string(c.ExitRoomMessageWrapper),
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> leaves towards the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, exitName),
			), true)
		// Tell the new room they have arrived
		response.SendRoomMessage(destRoom.RoomId,
			fmt.Sprintf(string(c.EnterRoomMessageWrapper),
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> enters from %s.`, mob.Character.Name, enterFromExit),
			), true)

		response.Handled = true
	}

	return response, nil
}
