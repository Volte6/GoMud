package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Macros(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if len(user.Macros) == 0 {
		user.SendText("You have no macros set.")
		return true, nil
	}

	user.SendText(`<ansi fg="yellow">Your macros:</ansi>`)
	for number, macroCommand := range user.Macros {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="yellow">%s:</ansi>`, number))
		user.SendText(fmt.Sprintf(`    <ansi fg="command">%s</ansi>`, macroCommand))
	}
	user.SendText(``)

	return true, nil
}
