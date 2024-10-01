package mobcommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
)

func Wander(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.IsCharmed() {
		return true, errors.New("friendly mobs don't wander")
	}

	// If they aren't home and need to go home, do it.
	if mob.Character.RoomId != mob.HomeRoomId {
		if mob.MaxWander > -1 { // -1 means they can wander forever and never go home. 0 means they never wander.
			if len(mob.RoomStack) > mob.MaxWander {

				mob.Command(`go home`)

				return true, nil
			}
		}
	}

	// If MaxWander is zero, they don't wander.
	if mob.MaxWander == 0 {
		return true, nil
	}

	exitOptions := make([]string, 0)
	restrictZone := true

	// First only consider adjacent rooms with loot in them
	if rest == `loot` {
		for exitName, exit := range room.Exits {
			exitRoom := rooms.LoadRoom(exit.RoomId)
			if len(exitRoom.Items) > 0 || exitRoom.Gold > 0 {
				exitOptions = append(exitOptions, exitName)
			}
		}
	}

	// First only consider adjacent rooms with loot in them
	if rest == `players` {
		for exitName, exit := range room.Exits {
			exitRoom := rooms.LoadRoom(exit.RoomId)
			if exitRoom.PlayerCt() > 0 {
				exitOptions = append(exitOptions, exitName)
			}
		}
	}

	if exitName, roomId := room.GetRandomExit(); roomId > 0 {
		if r := rooms.LoadRoom(roomId); r != nil {
			if !restrictZone || r.Zone == mob.Character.Zone {

				mob.Command(fmt.Sprintf("go %s", exitName))

			}
		}

	}

	return true, nil
}
