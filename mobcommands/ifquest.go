package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

// Expected format is:
// ifquest 1-start say has anyone seen my locket?
// or
// if -1-start Say has anyone seen my locket?
// The second message will only be executed if the quest is successfully given.
func IfQuest(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		response.Handled = true
		return response, nil
	}

	parts := strings.SplitN(rest, ` `, 2)

	if len(parts) < 2 {
		response.Handled = true
		return response, nil
	}

	questToken := parts[0]
	parts = parts[1:]
	cmd := strings.Join(parts, ` `)

	for _, uid := range room.GetPlayers() {
		if user := users.GetByUserId(uid); user != nil {

			if user.Character.HasQuestToken(questToken) {
				cmdQueue.QueueCommand(0, mob.InstanceId, cmd)
				break
			}

		}
	}

	response.Handled = true
	return response, nil
}

// Expected format is:
// ifquest 1-start say has anyone seen my locket?
// or
// if -1-start Say has anyone seen my locket?
// The second message will only be executed if the quest is successfully given.
func IfNotQuest(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		response.Handled = true
		return response, nil
	}

	parts := strings.SplitN(rest, ` `, 2)

	if len(parts) < 2 {
		response.Handled = true
		return response, nil
	}

	questToken := parts[0]
	parts = parts[1:]
	cmd := strings.Join(parts, ` `)

	for _, uid := range room.GetPlayers() {
		if user := users.GetByUserId(uid); user != nil {

			if !user.Character.HasQuestToken(questToken) {
				cmdQueue.QueueCommand(0, mob.InstanceId, cmd)
				break
			}

		}
	}

	response.Handled = true
	return response, nil
}
