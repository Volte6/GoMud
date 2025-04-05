package modules

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mapper"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {

	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	g := GMCPRoomModule{
		plug: plugins.New(`gmcp.Room`, `1.0`),
	}

	// Temporary for testing purposes.
	events.RegisterListener(events.RoomChange{}, g.roomChangeHandler)
	events.RegisterListener(events.PlayerDespawn{}, g.despawnHandler)

}

type GMCPRoomModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
}

func (g *GMCPRoomModule) despawnHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerDespawn", "Actual Type", e.Type())
		return events.Cancel
	}

	// If this isn't a user changing rooms, just pass it along.
	if evt.UserId == 0 {
		return events.Continue
	}

	room := rooms.LoadRoom(evt.RoomId)
	if room == nil {
		return events.Continue
	}

	//
	// Send GMCP Updates for players leaving
	//
	for _, uid := range room.GetPlayers() {

		if uid == evt.UserId {
			continue
		}

		u := users.GetByUserId(uid)
		if u == nil {
			continue
		}

		events.AddToQueue(GMCPOut{
			UserId:  uid,
			Payload: fmt.Sprintf(`Room.RemovePlayer "%s"`, evt.CharacterName),
		})

	}

	return events.Continue
}

func (g *GMCPRoomModule) roomChangeHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.RoomChange)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RoomChange", "Actual Type", e.Type())
		return events.Cancel
	}

	// If this isn't a user changing rooms, just pass it along.
	if evt.UserId == 0 {
		return events.Continue
	}

	// Get user... Make sure they still exist too.
	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Cancel
	}

	// Get the new room data... abort if doesn't exist.
	newRoom := rooms.LoadRoom(evt.ToRoomId)
	if newRoom == nil {
		return events.Cancel
	}

	// Get the old room data... abort if doesn't exist.
	oldRoom := rooms.LoadRoom(evt.FromRoomId)
	if oldRoom == nil {
		return events.Cancel
	}

	userGMCPEnabled := isGMCPEnabled(user.ConnectionId())

	if userGMCPEnabled {
		//
		// Send GMCP Updates
		//

		roomInfoStr := strings.Builder{}
		roomInfoStr.WriteString(`{ `)
		roomInfoStr.WriteString(`"num": ` + strconv.Itoa(newRoom.RoomId) + `, `)
		roomInfoStr.WriteString(`"name": "` + newRoom.Title + `", `)
		roomInfoStr.WriteString(`"area": "` + newRoom.Zone + `", `)
		roomInfoStr.WriteString(`"environment": "` + newRoom.GetBiome().Name() + `", `)

		// build coords
		// room coordinates (string of numbers separated by commas - area,X,Y,Z)
		// GoMud doesn't use numbers for areas, so will use string.
		roomInfoStr.WriteString(`"coords": "` + newRoom.Zone + `, `)
		m := mapper.GetZoneMapper(newRoom.Zone)
		x, y, z, err := m.GetCoordinates(newRoom.RoomId)
		if err != nil {
			roomInfoStr.WriteString(`999999999999999999,999999999999999999,999999999999999999`)
		} else {
			roomInfoStr.WriteString(strconv.Itoa(x))
			roomInfoStr.WriteString(`, `)
			roomInfoStr.WriteString(strconv.Itoa(y))
			roomInfoStr.WriteString(`, `)
			roomInfoStr.WriteString(strconv.Itoa(z))
		}
		roomInfoStr.WriteString(`", `)
		// end coords

		// build exits
		roomInfoStr.WriteString(`"exits": {`)
		exitCt := 0
		for name, exitInfo := range newRoom.Exits {
			if exitInfo.Secret {
				continue
			}

			if !mapper.IsCompassDirection(name) {
				continue
			}

			if exitCt > 0 {
				roomInfoStr.WriteString(`, `)
			}

			roomInfoStr.WriteString(`"` + name + `": ` + strconv.Itoa(exitInfo.RoomId))

			exitCt++
		}
		roomInfoStr.WriteString(`}, `)
		// End exits

		// build exits "extra" (exits that map to roomId AND include x/y/z delta information)
		roomInfoStr.WriteString(`"exitsv2": {`)
		exitCt = 0
		deltaX, deltaY, deltaZ := 0, 0, 0
		for name, exitInfo := range newRoom.Exits {

			if exitInfo.Secret {
				continue
			}

			if exitCt > 0 {
				roomInfoStr.WriteString(`, `)
			}

			if len(exitInfo.MapDirection) > 0 {
				deltaX, deltaY, deltaZ = mapper.GetDelta(exitInfo.MapDirection)
			} else {
				deltaX, deltaY, deltaZ = mapper.GetDelta(name)
			}

			roomInfoStr.WriteString(`"` + name + `": { "num": `)
			roomInfoStr.WriteString(strconv.Itoa(exitInfo.RoomId))
			roomInfoStr.WriteString(`, "dx": `)
			roomInfoStr.WriteString(strconv.Itoa(deltaX))
			roomInfoStr.WriteString(`, "dy": `)
			roomInfoStr.WriteString(strconv.Itoa(deltaY))
			roomInfoStr.WriteString(`, "dz": `)
			roomInfoStr.WriteString(strconv.Itoa(deltaZ))

			if exitInfo.HasLock() {
				roomInfoStr.WriteString(`, "locked": true`)

				lockId := fmt.Sprintf(`%d-%s`, newRoom.RoomId, name)
				hasKey, hasSequence := user.Character.HasKey(lockId, int(exitInfo.Lock.Difficulty))

				roomInfoStr.WriteString(`, "haskey": `)
				roomInfoStr.WriteString(strconv.FormatBool(hasKey))

				roomInfoStr.WriteString(`, "haspickcombo": `)
				roomInfoStr.WriteString(strconv.FormatBool(hasSequence))
			}

			roomInfoStr.WriteString(`}`)

			exitCt++
		}
		roomInfoStr.WriteString(`}, `)
		// End exits

		// build details
		roomInfoStr.WriteString(`"details": [`)

		detailCt := 0
		if len(newRoom.GetMobs(rooms.FindMerchant)) > 0 || len(newRoom.GetPlayers(rooms.FindMerchant)) > 0 {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"shop"`)
		}
		if len(newRoom.SkillTraining) > 0 {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"trainer"`)
		}
		if newRoom.IsBank {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"bank"`)
		}
		if newRoom.IsStorage {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"storage"`)
		}
		if newRoom.IsCharacterRoom {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"character"`)
		}
		roomInfoStr.WriteString(`], `)
		// end details

		// Room Contents
		roomInfoStr.WriteString(`"Contents": `)

		contents := GMCPRoomModule_Payload_Contents{}

		// Room.Contents.Containers
		contents.Containers = []GMCPRoomModule_Payload_Contents_Container{}
		for name, container := range newRoom.Containers {

			c := GMCPRoomModule_Payload_Contents_Container{
				Name:   name,
				Usable: len(container.Recipes) > 0,
			}

			if container.HasLock() {
				c.Locked = true
				lockId := fmt.Sprintf(`%d-%s`, newRoom.RoomId, name)
				c.HasKey, c.HasPickCombo = user.Character.HasKey(lockId, int(container.Lock.Difficulty))
			}

			contents.Containers = append(contents.Containers, c)
		}

		// Room.Contents.Items
		contents.Items = []GMCPRoomModule_Payload_Contents_Item{}
		for _, itm := range newRoom.Items {
			contents.Items = append(contents.Items, GMCPRoomModule_Payload_Contents_Item{
				Id:        itm.ShorthandId(),
				Name:      itm.Name(),
				QuestFlag: itm.GetSpec().QuestToken != ``,
			})
		}

		// Room.Contents.Players
		contents.Players = []GMCPRoomModule_Payload_Contents_Character{}
		for _, uId := range newRoom.GetPlayers() {

			// Exclude viewing player
			if uId == user.UserId {
				continue
			}

			u := users.GetByUserId(uId)
			if u == nil {
				continue
			}

			contents.Players = append(contents.Players, GMCPRoomModule_Payload_Contents_Character{
				Id:         u.ShorthandId(),
				Name:       u.Character.Name,
				Adjectives: u.Character.GetAdjectives(),
				Aggro:      u.Character.Aggro != nil,
				Shop:       len(u.Character.Shop) > 0,
			})
		}

		// Room.Contents.Npcs
		contents.Npcs = []GMCPRoomModule_Payload_Contents_Character{}
		for _, mIId := range newRoom.GetMobs() {
			mob := mobs.GetInstance(mIId)
			if mob == nil {
				continue
			}

			c := GMCPRoomModule_Payload_Contents_Character{
				Id:         mob.ShorthandId(),
				Name:       mob.Character.Name,
				Adjectives: mob.Character.GetAdjectives(),
				Aggro:      mob.Character.Aggro != nil,
				Shop:       len(mob.Character.Shop) > 0,
			}

			if len(mob.QuestFlags) > 0 {
				for _, qFlag := range mob.QuestFlags {
					if user.Character.HasQuest(qFlag) || (len(qFlag) >= 5 && qFlag[len(qFlag)-5:] == `start`) {
						c.QuestFlag = true
						break
					}
				}
			}

			contents.Npcs = append(contents.Npcs, c)
		}

		b, _ := json.Marshal(contents)
		roomInfoStr.WriteString(string(b))

		// End Room Contents

		roomInfoStr.WriteString(` }`)
		// End room info

		// send big 'ol room info object
		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  "Room.Info",
			Payload: roomInfoStr.String(),
		})

	} // end user specific room info

	newRoomPlayers := strings.Builder{}

	// Send to everyone in the new room that a player arrived
	for _, uid := range newRoom.GetPlayers() {

		if uid == user.UserId {
			continue
		}

		if u := users.GetByUserId(uid); u != nil {

			if newRoomPlayers.Len() > 0 {
				newRoomPlayers.WriteString(`, `)
			}

			newRoomPlayers.WriteString(`"` + u.Character.Name + `": ` + `"` + u.Character.Name + `"`)

			if !isGMCPEnabled(u.ConnectionId()) {
				continue
			}

			events.AddToQueue(GMCPOut{
				UserId:  uid,
				Module:  `Room.AddPlayer`,
				Payload: fmt.Sprintf(`{"name": "%s", "fullname": "%s"}`, user.Character.Name, user.Character.Name),
			})

		}
	}

	if userGMCPEnabled {
		// send player list for room
		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Room.Players`,
			Payload: "{" + newRoomPlayers.String() + `}`,
		})
	}

	// Send to everyone in the old room that a player left
	for _, uid := range oldRoom.GetPlayers() {

		if uid == user.UserId {
			continue
		}

		if u := users.GetByUserId(uid); u != nil {

			if !isGMCPEnabled(u.ConnectionId()) {
				continue
			}

			events.AddToQueue(GMCPOut{
				UserId:  uid,
				Module:  `Room.RemovePlayer`,
				Payload: fmt.Sprintf(`"%s"`, user.Character.Name),
			})

		}
	}

	return events.Continue
}

// /////////////////
// Room.Contents
// /////////////////
type GMCPRoomModule_Payload_Contents struct {
	Players    []GMCPRoomModule_Payload_Contents_Character `json:"Players"`
	Npcs       []GMCPRoomModule_Payload_Contents_Character `json:"Npcs"`
	Items      []GMCPRoomModule_Payload_Contents_Item      `json:"Items"`
	Containers []GMCPRoomModule_Payload_Contents_Container `json:"Containers"`
}

type GMCPRoomModule_Payload_Contents_Character struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Adjectives []string `json:"adjectives"`
	Aggro      bool     `json:"aggro"`
	Shop       bool     `json:"shop"`
	QuestFlag  bool     `json:"quest_flag"`
}

type GMCPRoomModule_Payload_Contents_Item struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	QuestFlag bool   `json:"quest_flag"`
}

type GMCPRoomModule_Payload_Contents_Container struct {
	Name         string `yaml:"name"`
	Locked       bool   `yaml:"locked"`
	HasKey       bool   `yaml:"haskey"`
	HasPickCombo bool   `yaml:"haspickcombo"`
	Usable       bool   `yaml:"usable"`
}
