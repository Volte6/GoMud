package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

func Break(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	if mob.Character.Aggro != nil {
		mob.Character.Aggro = nil
		response.SendRoomMessage(mob.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> breaks off combat.`, mob.Character.Name))
	}

	response.Handled = true
	return response, nil
}
