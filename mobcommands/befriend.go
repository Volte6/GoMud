package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Befriend(rest string, mobId int) (util.MessageQueue, error) {

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

	if rest == `revert` {

		if mob.Character.IsCharmed() {

			if charmedUserId := mob.Character.RemoveCharm(); charmedUserId > 0 {
				if charmedUser := users.GetByUserId(charmedUserId); charmedUser != nil {
					charmedUser.Character.TrackCharmed(mob.InstanceId, false)
				}
			}

		}

		response.Handled = true
		return response, nil
	}

	playerId, _ := room.FindByName(rest)

	if playerId > 0 {

		mob.Character.Charm(playerId, characters.CharmPermanent, characters.CharmExpiredRevert)

		if charmedUser := users.GetByUserId(playerId); charmedUser != nil {
			charmedUser.Character.TrackCharmed(mob.InstanceId, true)
		}

		//response.SendUserMessage(playerId,
		//	fmt.Sprintf(`<ansi fg="mobname">%s</ansi> looks at you with puppy dog eyes.`, mob.Character.Name),
		//	true)

		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> looks very friendly.`, mob.Character.Name),
			true)

	}

	response.Handled = true
	return response, nil
}
