package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
)

func Macros(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	if len(user.Macros) == 0 {
		user.SendText("You have no macros set.")
		return true, ``, nil
	}

	user.SendText(`<ansi fg="yellow">Your macros:</ansi>`)
	for number, macroCommand := range user.Macros {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="yellow">%s:</ansi>`, number))
		user.SendText(fmt.Sprintf(`    <ansi fg="command">%s</ansi>`, macroCommand))
	}
	user.SendText(``)

	return true, ``, nil
}
