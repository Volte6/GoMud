package mobcommands

import (
	"fmt"
	"math/rand"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
)

func Converse(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		return true, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone says, "<ansi fg="yellow">%s</ansi>"`, rest))
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest))
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

		targetMob := mobs.GetInstance(roomMobInstId)
		if targetMob == nil {
			continue
		}

		if handled, err := scripting.TryMobConverse(rest, targetMob.InstanceId, mob.InstanceId); err == nil {
			if handled {
				return true, nil
			}
		}
	}

	return true, nil
}
