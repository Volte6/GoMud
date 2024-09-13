package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Room(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	handled := true
	nextCommand := ``

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.room", nil)
		user.SendText(infoOutput)

		return handled, nextCommand, nil
	}

	var roomId int = 0
	roomCmd := strings.ToLower(args[0])

	if roomCmd == "copy" && len(args) >= 3 {

		property := args[1]

		if property == "spawninfo" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.SpawnInfo = sourceRoom.SpawnInfo
				rooms.SaveRoom(*room)

				user.SendText("Spawn info copied/overwritten.")
			}
		}

		if property == "idlemessages" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.IdleMessages = append(room.IdleMessages, sourceRoom.IdleMessages...)
				rooms.SaveRoom(*room)

				user.SendText("IdleMessages copied/overwritten.")
			}
		}

	} else if roomCmd == "info" {
		if len(args) == 1 {
			roomId = room.RoomId
		} else {
			roomId, _ = strconv.Atoi(args[1])
		}

		targetRoom := rooms.LoadRoom(roomId)
		if targetRoom == nil {
			user.SendText(fmt.Sprintf("Room %d not found.", roomId))
			return false, ``, fmt.Errorf("room %d not found", roomId)
		}

		infoOutput, _ := templates.Process("admincommands/ingame/roominfo", targetRoom)
		user.SendText(infoOutput)

	} else if len(args) >= 2 && roomCmd == "exit" {

		direction := strings.ToLower(args[1])
		roomId = 0

		if len(args) > 2 {
			roomId, _ = strconv.Atoi(args[2])
		}

		// Will be erasing it.
		if roomId == 0 {
			if _, ok := room.Exits[direction]; !ok {
				user.SendText(fmt.Sprintf("Exit %s does not exist.", direction))
				return handled, nextCommand, nil
			}
			delete(room.Exits, direction)
			return handled, nextCommand, nil
		}

		if _, ok := room.Exits[direction]; ok {
			user.SendText(fmt.Sprintf("Exit %s already exists (overwriting).", direction))
		}

		targetRoom := rooms.LoadRoom(roomId)
		if targetRoom == nil {
			err := fmt.Errorf(`room %d not found`, roomId)
			user.SendText(err.Error())
			return handled, nextCommand, nil
		}

		rooms.ConnectRoom(room.RoomId, targetRoom.RoomId, direction)
		user.SendText(fmt.Sprintf("Exit %s added.", direction))

	} else if len(args) >= 2 && roomCmd == "secretexit" {

		direction := args[1]
		if exit, ok := room.Exits[direction]; ok {
			if exit.Secret {
				exit.Secret = false
				room.Exits[direction] = exit
				rooms.SaveRoom(*room)
				user.SendText(fmt.Sprintf("Exit %s secrecy REMOVED.", direction))
			} else {
				exit.Secret = true
				room.Exits[direction] = exit
				rooms.SaveRoom(*room)
				user.SendText(fmt.Sprintf("Exit %s secrecy ADDED.", direction))
			}
		} else {
			user.SendText(fmt.Sprintf("Exit %s not found.", direction))
		}

	} else if len(args) >= 2 && roomCmd == "set" {

		propertyName := args[1]
		propertyValue := ``
		if len(args) > 2 {
			propertyValue = strings.Join(args[2:], ` `)
		}

		propertyValue = strings.Trim(propertyValue, `"`)

		if propertyName == "spawninfo" {
			if propertyValue == `clear` {
				room.SpawnInfo = room.SpawnInfo[:0]
				rooms.SaveRoom(*room)
			}

		} else if propertyName == "title" {
			if propertyValue == `` {
				propertyValue = `[no title]`
			}
			room.Title = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "description" {
			if propertyValue == `` {
				propertyValue = `[no description]`
			}
			room.Description = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "idlemessages" {
			room.IdleMessages = []string{}
			for _, idleMsg := range strings.Split(propertyValue, ";") {
				idleMsg = strings.TrimSpace(idleMsg)
				if len(idleMsg) < 1 {
					continue
				}
				room.IdleMessages = append(room.IdleMessages, idleMsg)
			}
			rooms.SaveRoom(*room)
		} else if propertyName == "symbol" || propertyName == "mapsymbol" {
			room.MapSymbol = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "legend" || propertyName == "maplegend" {
			room.MapLegend = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "zone" {
			// Try moving it to the new zone.
			if err := rooms.MoveToZone(room.RoomId, propertyValue); err != nil {
				user.SendText(err.Error())
				return handled, nextCommand, nil
			}

		} else if propertyName == "biome" {
			room.Biome = strings.ToLower(propertyValue)
		} else {
			user.SendText(
				`Invalid property provided to <ansi fg="command">room set</ansi>.`,
			)
			return false, ``, fmt.Errorf("room %d not found", roomId)
		}

	} else {

		var gotoRoomId int = 0

		if deltaD, ok := rooms.DirectionDeltas[roomCmd]; ok {

			rGraph := rooms.NewRoomGraph(100, 100, 0, rooms.MapModeAll)
			err := rGraph.Build(user.Character.RoomId, nil)
			if err != nil {
				user.SendText(err.Error())
				return true, nextCommand, nil
			}

			map2D, cX, cY := rGraph.Generate2DMap(61, 61, user.Character.RoomId)
			if len(map2D) < 1 {
				user.SendText("Error generating a 2d map")
				return true, ``, nil
			}

			for i := 1; i <= 30; i++ {
				dy := deltaD.Dy * i
				dx := deltaD.Dx * i
				if cY+dy < len(map2D) && cX+dx < len(map2D[0]) {
					if cY+dy >= 0 && cX+dx >= 0 {
						if map2D[cY+dy][cX+dx] != nil {
							gotoRoomId = map2D[cY+dy][cX+dx].RoomId
							break
						}
					}
				}
			}

			//dirDelta.Dx
			//dirDelta.Dy
		} else {
			// move to a new room
			gotoRoomId, _ = strconv.Atoi(args[0])
		}

		if gotoRoomId != 0 {
			if err := rooms.MoveToRoom(user.UserId, gotoRoomId); err != nil {
				user.SendText(err.Error())

			} else {
				user.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))

				gotoRoom := rooms.LoadRoom(gotoRoomId)
				gotoRoom.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, user.Character.Name),
					user.UserId,
				)

				if party := parties.Get(user.UserId); party != nil {

					newRoom := rooms.LoadRoom(gotoRoomId)
					for _, uid := range room.GetPlayers() {
						if party.IsMember(uid) {

							partyUser := users.GetByUserId(uid)
							if partyUser == nil {
								continue
							}

							rooms.MoveToRoom(partyUser.UserId, gotoRoomId)
							user.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))
							room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, partyUser.Character.Name), partyUser.UserId)

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

				nextCommand = "look" // Force them to look at the new room they are in.
			}
		} else {
			user.SendText(fmt.Sprintf("Invalid room comand: %s", args[0]))
		}
	}

	return handled, nextCommand, nil
}
