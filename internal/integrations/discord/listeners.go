package discord

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/users"
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
	err := SendMessage(message)
	if err != nil {
		return false
	}

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
		return false
	}

	return true
}
