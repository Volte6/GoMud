package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Shoot(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	if mob.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		response.Handled = true
		return response, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	// It's possible that they are shooting in a direction, so check whether multiple words were provided
	// And whether the last word is a direction.
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		response.Handled = true
		return response, nil
	}

	direction := args[len(args)-1]
	args = args[:len(args)-1]

	// Only shooting weapons can target adjacent rooms
	// "attack goblin east"
	exitName, attackRoomId := room.FindExitByName(direction)
	if attackRoomId > 0 {

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {
			response.Handled = true
			return response, nil
		}

		if adjacentRoom := rooms.LoadRoom(attackRoomId); adjacentRoom != nil {
			attackPlayerId, attackMobInstanceId = adjacentRoom.FindByName(strings.Join(args, ` `))
		}
	}

	if attackRoomId == 0 {
		response.Handled = true
		return response, nil
	}

	if attackPlayerId == 0 && attackMobInstanceId == 0 {
		response.Handled = true
		return response, nil
	}

	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)

	/*
		combatAddlWaitRounds := user.Character.Equipment.Weapon.GetSpec().WaitRounds + user.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if user.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			if m.Character.IsCharmed(mobId) {
				response.Handled = true
				return response, nil
			}

			mob.Character.SetAggroRemote(exitName, 0, attackMobInstanceId, characters.Shooting)

			if !isSneaking {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, m.Character.Name, exitName),
					true)
			}

		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			mob.Character.SetAggroRemote(exitName, attackPlayerId, 0, characters.Shooting)

			if !isSneaking {

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, p.Character.Name, exitName),
					true)

			}

		}

	}

	response.Handled = true
	return response, nil
}
