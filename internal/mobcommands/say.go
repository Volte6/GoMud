package mobcommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
)

func Say(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {

		return true, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone says, "<ansi fg="saytext">%s</ansi>"`, rest))
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> says, "<ansi fg="saytext">%s</ansi>"`, mob.Character.Name, rest))
	}

	room.SendTextToExits(`You hear someone talking.`, true)

	return true, nil
}
