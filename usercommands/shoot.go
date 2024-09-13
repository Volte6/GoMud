package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Shoot(rest string, userId int) (util.MessageQueue, error) {

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

	if user.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		user.SendText(`You don't have a shooting weapon.`)
		response.Handled = true
		return response, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	// It's possible that they are shooting in a direction, so check whether multiple words were provided
	// And whether the last word is a direction.
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		user.SendText(`Syntax: <ansi fg="command">shoot [target] [exit]</ansi>`)
		response.Handled = true
		return response, nil
	}

	direction := args[len(args)-1]
	args = args[:len(args)-1]

	// Only shooting weapons can target adjacent rooms
	// "attack goblin east"
	exitName, attackRoomId := room.FindExitByName(direction)
	if attackRoomId > 0 {

		exitInfo := room.Exits[exitName]
		if exitInfo.Lock.IsLocked() {
			user.SendText(fmt.Sprintf("The %s exit is locked.", exitName))
			response.Handled = true
			return response, nil
		}

		if adjacentRoom := rooms.LoadRoom(attackRoomId); adjacentRoom != nil {
			attackPlayerId, attackMobInstanceId = adjacentRoom.FindByName(strings.Join(args, ` `))
		}
	}

	if attackRoomId == 0 {
		user.SendText(`Could not find where you wanted to shoot`)
		response.Handled = true
		return response, nil
	}

	if attackPlayerId == 0 && attackMobInstanceId == 0 {
		user.SendText(`Could not find your target.`)
		response.Handled = true
		return response, nil
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
				response.Handled = true
				return response, nil
			}

			user.Character.SetAggroRemote(exitName, 0, attackMobInstanceId, characters.Shooting)

			user.SendText(
				fmt.Sprintf(`You prepare to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, m.Character.Name, exitName),
			)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, m.Character.Name, exitName),
					userId,
				)
			}

		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					response.Handled = true
					return response, nil
				}
			}

			user.Character.SetAggroRemote(exitName, attackPlayerId, 0, characters.Shooting)

			user.SendText(
				fmt.Sprintf(`You prepare to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, p.Character.Name, exitName),
			)

			if !isSneaking {

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, p.Character.Name, exitName),
					userId,
				)

			}

		}

	}

	response.Handled = true
	return response, nil
}
