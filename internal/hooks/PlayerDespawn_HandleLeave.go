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

func HandleLeave(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerDespawn", "Actual Type", e.Type())
		return events.Cancel
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		mudlog.Error("HandleLeave", "error", fmt.Sprintf(`user %d not found`, evt.UserId))
		return events.Cancel
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

	tplTxt, _ := templates.Process("goodbye", nil, evt.UserId)
	connections.SendTo([]byte(templates.AnsiParse(tplTxt)), connId)

	if err := users.LogOutUserByConnectionId(connId); err != nil {
		mudlog.Error("Log Out Error", "connectionId", connId, "error", err)
	}
	connections.Remove(connId)

	return events.Continue
}
