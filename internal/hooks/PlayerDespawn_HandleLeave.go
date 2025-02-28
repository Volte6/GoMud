package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

//
// Some clean up
//

func HandleLeave(e events.Event) bool {

	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "DeSpawned", "Actual Type", e.Type())
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	connId := user.ConnectionId()

	// Remove any zombie tracking for the user since they've been despawned from the world.
	if users.IsZombieConnection(connId) {
		users.RemoveZombieUser(evt.UserId)
	}

	room := rooms.LoadRoom(user.Character.RoomId)

	if currentParty := parties.Get(evt.UserId); currentParty != nil {
		currentParty.Leave(evt.UserId)
	}

	for _, mobInstId := range room.GetMobs(rooms.FindCharmed) {
		if mob := mobs.GetInstance(mobInstId); mob != nil {
			if mob.Character.IsCharmed(evt.UserId) {
				mob.Character.Charmed.Expire()
			}
		}
	}

	if _, ok := room.RemovePlayer(evt.UserId); ok {
		tplTxt, _ := templates.Process("player-despawn", user.Character.Name)
		room.SendText(tplTxt)
	}

	//
	// Send GMCP Updates for players leaving
	//
	for _, uid := range room.GetPlayers() {

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

	if err := users.LogOutUserByConnectionId(connId); err != nil {
		mudlog.Error("Log Out Error", "connectionId", connId, "error", err)
	}

	return true
}
