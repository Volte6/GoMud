package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func LookForAid(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	isCharmed := mob.Character.IsCharmed()
	if !isCharmed {
		return true, nil
	}

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
