package mobcommands

import (
	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/characters"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Backstab(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Must be sneaking
	isSneaking := mob.Character.HasBuffFlag(buffs.Hidden)
	if !isSneaking {
		return true, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {

		if mob.Character.Aggro != nil {
			mob.Character.Aggro.Type = characters.BackStab
			return true, nil
		} else {
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
		}

	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackMobInstanceId == mob.InstanceId {
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

	return true, nil
}
