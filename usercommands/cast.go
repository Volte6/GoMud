package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Cast(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	attackPlayerId := 0
	attackMobInstanceId := 0

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.SendUserMessage(userId, "Cast What? At Whom?", true)
		response.Handled = true
		return response, nil
	}

	spellName := args[0]

	if len(args) < 2 {
		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mobId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mobId)
			if m.Character.Aggro != nil && m.Character.Aggro.UserId == userId {
				attackMobInstanceId = mobId
				break
			}
		}
	} else {
		attackPlayerId, attackMobInstanceId = room.FindByName(strings.Join(args[1:], ` `))
	}

	if attackPlayerId == userId {
		attackPlayerId = 0
	}

	if !user.Character.HasSpell(spellName) {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't know a spell called <ansi fg="spellname">%s</ansi>.`, spellName), true)
		response.Handled = true
		return response, nil
	}

	spellInfo := spells.SpellBook[spellName]

	if attackMobInstanceId == 0 && attackPlayerId == 0 {
		response.SendUserMessage(userId, "No target found!", true)
		response.Handled = true
		return response, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m.Character.IsCharmed(userId) {
			response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name), true)
			response.Handled = true
			return response, nil
		}

		if m != nil {

			user.Character.SetCast(0, attackMobInstanceId, spellInfo.WaitRounds, spellName)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="mobname">%s</ansi>`, m.Character.Name),
				true)

			if !isSneaking {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, user.Character.Name, m.Character.Name),
					true)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.IsCharmed(userId) { // Charmed mobs help the player
						cmdQueue.QueueCommand(0, instId, fmt.Sprintf(`attack #%d`, attackMobInstanceId)) // # denotes a specific mob instanceId
					}
				}
			}
		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name), true)
					response.Handled = true
					return response, nil
				}
			}

			user.Character.SetCast(attackPlayerId, 0, spellInfo.WaitRounds, spellName)

			response.SendUserMessage(userId,
				fmt.Sprintf(`You prepare to enter into mortal combat with <ansi fg="username">%s</ansi>`, p.Character.Name),
				true)

			if !isSneaking {

				response.SendUserMessage(attackPlayerId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight you!`, user.Character.Name),
					true)

				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>`, user.Character.Name, p.Character.Name),
					true,
					userId, attackPlayerId)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				if m := mobs.GetInstance(instId); m != nil {
					if m.Character.IsCharmed(userId) { // Charmed mobs help the player
						cmdQueue.QueueCommand(0, instId, fmt.Sprintf(`attack @%d`, attackPlayerId)) // @ denotes a specific user id
					}
				}
			}

		}

	}

	response.Handled = true
	return response, nil
}
