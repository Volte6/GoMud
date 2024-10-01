package mobcommands

import (
	"fmt"
	"log/slog"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func Despawn(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if room == nil {
		return false, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	slog.Info("despawn", "mobname", mob.Character.Name, "reason", rest)

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
