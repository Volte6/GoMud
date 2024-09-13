package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func Shout(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		return true, ``, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	rest = strings.ToUpper(rest)

	if isSneaking {
		room.SendText(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), mobId)
	} else {
		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, mob.Character.Name, rest), mobId)
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendText(fmt.Sprintf(`Someone is shouting from the <ansi fg="exit">%s</ansi> direction.`, sourceExit))
			}
		}
	}

	return true, ``, nil
}
