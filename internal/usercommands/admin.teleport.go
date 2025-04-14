package usercommands

import (
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/scripting"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
* Role Permissions:
* teleport 				(All)
* teleport.direction	(Teleport through walls in a direction)
* teleport.playername	(Teleport to a player name)
* teleport.roomid		(Teleport to a roomId)
 */
func Teleport(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if len(rest) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.teleport", nil, user.UserId)
		user.SendText(infoOutput)

		return true, nil
	}

	gotoRoomId, numError := strconv.Atoi(rest)
	// If not a number, check if it's a direction
	if numError != nil {

		if mapper.IsCompassDirection(rest) {

			if !user.HasRolePermission(`teleport.direction`) {
				user.SendText(`you do not have <ansi fg="command">teleport.direction</ansi> permission`)
				return true, nil
			}

			zMapper := mapper.GetZoneMapper(room.Zone)
			if zMapper == nil {
				err := fmt.Errorf("Could not find mapper for zone: %s", room.Zone)
				mudlog.Error("Map", "error", err)
				user.SendText(`No map found (or an error occured)"`)
				return true, err
			}

			gotoRoomId, _ = zMapper.FindAdjacentRoom(user.Character.RoomId, rest)

		} else {

			// Finally, try a player name
			if locateUser := users.GetByCharacterName(rest); locateUser != nil {

				if !user.HasRolePermission(`teleport.playername`) {
					user.SendText(`you do not have <ansi fg="command">teleport.direction</ansi> permission`)
					return true, nil
				}

				gotoRoomId = locateUser.Character.RoomId
			}

		}

	} else {
		if !user.HasRolePermission(`teleport.roomid`) {
			user.SendText(`you do not have <ansi fg="command">teleport.direction</ansi> permission`)
			return true, nil
		}
	}

	if gotoRoomId != 0 || rest == `0` {

		previousRoomId := user.Character.RoomId

		if err := rooms.MoveToRoom(user.UserId, gotoRoomId); err != nil {
			user.SendText(err.Error())

		} else {

			scripting.TryRoomScriptEvent(`onExit`, user.UserId, previousRoomId)

			user.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))

			gotoRoom := rooms.LoadRoom(gotoRoomId)
			gotoRoom.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, user.Character.Name),
				user.UserId,
			)

			if party := parties.Get(user.UserId); party != nil {

				// Party leaders can move the whole party.
				if party.LeaderUserId == user.UserId {

					newRoom := rooms.LoadRoom(gotoRoomId)
					for _, uid := range room.GetPlayers() {
						if party.IsMember(uid) {

							partyUser := users.GetByUserId(uid)
							if partyUser == nil {
								continue
							}

							if partyUser.Character.RoomId != room.RoomId {
								continue
							}

							rooms.MoveToRoom(partyUser.UserId, gotoRoomId)
							partyUser.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))
							room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, partyUser.Character.Name), partyUser.UserId)

							Look(``, partyUser, gotoRoom, flags)

							for _, mInstanceId := range room.GetMobs(rooms.FindCharmed) {
								if mob := mobs.GetInstance(mInstanceId); mob != nil {
									if mob.Character.IsCharmed(partyUser.UserId) {
										room.RemoveMob(mob.InstanceId)
										newRoom.AddMob(mob.InstanceId)
									}
								}
							}
						}
					}

				}

			}

			Look(``, user, gotoRoom, flags)

			scripting.TryRoomScriptEvent(`onEnter`, user.UserId, gotoRoomId)

		}
	} else {
		user.SendText(fmt.Sprintf(`Invalid teleport command: <ansi fg="command">%s</ansi> (No RoomId, direction, or character name match)`, rest))
	}

	return true, nil
}
