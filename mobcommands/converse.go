package mobcommands

import (
	"fmt"
	"math/rand"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
)

func Converse(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		return true, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone says, "<ansi fg="yellow">%s</ansi>"`, rest), mobId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), mobId)
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

		if handled, err := scripting.TryMobConverse(rest, roomMobInstId, mobId); err == nil {
			if handled {
				return true, nil
			}
		}
	}

	return true, nil
}
