package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobs"
)

func Sneak(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if isSneaking {
		return true, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        0,
		MobInstanceId: mobId,
		BuffId:        9, // Buff 9 is sneak
	})

	return true, nil
}
