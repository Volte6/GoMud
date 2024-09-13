package mobcommands

import (
	"fmt"
	"math/rand"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/util"
)

func Converse(rest string, mobId int) (util.MessageQueue, error) {

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
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`someone says, "<ansi fg="yellow">%s</ansi>"`, rest), true)
	} else {
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), true)
	}

	roomMobs := room.GetMobs(rooms.FindIdle)

	// Randomize the mobs to determine who will potentially capture the message first
	for i := range roomMobs {
		j := rand.Intn(i + 1)
		roomMobs[i], roomMobs[j] = roomMobs[j], roomMobs[i]
	}

	for _, roomMobInstId := range roomMobs {

		if roomMobInstId == mob.InstanceId {
			continue
		}

		mob := mobs.GetInstance(roomMobInstId)
		if mob == nil {
			continue
		}

		if result, err := scripting.TryMobConverse(rest, roomMobInstId, mobId); err == nil {
			response.AbsorbMessages(result)
			if result.Handled {
				response.Handled = true
				return response, nil
			}
		}
	}

	response.Handled = true
	return response, nil
}
