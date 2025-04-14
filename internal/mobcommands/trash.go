package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Trash(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindInBackpack(rest)

	if found {
		mob.Character.RemoveItem(matchItem)

		events.AddToQueue(events.ItemOwnership{
			MobInstanceId: mob.InstanceId,
			Item:          matchItem,
			Gained:        false,
		})

	}

	return true, nil
}
