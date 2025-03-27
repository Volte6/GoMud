package usercommands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

/*
* Role Permissions:
* grant 				(All)
 */
func Grant(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.grant", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// args should look like one of the following:
	// [?target] 1000 experience - grant experience points to target, or self if unspecified target
	args := util.SplitButRespectQuotes(rest)

	targetUserId := 0
	targetMobInstanceId := 0

	lastWord := args[len(args)-1]

	if len(args) >= 2 && len(lastWord) >= 3 && lastWord[0:3] == `exp` || lastWord == `xp` {

		expAmt := 0

		if len(args) > 2 {

			targetUserId, targetMobInstanceId = room.FindByName(args[0])
			expAmt, _ = strconv.Atoi(args[1])

		} else {
			targetUserId = user.UserId
			expAmt, _ = strconv.Atoi(args[0])
		}

		if targetUserId > 0 {

			if u := users.GetByUserId(targetUserId); u != nil {
				u.GrantXP(expAmt, `admin grant`)
				user.SendText(fmt.Sprintf(`Granted <ansi fg="experience">%d experience</ansi> to <ansi fg="username">%s</ansi>.`, expAmt, u.Character.Name))
				return true, nil
			}

		} else if targetMobInstanceId > 0 {
			user.SendText(`Cannot grant experience to mobs.`)
			return true, nil
		}

		user.SendText(`Target not found.`)
		return true, errors.New(`target not found`)

	}

	user.SendText(`Invalid command.`)

	return false, errors.New(`unrecognized command`)
}
