package usercommands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/mapper"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
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

	var err error

	mapWidth := 65
	mapHeight := 21
	// assume 80x24 default?

	// Admin mapping gets a giant map
	borderWidth := 14
	borderHeight := 6 // Title, map top, map bottom, 2 legend, blank line.

	// Double the size
	//mapWidth = mapWidth << 1
	//mapHeight = mapHeight << 1

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

	zMapper := mapper.GetZoneMapper(zone)
	if zMapper == nil {
		mudlog.Error("Map", "error", "Could not find mapper for zone:"+zone)
		user.SendText(`No map found (or an error occured)"`)
		return true, err
	}

	c := mapper.Config{
		ZoomLevel: 5 - skillLevel,
		Width:     mapWidth,
		Height:    mapHeight,
		UserId:    user.UserId,
	}

	if skillLevel > 4 {
		for _, rid := range rooms.GetRoomsWithMobs() {
			if roomInfo := rooms.LoadRoom(rid); roomInfo != nil {
				if len(roomInfo.GetMobs(rooms.FindFighting|rooms.FindHostile)) > 0 {
					c.OverrideSymbol(rid, '☠', `Mob`)
				} else {
					c.OverrideSymbol(rid, '☺', `NPC`)
				}
			}
		}

		for _, rid := range rooms.GetRoomsWithPlayers() {
			c.OverrideSymbol(rid, '☺', `Player`)
		}
	}

	if p := parties.Get(user.UserId); p != nil {
		for _, uid := range p.GetMembers() {
			if tmpUser := users.GetByUserId(uid); tmpUser != nil {

				// Add any charmed mobs
				for _, mid := range tmpUser.Character.GetCharmIds() {
					if tmpMob := mobs.GetInstance(mid); tmpMob != nil {
						c.OverrideSymbol(tmpMob.Character.RoomId, '☹', `Friend`)
					}
				}

				c.OverrideSymbol(tmpUser.Character.RoomId, '☺', `Party Member`)
			}
		}
	}

	c.OverrideSymbol(user.Character.RoomId, '@', `You`)

	mapOutput := zMapper.GetLimitedMap(roomId, c)
	if skillLevel > 4 {
		//mapRender = m.GetFullMap(roomId, c)
	}

	legend := mapOutput.GetLegend(keywords.GetAllLegendAliases(room.Zone))

	width := 0

	displayLines := []string{}
	for i, line := range mapOutput.Render {
		displayLines = append(displayLines, string(line))
		if width == 0 {
			width = runewidth.StringWidth(displayLines[0])
		}
		for sym, txtLegend := range legend {
			txtLc := strings.ToLower(txtLegend)
			displayLines[i] = strings.Replace(displayLines[i], string(sym), fmt.Sprintf(`<ansi fg="map-room"><ansi fg="map-%s" bg="mapbg-%s">%c</ansi></ansi>`, txtLc, txtLc, sym), -1)
		}
	}

	mapData := map[string]any{
		"Title":        room.Zone,
		"DisplayLines": displayLines,
		"Height":       len(displayLines),
		"Width":        width,
		"Legend":       legend,
		"LegendWidth":  width,
		"LeftBorder": map[string]any{
			"Top":    ".-=~=-.",
			"Mid":    []string{"( _ __)", "(__  _)"},
			"Bottom": "`-._.-'",
		},
		"MidBorder": map[string]any{
			"Top":    "-._.-=",
			"Bottom": "-._.-=",
		},
		"RightBorder": map[string]any{
			"Top":    ".-=~=-.",
			"Mid":    []string{"( _ __)", "(__  _)"},
			"Bottom": "`-._.-'",
		},
	}

	mapTxt, err := templates.Process("maps/map", mapData, user.UserId)
	if err != nil {
		mudlog.Error("Map", "error", err.Error())
		user.SendText(`No map found (or an error occured)"`)
		return true, err
	}

	user.SendText(mapTxt)

	return true, nil
}
