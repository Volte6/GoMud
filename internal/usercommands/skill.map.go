package usercommands

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

/*
Skill Map
Level 1 - Map a 5x5 area
Level 2 - Map a 9x7 area
Level 3 - Map a 13x9 area
Level 4 - Map a 17x9 area, and enables the "wide" version.
*/
func Map(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Map)

	if skillLevel == 0 {
		user.SendText("You don't know how to map.")
		return true, errors.New(`you don't know how to map`)
	}

	if rest == "sprawl" {
		user.SendText(fmt.Sprintf("The reach of your maps is %d rooms.", user.Character.GetMapSprawlCapacity()))
		return true, nil
	}

	if rest == "wide" && skillLevel < 4 {
		user.SendText("You don't know how to create a wide map.")
		return true, errors.New(`you don't know how to create a wide map`)
	}

	if !user.Character.TryCooldown(skills.Map.String(), "1 round") {
		user.SendText(
			`You can only create 1 map per round.`,
		)
		return true, errors.New(`you're doing that too often`)
	}

	// replace any non alpha/numeric characters in "rest"
	zone := rest
	roomId := 0
	if zone != "" && zone != "wide" {
		zone = rooms.FindZoneName(zone)
		if zone != user.Character.Zone {
			roomId, _ = rooms.GetZoneRoot(zone)
		}
	}

	if zone == "" || roomId == 0 {
		zone = user.Character.Zone
		roomId = user.Character.RoomId
	}

	// First check for a premade map.
	if mapTxt, err := templates.Process("maps/"+rooms.ZoneNameSanitize(zone), zone); err == nil {
		user.SendText(mapTxt)
		return true, nil
	}

	var mapData rooms.MapData
	var err error

	mapWidth := 65
	mapHeight := 18
	//mapWidth := 200
	//mapHeight := 56
	// assume 80x24 default?

	// Admin mapping gets a giant map
	borderWidth := 14
	borderHeight := 6 // Title, map top, map bottom, 2 legend, blank line.

	// level 1 map size
	mapWidth = 5
	mapHeight = 5

	if skillLevel > 1 {
		mapWidth = mapWidth + (skillLevel-1)*4   // 3 * 3 = 9 more coverage?
		mapHeight = mapHeight + (skillLevel-1)*2 // 3 * 2 = 6 more coverage?
	}

	mapMaxWidth := 33
	//mapWidthDelta := mapMaxWidth - mapWidth // 16/?
	mapWidth += int(float64(user.Character.Stats.Perception.ValueAdj) / 5)
	if mapWidth > mapMaxWidth {
		mapWidth = mapMaxWidth
	}

	// Double the size
	mapWidth = mapWidth << 1
	mapHeight = mapHeight << 1

	// lets max the height:
	if mapHeight > 18 {
		mapHeight = 18
	}

	if skillLevel > 4 {

		sw := 80
		sh := 40
		if user.ClientSettings().Display.ScreenWidth > 0 {
			sw = int(user.ClientSettings().Display.ScreenWidth)
			sh = int(user.ClientSettings().Display.ScreenHeight)
		}

		mapWidth = int(sw) - borderWidth
		mapHeight = sh - borderHeight // extra 2 for the new lines after
		if mapHeight%2 != 0 {
			mapHeight--
		}

		if mapWidth > sw-borderWidth {
			mapWidth = sw - borderWidth
		}
		if mapHeight > sh-borderHeight {
			mapHeight = sh - borderHeight
		}

	}

	mapMode := rooms.MapModeAllButSecrets
	if skillLevel > 4 {
		mapMode = rooms.MapModeAll
	}

	var rGraph *rooms.RoomGraph
	if rest == "wide" {
		rGraph = rooms.GenerateZoneMap(zone, roomId, user.UserId, mapWidth, mapHeight, mapMode)
	} else {
		rGraph = rooms.GenerateZoneMap(zone, roomId, user.UserId, int(math.Ceil(float64(mapWidth)/2))<<1, int(math.Ceil(float64(mapHeight)/2))<<1, mapMode)
	}

	if skillLevel > 4 {
		for _, rid := range rooms.GetRoomsWithMobs() {
			if roomInfo := rooms.LoadRoom(rid); roomInfo != nil {
				if len(roomInfo.GetMobs(rooms.FindFighting|rooms.FindHostile)) > 0 {
					rGraph.AddRoomSymbolOverrides('☠', "Mob", rid)
				} else {
					rGraph.AddRoomSymbolOverrides('☺', "NPC", rid)
				}
			}
		}

		for _, rid := range rooms.GetRoomsWithPlayers() {
			rGraph.AddRoomSymbolOverrides('☺', "Player", rid)
		}
	}

	if p := parties.Get(user.UserId); p != nil {
		for _, uid := range p.GetMembers() {
			if tmpUser := users.GetByUserId(uid); tmpUser != nil {

				// Add any charmed mobs
				for _, mid := range tmpUser.Character.GetCharmIds() {
					if tmpMob := mobs.GetInstance(mid); tmpMob != nil {
						rGraph.AddRoomSymbolOverrides('☹', "Friend", tmpMob.Character.RoomId)
					}
				}

				rGraph.AddRoomSymbolOverrides('☺', "Player", tmpUser.Character.RoomId)
			}
		}
	}

	rGraph.AddRoomSymbolOverrides('@', "You", user.Character.RoomId)

	if rest == "wide" {
		mapData, err = rooms.DrawZoneMapWide(rGraph, zone, mapWidth, mapHeight)
	} else {
		mapData, err = rooms.DrawZoneMap(rGraph, zone, mapWidth, mapHeight)
	}

	if mapData.LegendWidth < 72 { // 80 - " Legend "
		mapData.LegendWidth = 72
	}

	//mapData, err := rooms.GenerateZoneMapZoomedOut(zone, roomId, 0, 65, 18)
	if err != nil {
		return false, err
	}

	mapTxt, err := templates.Process("maps/map", mapData)
	if err != nil {
		slog.Error("Map", "error", err.Error())
		user.SendText(`No map found (or an error occured)"`)
		return true, err
	}

	user.SendText(mapTxt)

	return true, nil
}
