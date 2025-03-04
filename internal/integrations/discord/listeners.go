package discord

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

// Player enters the world event
func HandlePlayerSpawn(e events.Event) bool {
	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	connDetails := connections.Get(user.ConnectionId())

	message := fmt.Sprintf(":white_check_mark: **%s** connected", user.Character.Name)

	if connDetails.IsWebsocket() {
		message += ` (via websocket)`
	}

	SendMessage(message)

	return true
}

// Player leaves the world event
func HandlePlayerDespawn(e events.Event) bool {
	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		return false
	}

	message := fmt.Sprintf(":x: **%s** disconnected", evt.CharacterName)
	err := SendMessage(message)
	if err != nil {
		mudlog.Warn(`Discord`, `error`, err)
	}

	return true
}

func HandleLogs(e events.Event) bool {
	evt, typeOk := e.(events.Log)
	if !typeOk {
		return false
	}

	if evt.Level != `ERROR` {
		return true
	}

	msgOut := util.StripANSI(fmt.Sprintln(evt.Data[1:]...))

	// Skip script timeout messages
	if strings.Contains(msgOut, `JSVM`) && strings.Contains(msgOut, `script timeout`) {
		return true
	}

	msgOut = strings.Replace(msgOut, evt.Level, `**`+evt.Level+`**`, 1)

	message := fmt.Sprintf(":sos: %s", msgOut)
	err := SendMessage(message)
	if err != nil {
		mudlog.Warn(`Discord`, `error`, err)
	}

	return true
}
