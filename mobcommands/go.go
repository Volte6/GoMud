package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Go(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// If has a buff that prevents combat, skip the player
	if mob.Character.HasBuffFlag(buffs.NoMovement) {
		return true, nil
	}

	exitName := ``
	goRoomId := 0

	if rest == `home` {

		if mob.Character.RoomId == mob.HomeRoomId {

			if len(mob.RoomStack) > 0 {
				mob.RoomStack = make([]int, 0)
			}
			mob.GoingHome = false

			return true, nil

		} else {

			mob.GoingHome = true

			if len(mob.RoomStack) == 0 {

				if util.Rand(50) == 0 {
					goRoomId = mob.HomeRoomId
					exitName = `mysterious`
				} else {

					mob.Command(`say I'm lost.`)

					return true, nil
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

			mob.Command(fmt.Sprintf(`emote tries to go the <ansi fg="exit">%s</ansi> exit, but it's locked.`, exitName))

			return true, nil
		}

	}

	if goRoomId > 0 {

		// Load current room details
		destRoom := rooms.LoadRoom(goRoomId)
		if destRoom == nil {
			return false, fmt.Errorf(`room %d not found`, goRoomId)
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
				return true, nil
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

		room.RemoveMob(mob.InstanceId)
		destRoom.AddMob(mob.InstanceId)

		c := configs.GetConfig()

		// Tell the old room they are leaving
		room.SendText(
			fmt.Sprintf(string(c.ExitRoomMessageWrapper),
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> leaves towards the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, exitName),
			))

		// Tell the new room they have arrived
		destRoom.SendText(
			fmt.Sprintf(string(c.EnterRoomMessageWrapper),
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> enters from %s.`, mob.Character.Name, enterFromExit),
			))

		destRoom.SendTextToExits(`You hear someone moving around.`, true, room.GetPlayers(rooms.FindAll)...)

		return true, nil
	}

	return false, nil
}
