package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Attack(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	if rest == `` {
		partyInfo := parties.Get(user.UserId)

		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro == nil {
				continue
			}

			if m.Character.Aggro.UserId == userId {
				attackMobInstanceId = m.InstanceId
				break
			}

			if partyInfo != nil {
				if partyInfo.IsMember(m.Character.Aggro.UserId) {
					attackMobInstanceId = m.InstanceId
					break
				}
			}
		}

		if attackMobInstanceId == 0 {
			for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
				u := users.GetByUserId(uId)
				if u.Character.Aggro == nil {
					continue
				}

				if u.Character.Aggro.UserId == userId {
					attackPlayerId = u.UserId
					break
				}

				if partyInfo != nil {
					if partyInfo.IsMember(u.Character.Aggro.UserId) {
						attackPlayerId = u.UserId
						break
					}
				}
			}
		}

		// Finally, if still no targets, check if any party members are aggroed and just glom onto that
		if attackMobInstanceId == 0 && attackPlayerId == 0 {
			if partyInfo != nil {
				for uId := range partyInfo.GetMembers() {
					if partyUser := users.GetByUserId(uId); partyUser != nil {
						if partyUser.Character.Aggro == nil {
							continue
						}

						if partyUser.Character.Aggro.MobInstanceId > 0 {
							attackMobInstanceId = partyUser.Character.Aggro.MobInstanceId
							break
						}

						if partyUser.Character.Aggro.UserId > 0 {
							attackPlayerId = partyUser.Character.Aggro.UserId
							break
						}

					}
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
		user.SendText("You attack the darkness!")
		return true, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	/*
		combatAddlWaitRounds := user.Character.Equipment.Weapon.GetSpec().WaitRounds + user.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if user.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {
			if m.Character.IsCharmed(userId) {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := users.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {

								partyUser.Command(fmt.Sprintf(`attack #%d`, attackMobInstanceId)) // # denotes a specific mob instanceId

							}
						}

					}
				}
			}

			user.Character.SetAggro(0, attackMobInstanceId, characters.DefaultAttack)

			user.SendText(
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="mobname">%s</ansi>`, m.Character.Name),
			)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, user.Character.Name, m.Character.Name),
					userId,
				)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.Aggro == nil && m.Character.IsCharmed(userId) { // Charmed mobs help the player

						m.Command(fmt.Sprintf(`attack #%d`, attackMobInstanceId)) // # denotes a specific mob instanceId

					}
				}
			}

		}

	} else if attackPlayerId > 0 {

		if !configs.GetConfig().PVPEnabled {
			user.SendText(`PVP is currently disabled.`)
			return true, nil
		}

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					return true, nil
				}
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := users.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {
								partyUser.Command(fmt.Sprintf(`attack @%d`, attackPlayerId)) // # denotes a specific mob instanceId
							}
						}
					}
				}
			}

			user.Character.SetAggro(attackPlayerId, 0, characters.DefaultAttack)

			user.SendText(
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="username">%s</ansi>`, p.Character.Name),
			)

			if !isSneaking {

				p.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight you!`, user.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, user.Character.Name, p.Character.Name),
					userId, attackPlayerId)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.Aggro == nil && m.Character.IsCharmed(userId) { // Charmed mobs help the player

						m.Command(fmt.Sprintf(`attack @%d`, attackPlayerId)) // @ denotes a specific user id

					}
				}
			}

		}

	}

	return true, nil
}
