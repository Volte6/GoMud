package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
SkullDuggery Skill
Level 3 - Backstab
*/
func Bump(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	pickPlayerId, pickMobInstanceId := room.FindByName(args[0])

	if pickPlayerId > 0 || pickMobInstanceId > 0 {

		if !user.Character.TryCooldown(skills.Brawling.String(`bump`), configs.GetConfig().MinutesToRounds(1)) {
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

		p := users.GetByUserId(pickPlayerId)

		if p != nil {

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
