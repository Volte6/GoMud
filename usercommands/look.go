package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Look(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	secretLook := false
	if strings.HasPrefix(rest, "secretly") {
		secretLook = true
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "secretly"))
	}

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// 0 = none. 1 = can see this room. 2 = can see this room and all exits
	visibility := 2
	if gametime.IsNight() {
		visibility -= 1
	}

	biome := room.GetBiome()
	if biome.IsDark() {
		visibility -= 2
	}
	if biome.IsLit() {
		visibility += 1
	}

	if visibility < 0 {
		visibility = 0
	} else if visibility > 2 {
		visibility = 2
	}

	// If someone has light, cancel the darkness
	if visibility < 2 {
		if mobInstanceIds := room.GetMobs(rooms.FindHasLight); len(mobInstanceIds) > 0 {
			visibility += 1
		} else if userIds := room.GetPlayers(rooms.FindHasLight); len(userIds) > 0 {
			visibility += 1
		}
	}

	if visibility < 1 {
		raceInfo := races.GetRace(user.Character.RaceId)
		if !raceInfo.NightVision {
			response.SendUserMessage(userId, `You can't see anything!`, true)
			response.Handled = true
			return response, nil
		}
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	// Looking AT something?
	if len(rest) > 0 {
		lookAt := rest

		containerName := room.FindContainerByName(lookAt)
		if containerName != `` {

			container := room.Containers[containerName]

			if container.Lock.IsLocked() {
				response.SendUserMessage(userId, ``, true)
				response.SendUserMessage(userId, `The chest is locked.`, true)
				response.SendUserMessage(userId, ``, true)
				response.Handled = true
				return response, nil
			}

			chestStuff := []string{}

			if container.Gold > 0 {
				chestStuff = append(chestStuff, fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, container.Gold))
			}

			for _, item := range container.Items {
				if !item.IsValid() {
					room.RemoveItem(item, false)
					continue
				}
				chestStuff = append(chestStuff, item.DisplayName())
			}

			textOut, _ := templates.Process("descriptions/insidecontainer", chestStuff)

			response.SendUserMessage(userId, ``, true)
			response.SendUserMessage(userId, textOut, false)

			response.Handled = true
			return response, nil
		}

		//
		// Check room exits
		//
		exitName, lookRoomId := room.FindExitByName(lookAt)

		// If nothing found, consider directional aliases
		if lookRoomId == 0 {

			if alias := keywords.TryDirectionAlias(lookAt); alias != lookAt {
				exitName, lookRoomId = room.FindExitByName(alias)
				if lookRoomId != 0 {
					lookAt = alias
				}
			}
		}

		if lookRoomId > 0 {

			if visibility < 2 {

				raceInfo := races.GetRace(user.Character.RaceId)
				if !raceInfo.NightVision {

					biome := room.GetBiome()
					if !biome.IsLit() {
						response.SendUserMessage(userId, `It's too dark to see anything in that direction.`, true)
						response.Handled = true
						return response, nil
					}

				}

			}

			exitInfo := room.Exits[exitName]
			if exitInfo.Lock.IsLocked() {
				response.SendUserMessage(userId, fmt.Sprintf("The %s exit is locked.", exitName), true)
				response.Handled = true
				return response, nil
			}

			response.SendUserMessage(userId, fmt.Sprintf("You peer toward the %s.", exitName), true)
			if !isSneaking {
				response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> peers toward the %s.`, user.Character.Name, exitName), true)
			}

			if lookRoomId > 0 {

				lookRoom(user.UserId, lookRoomId, &response, secretLook || isSneaking)

				response.Handled = true
				return response, nil
			}
		}

		//
		// Check for anything in their backpack they might want to look at
		//
		lookItem, found := user.Character.FindInBackpack(rest)
		lookDestination := `in your backpack`
		if !found {
			// Check for any equipment they are wearing they might want to look at
			lookItem, found = user.Character.FindOnBody(rest)
			lookDestination = `you are wearing`
		}

		if found {

			response.SendUserMessage(userId, ``, true)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You look at the <ansi fg="item">%s</ansi> %s:`, lookItem.DisplayName(), lookDestination),
				true)

			response.SendUserMessage(userId, ``, true)

			if !isSneaking {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> is admiring their <ansi fg="item">%s</ansi>.`, user.Character.Name, lookItem.DisplayName()),
					true)
			}

			response.SendUserMessage(userId,
				lookItem.GetLongDescription(),
				true)

			response.SendUserMessage(userId, ``, true)

			response.Handled = true
			return response, nil
		}

		//
		// look for any mobs, players, npcs
		//

		playerId, mobId := room.FindByName(lookAt)

		if playerId > 0 || mobId > 0 {

			statusTxt := ""
			invTxt := ""

			if playerId > 0 {

				u := *users.GetByUserId(playerId)

				if !isSneaking {
					response.SendUserMessage(u.UserId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking at you.`, user.Character.Name),
						true)

					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking at <ansi fg="username">%s</ansi>.`, user.Character.Name, u.Character.Name),
						true,
						u.UserId)
				}

				descTxt, _ := templates.Process("character/description", u)
				response.SendUserMessage(user.UserId, descTxt, false)

				itemNames := []string{}
				for _, item := range u.Character.Items {
					itemNames = append(itemNames, item.DisplayName())
				}

				invData := map[string]any{
					`Equipment`: &u.Character.Equipment,
					`ItemNames`: itemNames,
				}

				inventoryTxt, _ := templates.Process("character/inventory-look", invData)
				response.SendUserMessage(user.UserId, inventoryTxt, false)

			} else if mobId > 0 {

				m := mobs.GetInstance(mobId)

				if !isSneaking {
					targetName := m.Character.GetMobName(0).String()
					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking at %s.`, user.Character.Name, targetName),
						true)
				}

				descTxt, _ := templates.Process("character/description", m)
				response.SendUserMessage(user.UserId, descTxt, false)

				itemNames := []string{}
				for _, item := range m.Character.Items {
					itemNames = append(itemNames, item.DisplayName())
				}

				invData := map[string]any{
					`Equipment`: &m.Character.Equipment,
					`ItemNames`: itemNames,
				}

				inventoryTxt, _ := templates.Process("character/inventory-look", invData)
				response.SendUserMessage(user.UserId, inventoryTxt, false)
			}

			response.SendUserMessage(userId, statusTxt, false)
			response.SendUserMessage(userId, invTxt, false)

			response.Handled = true
			return response, nil

		}

		response.SendUserMessage(userId, "Look at what???", true)

		response.Handled = true
		return response, nil

	} else {

		if !secretLook && !isSneaking {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking around.`, user.Character.Name),
				true)

			// Make it a "secret looks" now because we don't want another look message sent out by the lookRoom() func
			secretLook = true
		}
		lookRoom(user.UserId, room.RoomId, &response, secretLook || isSneaking)
	}

	response.Handled = true
	return response, nil
}

func lookRoom(userId int, roomId int, response *util.MessageQueue, secretLook bool) {

	user := users.GetByUserId(userId)
	room := rooms.LoadRoom(roomId)

	if user == nil || room == nil {
		return
	}

	// Make sure to prepare the room before anyone looks in if this is the first time someone has dealt with it in a while.
	if room.PlayerCt() < 1 {
		room.Prepare(true)
	}

	if !secretLook {
		// Find the exit back
		lookFromName := room.FindExitTo(user.Character.RoomId)
		if lookFromName == "" {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking into the room from somewhere...`, user.Character.Name),
				true)
		} else {
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is looking into the room from the <ansi fg="exit">%s</ansi> exit`, user.Character.Name, lookFromName),
				true)
		}
	}

	details := room.GetRoomDetails(user)

	textOut, _ := templates.Process("descriptions/room-title", details)
	response.SendUserMessage(userId, textOut, false)

	textOut, _ = templates.Process("descriptions/room", details)
	response.SendUserMessage(userId, textOut, false)

	signCt := 0
	privateSigns := room.GetPrivateSigns()
	for _, sign := range privateSigns {
		if sign.VisibleUserId == userId {
			signCt++
			textOut, _ = templates.Process("descriptions/rune", sign)
			response.SendUserMessage(userId, textOut, true)
		}
	}

	publicSigns := room.GetPublicSigns()
	for _, sign := range publicSigns {
		signCt++
		textOut, _ = templates.Process("descriptions/sign", sign)
		response.SendUserMessage(userId, textOut, true)
	}

	if signCt > 0 {
		response.SendUserMessage(userId, "", true)
	}

	textOut, _ = templates.Process("descriptions/who", details)
	response.SendUserMessage(userId, textOut, false)

	groundStuff := []string{}
	for containerName, container := range room.Containers {

		chestName := fmt.Sprintf(`<ansi fg="container">%s</ansi>`, containerName)

		if container.HasLock() {
			if container.Lock.IsLocked() {
				chestName += ` <ansi fg="white">(locked)</ansi>`
			} else {
				chestName += ` <ansi fg="white">(unlocked)</ansi>`
			}
		}

		groundStuff = append(groundStuff, chestName)

	}

	if room.Gold > 0 {
		groundStuff = append(groundStuff, fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, room.Gold))
	}

	for _, item := range room.Items {
		if !item.IsValid() {
			room.RemoveItem(item, false)
			continue
		}
		groundStuff = append(groundStuff, item.DisplayName())
	}

	groundDetails := map[string]any{
		`GroundStuff`: groundStuff,
		`IsDark`:      room.GetBiome().IsDark(),
		`IsNight`:     gametime.IsNight(),
	}
	textOut, _ = templates.Process("descriptions/ontheground", groundDetails)
	response.SendUserMessage(userId, textOut, false)

	textOut, _ = templates.Process("descriptions/exits", details)
	response.SendUserMessage(userId, textOut, false)

}
