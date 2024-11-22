package mobcommands

import (
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
)

// This is a no-op, mob does nothing
func Noop(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {
	return true, nil
}
