package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Despawn(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	mudlog.Info("despawn", "mobname", mob.Character.Name, "reason", rest)

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.InstanceId)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns(true)
	}

	// Remove from current room
	room.RemoveMob(mob.InstanceId)

	return true, nil
}
