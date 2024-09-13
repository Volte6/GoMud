package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Look(rest string, mobId int) (util.MessageQueue, error) {
	response := NewMobCommandResponse(mobId)

	secretLook := false
	if strings.HasPrefix(rest, "secretly") {
		secretLook = true
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "secretly"))
	}

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

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	args := util.SplitButRespectQuotes(rest)

	// Looking AT something?
	if len(args) > 0 {
		lookAt := args[0]

		//
		// Check room exits
		//
		exitName, lookRoomId := room.FindExitByName(lookAt)
		if lookRoomId > 0 {

			exitInfo := room.Exits[exitName]
			if exitInfo.Lock.IsLocked() {
				response.Handled = true
				return response, nil
			}

			if !isSneaking {
				response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> peers toward the %s.`, mob.Character.Name, exitName), true)
			}

			if lookRoomId > 0 {

				lookRoom(mobId, lookRoomId, &response, secretLook || isSneaking)

				response.Handled = true
				return response, nil
			}
		}

		//
		// Check for anything in their backpack they might want to look at
		//
		if lookItem, found := mob.Character.FindInBackpack(rest); found {

			if !isSneaking {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is admiring their <ansi fg="item">%s</ansi>.`, mob.Character.Name, lookItem.DisplayName()),
					true)
			}

			response.Handled = true
			return response, nil
		}

		//
		// look for any mobs, players, npcs
		//

		playerId, mobId := room.FindByName(lookAt)

		if playerId > 0 || mobId > 0 {

			if playerId > 0 {

				u := *users.GetByUserId(playerId)

				if !isSneaking {
					response.SendUserMessage(u.UserId,
						fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is looking at you.`, mob.Character.Name),
						true)

					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is looking at <ansi fg="username">%s</ansi>.`, mob.Character.Name, u.Character.Name),
						true,
						u.UserId)
				}

			} else if mobId > 0 {

				m := mobs.GetInstance(mobId)

				if !isSneaking {
					targetName := m.Character.GetMobName(0).String()
					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is looking at %s.`, mob.Character.Name, targetName),
						true)
				}

			}

			response.Handled = true
			return response, nil

		}

		//
		// Check for any equipment they are wearing they might want to look at
		//
		if lookItem, found := mob.Character.FindOnBody(rest); found {

			if !isSneaking {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is admiring their <ansi fg="item">%s</ansi>.`, mob.Character.Name, lookItem.DisplayName()),
					true)
			}

			response.Handled = true
			return response, nil
		}

		response.Handled = true
		return response, nil

	} else {

		if !secretLook && !isSneaking {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is looking around.`, mob.Character.Name),
				true)

			// Make it a "secret looks" now because we don't want another look message sent out by the lookRoom() func
			secretLook = true
		}
		lookRoom(mobId, room.RoomId, &response, secretLook || isSneaking)
	}

	response.Handled = true
	return response, nil
}

func lookRoom(mobId int, roomId int, response *util.MessageQueue, secretLook bool) {

	mob := mobs.GetInstance(mobId)
	room := rooms.LoadRoom(roomId)

	if mob == nil || room == nil {
		return
	}

	// Make sure to prepare the room before anyone looks in if this is the first time someone has dealt with it in a while.
	if room.PlayerCt() < 1 {
		room.Prepare(true)
	}

	if !secretLook {
		// Find the exit back
		lookFromName := room.FindExitTo(mob.Character.RoomId)
		if lookFromName == "" {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking into the room from somewhere...`, mob.Character.Name),
				true)
		} else {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking into the room from the <ansi fg="exit">%s</ansi> exit`, mob.Character.Name, lookFromName),
				true)
		}
	}

}
