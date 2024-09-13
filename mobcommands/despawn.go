package mobcommands

import (
	"fmt"
	"log/slog"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func Despawn(rest string, mobId int) (bool, string, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
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

	return true, ``, nil
}
