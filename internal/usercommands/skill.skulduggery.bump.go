package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

/*
SkullDuggery Skill
Level 3 - Backstab
*/
func Bump(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 2 {
		return false, nil
	}

	if user.Character.Aggro != nil {
		user.SendText("You can't do that while in combat!")
		return true, nil
	}

	if room.AreMobsAttacking(user.UserId) {
		user.SendText("You can't do that while you are under attack!")
		return true, nil
	}

	if len(rest) == 0 {
		user.SendText("Who do you wanna bump?")
		return true, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	pickPlayerId, pickMobInstanceId := room.FindByName(args[0])

	if pickPlayerId > 0 || pickMobInstanceId > 0 {

		if !user.Character.TryCooldown(skills.Brawling.String(`bump`), "1 real minute") {
			user.SendText(fmt.Sprintf("You need to wait %d rounds before you can do that again!", user.Character.GetCooldown(skills.Brawling.String(`bump`))))
			return true, nil
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)
	}

	goldDropped := 0

	if pickMobInstanceId > 0 {

		m := mobs.GetInstance(pickMobInstanceId)

		if m != nil {

			levelDelta := user.Character.Level - m.Character.Level
			if levelDelta < 1 {
				levelDelta = 1
			}

			chanceIn100 := user.Character.Stats.Strength.ValueAdj / 2
			chanceIn100 /= levelDelta
			if chanceIn100 < 0 {
				chanceIn100 = 1
			}

			roll := util.Rand(100)

			util.LogRoll(`Bump`, roll, chanceIn100)

			if roll < chanceIn100 {

				if m.Character.Gold > 0 {
					goldDropped = util.Rand(m.Character.Gold >> 2)
					if goldDropped > 0 {
						m.Character.Gold -= goldDropped
					}
				}

			}

			user.SendText(
				fmt.Sprintf(`You "accidentally" bump into <ansi fg="mobname">%s</ansi>.`, m.Character.Name),
			)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into <ansi fg="mobname">%s</ansi>!`, user.Character.Name, m.Character.Name),
				user.UserId,
			)

		}

	} else if pickPlayerId > 0 {

		if p := users.GetByUserId(pickPlayerId); p != nil {

			if pvpErr := room.CanPvp(user, p); pvpErr != nil {
				user.SendText(pvpErr.Error())
				return true, nil
			}

			levelDelta := user.Character.Level - p.Character.Level
			if levelDelta < 1 {
				levelDelta = 1
			}
			chanceIn100 := user.Character.Stats.Strength.ValueAdj / 2
			chanceIn100 /= levelDelta
			if chanceIn100 < 0 {
				chanceIn100 = 1
			}

			roll := util.Rand(100)

			util.LogRoll(`Bump`, roll, chanceIn100)

			if roll < chanceIn100 {

				if p.Character.Gold > 0 {
					goldDropped = util.Rand(p.Character.Gold >> 2)
					if goldDropped > 0 {
						p.Character.Gold -= goldDropped
					}
				}
			}

			user.SendText(
				fmt.Sprintf(`You "accidentally" bump into <ansi fg="username">%s</ansi>.`, p.Character.Name),
			)

			p.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into you.`, user.Character.Name),
			)

			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into <ansi fg="username">%s</ansi>!`, user.Character.Name, p.Character.Name),
				user.UserId,
				pickPlayerId,
			)

		}

	} else {

		user.SendText("Pickpocket who?")
	}

	if goldDropped > 0 {

		room.SendText(
			fmt.Sprintf(`<ansi fg="gold">%d gold</ansi> jingles as it drops into the floor!`, goldDropped),
		)

		room.Gold += goldDropped

	}

	return true, nil
}
