package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

func Sneak(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if isSneaking {
		response.Handled = true
		return response, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        0,
		MobInstanceId: mobId,
		BuffId:        9, // Buff 9 is sneak
	})

	response.Handled = true
	return response, nil
}
