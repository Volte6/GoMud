package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/characters"
	"github.com/volte6/gomud/items"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
	"github.com/volte6/gomud/util"
)

func Shoot(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		return true, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	// It's possible that they are shooting in a direction, so check whether multiple words were provided
	// And whether the last word is a direction.
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		return true, nil
	}

	direction := args[len(args)-1]
	args = args[:len(args)-1]

	// Only shooting weapons can target adjacent rooms
	// "attack goblin east"
	exitName, attackRoomId := room.FindExitByName(direction)
	if exitName != `` {

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {
			return true, nil
		}

		if adjacentRoom := rooms.LoadRoom(attackRoomId); adjacentRoom != nil {
			attackPlayerId, attackMobInstanceId = adjacentRoom.FindByName(strings.Join(args, ` `))
		}
	} else {
		return true, nil
	}

	if attackPlayerId == 0 && attackMobInstanceId == 0 {
		return true, nil
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

			if m.Character.IsCharmed(mob.InstanceId) {
				return true, nil
			}

			mob.Character.SetAggroRemote(exitName, 0, attackMobInstanceId, characters.Shooting)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, m.Character.Name, exitName),
				)
			}

		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			mob.Character.SetAggroRemote(exitName, attackPlayerId, 0, characters.Shooting)

			if !isSneaking {

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, mob.Character.Name, p.Character.Name, exitName),
				)

			}

		}

	}

	return true, nil
}
