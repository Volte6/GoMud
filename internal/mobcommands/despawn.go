package mobcommands

import (
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
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
