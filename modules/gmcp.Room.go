package modules

import (
	"fmt"
	"strconv"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mapper"
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

	// connectionId to map[string]int
	g.cache, _ = lru.New[uint64, map[string]int](128)

	// Temporary for testing purposes.
	events.RegisterListener(events.RoomChange{}, g.roomChangeHandler)
	events.RegisterListener(events.PlayerDespawn{}, g.despawnHandler)

	events.RegisterListener(GMCPModules{}, func(e events.Event) events.ListenerReturn {
		if evt, ok := e.(GMCPModules); ok {
			g.cache.Add(evt.ConnectionId, evt.Modules)
		}
		return events.Continue
	})
}

type GMCPRoomModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug  *plugins.Plugin
	cache *lru.Cache[uint64, map[string]int]
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

	////////////////////////////////////
	// Check for support of this module
	////////////////////////////////////
	userHasModule := false
	if supportedModules, ok := g.cache.Get(user.ConnectionId()); ok {
		if _, ok := supportedModules[`Room`]; ok {
			userHasModule = true
		}
	}
	////////////////////////////////////
	// End check for support of this module
	////////////////////////////////////

	if userHasModule {
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
		roomInfoStr.WriteString(`]`)
		// end details

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

			if !g.supportsModule(u.ConnectionId(), `Room`) {
				continue
			}

			events.AddToQueue(GMCPOut{
				UserId:  uid,
				Module:  `Room.AddPlayer`,
				Payload: fmt.Sprintf(`{"name": "%s", "fullname": "%s"}`, user.Character.Name, user.Character.Name),
			})

		}
	}

	if userHasModule {
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

			if !g.supportsModule(u.ConnectionId(), `Room`) {
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

func (g *GMCPRoomModule) supportsModule(connectionId uint64, moduleName string) bool {
	supportedModules, ok := g.cache.Get(connectionId)
	if ok {
		if _, ok := supportedModules[moduleName]; ok {
			return true
		}
	} else {
		// Request that the gmcp module get the data and send the event
		events.AddToQueue(GMCPRequestModules{ConnectionId: connectionId})
	}
	return false
}
