package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Shoot(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	if user.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		user.SendText(`You don't have a shooting weapon.`)
		return true, nil
	}

	attackPlayerId := 0
	attackMobInstanceId := 0

	// It's possible that they are shooting in a direction, so check whether multiple words were provided
	// And whether the last word is a direction.
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		user.SendText(`Syntax: <ansi fg="command">shoot [target] [exit]</ansi>`)
		return true, nil
	}

	direction := args[len(args)-1]
	args = args[:len(args)-1]

	// Only shooting weapons can target adjacent rooms
	// "attack goblin east"
	exitName, attackRoomId := room.FindExitByName(direction)
	if exitName != `` {

		exitInfo, _ := room.GetExitInfo(exitName)
		if exitInfo.Lock.IsLocked() {
			user.SendText(fmt.Sprintf("The %s exit is locked.", exitName))
			return true, nil
		}

		if adjacentRoom := rooms.LoadRoom(attackRoomId); adjacentRoom != nil {
			attackPlayerId, attackMobInstanceId = adjacentRoom.FindByName(strings.Join(args, ` `))
		}
	} else {
		user.SendText(`Could not find where you wanted to shoot`)
		return true, nil
	}

	if attackPlayerId == 0 && attackMobInstanceId == 0 {
		user.SendText(`Could not find your target.`)
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

			if m.Character.IsCharmed(user.UserId) {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			user.Character.SetAggroRemote(exitName, 0, attackMobInstanceId, characters.Shooting)

			user.SendText(
				fmt.Sprintf(`You prepare to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, m.Character.Name, exitName),
			)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, m.Character.Name, exitName),
					user.UserId,
				)
			}

		}

	} else if attackPlayerId > 0 {

		p := users.GetByUserId(attackPlayerId)

		if p != nil {

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					return true, nil
				}
			}

			user.Character.SetAggroRemote(exitName, attackPlayerId, 0, characters.Shooting)

			user.SendText(
				fmt.Sprintf(`You prepare to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, p.Character.Name, exitName),
			)

			if !isSneaking {

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, p.Character.Name, exitName),
					user.UserId,
				)

			}

		}

	}

	return true, nil
}
