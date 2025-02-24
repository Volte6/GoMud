package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/combat"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Consider(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	// Looking AT something?
	if len(args) > 0 {
		lookAt := args[0]

		//
		// look for any mobs, players, npcs
		//

		playerId, mobId := room.FindByName(lookAt)
		if playerId == user.UserId {
			playerId = 0
		}

		if playerId > 0 || mobId > 0 {

			ratio := 0.0

			considerType := "mob"
			considerName := "nobody"

			if playerId > 0 {
				u := users.GetByUserId(playerId)

				p1 := combat.PowerRanking(*user.Character, *u.Character)
				p2 := combat.PowerRanking(*u.Character, *user.Character)

				ratio = p1 / p2
				considerType = "user"
				considerName = u.Character.Name

			} else if mobId > 0 {

				m := mobs.GetInstance(mobId)

				p1 := combat.PowerRanking(*user.Character, m.Character)
				p2 := combat.PowerRanking(m.Character, *user.Character)

				ratio = p1 / p2
				considerType = "mob"
				considerName = m.Character.Name
			}

			prediction := `Unknown`
			if ratio > 4 {
				prediction = `<ansi fg="blue-bold">Very Favorable</ansi>`
			} else if ratio > 3 {
				prediction = `<ansi fg="green">Favorable</ansi>`
			} else if ratio > 2 {
				prediction = `<ansi fg="green">Good</ansi>`
			} else if ratio > 1 {
				prediction = `<ansi fg="yellow">Okay</ansi>`
			} else if ratio > 0.5 {
				prediction = `<ansi fg="red-bold">Bad</ansi>`
			} else if ratio > 0 {
				prediction = `<ansi fg="red-bold">Very Bad</ansi>`
			} else {
				prediction = `<ansi fg="red-bold">YOU WILL DIE</ansi>`
			}

			user.SendText(
				fmt.Sprintf(`You consider <ansi fg="%sname">%s</ansi>...`, considerType, considerName),
			)
			user.SendText(
				fmt.Sprintf(`It is estimated that your chances to kill <ansi fg="%sname">%s</ansi> are %s (%f)`, considerType, considerName, prediction, ratio),
			)
		}
	}

	return true, nil
}
