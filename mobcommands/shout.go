package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Shout(rest string, mobId int) (util.MessageQueue, error) {

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

	rest = strings.ToUpper(rest)

	if isSneaking {
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest))
	} else {
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest))
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				response.SendRoomMessage(otherRoom.RoomId, fmt.Sprintf(`Someone is shouting from the <ansi fg="exit">%s</ansi> direction.`, sourceExit))
			}
		}
	}

	response.Handled = true
	return response, nil
}
