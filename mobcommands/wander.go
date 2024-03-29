package mobcommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

func Wander(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	if mob.Character.IsCharmed() {
		response.Handled = true
		return response, errors.New("friendly mobs don't wander")
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// If they aren't home and need to go home, do it.
	if mob.Character.RoomId != mob.HomeRoomId {
		if mob.MaxWander > -1 { // -1 means they can wander forever and never go home. 0 means they never wander.
			if len(mob.RoomStack) > mob.MaxWander {
				cmdQueue.QueueCommand(0, mobId, `go home`)

				response.Handled = true
				return response, nil
			}
		}
	}

	// If MaxWander is zero, they don't wander.
	if mob.MaxWander == 0 {
		response.Handled = true
		return response, nil
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
				cmdQueue.QueueCommand(0, mobId, fmt.Sprintf("go %s", exitName))
			}
		}

	}

	response.Handled = true
	return response, nil
}
