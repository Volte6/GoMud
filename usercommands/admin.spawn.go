package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Spawn(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.spawn", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	spawnType := args[0]
	args = args[1:]

	spawnTarget := ``
	if len(args) == 1 {
		spawnTarget = args[0]
		args = args[1:]
	} else {
		spawnTarget = strings.Join(args, ` `)
		args = []string{}

	}

	if len(spawnTarget) > 0 {

		if spawnType == `item` {

			itemId := items.FindItemByName(spawnTarget)

			if itemId < 1 {
				itemId, _ = strconv.Atoi(spawnTarget)
			}

			if itemId != 0 {

				itm := items.New(itemId)
				if itm.ItemId > 0 {
					room.AddItem(itm, false)

					user.SendText(
						fmt.Sprintf(`You wave your hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, itm.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, user.Character.Name, itm.DisplayName()),
						user.UserId,
					)

					return true, nil
				}

			}
		}

		if spawnType == `gold` || spawnTarget == `gold` {

			goldAmt := 0
			if spawnType == `gold` {
				goldAmt, _ = strconv.Atoi(spawnTarget)
			} else {
				goldAmt, _ = strconv.Atoi(spawnType)
			}

			if goldAmt < 1 {
				goldAmt = 1
			}

			room.Gold += goldAmt

			user.SendText(
				fmt.Sprintf(`You wave your hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, goldAmt),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, user.Character.Name, goldAmt),
				user.UserId,
			)

			return true, nil
		}

		if spawnType == `mob` {

			mobId := mobs.MobIdByName(spawnTarget)

			if mobId < 1 {
				mobIdInt, _ := strconv.Atoi(spawnTarget)
				mobId = mobs.MobId(mobs.MobId(mobIdInt))
			}

			if mobId > 0 {
				if mob := mobs.NewMobById(mobId, room.RoomId); mob != nil {
					room.AddMob(mob.InstanceId)

					user.SendText(
						fmt.Sprintf(`You wave your hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, mob.Character.Name),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, user.Character.Name, mob.Character.Name),
						user.UserId,
					)

					return true, nil
				}
			}

		}

	}

	user.SendText(
		"You wave your hands around pathetically.",
	)
	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around pathetically.`, user.Character.Name),
		user.UserId,
	)

	return true, nil
}
