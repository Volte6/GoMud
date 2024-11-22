package mobcommands

import (
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
)

func Sneak(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if isSneaking {
		return true, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        0,
		MobInstanceId: mob.InstanceId,
		BuffId:        9, // Buff 9 is sneak
	})

	return true, nil
}
