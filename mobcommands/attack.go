package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Attack(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.Handled = true
		return response, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
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
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackMobInstanceId == mobId { // Can't attack self!
		attackMobInstanceId = 0
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	/*
		combatAddlWaitRounds := mob.Character.Equipment.Weapon.GetSpec().WaitRounds + mob.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if mob.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if attackPlayerId > 0 {

		u := users.GetByUserId(attackPlayerId)

		if u != nil {

			mob.Character.SetAggro(attackPlayerId, 0, characters.DefaultAttack)

			if !isSneaking {

				response.SendUserMessage(u.UserId, fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight you!`, mob.Character.Name), true)

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to fight <ansi fg="username">%s</ansi>`, mob.Character.Name, u.Character.Name),
					true,
					u.UserId)

			}
		}

		response.Handled = true
		return response, nil

	} else if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			mob.Character.SetAggro(0, attackMobInstanceId, characters.DefaultAttack)

			if !isSneaking {

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, mob.Character.Name, m.Character.Name),
					true)

			}

		}

		response.Handled = true
		return response, nil
	}

	if !isSneaking {
		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> looks confused and upset.`, mob.Character.Name),
			true)
	}

	response.Handled = true
	return response, nil
}
