package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// This is a no-op, mob does nothing
func Noop(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {
	return true, nil
}
