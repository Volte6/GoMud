package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Peep Skill
Level 1 - Always visibly see the health % of an NPC
Level 2 - Reveals detailed stats of a player or mob.
Level 3 - Reveals detailed stats of the player or mob, plus equipment and items
Level 4 - eveals detailed stats of the player or mob, plus equipment and items, and tells you the % chance of dropping items.
*/
func Peep(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Peep)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to peep.", true)
		response.Handled = true
		return response, errors.New(`you don't know how to peep`)
	}

	if len(rest) == 0 {
		response.SendUserMessage(userId, "Type `help peep` for more information on the peep skill.", true)
		response.Handled = true
		return response, nil
	}

	if skillLevel < 2 {
		response.SendUserMessage(userId, "At level 1, peep is a passive skill.", true)
		response.SendUserMessage(userId, "Type `help peep` for more information on the peep skill.", true)
		response.Handled = true
		return response, errors.New(`at level 1, peep is a passive skill`)
	}

	if !user.Character.TryCooldown(skills.Peep.String(), 1) {
		response.SendUserMessage(userId,
			`You're using that skill just a little too fast.`,
			true)
		response.Handled = true
		return response, errors.New(`you're doing that too often`)
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// valid peep targets are: mobs, players
	playerId, mobId := room.FindByName(rest)

	if playerId > 0 || mobId > 0 {

		statusTxt := ""
		invTxt := ""
		dropTxt := ""

		if playerId > 0 {

			u := *users.GetByUserId(playerId)
			targetName := u.Character.GetPlayerName(user.UserId).String()

			if skillLevel >= 2 {
				statusTxt, _ = templates.Process("character/status-lite", u)
			}

			if skillLevel >= 3 {

				itemNames := []string{}
				itemNamesFormatted := []string{}

				for _, item := range u.Character.Items {

					iName := item.DisplayName()
					iNameFormatted := fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, iName)

					iSpec := item.GetSpec()
					if iSpec.Subtype == items.Drinkable || iSpec.Subtype == items.Edible {
						if iSpec.Uses > 0 { // Does the spec indicate a number of uses?
							iName = fmt.Sprintf(`%s (%d)`, iName, item.Uses)                                               // Display uses left
							iNameFormatted = fmt.Sprintf(`%s <ansi fg="uses-left">(%d)</ansi>`, iNameFormatted, item.Uses) // Display uses left
						}
					}
					itemNames = append(itemNames, iName)
					itemNamesFormatted = append(itemNamesFormatted, iNameFormatted)
				}

				raceInfo := races.GetRace(u.Character.RaceId)

				diceRoll := raceInfo.Damage.DiceRoll
				if u.Character.Equipment.Weapon.ItemId != 0 {
					iSpec := u.Character.Equipment.Weapon.GetSpec()
					diceRoll = iSpec.Damage.DiceRoll
				}

				invData := map[string]any{
					`Equipment`:          &u.Character.Equipment,
					`ItemNames`:          itemNames,
					`ItemNamesFormatted`: itemNamesFormatted,
					`AttackDamage`:       diceRoll,
					`RaceInfo`:           raceInfo,
				}

				invTxt, _ = templates.Process("character/inventory", invData)
			}

			if skillLevel >= 4 {
				dropTxt = fmt.Sprintf(` <ansi fg="username">%s</ansi> has a 100%% chance of dropping their equipment if killed.%s%s`, targetName, term.CRLFStr, term.CRLFStr)
			}

			response.SendUserMessage(playerId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is peeping at you.`, user.Character.Name),
				true)

			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is peeping at <ansi fg="username">%s</ansi>.`, user.Character.Name, u.Character.Name),
				true,
				u.UserId)

		} else if mobId > 0 {

			m := mobs.GetInstance(mobId)
			targetName := m.Character.GetMobName(user.UserId).String()

			if skillLevel >= 2 {
				statusTxt, _ = templates.Process("character/status-lite", m)
			}

			if skillLevel >= 3 {

				itemNames := []string{}
				itemNamesFormatted := []string{}

				for _, item := range m.Character.Items {

					iName := item.DisplayName()
					iNameFormatted := fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, iName)

					iSpec := item.GetSpec()
					if iSpec.Subtype == items.Drinkable || iSpec.Subtype == items.Edible {
						if iSpec.Uses > 0 { // Does the spec indicate a number of uses?
							iName = fmt.Sprintf(`%s (%d)`, iName, item.Uses)                                               // Display uses left
							iNameFormatted = fmt.Sprintf(`%s <ansi fg="uses-left">(%d)</ansi>`, iNameFormatted, item.Uses) // Display uses left
						}
					}
					itemNames = append(itemNames, iName)
					itemNamesFormatted = append(itemNamesFormatted, iNameFormatted)
				}

				raceInfo := races.GetRace(m.Character.RaceId)

				diceRoll := raceInfo.Damage.DiceRoll
				if m.Character.Equipment.Weapon.ItemId != 0 {
					iSpec := m.Character.Equipment.Weapon.GetSpec()
					diceRoll = iSpec.Damage.DiceRoll
				}

				invData := map[string]any{
					`Equipment`:          &m.Character.Equipment,
					`ItemNames`:          itemNames,
					`ItemNamesFormatted`: itemNamesFormatted,
					`AttackDamage`:       diceRoll,
					`RaceInfo`:           raceInfo,
				}

				invTxt, _ = templates.Process("character/inventory", invData)

			}

			if skillLevel >= 4 {
				dropTxt = fmt.Sprintf(`<ansi fg="mobname">%s</ansi> has a %d%% chance of dropping their equipment if killed.%s%s`, targetName, m.ItemDropChance, term.CRLFStr, term.CRLFStr)
			}

			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> is peeping at %s.`, user.Character.Name, targetName),
				true)

		}

		if statusTxt != `` {
			response.SendUserMessage(userId, statusTxt, false)
		}
		if invTxt != `` {
			response.SendUserMessage(userId, invTxt, false)
		}
		if dropTxt != `` {
			response.SendUserMessage(userId, dropTxt, false)
		}

		response.Handled = true
		return response, nil

	}

	response.SendUserMessage(userId, "You don't see that here.", true)
	response.Handled = true
	return response, errors.New(`you don't see that here`)

}
