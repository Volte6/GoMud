package modules

import (
	"math"

	"github.com/volte6/gomud/internal/events"
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
	g := GMCPPartyModule{
		plug: plugins.New(`gmcp.Comm`, `1.0`),
	}

	events.RegisterListener(events.PartyUpdated{}, g.onPartyChange)

}

type GMCPPartyModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
}

func (g *GMCPPartyModule) onPartyChange(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PartyUpdated)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PartyUpdated", "Actual Type", e.Type())
		return events.Cancel
	}

	if len(evt.UserIds) == 0 {
		return events.Cancel
	}

	partyPayload := GMCPPartyModule_Payload{
		Leader:  `None`,
		Members: []GMCPPartyModule_Payload_User{},
		Invited: []GMCPPartyModule_Payload_User{},
		Vitals:  map[string]GMCPPartyModule_Payload_Vitals{},
	}

	var party *parties.Party
	for _, uId := range evt.UserIds {
		if party = parties.Get(uId); party != nil {
			break
		}
	}

	inParty := map[int]string{}
	roomTitles := map[int]string{}

	if party != nil {

		for _, uId := range party.GetMembers() {

			if user := users.GetByUserId(uId); user != nil {

				if user.UserId == party.LeaderUserId {
					partyPayload.Leader = user.Character.Name
				}

				inParty[user.UserId] = user.Character.Name

				roomTitle, ok := roomTitles[user.Character.RoomId]
				if !ok {
					if uRoom := rooms.LoadRoom(user.Character.RoomId); uRoom != nil {
						roomTitle = uRoom.Title
						roomTitles[user.Character.RoomId] = roomTitle
					}
				}

				hPct := int(math.Floor((float64(user.Character.Health) / float64(user.Character.HealthMax.Value)) * 100))
				if hPct < 0 {
					hPct = 0
				}
				partyPayload.Members = append(partyPayload.Members,
					GMCPPartyModule_Payload_User{
						Name:     user.Character.Name,
						Status:   `In Party`,
						Position: party.GetRank(user.UserId),
					},
				)

				partyPayload.Vitals[user.Character.Name] = GMCPPartyModule_Payload_Vitals{
					Level:         user.Character.Level,
					HealthPercent: hPct,
					Location:      roomTitle,
				}
			}

		}

		for _, uId := range party.GetInvited() {

			if user := users.GetByUserId(uId); user != nil {

				inParty[user.UserId] = user.Character.Name

				roomTitle, ok := roomTitles[user.Character.RoomId]
				if !ok {
					if uRoom := rooms.LoadRoom(user.Character.RoomId); uRoom != nil {
						roomTitle = uRoom.Title
						roomTitles[user.Character.RoomId] = roomTitle
					}
				}

				partyPayload.Invited = append(partyPayload.Invited,
					GMCPPartyModule_Payload_User{
						Name:     user.Character.Name,
						Status:   `Invited`,
						Position: ``,
					},
				)

				partyPayload.Vitals[user.Character.Name] = GMCPPartyModule_Payload_Vitals{
					Level:         0,
					HealthPercent: 0,
					Location:      ``,
				}

			}

		}

	}

	for _, userId := range evt.UserIds {

		if _, ok := inParty[userId]; ok {

			events.AddToQueue(GMCPOut{
				UserId:  userId,
				Module:  `Party`,
				Payload: partyPayload,
			})

		} else {

			events.AddToQueue(GMCPOut{
				UserId:  userId,
				Module:  `Party`,
				Payload: GMCPPartyModule_Payload{},
			})

		}

	}

	return events.Continue
}

type GMCPPartyModule_Payload struct {
	Leader  string
	Members []GMCPPartyModule_Payload_User
	Invited []GMCPPartyModule_Payload_User
	Vitals  map[string]GMCPPartyModule_Payload_Vitals
}

type GMCPPartyModule_Payload_User struct {
	Name     string `json:"name"`
	Status   string `json:"status"`   // party/leader/invited
	Position string `json:"position"` // frontrank/middle/backrank
}

type GMCPPartyModule_Payload_Vitals struct {
	Level         int    `json:"level"`    // level of user
	HealthPercent int    `json:"health"`   // 1 = 1%, 23 = 23% etc.
	Location      string `json:"location"` // Title of room they are in
}
