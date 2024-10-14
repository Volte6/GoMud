package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Macros(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if len(user.Macros) == 0 {
		user.SendText("You have no macros set.")
		return true, nil
	}

	user.SendText(`<ansi fg="yellow">Your macros:</ansi>`)
	for macro, macroCommand := range user.Macros {

		user.SendText(``)

		user.SendText(fmt.Sprintf(`  <ansi fg="228">%s</ansi>:`, macro))

		commandParts := strings.Split(macroCommand, `;`)

		for i, cmd := range commandParts {

			cmdParts := strings.Split(cmd, ` `)
			cmdAlone := cmdParts[0]
			cmdRest := ``
			if len(cmdParts) > 1 {
				cmdRest = strings.Join(cmdParts[1:], ` `)
			}

			user.SendText(fmt.Sprintf(`      %s) <ansi fg="command">%s</ansi> %s`, string(97+i), cmdAlone, cmdRest))
		}
	}
	user.SendText(``)
	user.SendText(`To use a macro, type <ansi fg="command">={num}</ansi>.`)
	user.SendText(`Some terminals support pressing the associated F-Key (<ansi fg="228">F1</ansi>, <ansi fg="228">F2</ansi>, etc.)`)

	return true, nil
}
