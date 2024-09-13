package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Macros(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if len(user.Macros) == 0 {
		response.SendUserMessage(userId, "You have no macros set.")
		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, `<ansi fg="yellow">Your macros:</ansi>`)
	for number, macroCommand := range user.Macros {
		response.SendUserMessage(userId, ``)
		response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="yellow">%s:</ansi>`, number))
		response.SendUserMessage(userId, fmt.Sprintf(`    <ansi fg="command">%s</ansi>`, macroCommand))
	}
	response.SendUserMessage(userId, ``)

	response.Handled = true
	return response, nil
}
