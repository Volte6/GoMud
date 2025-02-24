package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

/*
SkullDuggery Skill
Level 2 - Backstab
*/
func Backstab(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 2 {
		return false, nil
	}

	// Must be sneaking
	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	if !isSneaking {
		user.SendText("You can't backstab unless you're hidden!")
		return true, nil
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
			user.SendText(fmt.Sprintf(`%s weapons can't be used to backstab.`, wpnSubType))
			return true, nil
		}
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro != nil && m.Character.Aggro.UserId == user.UserId {
				attackMobInstanceId = m.InstanceId
				break
			}
		}

		if attackMobInstanceId == 0 {
			for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
				u := users.GetByUserId(uId)
				if u.Character.Aggro != nil && u.Character.Aggro.UserId == user.UserId {
					attackPlayerId = u.UserId
					break
				}
			}
		}
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(rest)
	}

	if attackPlayerId == user.UserId { // Can't attack self!
		attackPlayerId = 0
	}

	if attackMobInstanceId == 0 && attackPlayerId == 0 {
		user.SendText("You attack the darkness!")
		return true, nil
	}

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m.Character.IsCharmed(user.UserId) {
			user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
			return true, nil
		}

		if m != nil {

			user.Character.SetAggro(0, attackMobInstanceId, characters.BackStab, 2)

			user.SendText(
				fmt.Sprintf(`You prepare to backstab <ansi fg="mobname">%s</ansi>`, m.Character.Name),
			)

		}

	} else if attackPlayerId > 0 {

		if p := users.GetByUserId(attackPlayerId); p != nil {

			if pvpErr := room.CanPvp(user, p); pvpErr != nil {
				user.SendText(pvpErr.Error())
				return true, nil
			}

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					return true, nil
				}
			}

			user.Character.SetAggro(attackPlayerId, 0, characters.BackStab, 2)

			user.SendText(
				fmt.Sprintf(`You prepare to backstab <ansi fg="username">%s</ansi>`, p.Character.Name),
			)
		}

	}

	return true, nil
}
