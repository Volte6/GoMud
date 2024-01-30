package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Backstab(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if !isSneaking {
		response.Handled = true
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {

		if mob.Character.Aggro != nil {
			mob.Character.Aggro.Type = characters.BackStab
			response.Handled = true
			return response, nil
		} else {
			// If no argument supplied, attack whoever is attacking the player currently.
			for _, mId := range room.GetMobs(rooms.FindFightingMob) {
				m := mobs.GetInstance(mId)
				if m.Character.Aggro != nil && m.Character.Aggro.MobInstanceId == mobId {
					attackMobInstanceId = m.InstanceId
					break
				}
			}

			if attackMobInstanceId == 0 {
				for _, uId := range room.GetPlayers(rooms.FindFightingMob) {
					u := users.GetByUserId(uId)
					if u.Character.Aggro != nil && u.Character.Aggro.MobInstanceId == mobId {
						attackPlayerId = u.UserId
						break
					}
				}
			}
		}

	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackMobInstanceId == mobId {
		attackMobInstanceId = 0
	}

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {
			mob.Character.SetAggro(0, attackMobInstanceId, characters.BackStab)
		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {
			mob.Character.SetAggro(attackPlayerId, 0, characters.BackStab)
		}

	}

	response.Handled = true
	return response, nil
}
