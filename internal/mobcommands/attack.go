package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Attack(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		return true, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingMob) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro != nil && m.Character.Aggro.MobInstanceId == mob.InstanceId {
				attackMobInstanceId = m.InstanceId
				break
			}
		}

		if attackMobInstanceId == 0 {
			for _, uId := range room.GetPlayers(rooms.FindFightingMob) {
				u := users.GetByUserId(uId)
				if u.Character.Aggro != nil && u.Character.Aggro.MobInstanceId == mob.InstanceId {
					attackPlayerId = u.UserId
					break
				}
			}
		}
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackMobInstanceId == mob.InstanceId { // Can't attack self!
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

				u.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to fight you!`, mob.Character.Name))

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to fight <ansi fg="username">%s</ansi>`, mob.Character.Name, u.Character.Name),
					u.UserId)

			}
		}

		return true, nil

	} else if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			mob.Character.SetAggro(0, attackMobInstanceId, characters.DefaultAttack)

			if !isSneaking {

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, mob.Character.Name, m.Character.Name))

			}

		}

		return true, nil
	}

	if !isSneaking {
		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> looks confused and upset.`, mob.Character.Name))
	}

	return true, nil
}
