package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Break(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	if mob.Character.Aggro != nil {
		mob.Character.Aggro = nil
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> breaks off combat.`, mob.Character.Name))
	}

	response.Handled = true
	return response, nil
}
