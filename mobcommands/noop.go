package mobcommands

import (
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
)

// This is a no-op, mob does nothing
func Noop(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {
	return true, nil
}
