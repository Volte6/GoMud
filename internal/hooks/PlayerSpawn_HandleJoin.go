package hooks

import (
	"fmt"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

//
// Execute on join commands
//

func HandleJoin(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerSpawn", "Actual Type", e.Type())
		return events.Cancel
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		mudlog.Error("HandleJoin", "error", fmt.Sprintf(`user %d not found`, evt.UserId))
		return events.Cancel
	}

	user.EventLog.Add(`conn`, fmt.Sprintf(`<ansi fg="username">%s</ansi> entered the world`, user.Character.Name))

	users.RemoveZombieUser(evt.UserId)

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {

		mudlog.Error("EnterWorld", "error", fmt.Sprintf(`room %d not found`, user.Character.RoomId))

		user.Character.RoomId = 1
		user.Character.Zone = "Frostfang"
		room = rooms.LoadRoom(user.Character.RoomId)
		if room == nil {
			mudlog.Error("EnterWorld", "error", fmt.Sprintf(`room %d not found`, user.Character.RoomId))
		}
	}

	// TODO HERE
	loginCmds := configs.GetConfig().Server.OnLoginCommands
	if len(loginCmds) > 0 {

		for _, cmd := range loginCmds {

			events.AddToQueue(events.Input{
				UserId:    evt.UserId,
				InputText: cmd,
				ReadyTurn: 0, // No delay between execution of commands
			})

		}

	}

	//
	// Send GMCP for their char name
	//
	if connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Char`) {

		events.AddToQueue(events.GMCPOut{
			UserId:  user.UserId,
			Payload: fmt.Sprintf(`Char.Name {"name": "%s", "fullname": "%s"}`, user.Character.Name, user.Character.Name),
		})

	}

	return events.Continue
}
