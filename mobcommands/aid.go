package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Aid(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	raceInfo := races.GetRace(mob.Character.RaceId)
	if !raceInfo.KnowsFirstAid {
		cmdQueue.QueueCommand(0, mobId, `emote doesn't know first aid.`)
		response.Handled = true
		return response, nil
	}

	if !room.IsCalm() {
		response.Handled = true
		return response, nil
	}

	if rest == `` {
		response.Handled = true
		return response, nil
	}

	aidPlayerId, _ := room.FindByName(rest, rooms.FindDowned)

	if aidPlayerId > 0 {

		p := users.GetByUserId(aidPlayerId)

		if p != nil {

			if p.Character.Health > 0 {
				response.Handled = true
				return response, nil
			}

			mob.Character.CancelBuffsWithFlag(buffs.Hidden)

			response.SendUserMessage(p.UserId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to apply first aid on you.`, mob.Character.Name), true)
			response.SendRoomMessage(mob.Character.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to provide aid to <ansi fg="username">%s</ansi>.`, mob.Character.Name, p.Character.Name), true)
		}

	}

	response.Handled = true
	return response, nil
}
