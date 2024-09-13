package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func LookForAid(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	isCharmed := mob.Character.IsCharmed()
	if !isCharmed {
		response.Handled = true
		return response, nil
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	for _, playerId := range room.GetPlayers(rooms.FindDowned) {

		user := users.GetByUserId(playerId)
		if user == nil {
			continue
		}

		if mob.Character.IsCharmed(playerId) {
			mob.Command(fmt.Sprintf("aid @%d", playerId)) // @ denotes a specific player id
			continue
		}

	}

	response.Handled = true
	return response, nil
}
