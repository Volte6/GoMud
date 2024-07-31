package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
SkullDuggery Skill
Level 1 - Sneak
*/
func Backstab(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return response, nil
	}

	// Must be sneaking
	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	if !isSneaking {
		response.SendUserMessage(userId, "You can't backstab unless you're hidden!", true)
		response.Handled = true
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// Do a check for whether these can backstab
	wpnSubtypeChecks := []items.ItemSubType{}

	if user.Character.Equipment.Weapon.ItemId != 0 {
		wpn := user.Character.Equipment.Weapon.GetSpec()
		wpnSubtypeChecks = append(wpnSubtypeChecks, wpn.Subtype)
	}

	if user.Character.Equipment.Offhand.ItemId != 0 {
		wpn := user.Character.Equipment.Weapon.GetSpec()
		if wpn.Type == items.Weapon {
			wpnSubtypeChecks = append(wpnSubtypeChecks, wpn.Subtype)
		}
	}

	for _, wpnSubType := range wpnSubtypeChecks {
		if !items.CanBackstab(wpnSubType) {
			response.SendUserMessage(userId, fmt.Sprintf(`%s weapons can't be used to backstab.`, wpnSubType), true)
			response.Handled = true
			return response, nil
		}
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro != nil && m.Character.Aggro.UserId == userId {
				attackMobInstanceId = m.InstanceId
				break
			}
		}

		if attackMobInstanceId == 0 {
			for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
				u := users.GetByUserId(uId)
				if u.Character.Aggro != nil && u.Character.Aggro.UserId == userId {
					attackPlayerId = u.UserId
					break
				}
			}
		}
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackPlayerId == userId { // Can't attack self!
		attackPlayerId = 0
	}

	if attackMobInstanceId == 0 && attackPlayerId == 0 {
		response.SendUserMessage(userId, "You attack the darkness!", true)
		response.Handled = true
		return response, nil
	}

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m.Character.IsCharmed(userId) {
			response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name), true)
			response.Handled = true
			return response, nil
		}

		if m != nil {

			user.Character.SetAggro(0, attackMobInstanceId, characters.BackStab, 2)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You prepare to backstab <ansi fg="mobname">%s</ansi>`, m.Character.Name),
				true)

		}

	} else if attackPlayerId > 0 {

		if !configs.GetConfig().PVPEnabled {
			response.SendUserMessage(userId, `PVP is currently disabled.`, true)
			response.Handled = true
			return response, nil
		}

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name), true)
					response.Handled = true
					return response, nil
				}
			}

			user.Character.SetAggro(attackPlayerId, 0, characters.BackStab, 2)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You prepare to backstab <ansi fg="username">%s</ansi>`, p.Character.Name),
				true)
		}

	}

	response.Handled = true
	return response, nil
}
