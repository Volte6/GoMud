package mobcommands

import (
	"fmt"

	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
)

func Break(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.Aggro != nil {
		mob.Character.Aggro = nil
		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> breaks off combat.`, mob.Character.Name))
	}

	return true, nil
}
