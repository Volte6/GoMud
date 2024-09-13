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
		response.SendUserMessage(userId, "You have no macros set.", true)
		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, `<ansi fg="yellow">Your macros:</ansi>`, true)
	for number, macroCommand := range user.Macros {
		response.SendUserMessage(userId, ``, true)
		response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="yellow">%s:</ansi>`, number), true)
		response.SendUserMessage(userId, fmt.Sprintf(`    <ansi fg="command">%s</ansi>`, macroCommand), true)
	}
	response.SendUserMessage(userId, ``, true)

	response.Handled = true
	return response, nil
}
