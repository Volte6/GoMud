package hooks

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

//
// RoomChangeHandler waits for RoomChange events
// It then sends out GMCP data updates
// Also sends music changes out
//

func LocationGMCPUpdates(e events.Event) bool {

	evt, typeOk := e.(events.RoomChange)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RoomChange", "Actual Type", e.Type())
		return false
	}

	// If this isn't a user changing rooms, just pass it along.
	if evt.UserId == 0 {
		return true
	}

	// Get user... Make sure they still exist too.
	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	// Get the new room data... abort if doesn't exist.
	newRoom := rooms.LoadRoom(evt.ToRoomId)
	if newRoom == nil {
		return false
	}

	// Get the old room data... abort if doesn't exist.
	oldRoom := rooms.LoadRoom(evt.FromRoomId)
	if oldRoom == nil {
		return false
	}

	//
	// Send GMCP Updates
	//
	if connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Room`) {

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

				if connections.GetClientSettings(u.ConnectionId()).GmcpEnabled(`Room`) {

					events.AddToQueue(events.GMCPOut{
						UserId:  uid,
						Payload: fmt.Sprintf(`Room.AddPlayer {"name": "%s", "fullname": "%s"}`, user.Character.Name, user.Character.Name),
					})

				}
			}
		}

		// Send to everyone in the old room that a player left
		for _, uid := range oldRoom.GetPlayers() {

			if uid == user.UserId {
				continue
			}

			if u := users.GetByUserId(uid); u != nil {
				if connections.GetClientSettings(u.ConnectionId()).GmcpEnabled(`Room`) {
					events.AddToQueue(events.GMCPOut{
						UserId:  uid,
						Payload: fmt.Sprintf(`Room.RemovePlayer "%s"`, user.Character.Name),
					})
				}
			}
		}

		// We only need to send the player list to the player who moved
		// The Room.Info is already sent by the PreRoomChange hook
		events.AddToQueue(events.GMCPOut{
			UserId:  user.UserId,
			Payload: "Room.Players {" + newRoomPlayers.String() + `}`,
		})
	}

	return true
}
