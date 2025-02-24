package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

/*
Brawling Skill
Level 3 - Attempt to tackle an opponent, making them miss a round.
*/
func Tackle(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 3 {
		return false, nil
	}

	if user.Character.Aggro == nil {
		user.SendText("Tackle is only used while in combat!")
		return true, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`tackle`), "5 rounds") {
		user.SendText("You are too tired to tackle again so soon!")
		return true, nil
	}

	attackMobInstanceId := user.Character.Aggro.MobInstanceId
	attackPlayerId := user.Character.Aggro.UserId

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			chanceIn100 := user.Character.Stats.Speed.ValueAdj - m.Character.Stats.Perception.ValueAdj
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 10
			roll := util.Rand(100)

			util.LogRoll(`Tackle`, roll, chanceIn100)

			if roll < chanceIn100 {

				user.SendText(
					fmt.Sprintf(`You lunge and tackle <ansi fg="mobname">%s</ansi>!`, m.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles <ansi fg="mobname">%s</ansi>!`, user.Character.Name, m.Character.Name),
					user.UserId,
				)

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: attackMobInstanceId,
					BuffId:        12, // buff 12 is tackled
				})

			} else {
				user.SendText(
					fmt.Sprintf(`You try to tackle <ansi fg="mobname">%s</ansi> and miss!`, m.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to tackle <ansi fg="mobname">%s</ansi> and misses!`, user.Character.Name, m.Character.Name),
					user.UserId,
				)

			}
		}
	} else if attackPlayerId > 0 {

		u := users.GetByUserId(attackPlayerId)

		if u != nil {

			chanceIn100 := user.Character.Stats.Speed.ValueAdj - u.Character.Stats.Perception.ValueAdj
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 10
			roll := util.Rand(100)

			util.LogRoll(`Tackle`, roll, chanceIn100)

			if roll < chanceIn100 {

				user.SendText(
					fmt.Sprintf(`You lunge and tackle <ansi fg="username">%s</ansi>!`, u.Character.Name),
				)

				if atkUser := users.GetByUserId(attackPlayerId); atkUser != nil {
					atkUser.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles you!`, user.Character.Name),
					)
				}

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles <ansi fg="username">%s</ansi>!`, user.Character.Name, u.Character.Name),
					user.UserId,
					attackPlayerId,
				)

				events.AddToQueue(events.Buff{
					UserId:        attackPlayerId,
					MobInstanceId: 0,
					BuffId:        12, // buff 12 is tackled
				})

			} else {
				user.SendText(
					fmt.Sprintf(`You lunge to tackle <ansi fg="username">%s</ansi> and miss!`, u.Character.Name),
				)

				if atkUser := users.GetByUserId(attackPlayerId); atkUser != nil {
					atkUser.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges to tackles you and misses!`, user.Character.Name),
					)
				}

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges to tackle <ansi fg="username">%s</ansi> and misses!`, user.Character.Name, u.Character.Name),
					user.UserId,
					attackPlayerId,
				)

			}
		}
	}

	return true, nil
}
