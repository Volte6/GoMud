package modules

import (
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/parties"
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
	g := GMCPCommModule{
		plug: plugins.New(`gmcp.Comm`, `1.0`),
	}

	// connectionId to map[string]int
	g.cache, _ = lru.New[uint64, map[string]int](128)

	events.RegisterListener(events.Communication{}, g.onComm)

	events.RegisterListener(GMCPModules{}, func(e events.Event) events.ListenerReturn {
		if evt, ok := e.(GMCPModules); ok {
			g.cache.Add(evt.ConnectionId, evt.Modules)
		}
		return events.Continue
	})

}

type GMCPCommModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug  *plugins.Plugin
	cache *lru.Cache[uint64, map[string]int]
}

func (g *GMCPCommModule) onComm(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.Communication)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RoomChange", "Actual Type", e.Type())
		return events.Cancel
	}

	msgPayload := `{channel: "` + evt.CommType + `", sender: "` + evt.Name + `", text: "` + strings.ReplaceAll(evt.Message, `"`, `\\"`) + `"}`

	// Sent to everyone.
	// say, party, broadcast, whisper

	sendToUserIds := []int{}

	if evt.CommType == `say` {

		roomId := 0

		if evt.SourceUserId > 0 {
			if user := users.GetByUserId(evt.SourceUserId); user != nil {
				roomId = user.Character.RoomId
			}
		}

		if evt.SourceMobInstanceId > 0 {
			if mob := mobs.GetInstance(evt.SourceMobInstanceId); mob != nil {
				roomId = mob.Character.RoomId
			}
		}

		if roomId > 0 {
			if room := rooms.LoadRoom(roomId); room != nil {
				sendToUserIds = append([]int{}, room.GetPlayers()...)
			}
		}

	} else if evt.CommType == `party` {

		if evt.SourceUserId > 0 {
			if party := parties.Get(evt.SourceUserId); party != nil {
				sendToUserIds = append([]int{}, party.UserIds...)
			}
		}

	} else if evt.CommType == `broadcast` {

		sendToUserIds = append([]int{}, users.GetOnlineUserIds()...)

	} else if evt.CommType == `whisper` {

		if evt.TargetUserId > 0 {
			sendToUserIds = append(sendToUserIds, evt.TargetUserId)
		}

	}

	for _, userId := range sendToUserIds {

		if userId == evt.SourceUserId && evt.CommType != `broadcast` {
			continue
		}

		events.AddToQueue(GMCPOut{
			UserId:  userId,
			Module:  `Comm.Channel`,
			Payload: msgPayload,
		})

	}

	return events.Continue
}

func (g *GMCPCommModule) supportsModule(connectionId uint64, moduleName string) bool {
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
