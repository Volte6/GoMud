package discord

import (
	"fmt"
	"strings"

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

	message := fmt.Sprintf(":white_check_mark: **%v** connected", user.Character.Name)
	SendMessage(message)

	return true
}

// Player leaves the world event
func HandlePlayerDespawn(e events.Event) bool {
	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		return false
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return false
	}

	message := fmt.Sprintf(":x: **%v** disconnected", user.Character.Name)
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
	msgOut = strings.Replace(msgOut, evt.Level, `**`+evt.Level+`**`, 1)

	message := fmt.Sprintf(":sos: %s", msgOut)
	err := SendMessage(message)
	if err != nil {
		mudlog.Warn(`Discord`, `error`, err)
	}

	return true
}
