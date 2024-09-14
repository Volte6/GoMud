package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func LookForAid(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	isCharmed := mob.Character.IsCharmed()
	if !isCharmed {
		return true, nil
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

	return true, nil
}
