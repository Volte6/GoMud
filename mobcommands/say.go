package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Say(rest string, mobId int) (util.MessageQueue, error) {

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

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone says, "<ansi fg="yellow">%s</ansi>"`, rest), mobId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), mobId)
	}

	response.Handled = true
	return response, nil
}
