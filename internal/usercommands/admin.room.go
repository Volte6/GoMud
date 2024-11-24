package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Room(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	handled := true

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.room", nil)
		user.SendText(infoOutput)

		return handled, nil
	}

	var roomId int = 0
	roomCmd := strings.ToLower(args[0])

	if roomCmd == `noun` || roomCmd == `nouns` {

		// room noun chair "a chair for sitting"
		if len(args) > 2 {
			noun := args[1]
			description := strings.Join(args[2:], ` `)

			if room.Nouns == nil {
				room.Nouns = map[string]string{}
			}
			room.Nouns[noun] = description

			user.SendText(`Noun Added:`)
			user.SendText(fmt.Sprintf(`<ansi fg="noun">%s</ansi> - %s`, strings.Repeat(` `, 20-len(noun))+noun, description))

			return true, nil
		}

		// room noun chair
		if len(args) == 2 || (len(args) == 3 && len(args[2]) == 0) {

			if _, ok := room.Nouns[args[1]]; ok {
				delete(room.Nouns, args[1])
				user.SendText(`Noun deleted.`)
			} else {
				user.SendText(`Noun not found.`)
			}

			return true, nil
		}

		// room noun
		// room nouns
		user.SendText(`Room Nouns:`)
		for noun, description := range room.Nouns {
			user.SendText(fmt.Sprintf(`<ansi fg="noun">%s</ansi> - %s`, strings.Repeat(` `, 20-len(noun))+noun, description))
		}
		return true, nil
	}

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

		if property == "mutator" || property == "mutators" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.Mutators = append(room.Mutators, sourceRoom.Mutators...)
				rooms.SaveRoom(*room)

				user.SendText("Mutators copied/overwritten.")
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
			return false, fmt.Errorf("room %d not found", roomId)
		}

		roomInfo := map[string]any{
			`room`: targetRoom,
			`zone`: rooms.GetZoneConfig(targetRoom.Zone),
		}

		infoOutput, _ := templates.Process("admincommands/ingame/roominfo", roomInfo)
		user.SendText(infoOutput)

	} else if len(args) >= 2 && roomCmd == "exit" {

		direction := strings.ToLower(args[1])
		roomId = 0
		var numError error = nil

		if len(args) > 2 {
			roomId, numError = strconv.Atoi(args[2])
		}

		// Will be erasing it.
		if numError != nil || len(args) < 3 { // If a bad room number or NO room number supplied, delete
			if _, ok := room.Exits[direction]; !ok {
				user.SendText(fmt.Sprintf("Exit %s does not exist.", direction))
				return handled, nil
			}
			delete(room.Exits, direction)
			return handled, nil
		}

		if _, ok := room.Exits[direction]; ok {
			user.SendText(fmt.Sprintf("Exit %s already exists (overwriting).", direction))
		}

		targetRoom := rooms.LoadRoom(roomId)
		if targetRoom == nil {
			err := fmt.Errorf(`room %d not found`, roomId)
			user.SendText(err.Error())
			return handled, nil
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

		if propertyName == "mutator" || propertyName == "mutators" {

			if propertyValue == `` { // If none specified, list all mutators

				user.SendText(`<ansi fg="table-title">Mutators:</ansi>`)
				if len(room.Mutators) == 0 {
					user.SendText(`  None.`)
				}
				for _, mut := range room.Mutators {
					user.SendText(`  <ansi fg="mutator">` + mut.MutatorId + `</ansi>`)
				}
				user.SendText(``)

			} else { // Otherwise, toggle the mentioned mutator on/off

				user.SendText(``)

				if !mutators.IsMutator(propertyValue) {
					user.SendText(`<ansi fg="table-title"><ansi fg="mutator">` + propertyValue + `</ansi> is an invalid mutator id.</ansi>`)
					user.SendText(`<ansi fg="table-title">  Here is a list of valid mutator id's:</ansi>`)
					for _, name := range mutators.GetAllMutatorIds() {
						user.SendText(`    <ansi fg="mutator">` + name + `</ansi>`)
					}
				} else if room.Mutators.Remove(propertyValue) {
					user.SendText(`<ansi fg="table-title">Mutator <ansi fg="mutator">` + propertyValue + `</ansi> Removed.</ansi>`)
				} else if room.Mutators.Add(propertyValue) {
					user.SendText(`<ansi fg="table-title">Mutator <ansi fg="mutator">` + propertyValue + `</ansi> Added.</ansi>`)
				}

				user.SendText(``)
			}

			return true, nil
		}

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
				return handled, nil
			}

		} else if propertyName == "biome" {
			room.Biome = strings.ToLower(propertyValue)
		} else {
			user.SendText(
				`Invalid property provided to <ansi fg="command">room set</ansi>.`,
			)
			return false, fmt.Errorf("room %d not found", roomId)
		}

	} else {

		var gotoRoomId int = 0
		var numError error = nil

		if deltaD, ok := rooms.DirectionDeltas[roomCmd]; ok {

			rGraph := rooms.NewRoomGraph(100, 100, 0, rooms.MapModeAll)
			err := rGraph.Build(user.Character.RoomId, nil)
			if err != nil {
				user.SendText(err.Error())
				return true, nil
			}

			map2D, cX, cY := rGraph.Generate2DMap(61, 61, user.Character.RoomId)
			if len(map2D) < 1 {
				user.SendText("Error generating a 2d map")
				return true, nil
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
			gotoRoomId, numError = strconv.Atoi(args[0])
		}

		if numError == nil {

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

				Look(`secretly`, user, gotoRoom)

				scripting.TryRoomScriptEvent(`onEnter`, user.UserId, gotoRoomId)

			}
		} else {
			user.SendText(fmt.Sprintf("Invalid room comand: %s", args[0]))
		}
	}

	return handled, nil
}
