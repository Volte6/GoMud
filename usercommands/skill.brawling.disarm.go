package usercommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Brawling Skill
Level 4 - Attempt to disarm an opponent.
*/
func Disarm(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 4 {
		return false, nil
	}

	if user.Character.Aggro == nil {
		user.SendText("Disarm is only used while in combat!")
		return true, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	attackMobInstanceId := user.Character.Aggro.MobInstanceId
	attackPlayerId := user.Character.Aggro.UserId

	if attackMobInstanceId > 0 || attackPlayerId > 0 {
		if !user.Character.TryCooldown(skills.Brawling.String(`disarm`), 15) {
			user.SendText(fmt.Sprintf("You can try disarming again in %d rounds.", user.Character.GetCooldown(skills.Brawling.String(`disarm`))))
			return true, nil
		}
	}

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			if m.Character.Equipment.Weapon.ItemId == 0 {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> has no weapon to disarm!`, m.Character.Name))
				return true, nil
			}

			chanceIn100 := (user.Character.Stats.Speed.ValueAdj + user.Character.Stats.Smarts.ValueAdj) - (m.Character.Stats.Strength.ValueAdj + m.Character.Stats.Perception.ValueAdj)
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 5
			roll := util.Rand(100)

			util.LogRoll(`Disarm`, roll, chanceIn100)

			if roll < chanceIn100 {

				user.SendText(
					fmt.Sprintf(`You disarm <ansi fg="mobname">%s</ansi>!`, m.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> disarms <ansi fg="mobname">%s</ansi>!`, user.Character.Name, m.Character.Name),
					userId,
				)

				removedItem := m.Character.Equipment.Weapon
				m.Character.RemoveFromBody(removedItem)
				m.Character.StoreItem(removedItem)

			} else {
				user.SendText(
					fmt.Sprintf(`You try to disarm <ansi fg="mobname">%s</ansi> and fail!`, m.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to disarm <ansi fg="mobname">%s</ansi> and fails!`, user.Character.Name, m.Character.Name),
					userId,
				)

			}
		}
	} else if attackPlayerId > 0 {

		u := users.GetByUserId(attackPlayerId)

		if u != nil {

			chanceIn100 := (user.Character.Stats.Speed.ValueAdj + user.Character.Stats.Smarts.ValueAdj) - (u.Character.Stats.Strength.ValueAdj + u.Character.Stats.Perception.ValueAdj)
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 5
			roll := util.Rand(100)

			util.LogRoll(`Disarm`, roll, chanceIn100)

			if roll < chanceIn100 {

				user.SendText(
					fmt.Sprintf(`You disarm <ansi fg="username">%s</ansi>!`, u.Character.Name),
				)

				if atkUser := users.GetByUserId(attackPlayerId); atkUser != nil {
					atkUser.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> disarms you!`, user.Character.Name),
					)
				}

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> disarms <ansi fg="username">%s</ansi>!`, user.Character.Name, u.Character.Name),
					userId,
					attackPlayerId,
				)

				removedItem := u.Character.Equipment.Weapon
				u.Character.RemoveFromBody(removedItem)
				u.Character.StoreItem(removedItem)

			} else {
				user.SendText(
					fmt.Sprintf(`You try to disarm <ansi fg="username">%s</ansi> and miss!`, u.Character.Name),
				)

				if atkUser := users.GetByUserId(attackPlayerId); atkUser != nil {
					atkUser.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to disarm you and misses!`, user.Character.Name),
					)
				}

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to disarm <ansi fg="username">%s</ansi> and misses!`, user.Character.Name, u.Character.Name),
					userId,
					attackPlayerId,
				)

			}
		}
	}

	return true, nil
}
