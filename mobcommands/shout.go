package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
)

func Shout(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		return true, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	rest = strings.ToUpper(rest)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), mob.InstanceId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), mob.InstanceId)
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendText(fmt.Sprintf(`Someone is shouting from the <ansi fg="exit">%s</ansi> direction.`, sourceExit))
			}
		}
	}

	return true, nil
}
