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
func Bump(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 2 {
		return response, nil
	}

	if user.Character.Aggro != nil {
		response.SendUserMessage(userId, "You can't do that while in combat!")
		response.Handled = true
		return response, nil
	}

	if room.AreMobsAttacking(userId) {
		response.SendUserMessage(userId, "You can't do that while you are under attack!")
		response.Handled = true
		return response, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	pickPlayerId, pickMobInstanceId := room.FindByName(args[0])

	if pickPlayerId > 0 || pickMobInstanceId > 0 {

		if !user.Character.TryCooldown(skills.Brawling.String(`bump`), configs.GetConfig().MinutesToRounds(1)) {
			response.SendUserMessage(userId, fmt.Sprintf("You need to wait %d rounds before you can do that again!", user.Character.GetCooldown(skills.Brawling.String(`bump`))))
			response.Handled = true
			return response, nil
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

			response.SendUserMessage(userId,
				fmt.Sprintf(`You "accidentally" bump into <ansi fg="mobname">%s</ansi>.`, m.Character.Name),
			)

			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into <ansi fg="mobname">%s</ansi>!`, user.Character.Name, m.Character.Name),
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

			response.SendUserMessage(userId,
				fmt.Sprintf(`You "accidentally" bump into <ansi fg="username">%s</ansi>.`, p.Character.Name),
			)

			response.SendUserMessage(pickPlayerId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into you.`, user.Character.Name),
			)

			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> accidentally bumps into <ansi fg="username">%s</ansi>!`, user.Character.Name, p.Character.Name),
				pickPlayerId,
			)

		}

	} else {

		response.SendUserMessage(userId, "Pickpocket who?")
	}

	if goldDropped > 0 {

		response.SendUserMessage(user.UserId,
			fmt.Sprintf(`<ansi fg="gold">%d gold</ansi> jingles as it drops into the floor!`, goldDropped),
		)

		response.SendRoomMessage(user.Character.RoomId,
			fmt.Sprintf(`<ansi fg="gold">%d gold</ansi> jingles as it drops into the floor!`, goldDropped),
		)

		room.Gold += goldDropped

	}

	response.Handled = true
	return response, nil
}
