package modules

import (
	"github.com/Volte6/ansitags"
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

	events.RegisterListener(events.Communication{}, g.onComm)

}

type GMCPCommModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
}

func (g *GMCPCommModule) onComm(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.Communication)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RoomChange", "Actual Type", e.Type())
		return events.Cancel
	}

	payload := GMCPCommModule_Payload{
		Channel: evt.CommType,
		Sender:  evt.Name,
		Text:    ansitags.Parse(evt.Message, ansitags.StripTags),
	}

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

		// Exclude user from receiving their own messages?
		//if userId == evt.SourceUserId && evt.CommType != `broadcast` {
		//continue
		//}

		events.AddToQueue(GMCPOut{
			UserId:  userId,
			Module:  `Comm.Channel`,
			Payload: payload,
		})

	}

	return events.Continue
}

type GMCPCommModule_Payload struct {
	Channel string `json:"channel"`
	Sender  string `json:"sender"`
	Text    string `json:"text"`
}
